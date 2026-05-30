package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpdateBugReportUseCaseExecute verifies the update bug report use case execute behavior and the expected outcome asserted below.
func TestUpdateBugReportUseCaseExecute(t *testing.T) {
	t.Parallel()

	reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("ReportIDFromString() error = %v", err)
	}

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300004")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	oldTitle, err := valueobjects.NewBugReportTitle("Old title")
	if err != nil {
		t.Fatalf("NewBugReportTitle() error = %v", err)
	}

	oldDescription, err := valueobjects.NewBugReportDescription("Old description")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	newTitle, err := valueobjects.NewBugReportTitle("New title")
	if err != nil {
		t.Fatalf("NewBugReportTitle() error = %v", err)
	}

	newDescription, err := valueobjects.NewBugReportDescription("New description")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	closedStatus, err := valueobjects.BugReportStatusFromString("Closed")
	if err != nil {
		t.Fatalf("BugReportStatusFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *reportRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.UpdateBugReportCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.BugReportResult, repository *reportRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "updates bug report",
			repository:  &reportRepositoryMock{findByID: entities.ReconstituteBugReport(reportID, userID, transactionID, oldTitle, oldDescription, valueobjects.OpenBugReportStatus(), testTime(), testTime())},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateBugReportCommand{ID: reportID, Title: newTitle, Description: newDescription, Status: closedStatus, ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.BugReportResult, repository *reportRepositoryMock, transaction *transactionManagerMock) {
				t.Helper()

				if repository.updatedReport == nil {
					t.Fatal("expected updated bug report")
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if result.Title != "New title" {
					t.Fatalf("result.Title = %q, want %q", result.Title, "New title")
				}

				if result.Status != "Closed" {
					t.Fatalf("result.Status = %q, want %q", result.Status, "Closed")
				}

				if !result.UpdatedAt.Equal(testTime()) {
					t.Fatalf("result.UpdatedAt = %v, want %v", result.UpdatedAt, testTime())
				}
			},
		},
		{
			name:        "returns not found",
			repository:  &reportRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateBugReportCommand{ID: reportID, Title: newTitle, Description: newDescription, Status: valueobjects.OpenBugReportStatus()},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:        "wraps repository update error",
			repository:  &reportRepositoryMock{findByID: entities.ReconstituteBugReport(reportID, userID, transactionID, oldTitle, oldDescription, valueobjects.OpenBugReportStatus(), testTime(), testTime()), updateErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateBugReportCommand{ID: reportID, Title: newTitle, Description: newDescription, Status: valueobjects.OpenBugReportStatus(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "updating bug report transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewUpdateBugReportUseCase(tc.repository, tc.transaction, tc.recorder)
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.transaction)
		})
	}
}
