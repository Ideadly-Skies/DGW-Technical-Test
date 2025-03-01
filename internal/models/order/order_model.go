package models

import "time"

// Order represents the structure of the orders table in the database
type Order struct {
	ID         int       `json:"id"`
	FarmerID   int       `json:"farmer_id"`
	Status     string    `json:"status"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Items     []OrderItem `json:"items"`
}

// OrderItem represents the structure of the order_items table in the database
type OrderItem struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}