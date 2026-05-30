package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep5Repository persists transaction step 5 screening records.
type TransactionStep5Repository interface {
	Create(ctx context.Context, step5 *entities.TransactionStep5) error
	FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep5, error)
}
