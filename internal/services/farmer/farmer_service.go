package services

import (
	"dgw-technical-test/internal/models/farmer"
	farmer_repo "dgw-technical-test/internal/repositories/farmer"
	product_repo "dgw-technical-test/internal/repositories/product"
	order_repo "dgw-technical-test/internal/repositories/order"

	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"context"
	"strings"	
)

type FarmerService struct {
	FarmerRepo  *farmer_repo.FarmerRepository
	ProductRepo *product_repo.ProductRepository
	OrderRepo   *order_repo.OrderRepository
}

func NewFarmerService(farmerRepo *farmer_repo.FarmerRepository, productRepo *product_repo.ProductRepository, orderRepo *order_repo.OrderRepository) *FarmerService {
	return &FarmerService{
		FarmerRepo:  farmerRepo,
		ProductRepo: productRepo,
		OrderRepo:   orderRepo,
	}
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

// IsFarmerRegistered checks if a farmer is registered by their ID
func (s *FarmerService) IsFarmerRegistered(farmerID int) (bool, error) {
	_, err := s.FarmerRepo.GetFarmerByID(farmerID)
	if err != nil {
		// farmer is not registered
		return false, err
	}
	// farmer is registered
	return true, nil
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
var CoreAPI coreapi.Client

func Init() {
	// retrieve server key from .env
	ServerKey := os.Getenv("MIDTRANS_SERVER_KEY")

	CoreAPI = coreapi.Client{}
	CoreAPI.New(ServerKey, midtrans.Sandbox)
}

// WithdrawMoney handles the process of withdrawing funds for a farmer
func (s *FarmerService) WithdrawMoney(farmerID int, amount float64, farmerName string) (string, string, string, error) {
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
	resp, err := CoreAPI.ChargeTransaction(request)
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
	resp , err := CoreAPI.CheckTransaction(orderID)
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

// process wallet payment for farmers
func (s *FarmerService) ProcessWalletPayment(ctx context.Context, farmerID, orderID int) error {
	// Process the order payment
	err := s.FarmerRepo.ProcessOrder(ctx, fmt.Sprintf("%d", orderID), farmerID)
	if err != nil {
		return fmt.Errorf("failed to process order: %v", err)
	}

	// return no error because there appears to be no error :)
	return nil
}

// prepare online statement
func (s *FarmerService) PrepareOnlinePayment(ctx context.Context, orderID int) (float64, []string, error) {
    // Fetch the order to calculate total cost and prepare item descriptions
    order, err := s.OrderRepo.GetOrderById(ctx, orderID)
    if err != nil {
        return 0, nil, err
    }

    var itemDescriptions []string
    totalCost := 0.0

    for _, item := range order.Items {
		product, err := s.ProductRepo.GetProductByID(ctx, item.ProductID)
        if err != nil {
            return 0, nil, err
        }
        itemDescription := fmt.Sprintf("%s x%d", product.Name, item.Quantity)
        itemDescriptions = append(itemDescriptions, itemDescription)
        totalCost += product.Price * float64(item.Quantity)
    }

    return totalCost, itemDescriptions, nil
}

// execute online statement for the farmer
func (s *FarmerService) ExecuteOnlinePayment(orderID int, totalCost float64, description []string) (*coreapi.ChargeResponse, error) {
	orderIDStr := fmt.Sprintf("store-%d-%d", orderID, time.Now().Unix())
	descriptionStr := strings.Join(description, ", ")

    req := &coreapi.ChargeReq{
        PaymentType: coreapi.PaymentTypeBankTransfer,
        TransactionDetails: midtrans.TransactionDetails{
            OrderID:  orderIDStr,
            GrossAmt: int64(totalCost),
        },
        BankTransfer: &coreapi.BankTransferDetails{
            Bank: midtrans.BankBca,
        },
        CustomField1: &descriptionStr,
    }

    response, err := CoreAPI.ChargeTransaction(req)
    if err != nil {
        return nil, err
    }

    return response, nil
}

// CheckIfOrderIsProcessed checks if a specific order has already been processed
func (s *FarmerService) CheckIfOrderIsProcessed(ctx context.Context, orderID int) (bool, error) {
	isProcessed, err := s.OrderRepo.CheckOrderProcessed(ctx, orderID)
	if err != nil {
		return false, fmt.Errorf("service failed to check if order is processed: %v", err)
	}
	return isProcessed, nil
}

// CheckTransaction checks the status of a transaction by order ID
func (s *FarmerService) CheckTransaction(orderID string) (*coreapi.TransactionStatusResponse, error) {
	resp, err := CoreAPI.CheckTransaction(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to check transaction: %v", err)
	}
	return resp, nil
}

// UpdateOrderStatus updates the order status for an order id
func (s *FarmerService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	// Call the repository function to update the order status
	err := s.OrderRepo.UpdateOrderStatus(ctx, orderID, status)
	if err != nil {
		return fmt.Errorf("error updating order status: %v", err)
	}
	return nil
}

// MarkOrderAsProcessed marks an order as processed in the database
func (s *FarmerService) MarkOrderAsProcessed(ctx context.Context, orderID int) error {
	err := s.OrderRepo.MarkOrderAsProcessed(ctx, orderID)
	if err != nil {
		return fmt.Errorf("service failed to mark order as processed: %v", err)
	}
	return nil
}

// update store quantity after transaction reaches settlement status
func (s *FarmerService) UpdateStoreQuantity(ctx context.Context, orderID int) error {
	err := s.OrderRepo.UpdateStoreQuantity(ctx, orderID)
	if err != nil {
		return fmt.Errorf("service failed to update store quantity: %v", err)
	}
	return nil
}