package models

import "time"

// Product struct
type Product struct {
	Id            	  string  	 `json:"id"`
	Admin_ID      	  string  	 `json:"admin_id"`
	Name     	  	  string  	 `json:"name"`
	Description   	  string  	 `json:"description"`
	Price		      float64 	 `json:"price"`
	Stock_Quantity 	  int 	  	 `json:"stock_quantity"`
	Category 		  string  	 `json:"category"`
	ImageURL	   	  string  	 `json:"image_url"`
	CreatedAt 	  	  time.Time  `json:"created_at"`
	UpdatedAt 	  	  time.Time  `json:"updated_at"`
}

// Product struct to bind to the request body
type ProductRequest struct {
	Name           string  `json:"name" validate:"required"`
	Description    string  `json:"description"`
	Price          float64 `json:"price" validate:"required"`
	StockQuantity  int     `json:"stock_quantity" validate:"required"`
	Category       string  `json:"category"`
	ImageURL       string  `json:"image_url"`
}