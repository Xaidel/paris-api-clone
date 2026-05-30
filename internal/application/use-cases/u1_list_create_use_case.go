package usecases

import (
	"context"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateU1ListUseCase creates a new U1 list entry.
type CreateU1ListUseCase struct {
	repository         outboundports.U1ListRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	newID              func() (valueobjects.U1ListID, error)
	now                func() time.Time
}

// NewCreateU1ListUseCase builds a CreateU1ListUseCase.
func NewCreateU1ListUseCase(repository outboundports.U1ListRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *CreateU1ListUseCase {
	return &CreateU1ListUseCase{
		repository:         repository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		newID:              valueobjects.NewU1ListID,
		now:                time.Now,
	}
}

// Execute creates and persists a U1 list entry.
func (uc *CreateU1ListUseCase) Execute(ctx context.Context, command inboundports.CreateU1ListCommand) (outboundports.U1ListResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.U1ListResult{}, err
	}

	createdBy, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("parsing actor user id: %w", err)
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("generating u1 list id: %w", err)
	}

	entry, err := entities.NewU1ListEntry(id, command.Sector, command.EligibleOperationType, command.ConditionGuidance)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("creating u1 list entry: %w", err)
	}

	now := uc.now()
	entry.SetCreatedBy(createdBy)
	if err := entry.SetAuditTimestamps(now, now); err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("setting u1 list audit timestamps: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, entry, command.ActorUserID); err != nil {
			return fmt.Errorf("creating u1 list entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createU1ListAdminEventType, map[string]any{
			"action":                  "create",
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
		return outboundports.U1ListResult{}, fmt.Errorf("creating u1 list transaction: %w", err)
	}

	return newU1ListResult(entry), nil
}
