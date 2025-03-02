package services

import (
	admin "dgw-technical-test/internal/models/admin"
	admin_repo "dgw-technical-test/internal/repositories/admin"
	review_repo "dgw-technical-test/internal/repositories/review"	

	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"context"
)

type AdminService struct {
	AdminRepo *admin_repo.AdminRepository
	ReviewRepo *review_repo.ReviewRepository
}

func NewAdminService(adminRepo *admin_repo.AdminRepository, reviewRepo *review_repo.ReviewRepository) *AdminService {
	return &AdminService{
		AdminRepo: adminRepo,
		ReviewRepo: reviewRepo,
	}
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

// GetAdminByEmail retrieves an admin by their email
func (s *AdminService) GetAdminByEmail(email string) (*admin.Admin, error) {
	ad, err := s.AdminRepo.GetAdminByEmail(email)
	if err != nil {
		return nil, err
	}
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

// update review status for admin
func (s *AdminService) UpdateReviewStatus(ctx context.Context, reviewID int, status string) error {
	return s.ReviewRepo.UpdateReviewStatus(ctx, reviewID, status)
}

// DeleteRejectedReview deletes a review if its status is 'rejected'.
func (s *AdminService) DeleteRejectedReview(ctx context.Context, reviewID int) error {
	// First, check the current status of the review.
	review, err := s.ReviewRepo.GetReviewByID(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("failed to retrieve review: %w", err)
	}

	// Check if the review status is 'rejected'.
	if review.Status != "rejected" {
		return fmt.Errorf("review status is not rejected: status=%s", review.Status)
	}

	// Proceed to delete the review as its status is 'rejected'.
	err = s.ReviewRepo.DeleteReview(ctx, reviewID)
	if err != nil {
		return fmt.Errorf("failed to delete review: %w", err)
	}

	return nil
}
