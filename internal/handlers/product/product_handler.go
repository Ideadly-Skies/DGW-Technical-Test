package handlers

import (
	"dgw-technical-test/internal/services/product"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	ProductService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{ProductService: productService}
}

// GetAllProducts godoc
// @Summary Retrieve all products
// @Description get all the products 
// @Tags products
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Product
// @Router /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.ProductService.GetAllProductsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seed products"})
		return
	}
	c.JSON(http.StatusOK, products)
}
