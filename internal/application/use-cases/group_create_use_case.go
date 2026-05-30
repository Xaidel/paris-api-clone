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

// CreateGroupUseCase creates a new group entry.
type CreateGroupUseCase struct {
	repository         outboundports.GroupRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
	newID              func() (valueobjects.GroupID, error)
}

// NewCreateGroupUseCase builds a CreateGroupUseCase.
func NewCreateGroupUseCase(repository outboundports.GroupRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		repository:         repository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		now:                time.Now,
		newID:              valueobjects.NewGroupID,
	}
}

// Execute creates and persists a group entry.
func (uc *CreateGroupUseCase) Execute(ctx context.Context, command inboundports.CreateGroupCommand) (outboundports.GroupResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.GroupResult{}, err
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("generating group id: %w", err)
	}

	group, err := entities.NewGroup(id, command.Name)
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("creating group entry: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, group); err != nil {
			return fmt.Errorf("creating group entry: %w", err)
		}

		if err := group.RecordCreated(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording group creation event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, group.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing group events: %w", err)
		}

		return nil
	}); err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("creating group transaction: %w", err)
	}

	return newGroupResult(group), nil
}
