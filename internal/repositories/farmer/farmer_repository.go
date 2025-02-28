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
func (r *FarmerRepository) LogWalletTransaction(farmerID int, orderID string, amount float64, description string) error {
	// Insert the transaction into the farmers' transaction table (wallet_transactions)
	transactionQuery := `
		INSERT INTO wallet_transactions (farmer_id, order_id, transaction_type, amount, status, description, created_at, updated_at)
		VALUES ($1, $2, 'Withdraw', $3, 'pending', $4, NOW(), NOW())
	`
	_, txnErr := r.DB.Exec(context.Background(), transactionQuery, farmerID, orderID, amount, description)
	if txnErr != nil {
		return fmt.Errorf("failed to log transaction: %v", txnErr)
	}
	return nil
}

// get transaction status for wallet withdraw transaction in here

// GetWithdrawalStatus retrieves the status of a withdrawal transaction
func (r *FarmerRepository) GetWithdrawalStatus(orderID string) (map[string]interface{}, error) {
	var status string
	var amount float64

	// Query the status and amount from wallet_transactions using the order_id (assuming order_id is unique for each transaction)
	query := `SELECT status, amount FROM wallet_transactions WHERE order_id = $1`
	err := r.DB.QueryRow(context.Background(), query, orderID).Scan(&status, &amount)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal status: %v", err)
	}

	// Return the transaction details
	return map[string]interface{}{
		"order_id": orderID,
		"status":   status,
		"amount":   amount,
	}, nil
}

// GetWithdrawalAmount retrieves the amount of the withdrawal transaction
func (r *FarmerRepository) GetWithdrawalAmount(orderID string) (float64, error) {
	var amount float64
	// Query the amount based on the order_id
	query := `SELECT amount FROM wallet_transactions WHERE order_id = $1`
	err := r.DB.QueryRow(context.Background(), query, orderID).Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get withdrawal amount: %v", err)
	}
	return amount, nil
}

// GetFarmerIDByOrderID retrieves the farmer ID associated with the given order ID
func (r *FarmerRepository) GetFarmerIDByOrderID(orderID string) (int, error) {
	var farmerID int
	query := `SELECT farmer_id FROM wallet_transactions WHERE order_id = $1`
	err := r.DB.QueryRow(context.Background(), query, orderID).Scan(&farmerID)
	if err != nil {
		return 0, fmt.Errorf("failed to get farmer ID: %v", err)
	}
	return farmerID, nil
}

// UpdateFarmerWalletBalance updates the farmer's wallet balance after the transaction
func (r *FarmerRepository) UpdateFarmerWalletBalance(farmerID int, amount float64) error {
	query := `UPDATE farmers SET wallet_balance = wallet_balance + $1 WHERE id = $2`
	_, err := r.DB.Exec(context.Background(), query, amount, farmerID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %v", err)
	}
	return nil
}

// MarkTransactionAsProcessed marks the transaction as processed (completed)
func (r *FarmerRepository) MarkTransactionAsProcessed(orderID string) error {
	query := `UPDATE wallet_transactions SET status = 'settlement' WHERE order_id = $1`
	_, err := r.DB.Exec(context.Background(), query, orderID)
	if err != nil {
		return fmt.Errorf("failed to mark transaction as processed: %v", err)
	}
	return nil
}