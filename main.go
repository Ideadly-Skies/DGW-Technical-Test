package main

import (
	config "dgw-technical-test/config/database"
	// cust_middleware "dgw-technical-test/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	// migrate data to supabase
	// config.MigrateData()

	// connect to db
	config.InitDB()
	defer config.CloseDB()

	// use echo-framework to simulate smart-city ecosystem
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Swagger route
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// start the server at 8080
	e.Logger.Fatal(e.Start(":8080"))
}