package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/sashabaranov/go-openai"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARNING] Error loading .env file")
	}

	supabase := db.NewSupabase()

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	handler := handler.NewHandler(supabase, openAi)

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	e := echo.New()

	e.Static("/static", "public")
	e.Renderer = t

	e.Use(middleware.Logger())
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   middleware.NewRateLimiterMemoryStore(20),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
	}))

	e.RouteNotFound("*", notFound)

	e.GET("/:slug", handler.ViewSplitBySlug)
	e.GET("/support/privacy", handler.ViewPrivacyPolicy)

	e.POST("/v1/recognize", handler.Recognize)

	e.GET("/v1/splits/:slug", handler.FindSplitBySlug)
	e.POST("/v1/splits", handler.SaveSplit)

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func notFound(c echo.Context) error {
	return c.Render(http.StatusNotFound, "404.html", nil)
}
