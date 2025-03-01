package handlers

import (
	admin "dgw-technical-test/internal/models/admin"
	"dgw-technical-test/internal/services/admin"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	AdminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{AdminService: adminService}
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