package handler

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/middleware"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (h *Handler) Login(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var req model.LoginRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(fmt.Errorf("failed to bind login request: %w", err))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
			"data":    nil,
		})
	}
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "email and password are required",
			"data":    nil,
		})
	}

	var user model.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "invalid email or password",
				"data":    nil,
			})
		}
		logger.Error(fmt.Errorf("failed to find user: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "invalid email or password",
			"data":    nil,
		})
	}

	token, err := h.issueToken(user.ID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to issue token: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data": model.AuthResponse{
			Token: token,
			User:  userToResponse(user),
		},
	})
}

func (h *Handler) Register(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var req model.RegisterRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(fmt.Errorf("failed to bind register request: %w", err))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
			"data":    nil,
		})
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "email, password and name are required",
			"data":    nil,
		})
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(fmt.Errorf("failed to hash password: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	user := model.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		Password: string(hashed),
		Name:     req.Name,
	}
	if err := h.db.Create(&user).Error; err != nil {
		if isUniqueViolation(err) {
			return c.JSON(http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "email already registered",
				"data":    nil,
			})
		}
		logger.Error(fmt.Errorf("failed to create user: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	token, err := h.issueToken(user.ID)
	if err != nil {
		logger.Error(fmt.Errorf("failed to issue token: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data": model.AuthResponse{
			Token: token,
			User:  userToResponse(user),
		},
	})
}

func (h *Handler) Me(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	var user model.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "user not found",
				"data":    nil,
			})
		}
		logger.Error(fmt.Errorf("failed to find user: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    userToResponse(user),
	})
}

func userToResponse(u model.User) model.UserResponse {
	return model.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) issueToken(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// isUniqueViolation returns true if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// GORM / pgx often wrap the error; check error message for unique_violation
	return strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate")
}
