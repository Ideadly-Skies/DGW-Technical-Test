package main

import (
	"log"
	"github.com/joho/godotenv"
	"dgw-technical-test/config/database"
	"dgw-technical-test/internal/handlers/farmer"
	"dgw-technical-test/internal/services/farmer"
	"dgw-technical-test/internal/repositories/farmer"
	"github.com/gin-gonic/gin"
)

func InitializeApp() *gin.Engine {
	// Initialize Gin
	router := gin.Default()

	// load the .env file
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

	// Initialize database connection
	config.InitDB()

	// Create the necessary repositories (dependency injection)
	farmerRepository := repositories.NewFarmerRepository(config.Pool)

	// Create the necessary services
	farmerService := services.NewFarmerService(farmerRepository)

	// create farmer handler and inject service
	farmerHandler := handlers.NewFarmerHandler(farmerService)

	// farmers route grouping under "farmers"
	farmerRoutes := router.Group("/farmers")
	{
		// Register farmer
		farmerRoutes.POST("/register", farmerHandler.RegisterFarmer)

		// Login farmer
		farmerRoutes.POST("/login", farmerHandler.LoginFarmer)
	}

	return router
}

func main() {
	// Migrate data to database
	// config.MigrateData()

	// Initialize the application with Gin and dependencies
	router := InitializeApp()

	// Start the Gin server on port 8080
	// close the database connection when the server exits
	defer config.CloseDB()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}