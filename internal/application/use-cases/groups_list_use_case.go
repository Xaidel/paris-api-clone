package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain/events"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListGroupsUseCase lists groups.
type ListGroupsUseCase struct {
	repository     outboundports.GroupRepository
	eventRecorder  adminEventRecorder
	actorDirectory outboundports.ActorDirectory
	now            func() time.Time
}

// NewListGroupsUseCase builds a ListGroupsUseCase.
func NewListGroupsUseCase(repository outboundports.GroupRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *ListGroupsUseCase {
	return &ListGroupsUseCase{repository: repository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute lists all groups.
func (uc *ListGroupsUseCase) Execute(ctx context.Context, query inboundports.ListGroupsQuery) (outboundports.ListGroupsResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.ListGroupsResult{}, err
	}

	groups, err := uc.repository.List(ctx)
	if err != nil {
		return outboundports.ListGroupsResult{}, fmt.Errorf("listing groups: %w", err)
	}

	results := make([]outboundports.GroupResult, 0, len(groups))
	for _, group := range groups {
		results = append(results, newGroupResult(group))
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, events.ListGroupsEventType, map[string]any{
		"action":       "list",
		"resource":     "group",
		"result_count": len(results),
	}); err != nil {
		return outboundports.ListGroupsResult{}, err
	}

	return outboundports.ListGroupsResult{Groups: results}, nil
}
