package usecases

import (
	"context"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateBugReportUseCase creates a new bug report.
type CreateBugReportUseCase struct {
	repository            outboundports.BugReportRepository
	transactionRepository outboundports.TransactionRepository
	transactionManager    outboundports.TransactionManager
	eventRecorder         adminEventRecorder
	newID                 func() (valueobjects.BugReportID, error)
	now                   func() time.Time
}

// NewCreateBugReportUseCase builds a CreateBugReportUseCase.
func NewCreateBugReportUseCase(repository outboundports.BugReportRepository, transactionRepository outboundports.TransactionRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *CreateBugReportUseCase {
	return &CreateBugReportUseCase{
		repository:            repository,
		transactionRepository: transactionRepository,
		transactionManager:    transactionManager,
		eventRecorder:         eventRecorder,
		newID:                 valueobjects.NewBugReportID,
		now:                   time.Now,
	}
}

// Execute creates and persists a bug report.
func (uc *CreateBugReportUseCase) Execute(ctx context.Context, command inboundports.CreateBugReportCommand) (outboundports.BugReportResult, error) {
	userID, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("parsing user id: %w", err)
	}

	tx, err := uc.transactionRepository.FindByID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("checking transaction existence: %w", err)
	}
	if tx == nil {
		return outboundports.BugReportResult{}, &NotFoundError{Resource: "transaction", ID: command.TransactionID.String()}
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("generating bug report id: %w", err)
	}

	now := uc.now()

	bugReport, err := entities.NewBugReport(id, userID, command.TransactionID, command.Title, command.Description, valueobjects.OpenBugReportStatus(), now)
	if err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("creating bug report: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, bugReport); err != nil {
			return fmt.Errorf("persisting bug report: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createBugReportAdminEventType, map[string]any{
			"action":         "create",
			"resource":       "bug_report",
			"target_id":      bugReport.ID().String(),
			"transaction_id": bugReport.TransactionID(),
			"title":          bugReport.Title(),
			"description":    bugReport.Description(),
			"status":         bugReport.Status(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.BugReportResult{}, fmt.Errorf("creating bug report transaction: %w", err)
	}

	return newBugReportResult(bugReport), nil
}
