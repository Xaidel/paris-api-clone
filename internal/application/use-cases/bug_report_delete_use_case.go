package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteBugReportUseCase deletes a bug report.
type DeleteBugReportUseCase struct {
	repository         outboundports.BugReportRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewDeleteBugReportUseCase builds a DeleteReportUseCase.
func NewDeleteBugReportUseCase(repository outboundports.BugReportRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *DeleteBugReportUseCase {
	return &DeleteBugReportUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute deletes an existing bug report.
func (uc *DeleteBugReportUseCase) Execute(ctx context.Context, command inboundports.DeleteBugReportCommand) (inboundports.DeleteBugReportResult, error) {
	bugReport, err := uc.repository.FindByID(ctx, command.ID)
	if err != nil {
		return inboundports.DeleteBugReportResult{}, fmt.Errorf("finding bug report by id: %w", err)
	}

	if bugReport == nil {
		return inboundports.DeleteBugReportResult{}, &NotFoundError{Resource: "bug_report", ID: command.ID.String()}
	}

	now := uc.now()

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.DeleteByID(txCtx, command.ID); err != nil {
			return fmt.Errorf("deleting bug report: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, deleteBugReportAdminEventType, map[string]any{
			"action":    "delete",
			"resource":  "bug_report",
			"target_id": bugReport.ID().String(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return inboundports.DeleteBugReportResult{}, fmt.Errorf("deleting bug report transaction: %w", err)
	}

	return inboundports.DeleteBugReportResult{ID: command.ID.String()}, nil
}
