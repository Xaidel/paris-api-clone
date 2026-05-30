package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetGroupQuery requests a group by identifier.
type GetGroupQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetGroupPort gets a group by identifier.
type GetGroupPort interface {
	Execute(ctx context.Context, query GetGroupQuery) (outboundports.GroupResult, error)
}
