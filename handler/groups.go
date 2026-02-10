package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/middleware"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"github.com/sirupsen/logrus"
)

// ListGroups returns the authenticated user's groups.
// GET /v1/groups
func (h *Handler) ListGroups(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	var groups []model.Group
	err := h.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&groups).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to list groups: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	list := make([]map[string]interface{}, 0, len(groups))
	for _, g := range groups {
		var count int64
		h.db.Model(&model.SplitEntity{}).Where("group_id = ?", g.ID).Count(&count)
		list = append(list, map[string]interface{}{
			"id":          g.ID,
			"name":        g.Name,
			"bankName":    g.BankName,
			"bankAccount": g.BankAccount,
			"bankNumber":  g.BankNumber,
			"shareSlug":   g.ShareSlug,
			"createdAt":   g.CreatedAt.Format(time.RFC3339),
			"splitCount":  count,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    list,
	})
}

// CreateGroup creates a new group for the authenticated user.
// POST /v1/groups
func (h *Handler) CreateGroup(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	userID, ok := c.Get(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"message": "unauthorized",
			"data":    nil,
		})
	}

	var req model.CreateGroupRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(fmt.Errorf("failed to bind request: %w", err))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request",
			"data":    nil,
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "name is required",
			"data":    nil,
		})
	}

	g := model.Group{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		BankNumber:  req.BankNumber,
	}
	if err := h.db.Create(&g).Error; err != nil {
		logger.Error(fmt.Errorf("failed to create group: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "failed to create group",
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data": map[string]interface{}{
			"id":          g.ID,
			"name":        g.Name,
			"bankName":    g.BankName,
			"bankAccount": g.BankAccount,
			"bankNumber":  g.BankNumber,
			"shareSlug":   g.ShareSlug,
			"createdAt":   g.CreatedAt.Format(time.RFC3339),
		},
	})
}

// UpdateGroup updates group name, bank, and/or share slug.
// PATCH /v1/groups/:id
func (h *Handler) UpdateGroup(c echo.Context) error {
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
	var g model.Group
	err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to get group: %w", err))
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "group not found",
			"data":    nil,
		})
	}

	var req model.UpdateGroupRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(fmt.Errorf("failed to bind request: %w", err))
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "invalid request",
			"data":    nil,
		})
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		if *req.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "name cannot be empty",
				"data":    nil,
			})
		}
		updates["name"] = *req.Name
	}
	if req.BankName != nil {
		updates["bank_name"] = *req.BankName
	}
	if req.BankAccount != nil {
		updates["bank_account"] = *req.BankAccount
	}
	if req.BankNumber != nil {
		updates["bank_number"] = *req.BankNumber
	}
	if req.GenerateShareSlug != nil {
		if *req.GenerateShareSlug {
			slug := utils.RandomSlug(10)
			for {
				var existing model.Group
				if h.db.Where("share_slug = ?", slug).First(&existing).Error != nil {
					break
				}
				slug = utils.RandomSlug(10)
			}
			updates["share_slug"] = slug
		} else {
			updates["share_slug"] = nil
		}
	}

	if len(updates) > 0 {
		if err := h.db.Model(&g).Updates(updates).Error; err != nil {
			logger.Error(fmt.Errorf("failed to update group: %w", err))
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "failed to update group",
				"data":    nil,
			})
		}
		// Reload to return current state
		_ = h.db.Where("id = ?", id).First(&g)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data": map[string]interface{}{
			"id":          g.ID,
			"name":        g.Name,
			"bankName":    g.BankName,
			"bankAccount": g.BankAccount,
			"bankNumber":  g.BankNumber,
			"shareSlug":   g.ShareSlug,
			"createdAt":   g.CreatedAt.Format(time.RFC3339),
		},
	})
}

// GetGroup returns group metadata and list of splits in the group.
// GET /v1/groups/:id
func (h *Handler) GetGroup(c echo.Context) error {
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
	var g model.Group
	err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to get group: %w", err))
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "group not found",
			"data":    nil,
		})
	}

	var entities []model.SplitEntity
	err = h.db.Where("group_id = ?", id).Order("created_at DESC").Find(&entities).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to list splits for group: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	splits := make([]map[string]interface{}, 0, len(entities))
	for _, e := range entities {
		splits = append(splits, map[string]interface{}{
			"id":         e.ID,
			"slug":       e.Slug,
			"name":       e.Name,
			"grandTotal": e.GrandTotal,
			"createdAt":  e.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data": map[string]interface{}{
			"id":          g.ID,
			"name":        g.Name,
			"bankName":    g.BankName,
			"bankAccount": g.BankAccount,
			"bankNumber":  g.BankNumber,
			"shareSlug":   g.ShareSlug,
			"createdAt":   g.CreatedAt.Format(time.RFC3339),
			"splits":      splits,
		},
	})
}

// GetGroupSummary returns aggregated totals per participant for all splits in the group.
// GET /v1/groups/:id/summary
func (h *Handler) GetGroupSummary(c echo.Context) error {
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
	var g model.Group
	err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to get group: %w", err))
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "group not found",
			"data":    nil,
		})
	}

	var entities []model.SplitEntity
	err = h.db.Where("group_id = ?", id).Order("created_at ASC").Find(&entities).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to list splits for group: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	// Aggregate by friendId: sum total across all splits
	type aggRow struct {
		friendId    string
		name        string
		accentColor string
		me          bool
		total       float64
	}
	byFriend := make(map[string]*aggRow)

	splitsOut := make([]model.GroupSummarySplit, 0, len(entities))

	for _, e := range entities {
		splitsOut = append(splitsOut, model.GroupSummarySplit{
			ID:         e.ID,
			Slug:       e.Slug,
			Name:       e.Name,
			GrandTotal: e.GrandTotal,
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		})

		var s model.Splitted
		if err := json.Unmarshal(e.Data, &s); err != nil {
			logger.Warn(fmt.Errorf("failed to unmarshal split data %s: %w", e.ID, err))
			continue
		}

		for _, f := range s.Friends {
			key := f.FriendID
			if key == "" {
				key = f.Name
			}
			if existing, ok := byFriend[key]; ok {
				existing.total += f.Total
			} else {
				byFriend[key] = &aggRow{
					friendId:    f.FriendID,
					name:        f.Name,
					accentColor: f.AccentColor,
					me:          f.Me,
					total:       f.Total,
				}
			}
		}
	}

	participants := make([]model.GroupSummaryParticipant, 0, len(byFriend))
	for _, row := range byFriend {
		participants = append(participants, model.GroupSummaryParticipant{
			FriendID:    row.friendId,
			Name:        row.name,
			AccentColor: row.accentColor,
			Me:          row.me,
			Total:       row.total,
		})
	}

	resp := model.GroupSummaryResponse{
		ID:           g.ID,
		Name:         g.Name,
		BankName:     g.BankName,
		BankAccount:  g.BankAccount,
		BankNumber:   g.BankNumber,
		ShareSlug:    strPtr(g.ShareSlug),
		Participants: participants,
		Splits:       splitsOut,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    resp,
	})
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// GetGroupByShareSlug returns group summary by share slug (public, no auth).
// GET /v1/groups/public/:slug
func (h *Handler) GetGroupByShareSlug(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	slug := c.Param("slug")
	if slug == "" {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "not found",
			"data":    nil,
		})
	}

	var g model.Group
	err := h.db.Where("share_slug = ?", slug).First(&g).Error
	if err != nil {
		logger.Warn(fmt.Errorf("group not found by share slug %q: %w", slug, err))
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": "group not found",
			"data":    nil,
		})
	}

	var entities []model.SplitEntity
	err = h.db.Where("group_id = ?", g.ID).Order("created_at ASC").Find(&entities).Error
	if err != nil {
		logger.Error(fmt.Errorf("failed to list splits for group: %w", err))
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "internal error",
			"data":    nil,
		})
	}

	type aggRow struct {
		friendId    string
		name        string
		accentColor string
		me          bool
		total       float64
	}
	byFriend := make(map[string]*aggRow)
	splitsOut := make([]model.GroupSummarySplit, 0, len(entities))

	for _, e := range entities {
		splitsOut = append(splitsOut, model.GroupSummarySplit{
			ID:         e.ID,
			Slug:       e.Slug,
			Name:       e.Name,
			GrandTotal: e.GrandTotal,
			CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		})

		var s model.Splitted
		if err := json.Unmarshal(e.Data, &s); err != nil {
			logger.Warn(fmt.Errorf("failed to unmarshal split data %s: %w", e.ID, err))
			continue
		}

		for _, f := range s.Friends {
			key := f.FriendID
			if key == "" {
				key = f.Name
			}
			if existing, ok := byFriend[key]; ok {
				existing.total += f.Total
			} else {
				byFriend[key] = &aggRow{
					friendId:    f.FriendID,
					name:        f.Name,
					accentColor: f.AccentColor,
					me:          f.Me,
					total:       f.Total,
				}
			}
		}
	}

	participants := make([]model.GroupSummaryParticipant, 0, len(byFriend))
	for _, row := range byFriend {
		participants = append(participants, model.GroupSummaryParticipant{
			FriendID:    row.friendId,
			Name:        row.name,
			AccentColor: row.accentColor,
			Me:          row.me,
			Total:       row.total,
		})
	}

	resp := model.GroupSummaryResponse{
		ID:           g.ID,
		Name:         g.Name,
		BankName:     g.BankName,
		BankAccount:  g.BankAccount,
		BankNumber:   g.BankNumber,
		ShareSlug:    strPtr(g.ShareSlug),
		Participants: participants,
		Splits:       splitsOut,
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    resp,
	})
}
