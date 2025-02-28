package repositories

import (
	"context"
	"dgw-technical-test/internal/models/farmer"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FarmerRepository interacts with the database to handle farmer-related queries
type FarmerRepository struct {
	DB *pgxpool.Pool
}

func NewFarmerRepository(db *pgxpool.Pool) *FarmerRepository {
	return &FarmerRepository{DB: db}
}

// CreateFarmer inserts a new farmer into the database
func (r *FarmerRepository) CreateFarmer(name, email, hashedPassword string) error {
	query := `INSERT INTO farmers (name, email, password, wallet_balance) VALUES ($1, $2, $3, 0) RETURNING id`
	_, err := r.DB.Exec(context.Background(), query, name, email, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create farmer: %w", err)
	}
	return nil
}

// GetFarmerByEmail fetches a farmer by email
func (r *FarmerRepository) GetFarmerByEmail(email string) (*models.Farmer, error) {
	query := `SELECT id, name, email, password, wallet_balance FROM farmers WHERE email = $1`
	var farmer models.Farmer
	err := r.DB.QueryRow(context.Background(), query, email).Scan(&farmer.ID, &farmer.Name, &farmer.Email, &farmer.Password, &farmer.WalletBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer: %w", err)
	}
	return &farmer, nil
}

// UpdateFarmerJWTToken updates the JWT token for the farmer in the database
func (r *FarmerRepository) UpdateFarmerJWTToken(farmerID int, token string) error {
	query := `UPDATE farmers SET jwt_token = $1 WHERE id = $2`
	_, err := r.DB.Exec(context.Background(), query, token, farmerID)
	if err != nil {
		return fmt.Errorf("failed to update farmer JWT token: %w", err)
	}
	return nil
}

// GetFarmerWalletBalance retrieves the wallet balance of the farmer by their ID
func (r *FarmerRepository) GetFarmerWalletBalance(farmerID int) (float64, error) {
    var walletBalance float64
	query := "SELECT wallet_balance FROM farmers WHERE id = $1"
	err := r.DB.QueryRow(context.Background(), query, farmerID).Scan(&walletBalance)
    if err != nil {
        return 0, err
    }
    return walletBalance, nil
}

// LogWalletTransaction logs a new transaction in the wallet_transactions table (PENDING)
func (r *FarmerRepository) LogWalletTransaction(farmerID int, amount float64, description string) error {
	// Insert the transaction into the farmers' transaction table (wallet_transactions)
	transactionQuery := `
		INSERT INTO wallet_transactions (farmer_id, transaction_type, amount, status, description, created_at, updated_at)
		VALUES ($1, 'Withdraw', $2, 'pending', $3, NOW(), NOW())
	`
	_, txnErr := r.DB.Exec(context.Background(), transactionQuery, farmerID, amount, description)
	if txnErr != nil {
		return fmt.Errorf("failed to log transaction: %v", txnErr)
	}
	return nil
}