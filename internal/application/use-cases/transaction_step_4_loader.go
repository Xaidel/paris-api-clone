package usecases

import (
	"context"
	"fmt"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func loadTransactionStep4Details(
	ctx context.Context,
	step4Repository ports.TransactionStep4Repository,
	sectorRepository ports.SectorRepository,
	transactionID valueobjects.TransactionID,
) (*entities.TransactionStep4, *entities.Sector, error) {
	if step4Repository == nil {
		return nil, nil, nil
	}

	step4, err := step4Repository.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, nil, fmt.Errorf("finding transaction step 4 by transaction id: %w", err)
	}
	if step4 == nil {
		return nil, nil, nil
	}

	if sectorRepository == nil {
		return step4, nil, nil
	}

	sector, err := sectorRepository.FindByID(ctx, step4.SectorID())
	if err != nil {
		return nil, nil, fmt.Errorf("finding sector by id: %w", err)
	}

	return step4, sector, nil
}
