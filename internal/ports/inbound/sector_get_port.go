package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetSectorQuery requests a single sector by identifier.
type GetSectorQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetSectorPort gets a single sector.
type GetSectorPort interface {
	Execute(ctx context.Context, query GetSectorQuery) (outboundports.SectorResult, error)
}
