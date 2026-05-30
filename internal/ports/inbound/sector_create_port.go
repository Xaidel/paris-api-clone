package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateSectorCommand requests the creation of a new sector entry.
type CreateSectorCommand struct {
	Type         string
	Name         string
	Description  string
	ActorUserID  string
	ActorGroupID string
}

// CreateSectorPort creates a new sector entry.
type CreateSectorPort interface {
	Execute(ctx context.Context, command CreateSectorCommand) (outboundports.SectorResult, error)
}
