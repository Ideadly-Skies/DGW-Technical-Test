package repositories

import (
	"context"
	"dgw-technical-test/internal/models/order"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	DB *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{DB: db}
}

// CreateOrder creates a new order in the database
func (r *OrderRepository) CreateOrder(ctx context.Context, farmerID int, totalPrice float64) (int, error) {
	var orderID int
	err := r.DB.QueryRow(ctx, "INSERT INTO orders (farmer_id, status, total_price) VALUES ($1, 'pending', $2) RETURNING id", farmerID, totalPrice).Scan(&orderID)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}
	return orderID, nil
}

// AddOrderItem adds an item to an existing order in the database
func (r *OrderRepository) AddOrderItem(ctx context.Context, orderID int, item models.OrderItem) error {
	_, err := r.DB.Exec(ctx, "INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)", orderID, item.ProductID, item.Quantity, item.Price)
	if err != nil {
		return fmt.Errorf("failed to add order item: %w", err)
	}
	return nil
}