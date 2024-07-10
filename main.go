package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/notblessy/model"
	"github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := db.NewPostgres()
	db.AutoMigrate(&model.SplitEntity{})

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	handler := handler.NewHandler(db, openAi)

	e := echo.New()
	e.POST("/v1/recognize", handler.Recognize)

	e.GET("/v1/splits/:slug", handler.FindSplitBySlug)
	e.POST("/v1/splits", handler.SaveSplit)

	e.Logger.Fatal(e.Start(":3200"))
}
