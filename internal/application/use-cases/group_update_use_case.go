package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateGroupUseCase updates a group.
type UpdateGroupUseCase struct {
	repository         outboundports.GroupRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
}

// NewUpdateGroupUseCase builds an UpdateGroupUseCase.
func NewUpdateGroupUseCase(repository outboundports.GroupRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *UpdateGroupUseCase {
	return &UpdateGroupUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute updates an existing group.
func (uc *UpdateGroupUseCase) Execute(ctx context.Context, command inboundports.UpdateGroupCommand) (outboundports.GroupResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.GroupResult{}, err
	}

	id, err := valueobjects.GroupIDFromString(command.ID)
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("parsing group id: %w", err)
	}

	group, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("finding group by id: %w", err)
	}

	if group == nil {
		return outboundports.GroupResult{}, &NotFoundError{Resource: "group", ID: command.ID}
	}

	if err := group.Update(command.Name); err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("updating group: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Update(txCtx, group); err != nil {
			return fmt.Errorf("updating group: %w", err)
		}

		if err := group.RecordUpdated(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording group update event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, group.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing group events: %w", err)
		}

		return nil
	}); err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("updating group transaction: %w", err)
	}

	return newGroupResult(group), nil
}
