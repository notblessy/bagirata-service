package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/middleware"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"github.com/sirupsen/logrus"
)

// ListSplits returns the authenticated user's splits with pagination and optional search.
// GET /v1/splits?page=1&size=10&search=term
func (h *Handler) ListSplits(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(c.QueryParam("size"))
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	offset := (page - 1) * size

	query := h.db.Where("user_id = ?", userID)
	if groupID := c.QueryParam("group_id"); groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	if search := c.QueryParam("search"); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var entities []model.SplitEntity
	err := query.Order("created_at DESC").
		Limit(size).
		Offset(offset).
		Find(&entities).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to list splits: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	list := make([]model.SplitSummary, 0, len(entities))
	for _, e := range entities {
		list = append(list, model.SplitSummary{
			ID:          e.ID,
			Slug:        e.Slug,
			Name:        e.Name,
			GrandTotal:  e.GrandTotal,
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
			FriendCount: e.FriendCount,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    list,
	})
}
