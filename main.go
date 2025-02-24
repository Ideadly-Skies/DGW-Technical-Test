package main

import (
	_ "dgw-technical-test/docs"
	"dgw-technical-test/config/database"
	admin_handler "dgw-technical-test/internal/adminHandler"
	cust_middleware "dgw-technical-test/internal/middleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title FTGO PlasCash Project
// @version 1.0
// @description API documentation for the FTGO PlashCash project.
// @termsOfService http://example.com/terms/
// @contact.Obie API Support
// @contact.url www.linkedin.com/in/obie-ananda-a87a64212 
// @contact.email Obie.kal22@gmail.com
// @license.name MIT
// @license.url http://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
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

	// public routes
	e.POST("/admin/register", admin_handler.RegisterAdmin)
	e.POST("/admin/login", admin_handler.LoginStoreAdmin)

	// protected routes for store admin using JWT middleware
	storeAdminGroup := e.Group("/admin")
	storeAdminGroup.Use(cust_middleware.JWTMiddleware)

	// Route for Admin
	

	
	// start the server at 8080
	e.Logger.Fatal(e.Start(":8080"))
}
