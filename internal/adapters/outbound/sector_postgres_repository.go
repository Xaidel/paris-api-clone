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
	createSectorEntryQuery = `
INSERT INTO sector (id, type, name, description, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	findSectorEntryByIDQuery = `
SELECT id, type, name, description, created_by, created_at, updated_at
FROM sector
WHERE id = $1
`
	listSectorEntriesQuery = `
SELECT id, type, name, description, created_by, created_at, updated_at
FROM sector
ORDER BY type ASC, name ASC, id ASC
`
	updateSectorEntryQuery = `
UPDATE sector
SET type = $2, name = $3, description = $4, updated_at = $5
WHERE id = $1
`
	deleteSectorEntryQuery = `
DELETE FROM sector
WHERE id = $1
`
)

// PostgresSectorRepository persists sector entries in PostgreSQL.
type PostgresSectorRepository struct {
	pool pgxQuerier
}

// NewPostgresSectorRepository builds a PostgresSectorRepository.
func NewPostgresSectorRepository(pool pgxQuerier) *PostgresSectorRepository {
	return &PostgresSectorRepository{pool: pool}
}

// Create inserts a new sector entry.
func (r *PostgresSectorRepository) Create(ctx context.Context, sector *entities.Sector, createdByUserID string) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createSectorEntryQuery, sector.ID().String(), sector.Type(), sector.Name(), sector.Description(), createdByUserID, sector.CreatedAt(), sector.UpdatedAt()); err != nil {
		return fmt.Errorf("executing create sector query: %w", err)
	}

	return nil
}

// FindByID returns a sector entry by identifier.
func (r *PostgresSectorRepository) FindByID(ctx context.Context, id valueobjects.SectorID) (*entities.Sector, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	sector, err := scanSectorEntry(querier.QueryRow(ctx, findSectorEntryByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning sector entry by id: %w", err)
	}

	return sector, nil
}

// List returns all sector entries.
func (r *PostgresSectorRepository) List(ctx context.Context) ([]*entities.Sector, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listSectorEntriesQuery)
	if err != nil {
		return nil, fmt.Errorf("querying sector entries: %w", err)
	}
	defer rows.Close()

	sectors := make([]*entities.Sector, 0)
	for rows.Next() {
		sector, scanErr := scanSectorEntry(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed sector entry: %w", scanErr)
		}

		sectors = append(sectors, sector)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating sector entries: %w", err)
	}

	return sectors, nil
}

// Update updates an existing sector entry.
func (r *PostgresSectorRepository) Update(ctx context.Context, sector *entities.Sector) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateSectorEntryQuery, sector.ID().String(), sector.Type(), sector.Name(), sector.Description(), sector.UpdatedAt())
	if err != nil {
		return fmt.Errorf("executing update sector query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing sector entry.
func (r *PostgresSectorRepository) DeleteByID(ctx context.Context, id valueobjects.SectorID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteSectorEntryQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete sector query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanSectorEntry(row scanner) (*entities.Sector, error) {
	var (
		sectorID        string
		sectorTypeValue string
		name            string
		description     string
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	if err := row.Scan(&sectorID, &sectorTypeValue, &name, &description, &createdBy, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedID, err := valueobjects.SectorIDFromString(sectorID)
	if err != nil {
		return nil, fmt.Errorf("parsing sector id: %w", err)
	}

	parsedType, err := valueobjects.SectorTypeFromString(sectorTypeValue)
	if err != nil {
		return nil, fmt.Errorf("parsing sector type: %w", err)
	}

	parsedCreatedBy, err := valueobjects.UserIDFromString(createdBy)
	if err != nil {
		return nil, fmt.Errorf("parsing sector created by: %w", err)
	}

	return entities.ReconstituteSectorWithAudit(parsedID, parsedType, name, description, parsedCreatedBy, createdAt, updatedAt), nil
}
