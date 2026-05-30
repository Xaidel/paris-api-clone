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
	createFeedbackQuery = `
INSERT INTO transaction_feedback (id, user_id, transaction_id, kind, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
`
	findFeedbackByUserAndTransactionQuery = `
SELECT id, user_id, transaction_id, kind, created_at, updated_at
FROM transaction_feedback
WHERE user_id = $1 AND transaction_id = $2
`
	updateFeedbackQuery = `
UPDATE transaction_feedback
SET kind = $3, updated_at = $4
WHERE user_id = $1 AND transaction_id = $2
`
	deleteFeedbackByUserAndTransactionQuery = `
DELETE FROM transaction_feedback
WHERE user_id = $1 AND transaction_id = $2
`
	findFeedbackByTransactionIDQuery = `
SELECT id, user_id, transaction_id, kind, created_at, updated_at
FROM transaction_feedback
WHERE transaction_id = $1
LIMIT 1
`
)

// PostgresFeedbackRepository persists transaction feedback in PostgreSQL.
type PostgresFeedbackRepository struct {
	pool pgxQuerier
}

// NewPostgresFeedbackRepository builds a PostgresFeedbackRepository.
func NewPostgresFeedbackRepository(pool pgxQuerier) *PostgresFeedbackRepository {
	return &PostgresFeedbackRepository{pool: pool}
}

// Create inserts a new feedback record.
func (r *PostgresFeedbackRepository) Create(ctx context.Context, feedback *entities.Feedback) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createFeedbackQuery,
		feedback.ID().String(),
		feedback.UserID(),
		feedback.TransactionID(),
		feedback.Kind(),
		feedback.CreatedAt(),
		feedback.UpdatedAt(),
	); err != nil {
		return fmt.Errorf("executing create feedback query: %w", err)
	}

	return nil
}

// FindByUserAndTransaction returns the feedback a user gave on a transaction.
func (r *PostgresFeedbackRepository) FindByUserAndTransaction(ctx context.Context, userID valueobjects.UserID, transactionID valueobjects.TransactionID) (*entities.Feedback, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	feedback, err := scanFeedback(querier.QueryRow(ctx, findFeedbackByUserAndTransactionQuery, userID.String(), transactionID.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning feedback by user and transaction: %w", err)
	}

	return feedback, nil
}

// FindByTransactionID returns a feedback for a transaction (any user).
func (r *PostgresFeedbackRepository) FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.Feedback, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	feedback, err := scanFeedback(querier.QueryRow(ctx, findFeedbackByTransactionIDQuery, transactionID.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning feedback by transaction id: %w", err)
	}

	return feedback, nil
}

// Update updates the kind and timestamp of an existing feedback record.
func (r *PostgresFeedbackRepository) Update(ctx context.Context, feedback *entities.Feedback) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateFeedbackQuery,
		feedback.UserID(),
		feedback.TransactionID(),
		feedback.Kind(),
		feedback.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("executing update feedback query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByUserAndTransaction removes a user's feedback on a transaction.
func (r *PostgresFeedbackRepository) DeleteByUserAndTransaction(ctx context.Context, userID valueobjects.UserID, transactionID valueobjects.TransactionID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteFeedbackByUserAndTransactionQuery, userID.String(), transactionID.String())
	if err != nil {
		return fmt.Errorf("executing delete feedback query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanFeedback(row scanner) (*entities.Feedback, error) {
	var (
		feedbackID    string
		userID        string
		transactionID string
		kind          string
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := row.Scan(&feedbackID, &userID, &transactionID, &kind, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedID, err := valueobjects.FeedbackIDFromString(feedbackID)
	if err != nil {
		return nil, fmt.Errorf("parsing feedback id: %w", err)
	}

	parsedUserID, err := valueobjects.UserIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing feedback user id: %w", err)
	}

	parsedTransactionID, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		return nil, fmt.Errorf("parsing feedback transaction id: %w", err)
	}

	parsedKind, err := valueobjects.FeedbackKindFromString(kind)
	if err != nil {
		return nil, fmt.Errorf("parsing feedback kind: %w", err)
	}

	return entities.ReconstituteFeedback(parsedID, parsedUserID, parsedTransactionID, parsedKind, createdAt, updatedAt), nil
}
