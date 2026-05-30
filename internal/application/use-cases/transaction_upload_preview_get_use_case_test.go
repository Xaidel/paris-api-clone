package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

func TestGetTransactionUploadPreviewUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixture := newTransactionUploadPreviewFixtures(t)
	uploadID := fixture.uploadID
	groupID := fixture.groupID

	tests := []struct {
		name        string
		buildRepos  func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock)
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		query       inboundports.GetTransactionUploadPreviewQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.GetTransactionUploadPreviewResult, previewRepo *transactionUploadPreviewRepositoryMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock)
	}{
		{
			name: "gets transaction upload preview",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				fixture := newTransactionUploadPreviewFixtures(t)
				return &transactionUploadRepositoryMock{findByIDResult: fixture.upload}, &transactionUploadPreviewRepositoryMock{findByUploadIDResult: fixture.preview}
			},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: groupID.String(),
			},
			assert: func(t *testing.T, result inboundports.GetTransactionUploadPreviewResult, previewRepo *transactionUploadPreviewRepositoryMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock) {
				t.Helper()

				if result.FileID != uploadID.String() {
					t.Fatalf("result.FileID = %q, want %q", result.FileID, uploadID.String())
				}

				if result.FileName != "transactions.csv" {
					t.Fatalf("result.FileName = %q, want %q", result.FileName, "transactions.csv")
				}

				if len(result.Columns) != 2 || result.Columns[0] != "Product" || result.Columns[1] != "Year" {
					t.Fatalf("result.Columns = %v, want %v", result.Columns, []string{"Product", "Year"})
				}

				if len(result.Rows) != 1 || len(result.Rows[0]) != 2 || result.Rows[0][0] != "CG" || result.Rows[0][1] != "2026" {
					t.Fatalf("result.Rows = %v, want %v", result.Rows, [][]string{{"CG", "2026"}})
				}

				if result.TotalRows != 1 {
					t.Fatalf("result.TotalRows = %d, want %d", result.TotalRows, 1)
				}

				if len(result.ValidationErrors) != 1 {
					t.Fatalf("len(result.ValidationErrors) = %d, want %d", len(result.ValidationErrors), 1)
				}

				if result.ValidationErrors[0].Code != "MISSING_FIELD" {
					t.Fatalf("result.ValidationErrors[0].Code = %q, want %q", result.ValidationErrors[0].Code, "MISSING_FIELD")
				}

				if previewRepo.findByUploadIDResult == nil {
					t.Fatal("expected preview record lookup")
				}

				if actors.userID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("actors.userID = %q, want %q", actors.userID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if actors.groupID != groupID.String() {
					t.Fatalf("actors.groupID = %q, want %q", actors.groupID, groupID.String())
				}

				if recorder.command.EventType != domainevents.PreviewTransactionUploadEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, domainevents.PreviewTransactionUploadEventType)
				}
			},
		},
		{
			name: "returns not found when upload is missing",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				return &transactionUploadRepositoryMock{}, &transactionUploadPreviewRepositoryMock{}
			},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: groupID.String(),
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name: "returns forbidden for wrong group",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				fixture := newTransactionUploadPreviewFixtures(t)
				return &transactionUploadRepositoryMock{findByIDResult: fixture.upload}, &transactionUploadPreviewRepositoryMock{}
			},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300004",
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var forbiddenErr *ForbiddenError
				if !errors.As(err, &forbiddenErr) {
					t.Fatalf("expected ForbiddenError, got %v", err)
				}
			},
		},
		{
			name: "returns actor validation failure",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				return &transactionUploadRepositoryMock{}, &transactionUploadPreviewRepositoryMock{}
			},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{err: errors.New("unknown actor")},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: groupID.String(),
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil || !strings.Contains(err.Error(), "validating actor ids") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
		{
			name: "returns not found when preview is missing",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				fixture := newTransactionUploadPreviewFixtures(t)
				return &transactionUploadRepositoryMock{findByIDResult: fixture.upload}, &transactionUploadPreviewRepositoryMock{}
			},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: groupID.String(),
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name: "wraps event publication failure",
			buildRepos: func(t *testing.T) (*transactionUploadRepositoryMock, *transactionUploadPreviewRepositoryMock) {
				t.Helper()
				fixture := newTransactionUploadPreviewFixtures(t)
				return &transactionUploadRepositoryMock{findByIDResult: fixture.upload}, &transactionUploadPreviewRepositoryMock{findByUploadIDResult: fixture.preview}
			},
			recorder: &adminEventRecorderMock{err: errors.New("audit failed")},
			actors:   &actorDirectoryMock{},
			query: inboundports.GetTransactionUploadPreviewQuery{
				ID:           uploadID.String(),
				ActorUserID:  "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID: groupID.String(),
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil || !strings.Contains(err.Error(), "publishing transaction upload preview events") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uploadRepo, previewRepo := tc.buildRepos(t)
			useCase := NewGetTransactionUploadPreviewUseCase(uploadRepo, previewRepo, tc.recorder, tc.actors)
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, previewRepo, tc.recorder, tc.actors)
		})
	}
}

func TestNewTransactionUploadPreviewFixturesReturnsDistinctInstances(t *testing.T) {
	t.Parallel()

	first := newTransactionUploadPreviewFixtures(t)
	second := newTransactionUploadPreviewFixtures(t)

	if first.upload == second.upload {
		t.Fatal("newTransactionUploadPreviewFixtures() returned the same upload pointer twice")
	}

	if first.preview == second.preview {
		t.Fatal("newTransactionUploadPreviewFixtures() returned the same preview pointer twice")
	}
}

type transactionUploadPreviewFixtures struct {
	uploadID valueobjects.UploadID
	groupID  valueobjects.GroupID
	upload   *entities.TransactionUpload
	preview  *outboundports.TransactionUploadPreviewRecord
}

func newTransactionUploadPreviewFixtures(t *testing.T) transactionUploadPreviewFixtures {
	t.Helper()

	uploadID := mustPreviewUploadID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")

	upload, err := entities.ReconstituteTransactionUpload(
		uploadID,
		groupID,
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"upload/file.csv",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		1,
		testTime(),
	)
	if err != nil {
		t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
	}

	return transactionUploadPreviewFixtures{
		uploadID: uploadID,
		groupID:  groupID,
		upload:   upload,
		preview: &outboundports.TransactionUploadPreviewRecord{
			UploadID:  uploadID.String(),
			Columns:   []string{"Product", "Year"},
			Rows:      [][]string{{"CG", "2026"}},
			TotalRows: 1,
			ValidationErrors: []outboundports.TransactionFileValidationError{{
				Code:        "MISSING_FIELD",
				Message:     "Year is required",
				RowNumber:   1,
				ColumnName:  "Year",
				ColumnIndex: 2,
			}},
		},
	}
}

func mustPreviewUploadID(t *testing.T, value string) valueobjects.UploadID {
	t.Helper()

	uploadID, err := valueobjects.UploadIDFromString(value)
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	return uploadID
}
