package ports

import (
	"context"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// AuditEventFilter describes supported audit event query filters.
type AuditEventFilter struct {
	EventOwner string
	EventType  string
	SessionID  string
	UserID     string
	StartedAt  *time.Time
	EndedAt    *time.Time
}

// AdminEventRepository persists and queries admin audit events.
type AdminEventRepository interface {
	Create(ctx context.Context, event *entities.AdminEvent) error
	FindByID(ctx context.Context, id valueobjects.EventID) (*entities.AdminEvent, error)
	List(ctx context.Context, filter AuditEventFilter) ([]*entities.AdminEvent, error)
}

// AdminEventOutboxRepository persists admin event outbox records.
type AdminEventOutboxRepository interface {
	Create(ctx context.Context, event *entities.AdminEvent) error
}
