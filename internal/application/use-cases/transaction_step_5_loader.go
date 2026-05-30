package usecases

import (
	"context"
	"fmt"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func loadTransactionStep5Details(
	ctx context.Context,
	step5Repository ports.TransactionStep5Repository,
	transactionID valueobjects.TransactionID,
) (*entities.TransactionStep5, error) {
	if step5Repository == nil {
		return nil, nil
	}

	step5, err := step5Repository.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("finding transaction step 5 by transaction id: %w", err)
	}

	return step5, nil
}
