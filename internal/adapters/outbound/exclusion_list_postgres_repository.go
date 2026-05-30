package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
)

const (
	createExclusionListEntryQuery = `
INSERT INTO exclusion_list (id, activity_type, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`
	findExclusionListEntryByIDQuery = `
SELECT id, activity_type, created_by, created_at, updated_at
FROM exclusion_list
WHERE id = $1
`
	listExclusionListEntriesQuery = `
SELECT id, activity_type, created_by, created_at, updated_at
FROM exclusion_list
ORDER BY activity_type ASC, id ASC
`
	updateExclusionListEntryQuery = `
UPDATE exclusion_list
SET activity_type = $2, updated_at = $3
WHERE id = $1
`
	deleteExclusionListEntryQuery = `
DELETE FROM exclusion_list
WHERE id = $1
`
)

// PostgresExclusionListRepository persists exclusion list entries in PostgreSQL.
type PostgresExclusionListRepository struct {
	pool pgxQuerier
}

// NewPostgresExclusionListRepository builds a PostgresExclusionListRepository.
func NewPostgresExclusionListRepository(pool pgxQuerier) *PostgresExclusionListRepository {
	return &PostgresExclusionListRepository{pool: pool}
}

// Create inserts a new exclusion list entry.
func (r *PostgresExclusionListRepository) Create(ctx context.Context, entry *entities.ExclusionListEntry, createdByUserID string) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createExclusionListEntryQuery, entry.ID().String(), entry.ActivityType(), createdByUserID, entry.CreatedAt(), entry.UpdatedAt()); err != nil {
		return fmt.Errorf("executing create exclusion list query: %w", err)
	}

	return nil
}

// FindByID returns an exclusion list entry by identifier.
func (r *PostgresExclusionListRepository) FindByID(ctx context.Context, id valueobjects.ExclusionListID) (*entities.ExclusionListEntry, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	entry, err := scanExclusionListEntry(querier.QueryRow(ctx, findExclusionListEntryByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning exclusion list entry by id: %w", err)
	}

	return entry, nil
}

// List returns all exclusion list entries.
func (r *PostgresExclusionListRepository) List(ctx context.Context) ([]*entities.ExclusionListEntry, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listExclusionListEntriesQuery)
	if err != nil {
		return nil, fmt.Errorf("querying exclusion list entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*entities.ExclusionListEntry, 0)
	for rows.Next() {
		entry, scanErr := scanExclusionListEntry(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed exclusion list entry: %w", scanErr)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating exclusion list entries: %w", err)
	}

	return entries, nil
}

// Update updates an existing exclusion list entry.
func (r *PostgresExclusionListRepository) Update(ctx context.Context, entry *entities.ExclusionListEntry) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateExclusionListEntryQuery, entry.ID().String(), entry.ActivityType(), entry.UpdatedAt())
	if err != nil {
		return fmt.Errorf("executing update exclusion list query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing exclusion list entry.
func (r *PostgresExclusionListRepository) DeleteByID(ctx context.Context, id valueobjects.ExclusionListID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteExclusionListEntryQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete exclusion list query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanExclusionListEntry(row scanner) (*entities.ExclusionListEntry, error) {
	var (
		entryID      string
		activityType string
		createdBy    string
		createdAt    time.Time
		updatedAt    time.Time
	)

	if err := row.Scan(&entryID, &activityType, &createdBy, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedID, err := valueobjects.ExclusionListIDFromString(entryID)
	if err != nil {
		return nil, fmt.Errorf("parsing exclusion list id: %w", err)
	}

	parsedCreatedBy, err := valueobjects.UserIDFromString(createdBy)
	if err != nil {
		return nil, fmt.Errorf("parsing exclusion list created by: %w", err)
	}

	return entities.ReconstituteExclusionListEntryWithAudit(parsedID, activityType, parsedCreatedBy, createdAt, updatedAt), nil
}
