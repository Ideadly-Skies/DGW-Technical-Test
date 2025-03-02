package services

import (
	"context"
	product_repo "dgw-technical-test/internal/repositories/product"
	order_repo 	 "dgw-technical-test/internal/repositories/order"
	log_repo 	 "dgw-technical-test/internal/repositories/log"
	order_model  "dgw-technical-test/internal/models/order"
	"fmt"
)

type PurchaseService struct {
	ProductRepo product_repo.ProductRepository 
	OrderRepo   order_repo.OrderRepository
	LogRepo		log_repo.LogRepository
}

func NewPurchaseService(productRepo product_repo.ProductRepository, orderRepo order_repo.OrderRepository, logRepo log_repo.LogRepository) *PurchaseService {
	return &PurchaseService{
		ProductRepo: productRepo,
		OrderRepo: orderRepo,
		LogRepo: logRepo,	// Initialize the log repo
	}
}

type FacilitatePurchaseRequest struct {
	FarmerID int
	Items    []order_model.OrderItem `json:"Items"`
}

func (s *PurchaseService) FacilitatePurchase(ctx context.Context, adminID int, req FacilitatePurchaseRequest) error {
	var total float64

	// validation for the product inputted
	for i, item := range req.Items {
		fmt.Println("item: ", item)
		fmt.Println("item.ProductID: ", item.ProductID)

		// check if the product is available at the store
		product, err := s.ProductRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return err
		}
		
		// check if the stock quantity is readily available 
		if product.StockQuantity < item.Quantity {
			return fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		// update the price and the total amount that you have to pay
		req.Items[i].Price = product.Price
		total += product.Price * float64(item.Quantity)
	}

	// create the order with the status pending
	orderID, err := s.OrderRepo.CreateOrder(ctx, req.FarmerID, total)
	if err != nil {
		return err
	}

	for _, item := range req.Items {
		// add each order item to the order
		err := s.OrderRepo.AddOrderItem(ctx, orderID, item)
		if err != nil {
			return err
		}
	}

	// log successful order creation
    logDetails := fmt.Sprintf("Admin %d facilitated a purchase for farmerID %d with total $%.2f", adminID, req.FarmerID, total)
    if err := s.LogRepo.LogAction(ctx, adminID, "Facilitate Purchase", logDetails); err != nil {
        return fmt.Errorf("failed to log purchase facilitation: %w", err)
    }
	
	return nil
}

// CancelOrder updates the status of an order to "cancelled"
func (s *PurchaseService) CancelOrder(ctx context.Context,adminID int, orderID int) error {
	if err := s.OrderRepo.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
		return fmt.Errorf("failed to cancel order: %v", err)
	}

	// Log this action
	action := "Cancel Order"
	details := fmt.Sprintf("Order ID %d cancelled by Admin ID %d", orderID, adminID)
	if err := s.LogRepo.LogAction(ctx, adminID, action, details); err != nil {
		return fmt.Errorf("failed to log cancel order action: %v", err)
	}

	return nil
}