package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateSectorCommand requests an update for an existing sector entry.
type UpdateSectorCommand struct {
	ID           string
	Type         string
	Name         string
	Description  string
	ActorUserID  string
	ActorGroupID string
}

// UpdateSectorPort updates an existing sector entry.
type UpdateSectorPort interface {
	Execute(ctx context.Context, command UpdateSectorCommand) (outboundports.SectorResult, error)
}
