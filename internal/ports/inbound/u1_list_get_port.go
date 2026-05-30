package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetU1ListQuery requests a single U1 list entry by identifier.
type GetU1ListQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetU1ListPort gets a single U1 list entry.
type GetU1ListPort interface {
	Execute(ctx context.Context, query GetU1ListQuery) (outboundports.U1ListResult, error)
}
