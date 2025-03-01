package services

import (
	admin "dgw-technical-test/internal/models/admin"
	"dgw-technical-test/internal/repositories/admin"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	AdminRepo *repositories.AdminRepository
}

func NewAdminService(adminRepo *repositories.AdminRepository) *AdminService {
	return &AdminService{AdminRepo: adminRepo}
}

// RegisterAdmin registers a new admin with the given data
func (s *AdminService) RegisterAdmin(name, email, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.AdminRepo.CreateAdmin(name, email, string(hashedPassword), role)
}

// LoginAdmin logs in an admin using email and password
func (s *AdminService) LoginAdmin(email, password string) (*admin.Admin, error) {
	ad, err := s.AdminRepo.GetAdminByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(ad.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate the JWT token for the admin
	token, err := s.GenerateJWT(ad)
	if err != nil {
		return nil, err
	}

	// Update the JWT token in the database
	err = s.AdminRepo.UpdateAdminJWTToken(ad.ID, token)
	if err != nil {
		return nil, err
	}

	// Return the admin with the new JWT token
	ad.JWTToken = token
	return ad, nil
}

// GenerateJWT generates a JWT token for the admin
func (s *AdminService) GenerateJWT(ad *admin.Admin) (string, error) {
	// Fetch the secret key from environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not set in environment variables")
	}

	claims := jwt.MapClaims{
		"admin_id":    ad.ID,
		"name":        ad.Name,
		"email":       ad.Email,
		"role":        ad.Role,
		"exp":         jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	return tokenString, err
}