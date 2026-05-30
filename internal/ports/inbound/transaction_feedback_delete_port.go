package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// DeleteTransactionFeedbackCommand requests removal of a feedback on a transaction.
type DeleteTransactionFeedbackCommand struct {
	TransactionID valueobjects.TransactionID
	ActorUserID   string
	ActorGroupID  string
}

// DeleteTransactionFeedbackPort removes a transaction feedback.
type DeleteTransactionFeedbackPort interface {
	Execute(ctx context.Context, command DeleteTransactionFeedbackCommand) error
}
