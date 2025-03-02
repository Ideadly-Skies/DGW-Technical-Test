package handlers

import (
	"dgw-technical-test/internal/models/farmer"
	"dgw-technical-test/internal/services/farmer"
	"net/http"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

// FarmerHandler contains services related to farmer operations
type FarmerHandler struct {
	FarmerService *services.FarmerService
}

// NewFarmerHandler creates a new FarmerHandler instance
func NewFarmerHandler(farmerService *services.FarmerService) *FarmerHandler {
	return &FarmerHandler{FarmerService: farmerService}
}

// RegisterFarmer godoc
// @Summary Register a new farmer
// @Description Registers a new farmer with name, email, and password.
// @Tags Farmer
// @Accept json
// @Produce json
// @Param farmer body models.RegisterRequest true "Registration information"
// @Success 200 {object} map[string]interface{} "message: Farmer registered successfully"
// @Failure 400 {object} map[string]string "message: Invalid request or missing fields"
// @Failure 500 {object} map[string]string "message: Internal server error"
// @Router /farmers/register [post]
func (h *FarmerHandler) RegisterFarmer(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		return
	}

	// Validate input
	if req.Name == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "All fields are required"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
		return
	}

	// Create the farmer in the database with initial wallet_balance set to 0
	err = h.FarmerService.RegisterFarmer(req.Name, req.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not register farmer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Farmer registered successfully"})
}

// LoginFarmer godoc
// @Summary Login a farmer
// @Description Logs in a farmer using email and password.
// @Tags Farmer
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "JWT token and farmer info"
// @Failure 400 {object} map[string]string "message: Invalid request"
// @Failure 401 {object} map[string]string "message: Invalid email or password"
// @Router /farmers/login [post]
func (h *FarmerHandler) LoginFarmer(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		return
	}

	// Login logic using FarmerService
	farmer, err := h.FarmerService.LoginFarmer(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := h.FarmerService.GenerateJWT(farmer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:         token,
		Name:          farmer.Name,
		Email:         farmer.Email,
		WalletBalance: farmer.WalletBalance,
	})
}

// GetWalletBalance godoc
// @Summary Retrieve wallet balance
// @Description Retrieves the current wallet balance for a logged-in farmer.
// @Tags Farmer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]float64 "wallet_balance: Current wallet balance"
// @Failure 500 {object} map[string]string "message: Failed to retrieve wallet balance"
// @Router /farmers/wallet-balance [get]
func (h *FarmerHandler) GetWalletBalance(c *gin.Context) {
	// Extract farmer ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)  // Directly get the claims here
	farmerID := int(user["farmer_id"].(float64)) // Access the "farmer_id" from the claims

	// Retrieve wallet balance using FarmerService
	balance, err := h.FarmerService.GetFarmerWalletBalance(int(farmerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve wallet balance"})
		return
	}

	// Return wallet balance
	c.JSON(http.StatusOK, gin.H{
		"wallet_balance": balance,
	})
}

// PaymentRequest contains structure for farmer transaction
type PaymentRequest struct {
	Amount float64 `json:"amount" validate:"required"`
}

// WithdrawMoney godoc
// @Summary Withdraw money from wallet
// @Description Allows a farmer to withdraw money from their wallet.
// @Tags Farmer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param amount body PaymentRequest true "Amount to withdraw"
// @Success 200 {object} map[string]interface{} "message: Withdrawal initiated successfully"
// @Failure 400 {object} map[string]string "message: Invalid request or amount must be greater than zero"
// @Failure 500 {object} map[string]string "message: Internal server error"
// @Router /farmers/withdraw [post]
func (h *FarmerHandler) WithdrawMoney(c *gin.Context) {
	// Extract farmer ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	farmerID := int(user["farmer_id"].(float64))  // Access the "farmer_id" from the claims
	farmerName := user["name"].(string)           // Access the "name" from the claims

	// Bind and validate request body
	var req PaymentRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Withdraw amount must be greater than zero"})
		return
	}

	// Call service to handle the withdrawal logic
	transactionID, orderID, vaNumber, err := h.FarmerService.WithdrawMoney(farmerID, req.Amount, farmerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Return withdrawal details with VA number
	c.JSON(http.StatusOK, gin.H{
		"message":        "Withdrawal initiated successfully",
		"transaction_id": transactionID,
		"order_id":       orderID,
		"va_number":      vaNumber,
		"gross_amount":   req.Amount,
		"status":         "Pending",
	})
}

// GetWithdrawalStatus godoc
// @Summary Check withdrawal status
// @Description Checks the status of a farmer's withdrawal transaction.
// @Tags Farmer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path string true "Order ID of the transaction"
// @Success 200 {object} map[string]interface{} "Transaction status retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid transaction request"
// @Failure 500 {object} map[string]string "Failed to fetch transaction status"
// @Router /farmers/withdrawal-status/{order_id} [get]
func (h *FarmerHandler) GetWithdrawalStatus(c *gin.Context) {
	orderID := c.Param("order_id")

	// Check withdrawal status in the service
	status, err := h.FarmerService.CheckWithdrawalStatus(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Return transaction status
	c.JSON(http.StatusOK, status)
}

// PayOrder godoc
// @Summary Pay for an order using wallet balance
// @Description Allows a farmer to pay for an order using the balance available in their wallet.
// @Tags Farmer
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "message: Payment successful"
// @Failure 400 {object} map[string]string "error: Invalid order ID or Farmer is not registered"
// @Failure 401 {object} map[string]string "error: Unauthorized access"
// @Failure 500 {object} map[string]string "error: Failed to process payment or check farmer registration"
// @Router /farmers/pay-order/{order_id} [post]
func (h *FarmerHandler) PayOrder(c *gin.Context) {
    orderIDParam := c.Param("order_id")
    orderID, err := strconv.Atoi(orderIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
        return
    }

	// Extract farmer ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)  // Directly get the claims here
	farmerID := int(user["farmer_id"].(float64)) // Access the "farmer_id" from the claims	

	// check if farmerID is registered in farmers db
	isRegistered, err := h.FarmerService.IsFarmerRegistered(farmerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check farmer registration", "details": err.Error()})
		return
	}

	if !isRegistered {
		// check if the farmer is registered
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer is not registered"})
		return
	}

	// process wallet payment of the farmer
    err = h.FarmerService.ProcessWalletPayment(c.Request.Context(), farmerID, orderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Payment successful"})
}

// ProcessOnlinePayment godoc
// @Summary Process online payment for an order
// @Description Allows a farmer to make an online payment for an order.
// @Tags Farmers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "message: Purchase initiated successfully along with transaction details"
// @Failure 400 {object} map[string]string "error: Payment not authorized or invalid request data"
// @Failure 401 {object} map[string]string "error: Unauthorized access"
// @Failure 500 {object} map[string]string "error: Failed to process online payment or check farmer registration"
// @Router /farmers/pay-online/{order_id} [post]
func (h *FarmerHandler) ProcessOnlinePayment(c *gin.Context) {
	// Extract farmer ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)  // Directly get the claims here
	farmerID := int(user["farmer_id"].(float64)) // Access the "farmer_id" from the claims	

	// check if farmerID is registered in farmers db
	isRegistered, err := h.FarmerService.IsFarmerRegistered(farmerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check farmer registration", "details": err.Error()})
		return
	}

	if !isRegistered {
		// check if the farmer is registered
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer is not registered"})
		return
	}

	// extract order id and prepare online payment statement
	orderID, _ := strconv.Atoi(c.Param("order_id"))
    totalCost, itemDescriptions, err := h.FarmerService.PrepareOnlinePayment(c, orderID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment preparation failed", "details": err.Error()})
        return
    }

	// prepare payment response statement to execute online payment
	paymentResponse, err := h.FarmerService.ExecuteOnlinePayment(orderID, totalCost, itemDescriptions)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process online payment", "details": err.Error()})
        return
    }

    if paymentResponse.TransactionStatus == "pending" {
        c.JSON(http.StatusOK, gin.H{
            "message":        "Purchase initiated successfully",
            "order_id":       paymentResponse.OrderID,
            "va_numbers":     paymentResponse.VaNumbers,
            "total_amount":   totalCost,
            "transaction_id": paymentResponse.TransactionID,
        })
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Payment not authorized"})
    }
}

// CheckAndProcessOrderStatus godoc
// @Summary Check and process the order status
// @Description Verifies and updates the order status based on transaction data from an external payment gateway.
// @Tags Farmers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path int true "Order ID"
// @Param midtrans_order_id query string true "Midtrans Order ID used for fetching transaction status"
// @Success 200 {object} map[string]interface{} "message: Purchase status checked successfully along with order and transaction details"
// @Failure 400 {object} map[string]string "error: Invalid order ID or transaction request"
// @Failure 409 {object} map[string]string "message: Transaction has already been processed"
// @Failure 500 {object} map[string]string "error: Failed to update order status, fetch transaction status, or process inventory update"
// @Router /farmers/check-status/{order_id} [get]
func (h *FarmerHandler) CheckAndProcessOrderStatus(c *gin.Context) {
	// derive order_id from parameter
	ctx := c.Request.Context()
	orderID := c.Param("order_id")
	midtrans_orderID := c.Param("midtrans_order_id")

	// Update the transaction status in the database
	orderIDInt, err := strconv.Atoi(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Check if the order has already been processed
	isProcessed, err := h.FarmerService.CheckIfOrderIsProcessed(ctx, orderIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check transaction processing status"})
		return
	}

	// check if the transaction is already processed
	if isProcessed {
		c.JSON(http.StatusConflict, gin.H{"message": "Transaction has already been processed"})
		return
	}

	// Fetch transaction status (Simulated call to external payment gateway like Midtrans)
	resp, err := h.FarmerService.CheckTransaction(midtrans_orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction status"})
		return
	}

	// update order status to settlement
	err = h.FarmerService.UpdateOrderStatus(ctx, orderIDInt, resp.TransactionStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Call the service to update the inventory if the order is in settlement status
	if resp.TransactionStatus == "settlement" {
		// update store quantity after settlement (need to make it into a transaction)
		err = h.FarmerService.UpdateStoreQuantity(ctx, orderIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Mark order as processed
		err = h.FarmerService.MarkOrderAsProcessed(ctx, orderIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark order as processed"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":        "Purchase status checked successfully",
		"order_id":       resp.OrderID,
		"transaction_id": resp.TransactionID,
		"status":         resp.TransactionStatus,
	})
}

// Extract review details from request
type ReviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// AddReview godoc
// @Summary Add a review for an order
// @Description Allows a farmer to add a review for an order that has reached 'settled' status.
// @Tags Farmers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order_id path int true "Order ID to which the review is being added"
// @Param review body ReviewRequest true "Review details including rating and comment"
// @Success 200 {object} map[string]interface{} "message: Review added successfully"
// @Failure 400 {object} map[string]string "error: Invalid order ID or review data; or the order is not eligible for review"
// @Failure 401 {object} map[string]string "error: Unauthorized access if JWT is missing or invalid"
// @Failure 500 {object} map[string]string "error: Internal server error if there's an issue registering the review in the database"
// @Router /farmers/review/{order_id} [post]
func (h *FarmerHandler) AddReview(c *gin.Context) {
	// Extract farmer ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	farmerID := int(user["farmer_id"].(float64))  // Access the "farmer_id" from the claims

	// check if farmer is registered in db
	isRegistered, err := h.FarmerService.IsFarmerRegistered(farmerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check farmer registration", "details": err.Error()})
		return
	}

	// if not registered
	if !isRegistered {
		// farmer is not registered
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer is not registered"})
		return
	}

	// derive order id from params
    orderID, err := strconv.Atoi(c.Param("order_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
        return
    }

    var req ReviewRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review data"})
        return
    }

    // Check if the order is settled
	var isProcessed, _ = h.FarmerService.CheckIfOrderIsProcessed(c.Request.Context(), orderID)	
	if !isProcessed {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Reviews can only be added for settled orders"})
        return
    }

    // Add the review
    if err := h.FarmerService.AddReview(c.Request.Context(), orderID, farmerID, req.Rating, req.Comment); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review", "details": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Review added successfully"})
}