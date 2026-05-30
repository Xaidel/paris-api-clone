package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetExclusionListQuery requests a single exclusion list entry by identifier.
type GetExclusionListQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetExclusionListPort gets a single exclusion list entry.
type GetExclusionListPort interface {
	Execute(ctx context.Context, query GetExclusionListQuery) (outboundports.ExclusionListResult, error)
}
