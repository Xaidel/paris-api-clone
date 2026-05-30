package ports

import "context"

// TransactionManager executes work inside a transaction boundary.
type TransactionManager interface {
	// WithinTransaction runs operation with a transaction-scoped context and
	// commits or rolls back based on the returned error.
	WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error
}
