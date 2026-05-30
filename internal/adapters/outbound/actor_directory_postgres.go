package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/jackc/pgx/v5"
)

const groupExistsQuery = `
SELECT 1
FROM user_group
WHERE id = $1
LIMIT 1
`

const actorExistsQuery = `
SELECT 1
FROM user_profile up
WHERE up.user_id = $1
  AND up.group_id = $2
LIMIT 1
`

// PostgresActorDirectory validates actor identifiers against PostgreSQL records.
type PostgresActorDirectory struct {
	pool pgxQuerier
}

// NewPostgresActorDirectory builds a PostgresActorDirectory.
func NewPostgresActorDirectory(pool pgxQuerier) *PostgresActorDirectory {
	return &PostgresActorDirectory{pool: pool}
}

// ActorExists validates that the given actor user and group pair exists.
func (d *PostgresActorDirectory) ActorExists(ctx context.Context, userID string, groupID string) error {
	querier := txQuerierFromContext(ctx, d.pool)
	var exists int
	if err := querier.QueryRow(ctx, groupExistsQuery, groupID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUnknownActorGroupID
		}

		return fmt.Errorf("querying group directory: %w", err)
	}

	if err := querier.QueryRow(ctx, actorExistsQuery, userID, groupID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUnknownActorUserID
		}

		return fmt.Errorf("querying actor directory: %w", err)
	}

	return nil
}
