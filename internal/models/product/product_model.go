package models

import "time"

// Product represents the structure of a product data stored in the database
type Product struct {
	ID            int       `json:"id"`
	SupplierID    int       `json:"supplier_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	Category      string    `json:"category"`
	Brand         string    `json:"brand"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}