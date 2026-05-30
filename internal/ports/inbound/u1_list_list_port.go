package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListU1ListQuery requests all U1 list entries.
type ListU1ListQuery struct {
	Sector       string
	ActorUserID  string
	ActorGroupID string
}

// ListU1ListPort lists all U1 list entries.
type ListU1ListPort interface {
	Execute(ctx context.Context, query ListU1ListQuery) (outboundports.ListU1ListResult, error)
}
