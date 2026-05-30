package ports

import (
	"context"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// EventPublisher publishes domain events to external systems.
type EventPublisher interface {
	Publish(ctx context.Context, events []domain.DomainEvent) error
}
