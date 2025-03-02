package repositories

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type LogRepository struct {
    DB *pgxpool.Pool
}

func NewLogRepository(db *pgxpool.Pool) *LogRepository {
    return &LogRepository{DB: db}
}

// LogAction logs an administrative action in the database.
func (r *LogRepository) LogAction(ctx context.Context, adminID int, action, details string) error {
    query := "INSERT INTO logs (admin_id, action, details) VALUES ($1, $2, $3)"
    _, err := r.DB.Exec(ctx, query, adminID, action, details)
    if err != nil {
        return fmt.Errorf("failed to log action: %w", err)
    }
    return nil
}