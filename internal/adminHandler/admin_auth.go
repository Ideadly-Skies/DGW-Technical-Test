package handler

import (
	"context"
	"errors"
	"fmt"
	config "dgw-technical-test/config/database"
	"dgw-technical-test/internal/adminHandler/models"
	"dgw-technical-test/utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// jwtSecret
var jwtSecret = os.Getenv("JWT_SECRET")

// RegisterAdmin godoc
// @Summary Register a store admin
// @Description Registers a new admin with the provided details
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Admin registration request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /store-admin/register [post]
func RegisterAdmin(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Request"})
	}

	// validate name, email, and password request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "All fields are required"})
	}

	// Validate email format
	if !utils.ValidateEmail(req.Email) {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid email format"})
	}

	// convert email to lowercase
	email := strings.ToLower(req.Email)

	// Validate password strength (e.g., min 8 chars)
	if len(req.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Password must be at least 8 characters long"})
	}

	// Hash the password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	}

	// Insert into store_admins table
	adminQuery := "INSERT INTO admins (name, email, password) VALUES ($1, $2, $3) RETURNING id"
	var adminID string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// QueryRow to insert the new admin into the db
	err = config.Pool.QueryRow(ctx, adminQuery, req.Name, email, string(hashPassword)).Scan(&adminID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return c.JSON(http.StatusBadRequest, map[string]string{"message": "Email already registered"})
			}
			log.Printf("PostgreSQL error: %v", err)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Store admin %s registered successfully", req.Name),
		"email":   email,
	})
}

// LoginAdmin godoc
// @Summary Login an admin
// @Description Authenticates an admin and returns a JWT token
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Admin login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /store-admin/login [post]
func LoginStoreAdmin(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Request"})
	}
	
	// Convert email to lowercase for case-insensitive comparison
	req.Email = strings.ToLower(req.Email)

	// Fetch admin details
	var admin models.Admin
	query := `SELECT id, name, email, password FROM admins WHERE email = $1`
	err := config.Pool.QueryRow(context.Background(), query, req.Email).Scan(
		&admin.ID, &admin.Name, &admin.Email, &admin.Password,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid email or password"})
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid email or password"})
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"name":     admin.Name,
		"email":    admin.Email,
		"exp":      jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to generate token"})
	}

	// Update the JWT token in the database
	updateQuery := "UPDATE admins SET jwt_token = $1 WHERE id = $2"
	_, err = config.Pool.Exec(context.Background(), updateQuery, tokenString, admin.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to update token"})
	}

	// Return login response
	return c.JSON(http.StatusOK, models.LoginResponse{
		Token: tokenString,
		Name:  admin.Name,
		Email: admin.Email,
	})
}