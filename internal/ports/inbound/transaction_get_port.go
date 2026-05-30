package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetTransactionQuery requests a single transaction by identifier.
type GetTransactionQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetTransactionPort gets a single transaction.
type GetTransactionPort interface {
	Execute(ctx context.Context, query GetTransactionQuery) (outboundports.TransactionResult, error)
}
