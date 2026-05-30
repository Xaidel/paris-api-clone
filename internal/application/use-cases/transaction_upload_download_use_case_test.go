package usecases

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDownloadTransactionUploadUseCaseExecute verifies the download transaction upload use case execute behavior and the expected outcome asserted below.
func TestDownloadTransactionUploadUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	otherGroupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
	actorValidationErr := errors.New("actor invalid")
	publishErr := errors.New("boom")

	tests := []struct {
		name        string
		uploadRepo  *transactionUploadRepositoryMock
		store       *rawFileStoreMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		query       inboundports.DownloadTransactionUploadQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DownloadTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, store *rawFileStoreMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock)
	}{
		{
			name:       "same-group download succeeds",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			store:      &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("a,b\n1,2\n"), ContentType: "text/custom"}},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, store *rawFileStoreMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock) {
				t.Helper()

				if actors.userID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("actors.userID = %q, want %q", actors.userID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if actors.groupID != groupID.String() {
					t.Fatalf("actors.groupID = %q, want %q", actors.groupID, groupID.String())
				}

				if store.readCommand.Key != "upload/file.csv" {
					t.Fatalf("store.readCommand.Key = %q, want %q", store.readCommand.Key, "upload/file.csv")
				}

				if result.FileName != "transactions.csv" {
					t.Fatalf("result.FileName = %q, want %q", result.FileName, "transactions.csv")
				}

				if result.ContentType != "text/custom" {
					t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "text/custom")
				}

				if result.ContentLength != len([]byte("a,b\n1,2\n")) {
					t.Fatalf("result.ContentLength = %d, want %d", result.ContentLength, len([]byte("a,b\n1,2\n")))
				}

				if string(result.FileBytes) != "a,b\n1,2\n" {
					t.Fatalf("string(result.FileBytes) = %q, want %q", string(result.FileBytes), "a,b\n1,2\n")
				}

				if recorder.command.EventType != domainevents.DownloadTransactionUploadEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, domainevents.DownloadTransactionUploadEventType)
				}
			},
		},
		{
			name:       "missing upload returns not found",
			uploadRepo: &transactionUploadRepositoryMock{},
			store:      &rawFileStoreMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:       "group mismatch returns forbidden",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			store:      &rawFileStoreMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: otherGroupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var forbiddenErr *ForbiddenError
				if !errors.As(err, &forbiddenErr) {
					t.Fatalf("expected ForbiddenError, got %v", err)
				}
			},
			assert: func(t *testing.T, _ inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, store *rawFileStoreMock, recorder *adminEventRecorderMock, _ *actorDirectoryMock) {
				t.Helper()

				if store.readCommand.Key != "" {
					t.Fatalf("store.readCommand.Key = %q, want empty", store.readCommand.Key)
				}

				if recorder.command.EventType != "" {
					t.Fatalf("recorder.command.EventType = %q, want empty", recorder.command.EventType)
				}
			},
		},
		{
			name:       "missing raw file returns not found",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			store:      &rawFileStoreMock{readErr: os.ErrNotExist},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}

				if notFoundErr.ID != uploadID.String() {
					t.Fatalf("notFoundErr.ID = %q, want %q", notFoundErr.ID, uploadID.String())
				}

				if err.Error() != "transaction upload file \""+uploadID.String()+"\" was not found" {
					t.Fatalf("err.Error() = %q, want %q", err.Error(), "transaction upload file \""+uploadID.String()+"\" was not found")
				}

				if !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("errors.Is(err, os.ErrNotExist) = false, want true (err = %v)", err)
				}
			},
		},
		{
			name:       "actor validation failure is propagated",
			uploadRepo: &transactionUploadRepositoryMock{},
			store:      &rawFileStoreMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{err: actorValidationErr},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil || !strings.Contains(err.Error(), "validating actor ids") {
					t.Fatalf("unexpected error = %v", err)
				}

				if !errors.Is(err, actorValidationErr) {
					t.Fatalf("errors.Is(err, actor invalid) = false, want true (err = %v)", err)
				}
			},
		},
		{
			name:       "download event publication failure is wrapped",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			store:      &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("content")}},
			recorder:   &adminEventRecorderMock{err: publishErr},
			actors:     &actorDirectoryMock{},
			query:      inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil || !strings.Contains(err.Error(), "publishing transaction upload events") {
					t.Fatalf("unexpected error = %v", err)
				}

				if !errors.Is(err, publishErr) {
					t.Fatalf("errors.Is(err, recorder.err) = false, want true (err = %v)", err)
				}
			},
		},
		{
			name: "filename extension fallback path is used before file format",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: func() *entities.TransactionUpload {
				upload, err := entities.ReconstituteTransactionUpload(uploadID, groupID, "transactions.csv", "unknown", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file.csv", "transaction-file-v1", uploadedTransactionUploadStatus, 1, testTime())
				if err != nil {
					t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
				}

				return upload
			}()},
			store:    &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("spreadsheet")}},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query:    inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *rawFileStoreMock, _ *adminEventRecorderMock, _ *actorDirectoryMock) {
				t.Helper()

				if result.ContentType != "text/csv; charset=utf-8" {
					t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "text/csv; charset=utf-8")
				}
			},
		},
		{
			name: "extensionless csv resolves from format when store content type is empty",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: func() *entities.TransactionUpload {
				upload, err := entities.ReconstituteTransactionUpload(uploadID, groupID, "transactions", "csv", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file", "transaction-file-v1", uploadedTransactionUploadStatus, 1, testTime())
				if err != nil {
					t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
				}

				return upload
			}()},
			store:    &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("col1,col2\n1,2\n")}},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query:    inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *rawFileStoreMock, _ *adminEventRecorderMock, _ *actorDirectoryMock) {
				t.Helper()

				if result.ContentType != "text/csv; charset=utf-8" {
					t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "text/csv; charset=utf-8")
				}
			},
		},
		{
			name: "extensionless known-format upload resolves content type from file format",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: func() *entities.TransactionUpload {
				upload, err := entities.ReconstituteTransactionUpload(uploadID, groupID, "transactions", "csv", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file", "transaction-file-v1", uploadedTransactionUploadStatus, 1, testTime())
				if err != nil {
					t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
				}

				return upload
			}()},
			store:    &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("mystery-bytes")}},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query:    inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *rawFileStoreMock, _ *adminEventRecorderMock, _ *actorDirectoryMock) {
				t.Helper()

				if result.ContentType != "text/csv; charset=utf-8" {
					t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "text/csv; charset=utf-8")
				}
			},
		},
		{
			name: "content type fallback to application octet-stream when no inference possible",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: func() *entities.TransactionUpload {
				upload, err := entities.ReconstituteTransactionUpload(uploadID, groupID, "transactions", "unknown", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file", "transaction-file-v1", uploadedTransactionUploadStatus, 1, testTime())
				if err != nil {
					t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
				}

				return upload
			}()},
			store:    &rawFileStoreMock{readResult: outboundports.ReadRawFileResult{FileBytes: []byte("mystery-bytes")}},
			recorder: &adminEventRecorderMock{},
			actors:   &actorDirectoryMock{},
			query:    inboundports.DownloadTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DownloadTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *rawFileStoreMock, _ *adminEventRecorderMock, _ *actorDirectoryMock) {
				t.Helper()

				if result.ContentType != "application/octet-stream" {
					t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "application/octet-stream")
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewDownloadTransactionUploadUseCase(tc.uploadRepo, tc.store, tc.recorder, tc.actors)
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				if tc.assert != nil {
					tc.assert(t, result, tc.uploadRepo, tc.store, tc.recorder, tc.actors)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.uploadRepo, tc.store, tc.recorder, tc.actors)
		})
	}
}

func TestResolveDownloadContentType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		fileName          string
		fileFormat        string
		storedContentType string
		want              string
	}{
		{name: "prefers stored content type", fileName: "transactions.csv", fileFormat: "csv", storedContentType: "text/custom", want: "text/custom"},
		{name: "uses deterministic known filename extension mapping", fileName: "transactions.csv", fileFormat: "unknown", want: "text/csv; charset=utf-8"},
		{name: "falls back to upload file format", fileName: "transactions", fileFormat: "csv", want: "text/csv; charset=utf-8"},
		{name: "unmapped extension and format fall back to application octet-stream", fileName: "transactions.weird", fileFormat: "unknown", want: "application/octet-stream"},
		{name: "falls back to application octet-stream", fileName: "transactions", fileFormat: "unknown", want: "application/octet-stream"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveDownloadContentType(tc.fileName, tc.fileFormat, tc.storedContentType); got != tc.want {
				t.Fatalf("resolveDownloadContentType(%q, %q, %q) = %q, want %q", tc.fileName, tc.fileFormat, tc.storedContentType, got, tc.want)
			}
		})
	}
}
