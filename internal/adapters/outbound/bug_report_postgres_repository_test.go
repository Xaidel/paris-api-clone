package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPostgresBugReportRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("BugReportIDFromString() error = %v", err)
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

	description, err := valueobjects.NewBugReportDescription("This transaction should be aligned.")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	status := valueobjects.OpenBugReportStatus()
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	bugReport, err := entities.NewBugReport(reportID, userID, transactionID, title, description, status, now)
	if err != nil {
		t.Fatalf("NewBugReport() error = %v", err)
	}

	repository := NewPostgresBugReportRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createBugReportQuery)).
		WithArgs(reportID.String(), userID.String(), transactionID.String(), title.String(), description.String(), status.String(), now, now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), bugReport); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestPostgresBugReportRepositoryFindByID(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantNil bool
	}{
		{
			name: "returns bug report",
			rows: pgxmock.NewRows([]string{"id", "user_id", "transaction_id", "title", "description", "status", "created_at", "updated_at"}).
				AddRow("01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300004", "01962b8f-aeb2-7e03-a8ff-1edce1300001", "Incorrect classification", "This transaction should be aligned.", "Open", now, now),
		},
		{
			name:    "returns nil when not found",
			rows:    pgxmock.NewRows([]string{"id", "user_id", "transaction_id", "title", "description", "status", "created_at", "updated_at"}),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("pgxmock.NewPool() error = %v", err)
			}
			defer mock.Close()

			reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
			if err != nil {
				t.Fatalf("BugReportIDFromString() error = %v", err)
			}

			mock.ExpectQuery(regexp.QuoteMeta(findBugReportByIDQuery)).WithArgs(reportID.String()).WillReturnRows(tt.rows)
			repository := NewPostgresBugReportRepository(mock)

			bugReport, err := repository.FindByID(context.Background(), reportID)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			if tt.wantNil {
				if bugReport != nil {
					t.Fatal("expected nil bug report")
				}
				return
			}

			if bugReport.ID().String() != reportID.String() {
				t.Fatalf("bugReport.ID().String() = %q, want %q", bugReport.ID().String(), reportID.String())
			}
			if bugReport.UserID() != "01962b8f-aeb2-7e03-a8ff-1edce1300004" {
				t.Fatalf("bugReport.UserID() = %q, want %q", bugReport.UserID(), "01962b8f-aeb2-7e03-a8ff-1edce1300004")
			}
			if bugReport.TransactionID() != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
				t.Fatalf("bugReport.TransactionID() = %q, want %q", bugReport.TransactionID(), "01962b8f-aeb2-7e03-a8ff-1edce1300001")
			}
		})
	}
}

func TestPostgresBugReportRepositoryListUpdateDelete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresBugReportRepository(mock)
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("BugReportIDFromString() error = %v", err)
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

	description, err := valueobjects.NewBugReportDescription("This transaction should be aligned.")
	if err != nil {
		t.Fatalf("NewBugReportDescription() error = %v", err)
	}

	status := valueobjects.OpenBugReportStatus()
	bugReport := entities.ReconstituteBugReport(reportID, userID, transactionID, title, description, status, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(listBugReportsQuery)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "user_id", "transaction_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow(reportID.String(), userID.String(), transactionID.String(), title.String(), description.String(), status.String(), now, now),
	)

	bugReports, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(bugReports) != 1 {
		t.Fatalf("len(bugReports) = %d, want 1", len(bugReports))
	}

	mock.ExpectExec(regexp.QuoteMeta(updateBugReportQuery)).
		WithArgs(reportID.String(), title.String(), description.String(), status.String(), now).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	if err := repository.Update(context.Background(), bugReport); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteBugReportQuery)).WithArgs(reportID.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), reportID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
}

func TestScanBugReportInvalidData(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	reportID, err := valueobjects.BugReportIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("BugReportIDFromString() error = %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(findBugReportByIDQuery)).WithArgs(reportID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id", "user_id", "transaction_id", "title", "description", "status", "created_at", "updated_at"}).
			AddRow("bad-id", "01962b8f-aeb2-7e03-a8ff-1edce1300004", "01962b8f-aeb2-7e03-a8ff-1edce1300001", "Incorrect classification", "This transaction should be aligned.", "Open", time.Now(), time.Now()),
	)

	repository := NewPostgresBugReportRepository(mock)
	_, err = repository.FindByID(context.Background(), reportID)
	if err == nil {
		t.Fatal("expected error for invalid scanned bug report data")
	}
}
