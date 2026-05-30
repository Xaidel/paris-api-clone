package adapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
)

const (
	createTransactionStep5Query = `
INSERT INTO transaction_step_5_data (
    transaction_id,
    screening_question_1_answer,
    screening_question_1_justification,
    screening_question_2_answer,
    screening_question_2_justification,
    reviewer_notes,
    is_final,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

	findTransactionStep5ByTransactionIDQuery = `
SELECT transaction_id,
       screening_question_1_answer,
       screening_question_1_justification,
       screening_question_2_answer,
       screening_question_2_justification,
       reviewer_notes,
       is_final,
       created_at,
       updated_at
FROM transaction_step_5_data
WHERE transaction_id = $1
`
)

// PostgresTransactionStep5Repository persists transaction step 5 records in PostgreSQL.
type PostgresTransactionStep5Repository struct {
	pool pgxQuerier
}

// NewPostgresTransactionStep5Repository builds a PostgresTransactionStep5Repository.
func NewPostgresTransactionStep5Repository(pool pgxQuerier) *PostgresTransactionStep5Repository {
	return &PostgresTransactionStep5Repository{pool: pool}
}

// Create inserts a new transaction step 5 record.
func (r *PostgresTransactionStep5Repository) Create(ctx context.Context, step5 *entities.TransactionStep5) error {
	querier := txQuerierFromContext(ctx, r.pool)

	var reviewerNotesValue any
	if reviewerNotes := step5.ReviewerNotes().String(); reviewerNotes != nil {
		reviewerNotesValue = *reviewerNotes
	}

	if _, err := querier.Exec(
		ctx,
		createTransactionStep5Query,
		step5.TransactionID().String(),
		step5.ScreeningQuestion1Answer(),
		step5.ScreeningQuestion1Justification().String(),
		step5.ScreeningQuestion2Answer(),
		step5.ScreeningQuestion2Justification().String(),
		reviewerNotesValue,
		step5.IsFinal(),
		step5.CreatedAt(),
		step5.UpdatedAt(),
	); err != nil {
		return fmt.Errorf("executing create transaction step 5 query: %w", err)
	}

	return nil
}

// FindByTransactionID returns one step 5 screening record by transaction identifier.
func (r *PostgresTransactionStep5Repository) FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep5, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	step5, err := scanTransactionStep5(querier.QueryRow(ctx, findTransactionStep5ByTransactionIDQuery, transactionID.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction step 5 by transaction id: %w", err)
	}

	return step5, nil
}

func scanTransactionStep5(row scanner) (*entities.TransactionStep5, error) {
	var (
		transactionIDValue              string
		screeningQuestion1Answer        bool
		screeningQuestion1Justification string
		screeningQuestion2Answer        bool
		screeningQuestion2Justification string
		reviewerNotesValue              sql.NullString
		isFinal                         bool
		createdAt                       time.Time
		updatedAt                       time.Time
	)

	if err := row.Scan(
		&transactionIDValue,
		&screeningQuestion1Answer,
		&screeningQuestion1Justification,
		&screeningQuestion2Answer,
		&screeningQuestion2Justification,
		&reviewerNotesValue,
		&isFinal,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	transactionID, err := valueobjects.TransactionIDFromString(transactionIDValue)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 5 transaction id: %w", err)
	}

	parsedQuestion1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification(screeningQuestion1Justification)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 5 screening question 1 justification: %w", err)
	}

	parsedQuestion2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification(screeningQuestion2Justification)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction step 5 screening question 2 justification: %w", err)
	}

	var reviewerNotes *string
	if reviewerNotesValue.Valid {
		reviewerNotes = &reviewerNotesValue.String
	}

	return entities.ReconstituteTransactionStep5(
		transactionID,
		screeningQuestion1Answer,
		parsedQuestion1Justification,
		screeningQuestion2Answer,
		parsedQuestion2Justification,
		valueobjects.NewTransactionStep5ReviewerNotes(reviewerNotes),
		isFinal,
		createdAt,
		updatedAt,
	), nil
}
