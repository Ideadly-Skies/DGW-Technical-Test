package handlers

import (
	admin_model    "dgw-technical-test/internal/models/admin"

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
// @Description Register a new administrator with name, email, password, and role.
// @Tags admin
// @Accept json
// @Produce json
// @Param admin body admin_model.RegisterRequest true "Admin Registration Data"
// @Success 200 {object} map[string]interface{} "message: Admin registered successfully"
// @Failure 400 {object} map[string]string "message: Invalid request"
// @Failure 500 {object} map[string]string "message: Could not register admin"
// @Router /admins/register [post]
func (h *AdminHandler) RegisterAdmin(c *gin.Context) {
	var req admin_model.RegisterRequest
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
// @Description Admin login with email and password.
// @Tags Admin
// @Accept json
// @Produce json
// @Param admin body admin_model.LoginRequest true "Admin Login Data"
// @Success 200 {object} map[string]interface{} "token, name, email: Admin login data"
// @Failure 400 {object} map[string]string "message: Invalid request"
// @Failure 401 {object} map[string]string "message: Invalid email or password"
// @Router /admins/login [post]
func (h *AdminHandler) LoginAdmin(c *gin.Context) {
	var req admin_model.LoginRequest
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
// @Param farmerID path int true "Farmer ID"
// @Param request body purchase_services.FacilitatePurchaseRequest true "Purchase Request Data"
// @Success 200 {object} map[string]interface{} "message: Purchase facilitated successfully"
// @Failure 400 {object} map[string]string "message: Invalid request body or farmer ID"
// @Failure 404 {object} map[string]string "message: Admin not found"
// @Failure 500 {object} map[string]string "message: Failed to facilitate purchase"
// @Router /admins/facilitate-purchase/{farmerID} [post]
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

// CancelOrderHandler godoc
// @Summary Cancel an order
// @Description Admin cancels an order
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param orderID path int true "Order ID"
// @Success 200 {object} map[string]interface{} "message: Order cancelled successfully"
// @Failure 400 {object} map[string]string "error: Invalid order ID"
// @Failure 404 {object} map[string]string "message: Admin not found"
// @Failure 500 {object} map[string]string "error: Failed to cancel order"
// @Router /admins/cancel-order/{orderID} [put]
func (h *AdminHandler) CancelOrderHandler(c *gin.Context) {
	// authentication - extract admin ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	adminEmail := user["email"].(string)

	// Retrieve admin from database to get adminID
	admin, err := h.AdminService.GetAdminByEmail(adminEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}

	// Extract orderId from URL parameter
	orderIDParam := c.Param("orderID")
	orderID, err := strconv.Atoi(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	// Invoke CancelOrder service by admin
	if err := h.PurchaseService.CancelOrder(c.Request.Context(), admin.ID, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// ApproveOrRejectReview godoc
// @Summary Approve or Reject a review
// @Description Admin approves or rejects a review based on review ID and status query parameter
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param review_id path int true "Review ID"
// @Param status query string true "Review status ('approved' or 'rejected')"
// @Success 200 {object} map[string]interface{} "message: Review status updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid review ID or status value"
// @Failure 404 {object} map[string]string "message: Admin not found"
// @Failure 500 {object} map[string]string "error: Failed to update review status"
// @Router /admins/reviews/{review_id} [post]
func (h *AdminHandler) ApproveOrRejectReview(c *gin.Context) {
	// authentication - extract admin ID from JWT claims
	user := c.MustGet("user").(jwt.MapClaims)
	adminEmail := user["email"].(string)

	// check if admin exists within the db
	_, err := h.AdminService.GetAdminByEmail(adminEmail)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Admin not found"})
		return
	}	
		
	// retrieve review_id from 
	reviewIDParam := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	status := c.Query("status")
	if status != "approved" && status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	if err := h.AdminService.UpdateReviewStatus(c.Request.Context(), reviewID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review status updated successfully"})
}


// HandleDeleteRejectedReview godoc
// @Summary Delete a rejected review
// @Description Admin deletes a review that has been marked as 'rejected'
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param review_id path int true "Review ID"
// @Success 200 {object} map[string]interface{} "message: Review deleted successfully"
// @Failure 400 {object} map[string]string "error: Invalid review ID"
// @Failure 500 {object} map[string]string "error: Failed to delete review"
// @Router /admins/delete-review/{review_id} [delete]
func (h *AdminHandler) HandleDeleteRejectedReview(c *gin.Context) {
	reviewIDParam := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	err = h.AdminService.DeleteRejectedReview(c.Request.Context(), reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted successfully"})
}
