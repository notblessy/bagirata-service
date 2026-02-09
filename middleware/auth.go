package middleware

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

const UserIDKey = "user_id"

// RequireAuth validates Bearer token and sets user_id in context. Returns 401 if missing/invalid.
func RequireAuth(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return c.JSON(401, map[string]interface{}{
					"success": false,
					"message": "missing authorization header",
					"data":    nil,
				})
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				return c.JSON(401, map[string]interface{}{
					"success": false,
					"message": "invalid authorization format",
					"data":    nil,
				})
			}
			tokenStr := strings.TrimPrefix(auth, prefix)
			token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				return secret, nil
			})
			if err != nil {
				return c.JSON(401, map[string]interface{}{
					"success": false,
					"message": "invalid or expired token",
					"data":    nil,
				})
			}
			claims, ok := token.Claims.(*JWTClaims)
			if !ok || !token.Valid {
				return c.JSON(401, map[string]interface{}{
					"success": false,
					"message": "invalid token",
					"data":    nil,
				})
			}
			c.Set(UserIDKey, claims.UserID)
			return next(c)
		}
	}
}

// OptionalAuth parses Bearer token if present and sets user_id in context. Does not return error if missing.
func OptionalAuth(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth != "" && strings.HasPrefix(auth, "Bearer ") {
				tokenStr := strings.TrimPrefix(auth, "Bearer ")
				token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return secret, nil
				})
				if err == nil {
					if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
						c.Set(UserIDKey, claims.UserID)
					}
				}
			}
			return next(c)
		}
	}
}
