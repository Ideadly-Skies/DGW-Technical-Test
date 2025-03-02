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

// CheckOrderProcessed checks if an order has been processed
func (r *OrderRepository) CheckOrderProcessed(ctx context.Context, orderID int) (bool, error) {
	var isProcessed bool
	err := r.DB.QueryRow(ctx, "SELECT is_processed FROM orders WHERE id = $1", orderID).Scan(&isProcessed)
	if err != nil {
		return false, fmt.Errorf("failed to check order processing status: %w", err)
	}
	return isProcessed, nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	_, err := r.DB.Exec(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

// MarkOrderAsProcessed marks the order as processed
func (r *OrderRepository) MarkOrderAsProcessed(ctx context.Context, orderID int) error {
	_, err := r.DB.Exec(ctx, "UPDATE orders SET is_processed = TRUE WHERE id = $1", orderID)
	if err != nil {
		return fmt.Errorf("failed to mark order as processed: %w", err)
	}
	return nil
}

// UpdateStoreQuantity updates the stock quantities for products based on an order without using transactions.
func (r *OrderRepository) UpdateStoreQuantity(ctx context.Context, orderID int) error {
    // First, fetch the order items.
    rows, err := r.DB.Query(ctx, "SELECT product_id, quantity FROM order_items WHERE order_id = $1", orderID)
    if err != nil {
        return fmt.Errorf("failed to fetch order items: %w", err)
    }
    defer rows.Close()

    // Iterate through each order item and update stock quantities directly.
    for rows.Next() {
        var productID, quantity int
        if err := rows.Scan(&productID, &quantity); err != nil {
            return fmt.Errorf("failed to scan order item: %w", err)
        }

        // Check current stock quantity to ensure there's enough stock.
        var stockQuantity int
        err = r.DB.QueryRow(ctx, "SELECT stock_quantity FROM products WHERE id = $1", productID).Scan(&stockQuantity)
        if err != nil {
            return fmt.Errorf("failed to fetch product stock quantity: %w", err)
        }

        if stockQuantity < quantity {
            // Optionally handle the situation, e.g., log or adjust business logic to prevent this path.
            return fmt.Errorf("insufficient stock for product ID %d", productID)
        }

        // Update the product stock quantity directly.
        _, err = r.DB.Exec(ctx, "UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2", quantity, productID)
        if err != nil {
            return fmt.Errorf("failed to update product stock: %w", err)
        }
    }

    return nil
}