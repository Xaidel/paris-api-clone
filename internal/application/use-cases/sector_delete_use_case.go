package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteSectorUseCase deletes a sector entry.
type DeleteSectorUseCase struct {
	repository         outboundports.SectorRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewDeleteSectorUseCase builds a DeleteSectorUseCase.
func NewDeleteSectorUseCase(repository outboundports.SectorRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *DeleteSectorUseCase {
	return &DeleteSectorUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute deletes an existing sector entry.
func (uc *DeleteSectorUseCase) Execute(ctx context.Context, command inboundports.DeleteSectorCommand) (inboundports.DeleteSectorResult, error) {
	id, err := valueobjects.SectorIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteSectorResult{}, fmt.Errorf("parsing sector id: %w", err)
	}

	sector, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return inboundports.DeleteSectorResult{}, fmt.Errorf("finding sector entry by id: %w", err)
	}

	if sector == nil {
		return inboundports.DeleteSectorResult{}, &NotFoundError{Resource: "sector entry", ID: command.ID}
	}

	now := uc.now()
	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.DeleteByID(txCtx, id); err != nil {
			return fmt.Errorf("deleting sector entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, deleteSectorAdminEventType, map[string]any{
			"action":      "delete",
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
		return inboundports.DeleteSectorResult{}, fmt.Errorf("deleting sector transaction: %w", err)
	}

	return inboundports.DeleteSectorResult{ID: command.ID}, nil
}
