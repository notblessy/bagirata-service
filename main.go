package main

import (
	"io"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/notblessy/middleware"
	"github.com/notblessy/model"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func init() {
	setupLogger()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("Error loading .env file")
	}

	postgres := db.NewPostgres()

	postgres.AutoMigrate(&model.User{}, &model.SplitEntity{}, &model.Group{}, &model.Dataset{})

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	h := handler.NewHandler(postgres, openAi)

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("change-me-in-production")
	}
	requireAuth := middleware.RequireAuth(jwtSecret)
	optionalAuth := middleware.OptionalAuth(jwtSecret)

	t := &Template{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}

	e := echo.New()

	e.Static("/static", "public")
	e.Static("", "ads")

	e.Renderer = t

	e.Use(echomw.Logger())
	e.Use(echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
		Skipper: echomw.DefaultSkipper,
		Store:   echomw.NewRateLimiterMemoryStore(20),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
	}))

	// implement cors (include PATCH for /auth/me profile update)
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders: []string{echo.HeaderContentType, echo.HeaderContentLength},
	}))

	e.RouteNotFound("*", notFound)

	e.GET("/", h.ViewLandingPage)

	e.GET("/view/:slug", h.ViewSplitBySlug)
	e.GET("/support/privacy", h.ViewPrivacyPolicy)

	// Auth (no auth required for login/register)
	e.POST("/login", h.Login)
	e.POST("/register", h.Register)
	// /auth/* routes for app compatibility (same handlers)
	e.POST("/auth/login", h.Login)
	e.POST("/auth/register", h.Register)
	e.GET("/me", h.Me, requireAuth)

	e.GET("/v1/auth/me", h.Me, requireAuth)
	e.PATCH("/v1/auth/me", h.UpdateProfile, requireAuth)
	e.DELETE("/v1/auth/me", h.DeleteAccount, requireAuth)

	e.POST("/v1/recognize", h.Recognize)

	e.GET("/v1/splits", h.ListSplits, requireAuth)
	e.GET("/v1/splits/:slug", h.FindSplitBySlug)
	e.POST("/v1/splits", h.SaveSplit, optionalAuth)

	e.GET("/v1/groups", h.ListGroups, requireAuth)
	e.POST("/v1/groups", h.CreateGroup, requireAuth)
	e.GET("/v1/groups/:id", h.GetGroup, requireAuth)
	e.GET("/v1/groups/:id/summary", h.GetGroupSummary, requireAuth)

	e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
}

func notFound(c echo.Context) error {
	return c.Render(http.StatusNotFound, "404.html", nil)
}

func setupLogger() {
	formatter := logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     false,
	}

	logrus.SetFormatter(&formatter)
	logrus.SetOutput(os.Stdout)

	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.DebugLevel
	}

	logrus.SetLevel(logLevel)
}
