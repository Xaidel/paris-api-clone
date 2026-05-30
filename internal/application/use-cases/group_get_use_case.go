package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetGroupUseCase gets a group by identifier.
type GetGroupUseCase struct {
	repository     outboundports.GroupRepository
	eventRecorder  adminEventRecorder
	actorDirectory outboundports.ActorDirectory
	now            func() time.Time
}

// NewGetGroupUseCase builds a GetGroupUseCase.
func NewGetGroupUseCase(repository outboundports.GroupRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *GetGroupUseCase {
	return &GetGroupUseCase{repository: repository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute loads a group by identifier.
func (uc *GetGroupUseCase) Execute(ctx context.Context, query inboundports.GetGroupQuery) (outboundports.GroupResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.GroupResult{}, err
	}

	id, err := valueobjects.GroupIDFromString(query.ID)
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("parsing group id: %w", err)
	}

	group, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("finding group by id: %w", err)
	}

	if group == nil {
		return outboundports.GroupResult{}, &NotFoundError{Resource: "group", ID: query.ID}
	}

	if err := group.RecordRead(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("recording group read event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, group.PullDomainEvents()); err != nil {
		return outboundports.GroupResult{}, fmt.Errorf("publishing group events: %w", err)
	}

	return newGroupResult(group), nil
}
