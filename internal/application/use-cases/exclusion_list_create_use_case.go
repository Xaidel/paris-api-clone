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

// CreateExclusionListUseCase creates a new exclusion list entry.
type CreateExclusionListUseCase struct {
	repository         outboundports.ExclusionListRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	newID              func() (valueobjects.ExclusionListID, error)
	now                func() time.Time
}

// NewCreateExclusionListUseCase builds a CreateExclusionListUseCase.
func NewCreateExclusionListUseCase(repository outboundports.ExclusionListRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *CreateExclusionListUseCase {
	return &CreateExclusionListUseCase{
		repository:         repository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		newID:              valueobjects.NewExclusionListID,
		now:                time.Now,
	}
}

// Execute creates and persists an exclusion list entry.
func (uc *CreateExclusionListUseCase) Execute(ctx context.Context, command inboundports.CreateExclusionListCommand) (outboundports.ExclusionListResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.ExclusionListResult{}, err
	}

	createdBy, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("parsing actor user id: %w", err)
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("generating exclusion list id: %w", err)
	}

	entry, err := entities.NewExclusionListEntry(id, command.ActivityType)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("creating exclusion list entry: %w", err)
	}

	now := uc.now()
	entry.SetCreatedBy(createdBy)
	if err := entry.SetAuditTimestamps(now, now); err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("setting exclusion list audit timestamps: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, entry, command.ActorUserID); err != nil {
			return fmt.Errorf("creating exclusion list entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createExclusionListAdminEventType, map[string]any{
			"action":        "create",
			"resource":      "exclusion_list",
			"target_id":     entry.ID().String(),
			"activity_type": entry.ActivityType(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("creating exclusion list transaction: %w", err)
	}

	return newExclusionListResult(entry), nil
}
