package main

import (
	"dgw-technical-test/config/database"
	farmer_handler "dgw-technical-test/internal/handlers/farmer"
	admin_handler "dgw-technical-test/internal/handlers/admin"
	product_handler "dgw-technical-test/internal/handlers/product"
	
	"dgw-technical-test/internal/middleware"
	
	farmer_service "dgw-technical-test/internal/services/farmer"
	admin_service "dgw-technical-test/internal/services/admin"
	product_service "dgw-technical-test/internal/services/product"
	purchase_service "dgw-technical-test/internal/services/purchase"
	
	farmer_repo "dgw-technical-test/internal/repositories/farmer"
	admin_repo "dgw-technical-test/internal/repositories/admin"	
	product_repo "dgw-technical-test/internal/repositories/product"	
	order_repo "dgw-technical-test/internal/repositories/order"
	log_repo   "dgw-technical-test/internal/repositories/log"
	review_repo "dgw-technical-test/internal/repositories/review"

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
	farmerRepository := farmer_repo.NewFarmerRepository(config.Pool)
	adminRepository := admin_repo.NewAdminRepository(config.Pool)
	productRepository := product_repo.NewProductRepository(config.Pool)
	orderRepository := order_repo.NewOrderRepository(config.Pool)
	logRepository := log_repo.NewLogRepository(config.Pool)
	reviewRepository := review_repo.NewReviewRepository(config.Pool)

	// Create the necessary services
	farmerService := farmer_service.NewFarmerService(farmerRepository, productRepository, orderRepository, reviewRepository)
	adminService := admin_service.NewAdminService(adminRepository)
	productService := product_service.NewProductService(productRepository)
	purchaseService := purchase_service.NewPurchaseService(*productRepository, *orderRepository, *logRepository)

	// create farmer handler and inject service
	farmerHandler := farmer_handler.NewFarmerHandler(farmerService)
	adminHandler := admin_handler.NewAdminHandler(adminService, purchaseService)
	productHandler := product_handler.NewProductHandler(productService)

	// farmers route grouping under "farmers"
	farmerRoutes := router.Group("/farmers")
	{
		// Register farmer
		farmerRoutes.POST("/register", farmerHandler.RegisterFarmer)

		// Login farmer
		farmerRoutes.POST("/login", farmerHandler.LoginFarmer)

		// get wallet balance (protected by JWT middleware)
		farmerRoutes.GET("/wallet-balance", middleware.JWTAuthMiddleware(), farmerHandler.GetWalletBalance)

		// withdraw money from the bank (protected by JWT middleware)
		farmerRoutes.POST("/withdraw", middleware.JWTAuthMiddleware(), farmerHandler.WithdrawMoney)

		// route to check withdrawal status (Top-Up)
		farmerRoutes.GET("/withdrawal-status/:order_id", middleware.JWTAuthMiddleware(), farmerHandler.GetWithdrawalStatus)
		
		// route to pay the pending order using wallet payment
		farmerRoutes.POST("/pay-order/wallet/:order_id", middleware.JWTAuthMiddleware(), farmerHandler.PayOrder)

		// route to pay the pending order using online payment
		farmerRoutes.POST("/pay-order/online/:order_id", middleware.JWTAuthMiddleware(), farmerHandler.ProcessOnlinePayment)

		// route to check transaction status
		farmerRoutes.GET("/check-status/:order_id/:midtrans_order_id", middleware.JWTAuthMiddleware(), farmerHandler.CheckAndProcessOrderStatus)

		// route to leave a review 
		farmerRoutes.POST("/:order_id/add-review", middleware.JWTAuthMiddleware(), farmerHandler.AddReview)
	}

	// admin route grouping under "admins" hehe
	adminRoutes := router.Group("/admins")
	{
		// Register admin
		adminRoutes.POST("/register", adminHandler.RegisterAdmin)

		// Login admin
		adminRoutes.POST("/login", adminHandler.LoginAdmin)

		// protected route for admin facilitating purchase for farmers
		adminRoutes.POST("/facilitate-purchase/:farmerID", middleware.JWTAuthMiddleware(), adminHandler.FacilitatePurchase)
		
		// protected route for admin facilitating purchase for farmers
		adminRoutes.PUT("/cancel-order/:orderID", middleware.JWTAuthMiddleware(), adminHandler.CancelOrderHandler)
	}

	// product route grouping under "products"
	productRoutes := router.Group("/products")
	{
		// View products
		productRoutes.GET("/view-products", productHandler.GetAllProducts)
	}

	return router
}

func main() {
	// Migrate data to database
	// config.MigrateData()

	// Initialize the application with Gin and dependencies
	router := InitializeApp()
	
	// init midtrans
	farmer_service.Init()

	// Start the Gin server on port 8080
	// close the database connection when the server exits
	defer config.CloseDB()

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}