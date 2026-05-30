package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	createU1ListEntryQuery = `
INSERT INTO u1_list (id, sector, eligible_operation_type, condition_guidance, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	findU1ListEntryByIDQuery = `
SELECT id, sector, eligible_operation_type, condition_guidance, created_by, created_at, updated_at
FROM u1_list
WHERE id = $1
`
	listU1ListEntriesQuery = `
SELECT id, sector, eligible_operation_type, condition_guidance, created_by, created_at, updated_at
FROM u1_list
ORDER BY sector ASC, eligible_operation_type ASC, id ASC
`
	listU1ListEntriesBySectorQuery = `
SELECT id, sector, eligible_operation_type, condition_guidance, created_by, created_at, updated_at
FROM u1_list
WHERE LOWER(sector) = LOWER($1)
ORDER BY sector ASC, eligible_operation_type ASC, id ASC
`
	updateU1ListEntryQuery = `
UPDATE u1_list
SET sector = $2, eligible_operation_type = $3, condition_guidance = $4, updated_at = $5
WHERE id = $1
`
	deleteU1ListEntryQuery = `
DELETE FROM u1_list
WHERE id = $1
`
)

// PostgresU1ListRepository persists U1 list entries in PostgreSQL.
type PostgresU1ListRepository struct {
	pool pgxQuerier
}

// NewPostgresU1ListRepository builds a PostgresU1ListRepository.
func NewPostgresU1ListRepository(pool pgxQuerier) *PostgresU1ListRepository {
	return &PostgresU1ListRepository{pool: pool}
}

// Create inserts a new U1 list entry.
func (r *PostgresU1ListRepository) Create(ctx context.Context, entry *entities.U1ListEntry, createdByUserID string) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createU1ListEntryQuery, entry.ID().String(), entry.Sector(), entry.EligibleOperationType(), entry.ConditionGuidance(), createdByUserID, entry.CreatedAt(), entry.UpdatedAt()); err != nil {
		return fmt.Errorf("executing create u1 list query: %w", err)
	}

	return nil
}

// FindByID returns a U1 list entry by identifier.
func (r *PostgresU1ListRepository) FindByID(ctx context.Context, id valueobjects.U1ListID) (*entities.U1ListEntry, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	entry, err := scanU1ListEntry(querier.QueryRow(ctx, findU1ListEntryByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning u1 list entry by id: %w", err)
	}

	return entry, nil
}

// List returns U1 list entries matching the supplied filter.
func (r *PostgresU1ListRepository) List(ctx context.Context, filter ports.U1ListFilter) ([]*entities.U1ListEntry, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := r.queryListRows(ctx, querier, filter)
	if err != nil {
		return nil, fmt.Errorf("querying u1 list entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*entities.U1ListEntry, 0)
	for rows.Next() {
		entry, scanErr := scanU1ListEntry(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed u1 list entry: %w", scanErr)
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating u1 list entries: %w", err)
	}

	return entries, nil
}

// Update updates an existing U1 list entry.
func (r *PostgresU1ListRepository) Update(ctx context.Context, entry *entities.U1ListEntry) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateU1ListEntryQuery, entry.ID().String(), entry.Sector(), entry.EligibleOperationType(), entry.ConditionGuidance(), entry.UpdatedAt())
	if err != nil {
		return fmt.Errorf("executing update u1 list query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing U1 list entry.
func (r *PostgresU1ListRepository) DeleteByID(ctx context.Context, id valueobjects.U1ListID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteU1ListEntryQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete u1 list query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *PostgresU1ListRepository) queryListRows(ctx context.Context, querier pgxQuerier, filter ports.U1ListFilter) (pgx.Rows, error) {
	if filter.Sector != "" {
		return querier.Query(ctx, listU1ListEntriesBySectorQuery, filter.Sector)
	}

	return querier.Query(ctx, listU1ListEntriesQuery)
}

func scanU1ListEntry(row scanner) (*entities.U1ListEntry, error) {
	var (
		entryID               string
		sector                string
		eligibleOperationType string
		conditionGuidance     string
		createdBy             string
		createdAt             time.Time
		updatedAt             time.Time
	)

	if err := row.Scan(&entryID, &sector, &eligibleOperationType, &conditionGuidance, &createdBy, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedID, err := valueobjects.U1ListIDFromString(entryID)
	if err != nil {
		return nil, fmt.Errorf("parsing u1 list id: %w", err)
	}

	parsedCreatedBy, err := valueobjects.UserIDFromString(createdBy)
	if err != nil {
		return nil, fmt.Errorf("parsing u1 list created by: %w", err)
	}

	return entities.ReconstituteU1ListEntryWithAudit(parsedID, sector, eligibleOperationType, conditionGuidance, parsedCreatedBy, createdAt, updatedAt), nil
}
