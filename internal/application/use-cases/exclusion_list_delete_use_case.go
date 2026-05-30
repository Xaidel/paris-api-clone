package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteExclusionListUseCase deletes an exclusion list entry.
type DeleteExclusionListUseCase struct {
	repository    outboundports.ExclusionListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewDeleteExclusionListUseCase builds a DeleteExclusionListUseCase.
func NewDeleteExclusionListUseCase(repository outboundports.ExclusionListRepository, eventRecorder adminEventRecorder) *DeleteExclusionListUseCase {
	return &DeleteExclusionListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute deletes an existing exclusion list entry.
func (uc *DeleteExclusionListUseCase) Execute(ctx context.Context, command inboundports.DeleteExclusionListCommand) (inboundports.DeleteExclusionListResult, error) {
	id, err := valueobjects.ExclusionListIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteExclusionListResult{}, fmt.Errorf("parsing exclusion list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return inboundports.DeleteExclusionListResult{}, fmt.Errorf("finding exclusion list entry by id: %w", err)
	}

	if entry == nil {
		return inboundports.DeleteExclusionListResult{}, &NotFoundError{Resource: "exclusion list entry", ID: command.ID}
	}

	if err := uc.repository.DeleteByID(ctx, id); err != nil {
		return inboundports.DeleteExclusionListResult{}, fmt.Errorf("deleting exclusion list entry: %w", err)
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), command.ActorUserID, command.ActorGroupID, deleteExclusionListAdminEventType, map[string]any{
		"action":        "delete",
		"resource":      "exclusion_list",
		"target_id":     entry.ID().String(),
		"activity_type": entry.ActivityType(),
	}); err != nil {
		return inboundports.DeleteExclusionListResult{}, err
	}

	return inboundports.DeleteExclusionListResult{ID: command.ID}, nil
}
