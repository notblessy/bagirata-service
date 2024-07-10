package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/model"
	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	openAi *openai.Client
}

func NewHandler(db *gorm.DB, openAi *openai.Client) *Handler {
	return &Handler{
		db:     db,
		openAi: openAi,
	}
}

func (h *Handler) Recognize(c echo.Context) error {
	var request model.RecognizeRequest

	if err := c.Bind(&request); err != nil {
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
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": fmt.Errorf("failed to unmarshal recognized: %w", err).Error,
			"data":    nil,
		})
	}

	recognized.ID = uuid.New().String()

	r := map[string]interface{}{
		"success": true,
		"message": "success",
		"data":    recognized,
	}

	js, _ := json.MarshalIndent(r, "", "    ")
	fmt.Println(string(js))

	return c.JSON(http.StatusOK, r)
}
