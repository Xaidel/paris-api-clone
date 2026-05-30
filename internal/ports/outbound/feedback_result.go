package ports

import "time"

// FeedbackResult exposes a transaction feedback to inbound adapters.
type FeedbackResult struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	Kind          string    `json:"kind"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
