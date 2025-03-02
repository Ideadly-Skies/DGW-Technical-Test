package services

import (
	"context"
	product_repo "dgw-technical-test/internal/repositories/product"
	order_repo 	 "dgw-technical-test/internal/repositories/order"
	order_model  "dgw-technical-test/internal/models/order"
	"fmt"
)

type PurchaseService struct {
	ProductRepo product_repo.ProductRepository 
	OrderRepo   order_repo.OrderRepository 
}

func NewPurchaseService(productRepo product_repo.ProductRepository, orderRepo order_repo.OrderRepository) *PurchaseService {
	return &PurchaseService{ProductRepo: productRepo, OrderRepo: orderRepo}
}

type FacilitatePurchaseRequest struct {
	FarmerID int
	Items    []order_model.OrderItem `json:"Items"`
}

func (s *PurchaseService) FacilitatePurchase(ctx context.Context, req FacilitatePurchaseRequest) error {
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
	return nil
}

// CancelOrder updates the status of an order to "cancelled"
func (s *PurchaseService) CancelOrder(ctx context.Context, orderID int) error {
    return s.OrderRepo.UpdateOrderStatus(ctx, orderID, "cancelled")
}