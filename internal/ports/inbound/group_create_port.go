package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateGroupCommand requests the creation of a new group entry.
type CreateGroupCommand struct {
	Name         string
	ActorUserID  string
	ActorGroupID string
}

// CreateGroupPort creates a new group entry.
type CreateGroupPort interface {
	Execute(ctx context.Context, command CreateGroupCommand) (outboundports.GroupResult, error)
}
