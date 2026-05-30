package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateExclusionListUseCase updates an exclusion list entry.
type UpdateExclusionListUseCase struct {
	repository    outboundports.ExclusionListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewUpdateExclusionListUseCase builds an UpdateExclusionListUseCase.
func NewUpdateExclusionListUseCase(repository outboundports.ExclusionListRepository, eventRecorder adminEventRecorder) *UpdateExclusionListUseCase {
	return &UpdateExclusionListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute updates an existing exclusion list entry.
func (uc *UpdateExclusionListUseCase) Execute(ctx context.Context, command inboundports.UpdateExclusionListCommand) (outboundports.ExclusionListResult, error) {
	id, err := valueobjects.ExclusionListIDFromString(command.ID)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("parsing exclusion list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("finding exclusion list entry by id: %w", err)
	}

	if entry == nil {
		return outboundports.ExclusionListResult{}, &NotFoundError{Resource: "exclusion list entry", ID: command.ID}
	}

	if err := entry.UpdateActivityType(command.ActivityType); err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("updating exclusion list activity type: %w", err)
	}

	now := uc.now()
	if err := entry.Touch(now); err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("updating exclusion list timestamp: %w", err)
	}

	if err := uc.repository.Update(ctx, entry); err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("updating exclusion list entry: %w", err)
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, updateExclusionListAdminEventType, map[string]any{
		"action":        "update",
		"resource":      "exclusion_list",
		"target_id":     entry.ID().String(),
		"activity_type": entry.ActivityType(),
	}); err != nil {
		return outboundports.ExclusionListResult{}, err
	}

	return newExclusionListResult(entry), nil
}
