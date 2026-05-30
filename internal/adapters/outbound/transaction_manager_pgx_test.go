package adapters

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
)

// TestPgxTransactionManagerWithinTransaction verifies the pgx transaction manager within transaction behavior and the expected outcome asserted below.
func TestPgxTransactionManagerWithinTransaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		configure   func(mock pgxmock.PgxPoolIface)
		operation   func(context.Context) error
		assertError func(t *testing.T, err error)
	}{
		{
			name: "commits transaction",
			configure: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta("SELECT 1")).WillReturnResult(pgxmock.NewResult("SELECT", 1))
				mock.ExpectCommit()
			},
			operation: func(ctx context.Context) error {
				querier := txQuerierFromContext(ctx, nil)
				if querier == nil {
					return errors.New("expected transaction querier")
				}
				_, err := querier.Exec(ctx, "SELECT 1")
				return err
			},
		},
		{
			name: "rolls back transaction on error",
			configure: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			operation: func(context.Context) error {
				return errors.New("boom")
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("pgxmock.NewPool() error = %v", err)
			}
			defer mock.Close()

			tt.configure(mock)
			manager := NewPgxTransactionManager(mock)
			err = manager.WithinTransaction(context.Background(), tt.operation)
			if tt.assertError != nil {
				tt.assertError(t, err)
			} else if err != nil {
				t.Fatalf("WithinTransaction() error = %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("ExpectationsWereMet() error = %v", err)
			}
		})
	}
}
