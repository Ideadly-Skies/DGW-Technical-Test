package models

// RegisterRequest represents the data needed to register a farmer
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the data needed to login a farmer
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the data returned upon successful login
type LoginResponse struct {
	Token         string  `json:"token"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	WalletBalance float64 `json:"wallet_balance"`
}

// Farmer represents the structure of the farmer data stored in the database
type Farmer struct {
	ID           int     `json:"id"`            
	Name         string  `json:"name"`          
	Email        string  `json:"email"`         
	Password     string  `json:"password"`      
	Address      string  `json:"address"`       
	PhoneNumber  string  `json:"phone_number"`  
	FarmType     string  `json:"farm_type"`     
	WalletBalance float64 `json:"wallet_balance"` 
	CreatedAt    string  `json:"created_at"`    
	UpdatedAt    string  `json:"updated_at"`
	JWTToken     string    
}