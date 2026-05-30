package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpsertTransactionFeedbackCommand requests a thumbs up/down on a transaction.
type UpsertTransactionFeedbackCommand struct {
	TransactionID valueobjects.TransactionID
	Kind          valueobjects.FeedbackKind
	ActorUserID   string
	ActorGroupID  string
}

// UpsertTransactionFeedbackPort creates or updates a transaction feedback.
type UpsertTransactionFeedbackPort interface {
	Execute(ctx context.Context, command UpsertTransactionFeedbackCommand) (outboundports.FeedbackResult, error)
}
