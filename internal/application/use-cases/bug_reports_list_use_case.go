package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListBugReportsUseCase lists all bug reoutboundports.
type ListBugReportsUseCase struct {
	repository    outboundports.BugReportRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewListBugReportsUseCase builds a ListBugReportsUseCase.
func NewListBugReportsUseCase(repository outboundports.BugReportRepository, eventRecorder adminEventRecorder) *ListBugReportsUseCase {
	return &ListBugReportsUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute lists all bug reoutboundports.
func (uc *ListBugReportsUseCase) Execute(ctx context.Context, query inboundports.ListBugReportsQuery) (outboundports.ListBugReportsResult, error) {
	bugReports, err := uc.repository.List(ctx)
	if err != nil {
		return outboundports.ListBugReportsResult{}, fmt.Errorf("listing bug reports: %w", err)
	}

	results := make([]outboundports.BugReportResult, 0, len(bugReports))
	for _, bugReport := range bugReports {
		results = append(results, newBugReportResult(bugReport))
	}

	now := uc.now()

	if err := recordAdminEvent(ctx, uc.eventRecorder, now, query.ActorUserID, query.ActorGroupID, listBugReportsAdminEventType, map[string]any{
		"action":       "list",
		"resource":     "bug_report",
		"result_count": len(results),
	}); err != nil {
		return outboundports.ListBugReportsResult{}, err
	}

	return outboundports.ListBugReportsResult{BugReports: results}, nil
}
