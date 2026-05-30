package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// FeedbackRepository persists transaction feedback.
type FeedbackRepository interface {
	Create(ctx context.Context, feedback *entities.Feedback) error
	FindByUserAndTransaction(ctx context.Context, userID valueobjects.UserID, transactionID valueobjects.TransactionID) (*entities.Feedback, error)
	FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.Feedback, error)
	Update(ctx context.Context, feedback *entities.Feedback) error
	DeleteByUserAndTransaction(ctx context.Context, userID valueobjects.UserID, transactionID valueobjects.TransactionID) error
}
