package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/sashabaranov/go-openai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	supabase := db.NewSupabase()

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	handler := handler.NewHandler(supabase, openAi)

	e := echo.New()
	e.POST("/v1/recognize", handler.Recognize)

	e.GET("/v1/splits/:slug", handler.FindSplitBySlug)
	e.POST("/v1/splits", handler.SaveSplit)

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}
