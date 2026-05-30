package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetUserQuery identifies the user to load together with the acting admin
// context used for audit recording.
type GetUserQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetUserPort gets a user by identifier.
type GetUserPort interface {
	Execute(ctx context.Context, query GetUserQuery) (outboundports.UserResult, error)
}
