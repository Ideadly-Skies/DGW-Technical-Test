package repositories

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ReviewRepository struct {
    DB *pgxpool.Pool
}

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository {
    return &ReviewRepository{DB: db}
}

// CreateReview logs a new review in the database
func (r *ReviewRepository) CreateReview(ctx context.Context, orderID, farmerID int, rating int, comment, status string) error {
    _, err := r.DB.Exec(ctx, "INSERT INTO reviews (order_id, farmer_id, rating, comment, status) VALUES ($1, $2, $3, $4, $5)", orderID, farmerID, rating, comment, status)
    if err != nil {
        return fmt.Errorf("failed to insert review: %w", err)
    }
    return nil
}