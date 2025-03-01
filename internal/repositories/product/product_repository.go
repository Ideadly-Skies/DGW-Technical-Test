package repositories

import (
	"context"
	"dgw-technical-test/internal/models/product"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	DB *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{DB: db}
}

// GetAllProducts retrieves all products from the database that are available on the online store
func (r *ProductRepository) GetAllProductsRepo() ([]models.Product, error) {
	query := `SELECT id, supplier_id, name, description, price, stock_quantity, category, brand, created_at, updated_at FROM products`
	rows, err := r.DB.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.SupplierID, &p.Name, &p.Description, &p.Price, &p.StockQuantity, &p.Category, &p.Brand, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}