package services

import (
	"errors"
	"dgw-technical-test/internal/repositories/farmer"
	"dgw-technical-test/internal/models/farmer"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"time"
	"os"
	"fmt"
)

type FarmerService struct {
	FarmerRepo *repositories.FarmerRepository
}

func NewFarmerService(farmerRepo *repositories.FarmerRepository) *FarmerService {
	return &FarmerService{FarmerRepo: farmerRepo}
}

// RegisterFarmer registers a new farmer with the given data
func (s *FarmerService) RegisterFarmer(name, email, hashedPassword string) error {
	return s.FarmerRepo.CreateFarmer(name, email, hashedPassword)
}

// LoginFarmer logs in a farmer using email and password
func (s *FarmerService) LoginFarmer(email, password string) (*models.Farmer, error) {
	farmer, err := s.FarmerRepo.GetFarmerByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(farmer.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate the JWT token for the farmer
	token, err := s.GenerateJWT(farmer)
	if err != nil {
		return nil, err
	}

	// Update the JWT token in the database
	err = s.FarmerRepo.UpdateFarmerJWTToken(farmer.ID, token)
	if err != nil {
		return nil, err
	}

	// Return the farmer with the new JWT token
	farmer.JWTToken = token
	return farmer, nil
}	

// GenerateJWT generates a JWT token for the farmer
func (s *FarmerService) GenerateJWT(farmer *models.Farmer) (string, error) {
	// Fetch the secret key from environment variables
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return "", fmt.Errorf("JWT_SECRET not set in environment variables")
    }

	claims := jwt.MapClaims{
		"farmer_id":    farmer.ID,
		"name":         farmer.Name,
		"email":        farmer.Email,
		"wallet_balance": farmer.WalletBalance,
		"exp":          jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	return tokenString, err
}