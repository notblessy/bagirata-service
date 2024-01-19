package utils

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/labstack/echo/v4"
)

// DefaultMessage :nodoc:
const DefaultMessage string = "success"

// HTTPResponse :nodoc:
type HTTPResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Response :nodoc:
func Response(c echo.Context, status int, response *HTTPResponse) error {
	if response.Message == "" {
		response.Message = DefaultMessage
	}

	return c.JSON(status, response)
}

func GenerateID(length int) string {
	rand.NewSource(time.Now().UnixNano())

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}

	return string(id)
}

func Dump(v interface{}) string {
	js, _ := json.Marshal(v)
	return string(js)
}
