package repositories

import (
	"context"
	admin "dgw-technical-test/internal/models/admin"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// handle admin related query
type AdminRepository struct {
	DB *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{DB: db}
}

// CreateAdmin inserts a new admin into the database
func (r *AdminRepository) CreateAdmin(name, email, hashedPassword, role string) error {
	query := `INSERT INTO admins (name, email, password, role) VALUES ($1, $2, $3, $4) RETURNING id`
	_, err := r.DB.Exec(context.Background(), query, name, email, hashedPassword, role)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	return nil
}

// GetAdminByEmail fetches an admin by email
func (r *AdminRepository) GetAdminByEmail(email string) (*admin.Admin, error) {
	query := `SELECT id, name, email, password, role FROM admins WHERE email = $1`
	var ad admin.Admin
	err := r.DB.QueryRow(context.Background(), query, email).Scan(&ad.ID, &ad.Name, &ad.Email, &ad.Password, &ad.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin: %w", err)
	}
	return &ad, nil
}

// UpdateAdminJWTToken updates the JWT token for the admin in the database
func (r *AdminRepository) UpdateAdminJWTToken(adminID int, token string) error {
	query := `UPDATE admins SET jwt_token = $1 WHERE id = $2`
	_, err := r.DB.Exec(context.Background(), query, token, adminID)
	if err != nil {
		return fmt.Errorf("failed to update admin JWT token: %w", err)
	}
	return nil
}