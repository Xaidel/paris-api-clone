package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteGroupUseCase deletes a group.
type DeleteGroupUseCase struct {
	repository         outboundports.GroupRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
}

// NewDeleteGroupUseCase builds a DeleteGroupUseCase.
func NewDeleteGroupUseCase(repository outboundports.GroupRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *DeleteGroupUseCase {
	return &DeleteGroupUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute deletes an existing group.
func (uc *DeleteGroupUseCase) Execute(ctx context.Context, command inboundports.DeleteGroupCommand) (inboundports.DeleteGroupResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.DeleteGroupResult{}, err
	}

	id, err := valueobjects.GroupIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteGroupResult{}, fmt.Errorf("parsing group id: %w", err)
	}

	group, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return inboundports.DeleteGroupResult{}, fmt.Errorf("finding group by id: %w", err)
	}

	if group == nil {
		return inboundports.DeleteGroupResult{}, &NotFoundError{Resource: "group", ID: command.ID}
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.DeleteByID(txCtx, id); err != nil {
			return fmt.Errorf("deleting group: %w", err)
		}

		if err := group.RecordDeleted(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording group deletion event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, group.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing group events: %w", err)
		}

		return nil
	}); err != nil {
		return inboundports.DeleteGroupResult{}, fmt.Errorf("deleting group transaction: %w", err)
	}

	return inboundports.DeleteGroupResult{ID: command.ID}, nil
}
