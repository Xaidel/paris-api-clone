package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListExclusionListUseCase lists exclusion list entries.
type ListExclusionListUseCase struct {
	repository    outboundports.ExclusionListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewListExclusionListUseCase builds a ListExclusionListUseCase.
func NewListExclusionListUseCase(repository outboundports.ExclusionListRepository, eventRecorder adminEventRecorder) *ListExclusionListUseCase {
	return &ListExclusionListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute lists all exclusion list entries.
func (uc *ListExclusionListUseCase) Execute(ctx context.Context, query inboundports.ListExclusionListQuery) (outboundports.ListExclusionListResult, error) {
	entries, err := uc.repository.List(ctx)
	if err != nil {
		return outboundports.ListExclusionListResult{}, fmt.Errorf("listing exclusion list entries: %w", err)
	}

	results := make([]outboundports.ExclusionListResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, newExclusionListResult(entry))
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, listExclusionListAdminEventType, map[string]any{
		"action":       "list",
		"resource":     "exclusion_list",
		"result_count": len(results),
	}); err != nil {
		return outboundports.ListExclusionListResult{}, err
	}

	return outboundports.ListExclusionListResult{Entries: results}, nil
}
