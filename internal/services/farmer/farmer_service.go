package services

import (
	"dgw-technical-test/internal/models/farmer"
	"dgw-technical-test/internal/repositories/farmer"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
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

// GetFarmerIDByEmail retrieves the farmer ID using the email
func (s *FarmerService) GetFarmerIDByEmail(email string) (int, error) {
	farmer, err := s.FarmerRepo.GetFarmerByEmail(email)
	if err != nil {
		return 0, err
	}
	return farmer.ID, nil
}

// GetFarmerWalletBalance retrieves the wallet balance of the farmer by their ID
func (s *FarmerService) GetFarmerWalletBalance(farmerID int) (float64, error) {
	walletBalance, err := s.FarmerRepo.GetFarmerWalletBalance(farmerID)
	if err != nil {
		return 0, err
	}
	return walletBalance, nil
}

// transaction component for farmer
var coreAPI coreapi.Client

func Init() {
	// retrieve server key from .env
	ServerKey := os.Getenv("MIDTRANS_SERVER_KEY")

	coreAPI = coreapi.Client{}
	coreAPI.New(ServerKey, midtrans.Sandbox)
}

// WithdrawMoney handles the process of withdrawing funds for a farmer
func (s *FarmerService) WithdrawMoney(farmerID int, amount float64, farmerName string) (string, string, string, error) {
	// Initialize Midtrans
	Init()

	// Generate order ID
	orderID := fmt.Sprintf("wd-%d-%d", farmerID, time.Now().Unix())

	// Generate Customer Field Value
	customFieldValue := fmt.Sprintf("facilitating withdraw request for %s", farmerName)

	// Create a Midtrans charge request
	request := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeBankTransfer,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  orderID,
			GrossAmt: int64(amount), // Midtrans uses IDR natively
		},
		BankTransfer: &coreapi.BankTransferDetails{
			Bank: midtrans.BankBca, // Use a specific bank for withdrawals
		},
		CustomField1: &customFieldValue,
	}

	// Send the charge request to Midtrans
	resp, err := coreAPI.ChargeTransaction(request)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to process withdrawal: %v", err)
	}

	// Check if VA numbers exist
	var vaNumber string
	if len(resp.VaNumbers) > 0 {
		vaNumber = resp.VaNumbers[0].VANumber // Get the first VA number
	} else {
		vaNumber = "No virtual account number available" // Fallback if no VA is provided
	}

	// Log the transaction in the wallet_transactions table
	description := fmt.Sprintf("Withdrawal initiated for %s", farmerName)
	if err := s.FarmerRepo.LogWalletTransaction(farmerID, orderID, amount, description); err != nil {
		return "", "", "", fmt.Errorf("Failed to log transaction: %v", err)
	}

	return resp.TransactionID, resp.OrderID, vaNumber, nil
}

// CheckWithdrawalStatus checks the withdrawal status and updates the farmer's wallet if successful
func (s *FarmerService) CheckWithdrawalStatus(orderID string) (map[string]interface{}, error) {
	resp , err := coreAPI.CheckTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction status: %v", err)
	}

	// If the transaction is successful (settlement), update the wallet balance
	if resp.TransactionStatus == "settlement" {
		// Get the withdrawal amount
		amount, err := s.FarmerRepo.GetWithdrawalAmount(orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch withdrawal amount: %v", err)
		}

		// Get the farmer ID associated with this order
		farmerID, err := s.FarmerRepo.GetFarmerIDByOrderID(orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch farmer ID: %v", err)
		}

		// Update the farmer's wallet balance
		err = s.FarmerRepo.UpdateFarmerWalletBalance(farmerID, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to update wallet balance: %v", err)
		}

		// Mark the transaction as processed
		err = s.FarmerRepo.MarkTransactionAsProcessed(orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to mark transaction as processed: %v", err)
		}
	}

	// Return the updated status of the transaction
	return map[string]interface{}{"transaction_status": resp.TransactionStatus}, nil
}
