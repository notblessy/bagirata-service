package main

import (
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/notblessy/middleware"
	"github.com/notblessy/model"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Paths matching these patterns are almost always vulnerability scanners
// (WordPress/PHP/config probes). Reject before routing so they cost nothing
// and don't pollute access logs.
var scannerSuffixes = []string{".php", ".asp", ".aspx", ".jsp", ".env", ".sql", ".bak", ".old", ".log"}
var scannerSubstrings = []string{"wp-", "xmlrpc", "phpmyadmin", "/.env", "/.git", "/htaccess", "/wordpress", "/adminer"}

func isScannerProbe(path string) bool {
	p := strings.ToLower(path)
	for _, s := range scannerSuffixes {
		if strings.HasSuffix(p, s) {
			return true
		}
	}
	for _, s := range scannerSubstrings {
		if strings.Contains(p, s) {
			return true
		}
	}
	return false
}

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

	// Drop scanner probes before they hit the router or logger.
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if isScannerProbe(c.Request().URL.Path) {
				return c.NoContent(http.StatusForbidden)
			}
			return next(c)
		}
	})

	e.Use(echomw.Logger())

	ipRateLimiter := func(r rate.Limit, burst int, expiresIn time.Duration) echo.MiddlewareFunc {
		return echomw.RateLimiterWithConfig(echomw.RateLimiterConfig{
			Skipper: echomw.DefaultSkipper,
			Store: echomw.NewRateLimiterMemoryStoreWithConfig(echomw.RateLimiterMemoryStoreConfig{
				Rate:      r,
				Burst:     burst,
				ExpiresIn: expiresIn,
			}),
			IdentifierExtractor: func(c echo.Context) (string, error) {
				return c.RealIP(), nil
			},
		})
	}

	// Baseline global limit per IP.
	e.Use(ipRateLimiter(10, 20, 5*time.Minute))

	// Expensive OpenAI-backed endpoint: ~30 req/min per IP, burst 3.
	recognizeLimiter := ipRateLimiter(rate.Every(2*time.Second), 3, 10*time.Minute)

	// Auth endpoints: ~12 req/min per IP, burst 5. Blocks brute force / credential stuffing.
	authLimiter := ipRateLimiter(rate.Every(5*time.Second), 5, 15*time.Minute)

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
	e.POST("/login", h.Login, authLimiter)
	e.POST("/register", h.Register, authLimiter)
	// /auth/* routes for app compatibility (same handlers)
	e.POST("/auth/login", h.Login, authLimiter)
	e.POST("/auth/register", h.Register, authLimiter)
	e.GET("/me", h.Me, requireAuth)

	e.GET("/v1/auth/me", h.Me, requireAuth)
	e.PATCH("/v1/auth/me", h.UpdateProfile, requireAuth)
	e.PATCH("/v1/auth/me/password", h.ChangePassword, requireAuth)
	e.DELETE("/v1/auth/me", h.DeleteAccount, requireAuth)

	e.POST("/v1/recognize", h.Recognize, recognizeLimiter)

	e.GET("/v1/splits", h.ListSplits, requireAuth)
	e.GET("/v1/splits/:slug", h.FindSplitBySlug)
	e.POST("/v1/splits", h.SaveSplit, optionalAuth)
	e.PATCH("/v1/splits/:id", h.UpdateSplitGroup, requireAuth)
	e.DELETE("/v1/splits/:id", h.DeleteSplit, requireAuth)

	e.GET("/v1/groups", h.ListGroups, requireAuth)
	e.POST("/v1/groups", h.CreateGroup, requireAuth)
	e.GET("/v1/groups/public/:slug", h.GetGroupByShareSlug)
	e.GET("/v1/groups/:id", h.GetGroup, requireAuth)
	e.PATCH("/v1/groups/:id", h.UpdateGroup, requireAuth)
	e.DELETE("/v1/groups/:id", h.DeleteGroup, requireAuth)
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
