package main

import (
	"dgw-technical-test/config/database"
	"dgw-technical-test/internal/handlers/farmer"
	"dgw-technical-test/internal/middleware"
	"dgw-technical-test/internal/repositories/farmer"
	"dgw-technical-test/internal/services/farmer"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

		// get wallet balance (protected by JWT middleware)
		farmerRoutes.GET("/wallet-balance", middleware.JWTAuthMiddleware(), farmerHandler.GetWalletBalance)

		farmerRoutes.POST("/withdraw", middleware.JWTAuthMiddleware(), farmerHandler.WithdrawMoney)
	}

	return router
}

func main() {
	// Migrate data to database
	config.MigrateData()

	// Initialize the application with Gin and dependencies
	router := InitializeApp()

	// Start the Gin server on port 8080
	// close the database connection when the server exits
	defer config.CloseDB()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}