package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListSectorsQuery requests all sector entries.
type ListSectorsQuery struct {
	ActorUserID  string
	ActorGroupID string
}

// ListSectorsPort lists all sector entries.
type ListSectorsPort interface {
	Execute(ctx context.Context, query ListSectorsQuery) (outboundports.ListSectorsResult, error)
}
