package repositories

import (
    "context"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
	review_model "dgw-technical-test/internal/models/review"
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

// update review status (for admin use)
func (r *ReviewRepository) UpdateReviewStatus(ctx context.Context, reviewID int, status string) error {
	query := `UPDATE reviews SET status = $1 WHERE id = $2 AND status = 'pending'`
	if _, err := r.DB.Exec(ctx, query, status, reviewID); err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}
	return nil
}

// GetReviewByID retrieves a review by its ID.
func (r *ReviewRepository) GetReviewByID(ctx context.Context, reviewID int) (*review_model.Review, error) {
	var rev review_model.Review
	err := r.DB.QueryRow(ctx, "SELECT id, status FROM reviews WHERE id = $1", reviewID).Scan(&rev.ID, &rev.Status)
	if err != nil {
		return nil, fmt.Errorf("error fetching review: %w", err)
	}
	return &rev, nil
}

// DeleteReview deletes a review from the database.
func (r *ReviewRepository) DeleteReview(ctx context.Context, reviewID int) error {
	_, err := r.DB.Exec(ctx, "DELETE FROM reviews WHERE id = $1", reviewID)
	if err != nil {
		return fmt.Errorf("error deleting review: %w", err)
	}
	return nil
}
