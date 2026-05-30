package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListExclusionListQuery requests all exclusion list entries.
type ListExclusionListQuery struct {
	ActorUserID  string
	ActorGroupID string
}

// ListExclusionListPort lists all exclusion list entries.
type ListExclusionListPort interface {
	Execute(ctx context.Context, query ListExclusionListQuery) (outboundports.ListExclusionListResult, error)
}
