package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateBugReportUseCase updates a bug report.
type UpdateBugReportUseCase struct {
	repository         outboundports.BugReportRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewUpdateBugReportUseCase builds an UpdateReportUseCase.
func NewUpdateBugReportUseCase(repository outboundports.BugReportRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *UpdateBugReportUseCase {
	return &UpdateBugReportUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute updates an existing bug report.
func (uc *UpdateBugReportUseCase) Execute(ctx context.Context, command inboundports.UpdateBugReportCommand) (outboundports.BugReportResult, error) {
	bugReport, err := uc.repository.FindByID(ctx, command.ID)
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("finding bug report by id: %w", err)
	}

	if bugReport == nil {
		return outboundports.BugReportResult{}, &NotFoundError{Resource: "bug_report", ID: command.ID.String()}
	}

	if err := bugReport.Update(command.Title, command.Description, command.Status, uc.now()); err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("updating bug report: %w", err)
	}

	now := uc.now()
	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Update(txCtx, bugReport); err != nil {
			return fmt.Errorf("persisting bug report update: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, updateBugReportAdminEventType, map[string]any{
			"action":      "update",
			"resource":    "bug_report",
			"target_id":   bugReport.ID().String(),
			"title":       bugReport.Title(),
			"description": bugReport.Description(),
			"status":      bugReport.Status(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("updating bug report transaction: %w", err)
	}

	return newBugReportResult(bugReport), nil
}
