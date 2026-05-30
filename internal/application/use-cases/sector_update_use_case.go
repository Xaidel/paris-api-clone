package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateSectorUseCase updates a sector entry.
type UpdateSectorUseCase struct {
	repository         outboundports.SectorRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewUpdateSectorUseCase builds an UpdateSectorUseCase.
func NewUpdateSectorUseCase(repository outboundports.SectorRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *UpdateSectorUseCase {
	return &UpdateSectorUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute updates an existing sector entry.
func (uc *UpdateSectorUseCase) Execute(ctx context.Context, command inboundports.UpdateSectorCommand) (outboundports.SectorResult, error) {
	id, err := valueobjects.SectorIDFromString(command.ID)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("parsing sector id: %w", err)
	}

	sector, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("finding sector entry by id: %w", err)
	}

	if sector == nil {
		return outboundports.SectorResult{}, &NotFoundError{Resource: "sector entry", ID: command.ID}
	}

	if err := sector.Update(command.Type, command.Name, command.Description); err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("updating sector entry: %w", err)
	}

	now := uc.now()
	if err := sector.Touch(now); err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("updating sector timestamp: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Update(txCtx, sector); err != nil {
			return fmt.Errorf("updating sector entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, updateSectorAdminEventType, map[string]any{
			"action":      "update",
			"resource":    "sector",
			"target_id":   sector.ID().String(),
			"type":        sector.Type(),
			"name":        sector.Name(),
			"description": sector.Description(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("updating sector transaction: %w", err)
	}

	return newSectorResult(sector), nil
}
