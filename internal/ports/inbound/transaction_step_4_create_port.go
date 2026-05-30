package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateTransactionStep4Command requests creation of an aggregated step 4 review.
type CreateTransactionStep4Command struct {
	TransactionID     valueobjects.TransactionID
	SectorID          valueobjects.SectorID
	AdditionalContext valueobjects.TransactionStep4AdditionalContext
	IsHighEmitting    *bool
	ActorUserID       string
	ActorGroupID      string
}

// CreateTransactionStep4Port creates an aggregated transaction step 4 review.
type CreateTransactionStep4Port interface {
	Execute(ctx context.Context, command CreateTransactionStep4Command) (outboundports.TransactionStep4Result, error)
}
