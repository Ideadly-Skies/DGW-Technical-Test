package handlers

import (
	"dgw-technical-test/internal/models/farmer"
	"dgw-technical-test/internal/services/farmer"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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
// @Description This endpoint registers a new farmer with name, email, and password.
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