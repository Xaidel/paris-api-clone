package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListReportsUseCaseExecute verifies the list reports use case execute behavior and the expected outcome asserted below.
func TestListReportsUseCaseExecute(t *testing.T) {
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

	repository := &reportRepositoryMock{listReports: []*entities.BugReport{bugReport}}
	useCase := NewListBugReportsUseCase(repository, &adminEventRecorderMock{})

	result, err := useCase.Execute(context.Background(), inboundports.ListBugReportsQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.BugReports) != 1 {
		t.Fatalf("len(result.BugReports) = %d, want 1", len(result.BugReports))
	}

	if result.BugReports[0].ID != reportID.String() {
		t.Fatalf("result.BugReports[0].ID = %q, want %q", result.BugReports[0].ID, reportID.String())
	}

	if result.BugReports[0].Status != "Open" {
		t.Fatalf("result.BugReports[0].Status = %q, want %q", result.BugReports[0].Status, "Open")
	}
}
