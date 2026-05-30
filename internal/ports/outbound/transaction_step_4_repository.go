package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep4Repository persists transaction step 4 review records.
type TransactionStep4Repository interface {
	Create(ctx context.Context, step4 *entities.TransactionStep4) error
	FindByTransactionID(ctx context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep4, error)
}
