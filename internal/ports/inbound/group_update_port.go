package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateGroupCommand requests an update for an existing group.
type UpdateGroupCommand struct {
	ID           string
	Name         string
	ActorUserID  string
	ActorGroupID string
}

// UpdateGroupPort updates an existing group.
type UpdateGroupPort interface {
	Execute(ctx context.Context, command UpdateGroupCommand) (outboundports.GroupResult, error)
}
