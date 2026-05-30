package adapters

import (
	"context"
	"fmt"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
)

const createAdminEventOutboxQuery = `
INSERT INTO admin_event_outbox (event_id, event_type, payload, created_at)
VALUES ($1, $2, $3::jsonb, $4)
`

// PostgresAdminEventOutboxRepository persists admin event outbox records in PostgreSQL.
type PostgresAdminEventOutboxRepository struct {
	pool pgxQuerier
}

// NewPostgresAdminEventOutboxRepository builds a PostgresAdminEventOutboxRepository.
func NewPostgresAdminEventOutboxRepository(pool pgxQuerier) *PostgresAdminEventOutboxRepository {
	return &PostgresAdminEventOutboxRepository{pool: pool}
}

// Create inserts a new admin event outbox record.
func (r *PostgresAdminEventOutboxRepository) Create(ctx context.Context, event *entities.AdminEvent) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createAdminEventOutboxQuery, event.ID().String(), event.EventType(), string(event.EventData()), event.OccurredAt()); err != nil {
		return fmt.Errorf("executing create admin event outbox query: %w", err)
	}

	return nil
}
