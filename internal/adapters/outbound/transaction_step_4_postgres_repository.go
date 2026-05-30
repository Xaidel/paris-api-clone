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
	createTransactionStep4Query = `
INSERT INTO transaction_step_4 (
    transaction_id,
    sector_id,
    additional_context,
    is_high_emitting,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
`

	findTransactionStep4ByTransactionIDQuery = `
SELECT transaction_id, sector_id, additional_context, is_high_emitting, created_at, updated_at
FROM transaction_step_4
WHERE transaction_id = $1
`
)

// PostgresTransactionStep4Repository persists transaction step 4 records in PostgreSQL.
type PostgresTransactionStep4Repository struct {
	pool pgxQuerier
}

// NewPostgresTransactionStep4Repository builds a PostgresTransactionStep4Repository.
func NewPostgresTransactionStep4Repository(pool pgxQuerier) *PostgresTransactionStep4Repository {
	return &PostgresTransactionStep4Repository{pool: pool}
}

// Create inserts a new transaction step 4 record.
func (r *PostgresTransactionStep4Repository) Create(ctx context.Context, step4 *entities.TransactionStep4) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(
		ctx,
		createTransactionStep4Query,
		step4.TransactionID().String(),
		step4.SectorID().String(),
		step4.AdditionalContext().String(),
		step4.IsHighEmitting(),
		step4.CreatedAt(),
		step4.UpdatedAt(),
	); err != nil {
		return fmt.Errorf("executing create transaction step 4 query: %w", err)
	}

	return nil
}

// FindByTransactionID returns one step 4 review by transaction identifier.
func (r *PostgresTransactionStep4Repository) FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep4, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	step4, err := scanTransactionStep4(querier.QueryRow(ctx, findTransactionStep4ByTransactionIDQuery, transactionID.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction step 4 by transaction id: %w", err)
	}

	return step4, nil
}

func scanTransactionStep4(row scanner) (*entities.TransactionStep4, error) {
	var (
		transactionIDValue     string
		sectorIDValue          string
		additionalContextValue string
		isHighEmitting         bool
		createdAt              time.Time
		updatedAt              time.Time
	)

	if err := row.Scan(&transactionIDValue, &sectorIDValue, &additionalContextValue, &isHighEmitting, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	transactionID, err := valueobjects.TransactionIDFromString(transactionIDValue)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 4 transaction id: %w", err)
	}

	sectorID, err := valueobjects.SectorIDFromString(sectorIDValue)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 4 sector id: %w", err)
	}

	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext(additionalContextValue)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 4 additional context: %w", err)
	}

	return entities.ReconstituteTransactionStep4(transactionID, sectorID, additionalContext, isHighEmitting, createdAt, updatedAt), nil
}
