package adapters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	createAdminEventQuery = `
INSERT INTO admin_event (id, timestamp, user_id, group_id, event_type, event_data)
VALUES ($1, $2, $3, $4, $5, $6::jsonb)
`
	findAdminEventByIDQuery = `
SELECT id, timestamp, user_id, group_id, event_type, event_data
FROM admin_event
WHERE id = $1
`
	baseListAdminEventsQuery = `
SELECT id, timestamp, user_id, group_id, event_type, event_data
FROM admin_event
`
)

// PostgresAdminEventRepository persists admin audit events in PostgreSQL.
type PostgresAdminEventRepository struct {
	pool pgxQuerier
}

// NewPostgresAdminEventRepository builds a PostgresAdminEventRepository.
func NewPostgresAdminEventRepository(pool pgxQuerier) *PostgresAdminEventRepository {
	return &PostgresAdminEventRepository{pool: pool}
}

// Create inserts a new admin event.
func (r *PostgresAdminEventRepository) Create(ctx context.Context, event *entities.AdminEvent) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createAdminEventQuery, event.ID().String(), event.OccurredAt(), event.UserID(), event.GroupID(), event.EventType(), string(event.EventData())); err != nil {
		return fmt.Errorf("executing create admin event query: %w", err)
	}

	return nil
}

// FindByID returns an admin event by identifier.
func (r *PostgresAdminEventRepository) FindByID(ctx context.Context, id valueobjects.EventID) (*entities.AdminEvent, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	event, err := scanAdminEvent(querier.QueryRow(ctx, findAdminEventByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning admin event by id: %w", err)
	}

	return event, nil
}

// List returns filtered admin events.
func (r *PostgresAdminEventRepository) List(ctx context.Context, filter ports.AuditEventFilter) ([]*entities.AdminEvent, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	query, args := buildListAdminEventsQuery(filter)
	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying admin events: %w", err)
	}
	defer rows.Close()

	events := make([]*entities.AdminEvent, 0)
	for rows.Next() {
		event, scanErr := scanAdminEvent(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed admin event: %w", scanErr)
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating admin events: %w", err)
	}

	return events, nil
}

func scanAdminEvent(row scanner) (*entities.AdminEvent, error) {
	var (
		eventID   string
		timestamp time.Time
		userID    string
		groupID   string
		eventType string
		eventData []byte
	)

	if err := row.Scan(&eventID, &timestamp, &userID, &groupID, &eventType, &eventData); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedEventID, err := valueobjects.EventIDFromString(eventID)
	if err != nil {
		return nil, fmt.Errorf("parsing event id: %w", err)
	}

	event, err := entities.NewAdminEvent(parsedEventID, timestamp, userID, groupID, eventType, eventData)
	if err != nil {
		return nil, fmt.Errorf("reconstituting admin event: %w", err)
	}

	return event, nil
}

func buildListAdminEventsQuery(filter ports.AuditEventFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if strings.TrimSpace(filter.EventOwner) != "" {
		conditions = append(conditions, fmt.Sprintf("'user' = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(filter.EventOwner))
	}

	if strings.TrimSpace(filter.EventType) != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(filter.EventType))
	}

	if strings.TrimSpace(filter.UserID) != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, strings.TrimSpace(filter.UserID))
	}

	if filter.StartedAt != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", len(args)+1))
		args = append(args, *filter.StartedAt)
	}

	if filter.EndedAt != nil {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", len(args)+1))
		args = append(args, *filter.EndedAt)
	}

	query := baseListAdminEventsQuery
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}

	query += "ORDER BY id ASC"
	return query, args
}
