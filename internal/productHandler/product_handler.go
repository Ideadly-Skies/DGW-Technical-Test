package handler

import (
	"context"
	"dgw-technical-test/config/database"
	product_models "dgw-technical-test/internal/productHandler/models"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
)

// CreateProduct godoc
// @Summary Create a new product
// @Description Creates a new product with the provided details and stores it in the database.
// @Tags products
// @Accept json
// @Produce json
// @Param product body product_models.ProductRequest true "Product information"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}{"message": "Product created successfully", "product_id": "string"}
// @Failure 400 {object} map[string]string{"message": "Invalid request body"}
// @Failure 400 {object} map[string]string{"message": "Missing required fields: Name, Price, Stock Quantity"}
// @Failure 500 {object} map[string]string{"message": "Internal server error"}
// @Router /products [post]
func CreateProduct(c echo.Context) error {
	// Extract admnin claims
	admin := c.Get("user").(*jwt.Token)
	adminClaims := admin.Claims.(jwt.MapClaims)
	adminID := adminClaims["admin_id"].(string) // extract from claims

	// Bind the incoming request body to the ProductRequest struct
	var req product_models.ProductRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(400, "Invalid request body")
	}
	
	// validate the request data
	if req.Name == "" || req.Price == 0 || req.StockQuantity == 0 {
		return echo.NewHTTPError(400, "Missing required fields: Name, Price, Stock Quantity")
	}

	// create a new Product object from the request data
	product := product_models.Product {
		Id: 		 	uuid.New().String(),
		Admin_ID:    	adminID,
		Name: 		 	req.Name,
		Description: 	req.Description,
		Price: 		    req.Price,
		Stock_Quantity: req.StockQuantity,
		Category: 		req.Category,
		ImageURL: 		req.ImageURL,
		CreatedAt: 		time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert the new product into the database
	productQuery := `
		INSERT INTO products (id, admin_id, name, description, price, stock_quantity, category, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id	
	`
	
	// productID string
	var productID string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	// Execute the query to insert the new product and retrieve the product ID
    err := config.Pool.QueryRow(ctx, productQuery, product.Id, product.Admin_ID, product.Name, product.Description, product.Price, product.Stock_Quantity, product.Category, product.ImageURL, product.CreatedAt, product.UpdatedAt).Scan(&productID)
    if err != nil {
        // Handle PostgreSQL errors, e.g., unique constraint violations, etc.
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            log.Printf("PostgreSQL error: %v", err)
            return echo.NewHTTPError(500, "Failed to insert product into the database")
        }
        log.Printf("Unexpected error: %v", err)
        return echo.NewHTTPError(500, "Internal server error")
    }

    // Optionally, return the product ID or confirmation response
    return c.JSON(201, map[string]interface{}{
        "message":  "Product created successfully",
        "product_id": productID,
    })

}

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Updates the details of an existing product with the provided information and stores it in the database.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body product_models.ProductRequest true "Product information"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}{"message": "Product updated successfully", "product_id": "string"}
// @Failure 400 {object} map[string]string{"message": "Invalid request body"}
// @Failure 404 {object} map[string]string{"message": "Product not found"}
// @Failure 500 {object} map[string]string{"message": "Internal server error"}
// @Router /products/{id} [put]
func UpdateProduct(c echo.Context) error {
    // Extract admin claims
    admin := c.Get("user").(*jwt.Token)
    adminClaims := admin.Claims.(jwt.MapClaims)
    adminID := adminClaims["admin_id"].(string) // Extract from claims

    // Get the product ID from the URL path parameter
    productID := c.Param("id")

    // Bind the incoming request body to the ProductRequest struct
    var req product_models.ProductRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(400, "Invalid request body")
    }

    // Validate the request data
    if req.Name == "" || req.Price == 0 || req.StockQuantity == 0 {
        return echo.NewHTTPError(400, "Missing required fields: Name, Price, Stock Quantity")
    }

    // Create an update query
    updateQuery := `
        UPDATE products 
        SET 
            name = $1, 
            description = $2, 
            price = $3, 
            stock_quantity = $4, 
            category = $5, 
            image_url = $6, 
            updated_at = $7 
        WHERE id = $8 AND admin_id = $9 
        RETURNING id
    `

    // Prepare the current timestamp for updating
    updatedAt := time.Now()

    var updatedProductID string
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Execute the query to update the product and retrieve the updated product ID
    err := config.Pool.QueryRow(ctx, updateQuery, req.Name, req.Description, req.Price, req.StockQuantity, req.Category, req.ImageURL, updatedAt, productID, adminID).Scan(&updatedProductID)
    if err != nil {
        // Handle PostgreSQL errors, e.g., product not found or unique constraint violations
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // If no rows are affected, the product may not exist
            if pgErr.Code == "23503" {
                return echo.NewHTTPError(404, "Product not found")
            }
            log.Printf("PostgreSQL error: %v", err)
            return echo.NewHTTPError(500, "Failed to update product in the database")
        }
        log.Printf("Unexpected error: %v", err)
        return echo.NewHTTPError(500, "Cannot Update Product That Does Not Belong To You!")
    }

    // Return the updated product ID or confirmation response
    return c.JSON(200, map[string]interface{}{
        "message":    "Product updated successfully",
        "product_id": updatedProductID,
    })
}

// DeleteProduct godoc
// @Summary Delete an existing product
// @Description Deletes the product with the given ID from the database.
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string{"message": "Product deleted successfully"}
// @Failure 404 {object} map[string]string{"message": "Product not found"}
// @Failure 500 {object} map[string]string{"message": "Internal server error"}
// @Router /products/{id} [delete]
func DeleteProduct(c echo.Context) error {
    // Extract admin claims
    admin := c.Get("user").(*jwt.Token)
    adminClaims := admin.Claims.(jwt.MapClaims)
    adminID := adminClaims["admin_id"].(string) // Extract from claims

    // Get the product ID from the URL path parameter
    productID := c.Param("id")

    // Prepare the delete query to check if the product exists and delete it
    deleteQuery := `
        DELETE FROM products 
        WHERE id = $1 AND admin_id = $2 
        RETURNING id
    `

    var deletedProductID string
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Execute the delete query and retrieve the deleted product ID
    err := config.Pool.QueryRow(ctx, deleteQuery, productID, adminID).Scan(&deletedProductID)
    if err != nil {
        // Handle PostgreSQL errors, e.g., product not found or unique constraint violations
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) {
            // If no rows are affected, the product may not exist
            if pgErr.Code == "23503" {
                return echo.NewHTTPError(404, "Product not found")
            }
            log.Printf("PostgreSQL error: %v", err)
            return echo.NewHTTPError(500, "Failed to delete product from the database")
        }
        log.Printf("Unexpected error: %v", err)
        return echo.NewHTTPError(500, "Product Not Found!")
    }

    // Return the success response after deletion
    return c.JSON(200, map[string]string{
        "message": "Product deleted successfully",
    })
}

// GetAllProducts godoc
// @Summary Get all products for the logged-in admin
// @Description Retrieves all products belonging to the admin and returns them in the response.
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} product_models.Product{"id": "string", "name": "string", "description": "string", "price": "float64", "stock_quantity": "int", "category": "string", "image_url": "string"}
// @Failure 500 {object} map[string]string{"message": "Internal server error"}
// @Router /products [get]
func GetAllProducts(c echo.Context) error {
	// Extract admin claims
	admin := c.Get("user").(*jwt.Token)
	adminClaims := admin.Claims.(jwt.MapClaims)
	adminID := adminClaims["admin_id"].(string) // Extract admin ID from claims

	// Query to retrieve all products belonging to the logged-in admin
	query := `
		SELECT id, name, description, price, stock_quantity, category, image_url, created_at, updated_at
		FROM products
		WHERE admin_id = $1
	`

	// Prepare the context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query the database for all products belonging to the admin
	rows, err := config.Pool.Query(ctx, query, adminID)
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		return echo.NewHTTPError(500, "Internal server error")
	}
	defer rows.Close()

	// Slice to hold the results
	var products []product_models.Product

	// Iterate over the rows and scan them into the Product slice
	for rows.Next() {
		var product product_models.Product
		if err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.Stock_Quantity, &product.Category, &product.ImageURL, &product.CreatedAt, &product.UpdatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			return echo.NewHTTPError(500, "Internal server error")
		}
		products = append(products, product)
	}

	// Check for any error during iteration
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		return echo.NewHTTPError(500, "Internal server error")
	}

	// Return the list of products
	return c.JSON(200, products)
}
