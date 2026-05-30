package ports

import "context"

// ActorDirectory validates standardized actor identifiers.
type ActorDirectory interface {
	ActorExists(ctx context.Context, userID string, groupID string) error
}
