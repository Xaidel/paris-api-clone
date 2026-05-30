package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetBugReportUseCase gets a bug report by identifier.
type GetBugReportUseCase struct {
	repository    outboundports.BugReportRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewGetBugReportUseCase builds a GetBugReportUseCase.
func NewGetBugReportUseCase(repository outboundports.BugReportRepository, eventRecorder adminEventRecorder) *GetBugReportUseCase {
	return &GetBugReportUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute loads a bug report by identifier.
func (uc *GetBugReportUseCase) Execute(ctx context.Context, query inboundports.GetBugReportQuery) (outboundports.BugReportResult, error) {
	bugReport, err := uc.repository.FindByID(ctx, query.ID)
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("finding bug report by id: %w", err)
	}

	if bugReport == nil {
		return outboundports.BugReportResult{}, &NotFoundError{Resource: "bug_report", ID: query.ID.String()}
	}

	now := uc.now()

	if err := recordAdminEvent(ctx, uc.eventRecorder, now, query.ActorUserID, query.ActorGroupID, getBugReportAdminEventType, map[string]any{
		"action":    "read",
		"resource":  "bug_report",
		"target_id": bugReport.ID().String(),
	}); err != nil {
		return outboundports.BugReportResult{}, err
	}

	return newBugReportResult(bugReport), nil
}
