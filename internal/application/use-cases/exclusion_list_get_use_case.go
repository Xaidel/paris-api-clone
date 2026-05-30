package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetExclusionListUseCase gets an exclusion list entry by identifier.
type GetExclusionListUseCase struct {
	repository    outboundports.ExclusionListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewGetExclusionListUseCase builds a GetExclusionListUseCase.
func NewGetExclusionListUseCase(repository outboundports.ExclusionListRepository, eventRecorder adminEventRecorder) *GetExclusionListUseCase {
	return &GetExclusionListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute loads an exclusion list entry by identifier.
func (uc *GetExclusionListUseCase) Execute(ctx context.Context, query inboundports.GetExclusionListQuery) (outboundports.ExclusionListResult, error) {
	id, err := valueobjects.ExclusionListIDFromString(query.ID)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("parsing exclusion list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.ExclusionListResult{}, fmt.Errorf("finding exclusion list entry by id: %w", err)
	}

	if entry == nil {
		return outboundports.ExclusionListResult{}, &NotFoundError{Resource: "exclusion list entry", ID: query.ID}
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, getExclusionListAdminEventType, map[string]any{
		"action":        "read",
		"resource":      "exclusion_list",
		"target_id":     entry.ID().String(),
		"activity_type": entry.ActivityType(),
	}); err != nil {
		return outboundports.ExclusionListResult{}, err
	}

	return newExclusionListResult(entry), nil
}
