package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteBugReportUseCaseExecute verifies the delete bug report use case execute behavior and the expected outcome asserted below.
func TestDeleteBugReportUseCaseExecute(t *testing.T) {
	t.Parallel()

	bugReportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
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

	title, err := valueobjects.NewBugReportTitle("Incorrect classification")
	if err != nil {
		t.Fatalf("NewBugReportTitle() error = %v", err)
	}

	description, err := valueobjects.NewBugReportDescription("Some description")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	bugReport := entities.ReconstituteBugReport(bugReportID, userID, transactionID, title, description, valueobjects.OpenBugReportStatus(), testTime(), testTime())

	tests := []struct {
		name        string
		repository  *reportRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.DeleteBugReportCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DeleteBugReportResult, repository *reportRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "deletes bug report",
			repository:  &reportRepositoryMock{findByID: bugReport},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteBugReportCommand{ID: bugReportID, ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.DeleteBugReportResult, repository *reportRepositoryMock, transaction *transactionManagerMock) {
				t.Helper()

				if repository.deletedID != bugReportID.String() {
					t.Fatalf("repository.deletedID = %q, want %q", repository.deletedID, bugReportID.String())
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if result.ID != bugReportID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, bugReportID.String())
				}
			},
		},
		{
			name:        "returns not found",
			repository:  &reportRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteBugReportCommand{ID: bugReportID},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:        "wraps repository delete error",
			repository:  &reportRepositoryMock{findByID: bugReport, deleteErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteBugReportCommand{ID: bugReportID, ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "deleting bug report transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewDeleteBugReportUseCase(tc.repository, tc.transaction, tc.recorder).Execute(context.Background(), tc.command)
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
