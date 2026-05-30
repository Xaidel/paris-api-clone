package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestCreateReportUseCaseExecute verifies the create report use case execute behavior and the expected outcome asserted below.
func TestCreateReportUseCaseExecute(t *testing.T) {
	t.Parallel()

	reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("ReportIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	title, err := valueobjects.NewBugReportTitle("Incorrect classification")
	if err != nil {
		t.Fatalf("NewBugReportTitle() error = %v", err)
	}

	description, err := valueobjects.NewBugReportDescription("Transaction was incorrectly classified")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	tests := []struct {
		name            string
		reportRepo      *reportRepositoryMock
		transactionRepo *transactionRepositoryMock
		transaction     *transactionManagerMock
		recorder        *adminEventRecorderMock
		command         inboundports.CreateBugReportCommand
		assertError     func(t *testing.T, err error)
		assert          func(t *testing.T, result outboundports.BugReportResult, reportRepo *reportRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:            "creates bug report",
			reportRepo:      &reportRepositoryMock{},
			transactionRepo: &transactionRepositoryMock{findByID: transaction},
			transaction:     &transactionManagerMock{},
			recorder:        &adminEventRecorderMock{},
			command:         inboundports.CreateBugReportCommand{TransactionID: transactionID, Title: title, Description: description, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.BugReportResult, reportRepo *reportRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if reportRepo.createdReport == nil {
					t.Fatal("expected bug report to be created")
				}

				if reportRepo.createdReport.UserID() != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("bugReport.UserID() = %q, want %q", reportRepo.createdReport.UserID(), "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if reportRepo.createdReport.TransactionID() != transactionID.String() {
					t.Fatalf("bugReport.TransactionID() = %q, want %q", reportRepo.createdReport.TransactionID(), transactionID.String())
				}

				if result.UserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("result.UserID = %q, want %q", result.UserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.Title != "Incorrect classification" {
					t.Fatalf("result.Title = %q, want %q", result.Title, "Incorrect classification")
				}

				if result.Status != "Open" {
					t.Fatalf("result.Status = %q, want %q", result.Status, "Open")
				}

				if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
					t.Fatalf("expected non-zero timestamps, got created_at=%v updated_at=%v", result.CreatedAt, result.UpdatedAt)
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.EventType != createBugReportAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createBugReportAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["title"] != "Incorrect classification" {
					t.Fatalf("payload[title] = %v, want %q", payload["title"], "Incorrect classification")
				}
			},
		},
		{
			name:            "returns not found when transaction does not exist",
			reportRepo:      &reportRepositoryMock{},
			transactionRepo: &transactionRepositoryMock{},
			transaction:     &transactionManagerMock{},
			recorder:        &adminEventRecorderMock{},
			command:         inboundports.CreateBugReportCommand{TransactionID: transactionID, Title: title, Description: description, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
				if notFoundErr.Resource != "transaction" {
					t.Fatalf("notFoundErr.Resource = %q, want %q", notFoundErr.Resource, "transaction")
				}
			},
		},
		{
			name:            "wraps transaction existence check error",
			reportRepo:      &reportRepositoryMock{},
			transactionRepo: &transactionRepositoryMock{findByIDErr: errors.New("db down")},
			transaction:     &transactionManagerMock{},
			recorder:        &adminEventRecorderMock{},
			command:         inboundports.CreateBugReportCommand{TransactionID: transactionID, Title: title, Description: description, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "checking transaction existence") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
		{
			name:            "wraps repository error",
			reportRepo:      &reportRepositoryMock{createErr: errors.New("boom")},
			transactionRepo: &transactionRepositoryMock{findByID: transaction},
			transaction:     &transactionManagerMock{},
			recorder:        &adminEventRecorderMock{},
			command:         inboundports.CreateBugReportCommand{TransactionID: transactionID, Title: title, Description: description, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "creating bug report transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateBugReportUseCase(tc.reportRepo, tc.transactionRepo, tc.transaction, tc.recorder)
			useCase.newID = func() (valueobjects.BugReportID, error) { return reportID, nil }
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.reportRepo, tc.transaction, tc.recorder)
		})
	}
}
