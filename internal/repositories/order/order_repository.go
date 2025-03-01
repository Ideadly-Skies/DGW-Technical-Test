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

// UpdateOrderStatus changes the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
    query := `UPDATE orders SET status = $1 WHERE id = $2`
    _, err := r.DB.Exec(ctx, query, status, orderID)
    if err != nil {
        return fmt.Errorf("failed to update order status: %w", err)
    }
    return nil
}

// GetOrderById retrieves an order by its ID
func (r *OrderRepository) GetOrderById(ctx context.Context, orderID int) (*models.Order, error) {
	var o models.Order
	query := `SELECT id, farmer_id, status, total_price, created_at, updated_at FROM orders WHERE id = $1 AND status = 'pending'`
	err := r.DB.QueryRow(ctx, query, orderID).Scan(&o.ID, &o.FarmerID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}

	// Fetch order items
	rows, err := r.DB.Query(ctx, "SELECT id, product_id, quantity, price, created_at, updated_at FROM order_items WHERE order_id = $1", orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		o.Items = append(o.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over order items: %w", err)
	}

	return &o, nil
}