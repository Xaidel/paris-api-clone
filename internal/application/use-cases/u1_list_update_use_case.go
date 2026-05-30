package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateU1ListUseCase updates a U1 list entry.
type UpdateU1ListUseCase struct {
	repository         outboundports.U1ListRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewUpdateU1ListUseCase builds an UpdateU1ListUseCase.
func NewUpdateU1ListUseCase(repository outboundports.U1ListRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *UpdateU1ListUseCase {
	return &UpdateU1ListUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute updates an existing U1 list entry.
func (uc *UpdateU1ListUseCase) Execute(ctx context.Context, command inboundports.UpdateU1ListCommand) (outboundports.U1ListResult, error) {
	id, err := valueobjects.U1ListIDFromString(command.ID)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("parsing u1 list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("finding u1 list entry by id: %w", err)
	}

	if entry == nil {
		return outboundports.U1ListResult{}, &NotFoundError{Resource: "u1 list entry", ID: command.ID}
	}

	if err := entry.Update(command.Sector, command.EligibleOperationType, command.ConditionGuidance); err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("updating u1 list entry: %w", err)
	}

	now := uc.now()
	if err := entry.Touch(now); err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("updating u1 list timestamp: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Update(txCtx, entry); err != nil {
			return fmt.Errorf("updating u1 list entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, updateU1ListAdminEventType, map[string]any{
			"action":                  "update",
			"resource":                "u1_list",
			"target_id":               entry.ID().String(),
			"sector":                  entry.Sector(),
			"eligible_operation_type": entry.EligibleOperationType(),
			"condition_guidance":      entry.ConditionGuidance(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("updating u1 list transaction: %w", err)
	}

	return newU1ListResult(entry), nil
}
