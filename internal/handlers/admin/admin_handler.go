package handlers

import (
	admin "dgw-technical-test/internal/models/admin"
	admin_services "dgw-technical-test/internal/services/admin"
	purchase_services "dgw-technical-test/internal/services/purchase"	
	
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type AdminHandler struct {
	AdminService *admin_services.AdminService
	PurchaseService *purchase_services.PurchaseService
}

func NewAdminHandler(adminService *admin_services.AdminService, purchaseService *purchase_services.PurchaseService) *AdminHandler {
	return &AdminHandler{
		AdminService: adminService,
		PurchaseService: purchaseService,
	}
}

// RegisterAdmin godoc
// @Summary Register a new admin
func (h *AdminHandler) RegisterAdmin(c *gin.Context) {
	var req admin.RegisterRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		return
	}

	err := h.AdminService.RegisterAdmin(req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not register admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin registered successfully"})
}

// LoginAdmin godoc
// @Summary Login an admin
func (h *AdminHandler) LoginAdmin(c *gin.Context) {
	var req admin.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Request"})
		return
	}

	adminData, err := h.AdminService.LoginAdmin(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": adminData.JWTToken,
		"name":  adminData.Name,
		"email": adminData.Email,
	})
}

// FacilitatePurchase godoc
// @Summary Facilitate a purchase for a farmer
// @Description Admin facilitates a purchase by logging the order and adjusting inventory
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body FacilitatePurchaseRequest true "Purchase request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/purchase [post]
func (h *AdminHandler) FacilitatePurchase(c *gin.Context) {
	// authentication - extract admin ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	adminEmail := user["email"].(string) // access the adminEmail

	// Assuming admin.ID exists and GetAdminByEmail returns an admin object which includes an ID
	if admin, err := h.AdminService.GetAdminByEmail(adminEmail); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	} else {

		// Extract farmerID from URL parameter
		farmerID := c.Param("farmerID")
		if farmerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Farmer ID is required"})
			return
		}

		// Extract items from JSON body
		var req purchase_services.FacilitatePurchaseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
			return
		}
		
		// convert farmerId into int and bind it with the request
		farmerIDInt, err := strconv.Atoi(farmerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid Farmer ID"})
			return
		}
		req.FarmerID = farmerIDInt 

		// Now use admin.ID in your service call
		if err := h.PurchaseService.FacilitatePurchase(c.Request.Context(), admin.ID, req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to facilitate purchase", "error": err.Error()})
			return
		}
	}

	// purchase facilitated successfully!
	c.JSON(http.StatusOK, gin.H{"message": "Purchase facilitated successfully"})
}

// cancel an order made by an admin
func (h *AdminHandler) CancelOrderHandler(c *gin.Context) {
	// authentication - extract admin ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	adminEmail := user["email"].(string) // access the adminEmail

	// check if admin exists in the database
	if _, err := h.AdminService.GetAdminByEmail(adminEmail); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}
	
	// derive orderId from URL parameter 
	orderIDParam := c.Param("orderID")
    orderID, err := strconv.Atoi(orderIDParam)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
        return
    }

	// invoke CancelOrder service by admin 
    if err := h.PurchaseService.CancelOrder(c.Request.Context(), orderID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order", "details": err.Error()})
        return
    }

	// return the appropriate failure message
    c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}