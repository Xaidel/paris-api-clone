package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateU1ListCommand requests an update for an existing U1 list entry.
type UpdateU1ListCommand struct {
	ID                    string
	Sector                string
	EligibleOperationType string
	ConditionGuidance     string
	ActorUserID           string
	ActorGroupID          string
}

// UpdateU1ListPort updates an existing U1 list entry.
type UpdateU1ListPort interface {
	Execute(ctx context.Context, command UpdateU1ListCommand) (outboundports.U1ListResult, error)
}
