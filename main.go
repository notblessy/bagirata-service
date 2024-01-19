package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/notblessy/config"
	"github.com/notblessy/db"
	"github.com/notblessy/handler"
	"github.com/notblessy/http"
	"github.com/notblessy/model"
	"github.com/notblessy/utils"
	"github.com/sirupsen/logrus"
)

func init() {
	config.LoadConfig()
}

func main() {
	mysql := db.NewMysql()
	mysql.AutoMigrate(&model.SplitEntity{})

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Validator = &utils.CustomValidator{Validator: validator.New()}

	handler := handler.New(mysql)
	httpSvc := http.New(handler)

	httpSvc.Routes(e)

	logrus.Fatal(e.Start(":" + config.HTTPPort()))
}
