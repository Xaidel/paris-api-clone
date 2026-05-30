package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetTransactionFeedbackQuery requests the feedback a user gave on a transaction.
type GetTransactionFeedbackQuery struct {
	TransactionID valueobjects.TransactionID
	ActorUserID   string
	ActorGroupID  string
}

// GetTransactionFeedbackPort retrieves the current feedback for a transaction/user pair.
type GetTransactionFeedbackPort interface {
	Execute(ctx context.Context, query GetTransactionFeedbackQuery) (*outboundports.FeedbackResult, error)
}
