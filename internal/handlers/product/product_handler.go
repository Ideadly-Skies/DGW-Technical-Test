package handlers

import (
	"dgw-technical-test/internal/services/product"
	"net/http"
	_ "dgw-technical-test/internal/models/product"

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
// @Description Retrieves a list of all available products from the database, providing detailed information about each product.
// @Tags products
// @Accept  json
// @Produce  json
// @Success 200 {array} models.Product "An array of products with detailed information including ID, name, description, price, and stock quantity"
// @Failure 500 {object} map[string]string "error: Unable to fetch product data due to internal server error"
// @Router /products [get]
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	products, err := h.ProductService.GetAllProductsService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve seed products"})
		return
	}
	c.JSON(http.StatusOK, products)
}