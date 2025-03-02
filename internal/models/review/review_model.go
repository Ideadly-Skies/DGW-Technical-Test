package models

import "time"

// Review represents a review left by a farmer on an order item.
type Review struct {
    ID        int       `json:"id"`        // Unique identifier for the review
    OrderID   int       `json:"order_id"`  // ID of the order associated with this review
    FarmerID  int       `json:"farmer_id"` // ID of the farmer who made the review
    Rating    int       `json:"rating"`    // Rating given by the farmer, between 1 and 5
    Comment   string    `json:"comment"`   // Comment text of the review
    CreatedAt time.Time `json:"created_at"`// Timestamp when the review was created
    UpdatedAt time.Time `json:"updated_at"`// Timestamp when the review was last updated
    Status    string    `json:"status"`    // Status of the review, can be 'pending', 'approved', or 'rejected'
}