package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateTransactionStep5Command requests step 5 screening for a transaction.
type CreateTransactionStep5Command struct {
	TransactionID                   valueobjects.TransactionID
	ScreeningQuestion1Answer        *bool
	ScreeningQuestion1Justification valueobjects.TransactionStep5ScreeningQuestion1Justification
	ScreeningQuestion2Answer        *bool
	ScreeningQuestion2Justification valueobjects.TransactionStep5ScreeningQuestion2Justification
	ReviewerNotes                   valueobjects.TransactionStep5ReviewerNotes
	IsFinal                         *bool
	ActorUserID                     string
	ActorGroupID                    string
}

// CreateTransactionStep5Port creates a step 5 screening result for a transaction.
type CreateTransactionStep5Port interface {
	Execute(ctx context.Context, command CreateTransactionStep5Command) (outboundports.TransactionStep5Result, error)
}
