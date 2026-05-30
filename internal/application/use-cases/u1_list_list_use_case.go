package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListU1ListUseCase lists U1 list entries.
type ListU1ListUseCase struct {
	repository    outboundports.U1ListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewListU1ListUseCase builds a ListU1ListUseCase.
func NewListU1ListUseCase(repository outboundports.U1ListRepository, eventRecorder adminEventRecorder) *ListU1ListUseCase {
	return &ListU1ListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute lists all U1 list entries.
func (uc *ListU1ListUseCase) Execute(ctx context.Context, query inboundports.ListU1ListQuery) (outboundports.ListU1ListResult, error) {
	filter := outboundports.U1ListFilter{Sector: strings.ToLower(strings.TrimSpace(query.Sector))}

	entries, err := uc.repository.List(ctx, filter)
	if err != nil {
		return outboundports.ListU1ListResult{}, fmt.Errorf("listing u1 list entries: %w", err)
	}

	results := make([]outboundports.U1ListResult, 0, len(entries))
	for _, entry := range entries {
		results = append(results, newU1ListResult(entry))
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, listU1ListAdminEventType, map[string]any{
		"action":        "list",
		"resource":      "u1_list",
		"filter_sector": filter.Sector,
		"result_count":  len(results),
	}); err != nil {
		return outboundports.ListU1ListResult{}, err
	}

	return outboundports.ListU1ListResult{Entries: results}, nil
}
