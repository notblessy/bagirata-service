package handler

import (
	"encoding/json"
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
		currencyCode := "IDR"
		if len(e.Data) > 0 {
			var partial struct {
				CurrencyCode string `json:"currencyCode"`
			}
			if err := json.Unmarshal(e.Data, &partial); err == nil && partial.CurrencyCode != "" {
				currencyCode = partial.CurrencyCode
			}
		}
		list = append(list, model.SplitSummary{
			ID:           e.ID,
			Slug:         e.Slug,
			Name:         e.Name,
			GrandTotal:   e.GrandTotal,
			CreatedAt:    e.CreatedAt.Format(time.RFC3339),
			FriendCount:  e.FriendCount,
			CurrencyCode: currencyCode,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    list,
	})
}

// UpdateSplitGroup assigns or removes a split from a group.
// PATCH /v1/splits/:id
func (h *Handler) UpdateSplitGroup(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "missing split id",
			"data":    nil,
		})
	}

	var req struct {
		GroupID *string `json:"groupId"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request body",
			"data":    nil,
		})
	}

	// Fetch the split to ensure it belongs to this user
	var entity model.SplitEntity
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&entity).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "split not found",
			"data":    nil,
		})
	}

	// If assigning to a group, validate the group exists and belongs to the user
	if req.GroupID != nil && *req.GroupID != "" {
		var g model.Group
		if err := h.db.Where("id = ? AND user_id = ?", *req.GroupID, userID).First(&g).Error; err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "group not found",
				"data":    nil,
			})
		}

		// Validate currency match
		if g.CurrencyCode != "" {
			currencyCode := "IDR"
			if len(entity.Data) > 0 {
				var partial struct {
					CurrencyCode string `json:"currencyCode"`
				}
				if err := json.Unmarshal(entity.Data, &partial); err == nil && partial.CurrencyCode != "" {
					currencyCode = partial.CurrencyCode
				}
			}
			if g.CurrencyCode != currencyCode {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "group currency does not match split currency",
					"data":    nil,
				})
			}
		}
	}

	// Update the group_id (nil removes from group)
	if err := h.db.Model(&entity).Update("group_id", req.GroupID).Error; err != nil {
		logger.Error(fmt.Errorf("failed to update split group %s: %w", id, err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to update split group",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    nil,
	})
}

// DeleteSplit deletes a split by ID for the authenticated user.
// DELETE /v1/splits/:id
func (h *Handler) DeleteSplit(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "missing split id",
			"data":    nil,
		})
	}

	// Only delete splits that belong to this user.
	res := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.SplitEntity{})
	if res.Error != nil {
		logger.Error(fmt.Errorf("failed to delete split %s: %w", id, res.Error))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to delete split",
			"data":    nil,
		})
	}

	if res.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "split not found",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    nil,
	})
}
