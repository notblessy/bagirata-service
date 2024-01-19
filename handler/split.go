package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

func (h *Handler) Create(c echo.Context) error {
	req := model.SplitEntity{}

	if err := c.Bind(&req); err != nil {
		return utils.Response(c, http.StatusBadRequest, &utils.HTTPResponse{
			Message: fmt.Sprintf("error bind request: %s", err.Error()),
		})
	}

	if req.ID == "" {
		req.ID = utils.GenerateID(10)
	}

	err := h.db.Save(&req).Error
	if err != nil {
		return utils.Response(c, http.StatusInternalServerError, &utils.HTTPResponse{
			Message: err.Error(),
		})
	}

	return utils.Response(c, http.StatusCreated, &utils.HTTPResponse{
		Data: req,
	})
}

func (h *Handler) Find(c echo.Context) error {
	id := c.Param("id")

	if id == "" {
		return utils.Response(c, http.StatusBadRequest, &utils.HTTPResponse{
			Message: fmt.Sprintf("error bind request: %s", errors.New("id is required")),
		})
	}

	var data model.SplitEntity
	err := h.db.Where("id = ?", id).First(&data).Error
	if err != nil {
		return utils.Response(c, http.StatusInternalServerError, &utils.HTTPResponse{
			Message: err.Error(),
		})
	}

	return utils.Response(c, http.StatusCreated, &utils.HTTPResponse{
		Data: data,
	})
}
