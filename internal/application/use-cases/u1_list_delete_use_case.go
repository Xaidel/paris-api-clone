package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteU1ListUseCase deletes a U1 list entry.
type DeleteU1ListUseCase struct {
	repository         outboundports.U1ListRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewDeleteU1ListUseCase builds a DeleteU1ListUseCase.
func NewDeleteU1ListUseCase(repository outboundports.U1ListRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *DeleteU1ListUseCase {
	return &DeleteU1ListUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute deletes an existing U1 list entry.
func (uc *DeleteU1ListUseCase) Execute(ctx context.Context, command inboundports.DeleteU1ListCommand) (inboundports.DeleteU1ListResult, error) {
	id, err := valueobjects.U1ListIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteU1ListResult{}, fmt.Errorf("parsing u1 list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return inboundports.DeleteU1ListResult{}, fmt.Errorf("finding u1 list entry by id: %w", err)
	}

	if entry == nil {
		return inboundports.DeleteU1ListResult{}, &NotFoundError{Resource: "u1 list entry", ID: command.ID}
	}

	now := uc.now()
	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.DeleteByID(txCtx, id); err != nil {
			return fmt.Errorf("deleting u1 list entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, deleteU1ListAdminEventType, map[string]any{
			"action":                  "delete",
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
		return inboundports.DeleteU1ListResult{}, fmt.Errorf("deleting u1 list transaction: %w", err)
	}

	return inboundports.DeleteU1ListResult{ID: command.ID}, nil
}
