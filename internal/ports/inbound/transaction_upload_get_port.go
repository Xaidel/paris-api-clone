package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetTransactionUploadQuery requests a single upload and its accepted transactions.
type GetTransactionUploadQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetTransactionUploadPort gets a single upload and its accepted transactions.
type GetTransactionUploadPort interface {
	Execute(ctx context.Context, query GetTransactionUploadQuery) (outboundports.TransactionUploadDetailsResult, error)
}
