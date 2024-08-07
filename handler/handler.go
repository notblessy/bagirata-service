package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/supabase-community/supabase-go"
)

type Handler struct {
	sb     *supabase.Client
	openAi *openai.Client
}

func NewHandler(sb *supabase.Client, openAi *openai.Client) *Handler {
	return &Handler{
		sb:     sb,
		openAi: openAi,
	}
}

func (h *Handler) Recognize(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var request model.RecognizeRequest

	if err := c.Bind(&request); err != nil {
		logger.Error(fmt.Errorf("failed to bind request: %w", err))

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": fmt.Errorf("failed to bind request: %w", err).Error,
			"data":    nil,
		})
	}

	resp, err := h.openAi.CreateChatCompletion(
		c.Request().Context(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: request.Model,
				},
			},
		},
	)
	if err != nil {
		logger.Error(fmt.Errorf("failed to create chat completion: %w", err))

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Errorf("failed to create chat completion: %w", err).Error,
			"data":    nil,
		})
	}

	var recognized model.Recognized

	recognized.CreatedAt = time.Now()

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &recognized)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal recognized: %w", err))

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Errorf("failed to unmarshal recognized: %w", err).Error,
			"data":    nil,
		})
	}

	recognized.ID = uuid.New().String()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    recognized,
	})
}

func (h *Handler) SaveSplit(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var splitted model.Splitted

	if err := c.Bind(&splitted); err != nil {
		logger.Error(fmt.Errorf("failed to bind request: %w", err))

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": fmt.Errorf("failed to bind request: %w", err).Error,
			"data":    nil,
		})
	}

	entity := splitted.ToData()

	_, _, err := h.sb.From("splits").Upsert(entity, "", "", "").Execute()
	if err != nil {
		logger.Error(fmt.Errorf("failed to upsert split: %w", err))

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    splitted.Slug,
	})
}

func (h *Handler) FindSplitBySlug(c echo.Context) error {
	slug := c.Param("slug")

	logger := logrus.WithFields(logrus.Fields{
		"ctx":  utils.Dump(c.Request().Context()),
		"slug": slug,
	})

	var splitted model.SplitEntity

	// find by slug to supabase
	data, _, err := h.sb.From("splits").Select("*", "exact", false).Eq("slug", slug).Single().Execute()
	if err != nil {
		logger.Error(fmt.Errorf("failed to find split by slug: %w", err))

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	err = json.Unmarshal(data, &splitted)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal split: %w", err))

		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    splitted.Data,
	})
}

func (h *Handler) ViewSplitBySlug(c echo.Context) error {
	slug := c.Param("slug")

	logger := logrus.WithFields(logrus.Fields{
		"ctx":  utils.Dump(c.Request().Context()),
		"slug": slug,
	})

	var splitted model.SplitEntity

	// find by slug to supabase
	data, _, err := h.sb.From("splits").Select("*", "exact", false).Eq("slug", slug).Single().Execute()
	if err != nil {
		logger.Error(fmt.Errorf("failed to find split by slug: %w", err))

		return c.Render(http.StatusNotFound, "404.html", nil)
	}

	err = json.Unmarshal(data, &splitted)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal split: %w", err))
		return c.Render(http.StatusNotFound, "404.html", nil)
	}

	var splittedData model.Splitted
	err = json.Unmarshal(splitted.Data, &splittedData)
	if err != nil {
		logger.Error(fmt.Errorf("failed to unmarshal split data: %w", err))
		return c.Render(http.StatusNotFound, "404.html", nil)
	}

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"data": splittedData,
	})
}

func (h *Handler) ViewPrivacyPolicy(c echo.Context) error {
	return c.Render(http.StatusOK, "privacy.html", nil)
}

func (h *Handler) ViewLandingPage(c echo.Context) error {
	return c.Render(http.StatusOK, "landing.html", nil)
}
