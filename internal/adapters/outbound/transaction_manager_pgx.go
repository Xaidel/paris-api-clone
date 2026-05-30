package adapters

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type txBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type transactionContextKey string

const transactionKey transactionContextKey = "pgx-transaction"

// PgxTransactionManager executes operations inside a pgx transaction.
type PgxTransactionManager struct {
	beginner txBeginner
}

// NewPgxTransactionManager builds a PgxTransactionManager.
func NewPgxTransactionManager(beginner txBeginner) *PgxTransactionManager {
	return &PgxTransactionManager{beginner: beginner}
}

// WithinTransaction executes an operation within a database transaction.
func (m *PgxTransactionManager) WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error {
	tx, err := m.beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	committed := false
	defer func() {
		// Roll back any uncommitted transaction so panics and early returns leave no
		// open transaction behind.
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	// Store the pgx transaction in the context so repositories can opt into the
	// active transaction without changing every repository method signature.
	txCtx := context.WithValue(ctx, transactionKey, tx)
	if err := operation(txCtx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !isTransactionClosedError(rollbackErr) {
			return fmt.Errorf("rolling back transaction: %w", rollbackErr)
		}

		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	committed = true
	return nil
}

func txQuerierFromContext(ctx context.Context, fallback pgxQuerier) pgxQuerier {
	// Repositories use the transaction when present and otherwise fall back to the
	// pool, which keeps the same query path usable inside and outside transactions.
	tx, ok := ctx.Value(transactionKey).(pgx.Tx)
	if ok && tx != nil {
		return tx
	}

	return fallback
}

func isTransactionClosedError(err error) bool {
	// pgx reports rollback-after-commit as ErrTxClosed; treat that as a benign
	// cleanup condition rather than masking the original application error.
	return errors.Is(err, pgx.ErrTxClosed)
}
