package usecases

import (
	"context"
	"fmt"

	"github.com/gyud-adb/paris-api/internal/domain"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func publishDomainEvents(ctx context.Context, publisher outboundports.EventPublisher, events []domain.DomainEvent) error {
	if publisher == nil || len(events) == 0 {
		return nil
	}

	if err := publisher.Publish(ctx, events); err != nil {
		return fmt.Errorf("publishing domain events: %w", err)
	}

	return nil
}
