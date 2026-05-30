package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateExclusionListCommand requests the creation of a new exclusion list entry.
type CreateExclusionListCommand struct {
	ActivityType string
	ActorUserID  string
	ActorGroupID string
}

// CreateExclusionListPort creates a new exclusion list entry.
type CreateExclusionListPort interface {
	Execute(ctx context.Context, command CreateExclusionListCommand) (outboundports.ExclusionListResult, error)
}
