package adapters

import (
	"context"
	"errors"
	"fmt"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
)

const (
	createGroupQuery = `
INSERT INTO user_group (id, name)
VALUES ($1, $2)
`
	findGroupByIDQuery = `
SELECT id, name
FROM user_group
WHERE id = $1
`
	listGroupsQuery = `
SELECT id, name
FROM user_group
ORDER BY name ASC, id ASC
`
	updateGroupQuery = `
UPDATE user_group
SET name = $2
WHERE id = $1
`
	deleteGroupQuery = `
DELETE FROM user_group
WHERE id = $1
`
)

// PostgresGroupRepository persists groups in PostgreSQL.
type PostgresGroupRepository struct {
	pool pgxQuerier
}

// NewPostgresGroupRepository builds a PostgresGroupRepository.
func NewPostgresGroupRepository(pool pgxQuerier) *PostgresGroupRepository {
	return &PostgresGroupRepository{pool: pool}
}

// Create inserts a new group.
func (r *PostgresGroupRepository) Create(ctx context.Context, group *entities.Group) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createGroupQuery, group.ID().String(), group.Name()); err != nil {
		return fmt.Errorf("executing create group query: %w", err)
	}

	return nil
}

// FindByID returns a group by identifier.
func (r *PostgresGroupRepository) FindByID(ctx context.Context, id valueobjects.GroupID) (*entities.Group, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	group, err := scanGroup(querier.QueryRow(ctx, findGroupByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning group by id: %w", err)
	}

	return group, nil
}

// List returns all groups.
func (r *PostgresGroupRepository) List(ctx context.Context) ([]*entities.Group, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listGroupsQuery)
	if err != nil {
		return nil, fmt.Errorf("querying groups: %w", err)
	}
	defer rows.Close()

	groups := make([]*entities.Group, 0)
	for rows.Next() {
		group, scanErr := scanGroup(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed group: %w", scanErr)
		}

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating listed groups: %w", err)
	}

	return groups, nil
}

// Update updates an existing group.
func (r *PostgresGroupRepository) Update(ctx context.Context, group *entities.Group) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateGroupQuery, group.ID().String(), group.Name())
	if err != nil {
		return fmt.Errorf("executing update group query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing group.
func (r *PostgresGroupRepository) DeleteByID(ctx context.Context, id valueobjects.GroupID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteGroupQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete group query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanGroup(row scanner) (*entities.Group, error) {
	var (
		groupID string
		name    string
	)

	if err := row.Scan(&groupID, &name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedGroupID, err := valueobjects.GroupIDFromString(groupID)
	if err != nil {
		return nil, fmt.Errorf("parsing group id: %w", err)
	}

	return entities.ReconstituteGroup(parsedGroupID, name), nil
}
