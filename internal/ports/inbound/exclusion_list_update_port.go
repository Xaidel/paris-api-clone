package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateExclusionListCommand requests an update for an existing exclusion list entry.
type UpdateExclusionListCommand struct {
	ID           string
	ActivityType string
	ActorUserID  string
	ActorGroupID string
}

// UpdateExclusionListPort updates an existing exclusion list entry.
type UpdateExclusionListPort interface {
	Execute(ctx context.Context, command UpdateExclusionListCommand) (outboundports.ExclusionListResult, error)
}
