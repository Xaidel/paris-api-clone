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

// TestGetReportUseCaseExecute verifies the get report use case execute behavior and the expected outcome asserted below.
func TestGetReportUseCaseExecute(t *testing.T) {
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

	title, err := valueobjects.NewBugReportTitle("Incorrect classification")
	if err != nil {
		t.Fatalf("NewBugReportTitle() error = %v", err)
	}

	description, err := valueobjects.NewBugReportDescription("Some description")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	bugReport := entities.ReconstituteBugReport(reportID, userID, transactionID, title, description, valueobjects.OpenBugReportStatus(), testTime(), testTime())

	tests := []struct {
		name        string
		repository  *reportRepositoryMock
		recorder    *adminEventRecorderMock
		query       inboundports.GetBugReportQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.BugReportResult)
	}{
		{
			name:       "gets bug report",
			repository: &reportRepositoryMock{findByID: bugReport},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetBugReportQuery{ID: reportID, ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.BugReportResult) {
				t.Helper()
				if result.ID != reportID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, reportID.String())
				}
				if result.Title != "Incorrect classification" {
					t.Fatalf("result.Title = %q, want %q", result.Title, "Incorrect classification")
				}
				if result.Status != "Open" {
					t.Fatalf("result.Status = %q, want %q", result.Status, "Open")
				}
			},
		},
		{
			name:       "returns not found",
			repository: &reportRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetBugReportQuery{ID: reportID},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:       "wraps repository error",
			repository: &reportRepositoryMock{findByIDErr: errors.New("boom")},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetBugReportQuery{ID: reportID},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "finding bug report by id") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := NewGetBugReportUseCase(tc.repository, tc.recorder).Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result)
		})
	}
}
