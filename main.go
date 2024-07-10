package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/handler"
	"github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	handler := handler.NewHandler(nil, openAi)

	e := echo.New()
	e.POST("/v1/recognize", handler.Recognize)

	e.Logger.Fatal(e.Start(":8080"))
}
