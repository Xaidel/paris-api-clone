package usecases

import (
	"context"
	"fmt"

	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func validateActor(ctx context.Context, directory ports.ActorDirectory, actorUserID string, actorGroupID string) error {
	if directory == nil {
		return nil
	}

	if err := directory.ActorExists(ctx, actorUserID, actorGroupID); err != nil {
		return fmt.Errorf("validating actor ids: %w", err)
	}

	return nil
}
