package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
)

const (
	transactionUploadActorGroupID = "01962b8f-aeb2-7e03-a8ff-1edce1300003"
	transactionUploadID1          = "01962b8f-aeb2-7e03-a8ff-1edce1300201"
	transactionUploadID2          = "01962b8f-aeb2-7e03-a8ff-1edce1300202"
)

type createTransactionUploadPortStub struct {
	result  inboundports.CreateTransactionUploadResult
	err     error
	command inboundports.CreateTransactionUploadCommand
}

func (s *createTransactionUploadPortStub) Execute(_ context.Context, command inboundports.CreateTransactionUploadCommand) (inboundports.CreateTransactionUploadResult, error) {
	s.command = command
	return s.result, s.err
}

type transactionUploadStreamExecutorStub struct {
	result  inboundports.CreateTransactionUploadResult
	err     error
	command inboundports.CreateTransactionUploadCommand
	updates []ports.TransactionUploadProgressUpdate
}

func (s *transactionUploadStreamExecutorStub) Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand, reporter ports.TransactionUploadProgressReporter) (inboundports.CreateTransactionUploadResult, error) {
	s.command = command
	for _, update := range s.updates {
		if err := reporter.Report(ctx, update); err != nil {
			return inboundports.CreateTransactionUploadResult{}, err
		}
	}

	return s.result, s.err
}

type getTransactionUploadPortStub struct {
	result ports.TransactionUploadDetailsResult
	err    error
	query  inboundports.GetTransactionUploadQuery
}

func (s *getTransactionUploadPortStub) Execute(_ context.Context, query inboundports.GetTransactionUploadQuery) (ports.TransactionUploadDetailsResult, error) {
	s.query = query
	return s.result, s.err
}

type getTransactionUploadPreviewPortStub struct {
	result inboundports.GetTransactionUploadPreviewResult
	err    error
	query  inboundports.GetTransactionUploadPreviewQuery
}

func (s *getTransactionUploadPreviewPortStub) Execute(_ context.Context, query inboundports.GetTransactionUploadPreviewQuery) (inboundports.GetTransactionUploadPreviewResult, error) {
	s.query = query
	return s.result, s.err
}

type downloadTransactionUploadPortStub struct {
	result inboundports.DownloadTransactionUploadResult
	err    error
	query  inboundports.DownloadTransactionUploadQuery
}

func (s *downloadTransactionUploadPortStub) Execute(_ context.Context, query inboundports.DownloadTransactionUploadQuery) (inboundports.DownloadTransactionUploadResult, error) {
	s.query = query
	return s.result, s.err
}

type listTransactionUploadsPortStub struct {
	result inboundports.ListTransactionUploadsResult
	err    error
	query  inboundports.ListTransactionUploadsQuery
}

func (s *listTransactionUploadsPortStub) Execute(_ context.Context, query inboundports.ListTransactionUploadsQuery) (inboundports.ListTransactionUploadsResult, error) {
	s.query = query
	return s.result, s.err
}

type deleteTransactionUploadPortStub struct {
	err     error
	command inboundports.DeleteTransactionUploadCommand
}

func (s *deleteTransactionUploadPortStub) Execute(_ context.Context, command inboundports.DeleteTransactionUploadCommand) (inboundports.DeleteTransactionUploadResult, error) {
	s.command = command
	return inboundports.DeleteTransactionUploadResult{ID: command.ID}, s.err
}

type retryTransactionUploadClassificationPortStub struct {
	result  inboundports.RetryTransactionUploadClassificationResult
	err     error
	command inboundports.RetryTransactionUploadClassificationCommand
}

func (s *retryTransactionUploadClassificationPortStub) Execute(_ context.Context, command inboundports.RetryTransactionUploadClassificationCommand) (inboundports.RetryTransactionUploadClassificationResult, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpTransactionUploadAdapterRoutes verifies the HTTP transaction upload adapter routes behavior and the expected outcome asserted below.
func TestHttpTransactionUploadAdapterRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		method             string
		target             string
		withFile           bool
		classificationTask string
		createStub         *createTransactionUploadPortStub
		streamStub         *transactionUploadStreamExecutorStub
		getStub            *getTransactionUploadPortStub
		previewStub        *getTransactionUploadPreviewPortStub
		listStub           *listTransactionUploadsPortStub
		deleteStub         *deleteTransactionUploadPortStub
		retryStub          *retryTransactionUploadClassificationPortStub
		downloadStub       *downloadTransactionUploadPortStub
		wantStatus         int
		wantErrorCode      string
		assert             func(t *testing.T, response *httptest.ResponseRecorder, createStub *createTransactionUploadPortStub, streamStub *transactionUploadStreamExecutorStub, getStub *getTransactionUploadPortStub, previewStub *getTransactionUploadPreviewPortStub, listStub *listTransactionUploadsPortStub, deleteStub *deleteTransactionUploadPortStub, retryStub *retryTransactionUploadClassificationPortStub, downloadStub *downloadTransactionUploadPortStub, payload map[string]any)
	}{
		{
			name:        "creates upload",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads",
			withFile:    true,
			createStub:  &createTransactionUploadPortStub{result: inboundports.CreateTransactionUploadResult{Upload: ports.TransactionUploadResult{ID: transactionUploadID1, FileName: "transactions.csv", FileFormat: "csv", Status: "uploaded", RowCount: 1}}},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusCreated,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, createStub *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()
				if createStub.command.FileName != "transactions.csv" {
					t.Fatalf("createStub.command.FileName = %q, want %q", createStub.command.FileName, "transactions.csv")
				}
				if createStub.command.ClassificationTask != "" {
					t.Fatalf("createStub.command.ClassificationTask = %q, want empty for default react path", createStub.command.ClassificationTask)
				}
				dataPayload := payload["data"].(map[string]any)
				uploadPayload := dataPayload["upload"].(map[string]any)
				if uploadPayload["id"] != transactionUploadID1 {
					t.Fatalf("uploadPayload[id] = %v, want %q", uploadPayload["id"], transactionUploadID1)
				}
				if uploadPayload["status"] != "uploaded" {
					t.Fatalf("uploadPayload[status] = %v, want %q", uploadPayload["status"], "uploaded")
				}
			},
		},
		{
			name:     "creates upload with skipped rows",
			method:   http.MethodPost,
			target:   "/api/v1/transaction-uploads",
			withFile: true,
			createStub: &createTransactionUploadPortStub{result: inboundports.CreateTransactionUploadResult{
				Upload: ports.TransactionUploadResult{ID: transactionUploadID1, FileName: "transactions.csv", FileFormat: "csv", Status: "uploaded", RowCount: 1},
				SkippedRows: []ports.TransactionUploadSkippedRow{{
					RowNumber: 3,
					Reason:    ports.TransactionUploadSkippedRowReasonMalformed,
				}},
			}},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusCreated,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				uploadPayload := dataPayload["upload"].(map[string]any)
				if uploadPayload["status"] != "uploaded" {
					t.Fatalf("uploadPayload[status] = %v, want %q", uploadPayload["status"], "uploaded")
				}
				skippedRows := dataPayload["skipped_rows"].([]any)
				if len(skippedRows) != 1 {
					t.Fatalf("len(skippedRows) = %d, want %d", len(skippedRows), 1)
				}

				skippedEntry := skippedRows[0].(map[string]any)
				if skippedEntry["row_number"] != float64(3) {
					t.Fatalf("skippedEntry[row_number] = %v, want %v", skippedEntry["row_number"], float64(3))
				}
				if skippedEntry["reason"] != ports.TransactionUploadSkippedRowReasonMalformed {
					t.Fatalf("skippedEntry[reason] = %v, want %q", skippedEntry["reason"], ports.TransactionUploadSkippedRowReasonMalformed)
				}
			},
		},
		{
			name:        "returns normalized invalid format error for failed upload",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads",
			withFile:    true,
			createStub:  &createTransactionUploadPortStub{result: inboundports.CreateTransactionUploadResult{Upload: ports.TransactionUploadResult{ID: transactionUploadID1, Status: "failed"}, ValidationErrors: []ports.TransactionFileValidationError{{Code: "type_mismatch", RowNumber: 2, ColumnName: "Year"}}}},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "INVALID_FORMAT" {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], "INVALID_FORMAT")
				}
				if errorPayload["message"] != "One or more values do not match the expected format." {
					t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "One or more values do not match the expected format.")
				}
				if _, ok := payload["data"]; ok {
					t.Fatalf("payload[data] present = %v, want omitted", payload["data"])
				}
			},
		},
		{
			name:        "returns normalized missing field error for failed upload",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads",
			withFile:    true,
			createStub:  &createTransactionUploadPortStub{result: inboundports.CreateTransactionUploadResult{Upload: ports.TransactionUploadResult{ID: transactionUploadID1, FileName: "transactions.csv", FileFormat: "csv", Status: "failed", RowCount: 0}, ValidationErrors: []ports.TransactionFileValidationError{{Code: "missing_required_value", RowNumber: 2, ColumnName: "Year"}}}},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "MISSING_FIELD" {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], "MISSING_FIELD")
				}
				if errorPayload["message"] != "Required field is empty or null." {
					t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "Required field is empty or null.")
				}
				if _, ok := payload["data"]; ok {
					t.Fatalf("payload[data] present = %v, want omitted", payload["data"])
				}
			},
		},
		{
			name:          "returns normalized combined missing field and invalid reference error for failed upload",
			method:        http.MethodPost,
			target:        "/api/v1/transaction-uploads",
			withFile:      true,
			createStub:    &createTransactionUploadPortStub{result: inboundports.CreateTransactionUploadResult{Upload: ports.TransactionUploadResult{ID: transactionUploadID1, FileName: "transactions.csv", FileFormat: "csv", Status: "failed", RowCount: 0}, ValidationErrors: []ports.TransactionFileValidationError{{Code: "missing_required_value", RowNumber: 2, ColumnName: "Year"}, {Code: "invalid_value", RowNumber: 3, ColumnName: "Applicant"}}}},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: "MISSING_FIELD_AND_INVALID_REFERENCE",
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				errorPayload := payload["error"].(map[string]any)
				if errorPayload["message"] != "The upload has missing required fields and values that do not match the allowed reference values." {
					t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "The upload has missing required fields and values that do not match the allowed reference values.")
				}
				if _, ok := payload["data"]; ok {
					t.Fatalf("payload[data] present = %v, want omitted", payload["data"])
				}
			},
		},
		{
			name:       "downloads upload",
			method:     http.MethodGet,
			target:     "/api/v1/transaction-uploads/" + transactionUploadID1 + "/download",
			createStub: &createTransactionUploadPortStub{},
			streamStub: &transactionUploadStreamExecutorStub{},
			getStub:    &getTransactionUploadPortStub{},
			listStub:   &listTransactionUploadsPortStub{},
			deleteStub: &deleteTransactionUploadPortStub{},
			retryStub:  &retryTransactionUploadClassificationPortStub{},
			downloadStub: &downloadTransactionUploadPortStub{result: inboundports.DownloadTransactionUploadResult{
				FileName:      "transactions.csv",
				ContentType:   "text/csv",
				ContentLength: len([]byte("a,b\n1,2\n")),
				FileBytes:     []byte("a,b\n1,2\n"),
			}},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, response *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, downloadStub *downloadTransactionUploadPortStub, _ map[string]any) {
				t.Helper()
				if downloadStub.query.ID != transactionUploadID1 {
					t.Fatalf("downloadStub.query.ID = %q, want %q", downloadStub.query.ID, transactionUploadID1)
				}
				if downloadStub.query.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("downloadStub.query.ActorUserID = %q, want %q", downloadStub.query.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if downloadStub.query.ActorGroupID != transactionUploadActorGroupID {
					t.Fatalf("downloadStub.query.ActorGroupID = %q, want %q", downloadStub.query.ActorGroupID, transactionUploadActorGroupID)
				}
				if got := response.Header().Get("Content-Disposition"); !strings.Contains(got, `attachment; filename="transactions.csv"`) {
					t.Fatalf("Content-Disposition = %q, want attachment filename", got)
				}
				if got := response.Header().Get("Content-Type"); !strings.Contains(got, "text/csv") {
					t.Fatalf("Content-Type = %q, want to contain %q", got, "text/csv")
				}
				if got := response.Header().Get("Content-Length"); got != strconv.Itoa(len([]byte("a,b\n1,2\n"))) {
					t.Fatalf("Content-Length = %q, want %q", got, strconv.Itoa(len([]byte("a,b\n1,2\n"))))
				}
				if got := response.Body.Bytes(); !bytes.Equal(got, []byte("a,b\n1,2\n")) {
					t.Fatalf("response body = %q, want %q", got, []byte("a,b\n1,2\n"))
				}
			},
		},
		{
			name:        "gets upload",
			method:      http.MethodGet,
			target:      "/api/v1/transaction-uploads/" + transactionUploadID1,
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{result: ports.TransactionUploadDetailsResult{TransactionUploadResult: ports.TransactionUploadResult{ID: transactionUploadID1, Status: "uploaded"}, Transactions: []ports.TransactionResult{{RowNumber: 2, Product: "CG", ReferenceNumber: "REF-1", GoodsDescription: "Goods", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}}}},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, getStub *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				if getStub.query.ID != transactionUploadID1 {
					t.Fatalf("getStub.query.ID = %q, want %q", getStub.query.ID, transactionUploadID1)
				}
				if getStub.query.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("getStub.query.ActorUserID = %q, want %q", getStub.query.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["status"] != "uploaded" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "uploaded")
				}
				transactionsPayload := dataPayload["transactions"].([]any)
				transactionPayload := transactionsPayload[0].(map[string]any)
				transactionData := transactionPayload["transaction_data"].(map[string]any)
				if transactionPayload["exit_classification"] != "" {
					t.Fatalf("transactionPayload[exit_classification] = %v, want empty string", transactionPayload["exit_classification"])
				}
				if transactionPayload["status"] != "" {
					t.Fatalf("transactionPayload[status] = %v, want empty string", transactionPayload["status"])
				}
				if transactionData["product"] != "CG" {
					t.Fatalf("transactionData[product] = %v, want %q", transactionData["product"], "CG")
				}
				if transactionData["reference_number"] != "REF-1" {
					t.Fatalf("transactionData[reference_number] = %v, want %q", transactionData["reference_number"], "REF-1")
				}
				if transactionData["goods_description"] != "Goods" {
					t.Fatalf("transactionData[goods_description] = %v, want %q", transactionData["goods_description"], "Goods")
				}
				if transactionData["created_by"] != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("transactionData[created_by] = %v, want %q", transactionData["created_by"], "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if _, exists := transactionData["classification"]; exists {
					t.Fatalf("transactionData[classification] present = true, want false")
				}
				if _, exists := transactionData["status"]; exists {
					t.Fatalf("transactionData[status] present = true, want false")
				}
				if _, exists := transactionPayload["pipeline_result"]; exists {
					t.Fatalf("transactionPayload[pipeline_result] present = true, want false")
				}
				if _, exists := transactionPayload["transaction_id"]; exists {
					t.Fatalf("transactionPayload[transaction_id] present = true, want false")
				}
			},
		},
		{
			name:       "gets upload preview",
			method:     http.MethodGet,
			target:     "/api/v1/transaction-uploads/" + transactionUploadID1 + "/preview",
			createStub: &createTransactionUploadPortStub{},
			streamStub: &transactionUploadStreamExecutorStub{},
			getStub:    &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{result: inboundports.GetTransactionUploadPreviewResult{
				FileID:    transactionUploadID1,
				FileName:  "transactions.csv",
				Columns:   []string{"Product", "Year"},
				Rows:      [][]string{{"CG", "2026"}},
				TotalRows: 1,
				ValidationErrors: []ports.TransactionFileValidationError{
					{
						Code:        "missing_required_value",
						Message:     "Year is required",
						RowNumber:   1,
						ColumnName:  "Year",
						ColumnIndex: 2,
					},
					{
						Code:        "type_mismatch",
						Message:     "Year must be an integer",
						RowNumber:   2,
						ColumnName:  "Year",
						ColumnIndex: 2,
						Value:       "20XX",
					},
					{
						Code:        "invalid_value",
						Message:     "PA Alignment must be one of PA Aligned, Not PA Aligned",
						RowNumber:   3,
						ColumnName:  "PA Alignment",
						ColumnIndex: 18,
						Value:       "Unknown",
					},
				},
			}},
			listStub:   &listTransactionUploadsPortStub{},
			deleteStub: &deleteTransactionUploadPortStub{},
			retryStub:  &retryTransactionUploadClassificationPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, previewStub *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				if previewStub.query.ID != transactionUploadID1 {
					t.Fatalf("previewStub.query.ID = %q, want %q", previewStub.query.ID, transactionUploadID1)
				}
				if previewStub.query.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("previewStub.query.ActorUserID = %q, want %q", previewStub.query.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if previewStub.query.ActorGroupID != transactionUploadActorGroupID {
					t.Fatalf("previewStub.query.ActorGroupID = %q, want %q", previewStub.query.ActorGroupID, transactionUploadActorGroupID)
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["file_id"] != transactionUploadID1 {
					t.Fatalf("dataPayload[file_id] = %v, want %q", dataPayload["file_id"], transactionUploadID1)
				}
				if dataPayload["file_name"] != "transactions.csv" {
					t.Fatalf("dataPayload[file_name] = %v, want %q", dataPayload["file_name"], "transactions.csv")
				}
				columns := dataPayload["columns"].([]any)
				if len(columns) != 2 {
					t.Fatalf("len(columns) = %d, want %d", len(columns), 2)
				}
				rows := dataPayload["rows"].([]any)
				if len(rows) != 1 {
					t.Fatalf("len(rows) = %d, want %d", len(rows), 1)
				}
				firstRow := rows[0].(map[string]any)
				if len(firstRow) != 2 {
					t.Fatalf("len(firstRow) = %d, want %d", len(firstRow), 2)
				}
				if firstRow["Product"] != "CG" {
					t.Fatalf("firstRow[Product] = %v, want %q", firstRow["Product"], "CG")
				}
				if firstRow["Year"] != "2026" {
					t.Fatalf("firstRow[Year] = %v, want %q", firstRow["Year"], "2026")
				}
				if dataPayload["total_rows"] != float64(1) {
					t.Fatalf("dataPayload[total_rows] = %v, want %v", dataPayload["total_rows"], float64(1))
				}
				validationErrors := dataPayload["validation_errors"].([]any)
				if len(validationErrors) != 3 {
					t.Fatalf("len(validationErrors) = %d, want %d", len(validationErrors), 3)
				}
				firstValidationError := validationErrors[0].(map[string]any)
				if firstValidationError["Code"] != "MISSING_FIELD" {
					t.Fatalf("firstValidationError[Code] = %v, want %q", firstValidationError["Code"], "MISSING_FIELD")
				}
				if firstValidationError["Message"] != "Year is required" {
					t.Fatalf("firstValidationError[Message] = %v, want %q", firstValidationError["Message"], "Year is required")
				}
				if firstValidationError["RowNumber"] != float64(1) {
					t.Fatalf("firstValidationError[RowNumber] = %v, want %v", firstValidationError["RowNumber"], float64(1))
				}
				if firstValidationError["ColumnName"] != "Year" {
					t.Fatalf("firstValidationError[ColumnName] = %v, want %q", firstValidationError["ColumnName"], "Year")
				}
				if firstValidationError["ColumnIndex"] != float64(2) {
					t.Fatalf("firstValidationError[ColumnIndex] = %v, want %v", firstValidationError["ColumnIndex"], float64(2))
				}
				if firstValidationError["Value"] != "" {
					t.Fatalf("firstValidationError[Value] = %v, want %q", firstValidationError["Value"], "")
				}

				secondValidationError := validationErrors[1].(map[string]any)
				if secondValidationError["Code"] != "INVALID_FORMAT" {
					t.Fatalf("secondValidationError[Code] = %v, want %q", secondValidationError["Code"], "INVALID_FORMAT")
				}
				if secondValidationError["Value"] != "20XX" {
					t.Fatalf("secondValidationError[Value] = %v, want %q", secondValidationError["Value"], "20XX")
				}

				thirdValidationError := validationErrors[2].(map[string]any)
				if thirdValidationError["Code"] != "invalid_value" {
					t.Fatalf("thirdValidationError[Code] = %v, want %q", thirdValidationError["Code"], "invalid_value")
				}
				if _, exists := dataPayload["group_id"]; exists {
					t.Fatalf("dataPayload[group_id] present = true, want false")
				}
			},
		},
		{
			name:        "lists uploads",
			method:      http.MethodGet,
			target:      "/api/v1/transaction-uploads?file_name=transactions",
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{result: inboundports.ListTransactionUploadsResult{Uploads: []ports.TransactionUploadDetailsResult{{TransactionUploadResult: ports.TransactionUploadResult{ID: transactionUploadID1, Status: "uploaded"}, Transactions: []ports.TransactionResult{{RowNumber: 2, Product: "CG", ReferenceNumber: "REF-1", GoodsDescription: "Goods", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}}}}}},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, listStub *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()
				if listStub.query.FileName != "transactions" {
					t.Fatalf("listStub.query.FileName = %q, want %q", listStub.query.FileName, "transactions")
				}
				if listStub.query.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("listStub.query.ActorUserID = %q, want %q", listStub.query.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				dataPayload := payload["data"].(map[string]any)
				uploadsPayload := dataPayload["uploads"].([]any)
				uploadPayload := uploadsPayload[0].(map[string]any)
				if uploadPayload["id"] != transactionUploadID1 {
					t.Fatalf("uploadPayload[id] = %v, want %q", uploadPayload["id"], transactionUploadID1)
				}
				if uploadPayload["status"] != "uploaded" {
					t.Fatalf("uploadPayload[status] = %v, want %q", uploadPayload["status"], "uploaded")
				}
				transactionsPayload := uploadPayload["transactions"].([]any)
				transactionPayload := transactionsPayload[0].(map[string]any)
				transactionData := transactionPayload["transaction_data"].(map[string]any)
				if transactionPayload["exit_classification"] != "" {
					t.Fatalf("transactionPayload[exit_classification] = %v, want empty string", transactionPayload["exit_classification"])
				}
				if transactionPayload["status"] != "" {
					t.Fatalf("transactionPayload[status] = %v, want empty string", transactionPayload["status"])
				}
				if transactionData["product"] != "CG" {
					t.Fatalf("transactionData[product] = %v, want %q", transactionData["product"], "CG")
				}
				if transactionData["reference_number"] != "REF-1" {
					t.Fatalf("transactionData[reference_number] = %v, want %q", transactionData["reference_number"], "REF-1")
				}
				if transactionData["goods_description"] != "Goods" {
					t.Fatalf("transactionData[goods_description] = %v, want %q", transactionData["goods_description"], "Goods")
				}
				if transactionData["created_by"] != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("transactionData[created_by] = %v, want %q", transactionData["created_by"], "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if _, exists := transactionData["classification"]; exists {
					t.Fatalf("transactionData[classification] present = true, want false")
				}
				if _, exists := transactionData["status"]; exists {
					t.Fatalf("transactionData[status] present = true, want false")
				}
				if _, exists := transactionPayload["pipeline_result"]; exists {
					t.Fatalf("transactionPayload[pipeline_result] present = true, want false")
				}
				if _, exists := transactionPayload["transaction_id"]; exists {
					t.Fatalf("transactionPayload[transaction_id] present = true, want false")
				}
			},
		},
		{
			name:        "deletes upload",
			method:      http.MethodDelete,
			target:      "/api/v1/transaction-uploads/" + transactionUploadID1,
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusNoContent,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, deleteStub *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, _ map[string]any) {
				t.Helper()
				if deleteStub.command.ID != transactionUploadID1 {
					t.Fatalf("deleteStub.command.ID = %q, want %q", deleteStub.command.ID, transactionUploadID1)
				}
			},
		},
		{
			name:        "retries upload classification",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads/" + transactionUploadID1 + "/retry-classification",
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub: &retryTransactionUploadClassificationPortStub{result: inboundports.RetryTransactionUploadClassificationResult{
				UploadID:                   transactionUploadID1,
				EligibleFailedTransactions: 3,
				RetriedTransactions:        2,
				SkippedTransactions:        1,
				Skipped: []inboundports.RetryTransactionUploadClassificationSkippedTransaction{{
					TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62",
					Reason:        "already_queued_or_not_failed",
				}},
				FailedRetryCreations: 0,
			}},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, retryStub *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()

				if retryStub.command.UploadID != transactionUploadID1 {
					t.Fatalf("retryStub.command.UploadID = %q, want %q", retryStub.command.UploadID, transactionUploadID1)
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["upload_id"] != transactionUploadID1 {
					t.Fatalf("dataPayload[upload_id] = %v, want %q", dataPayload["upload_id"], transactionUploadID1)
				}
				if dataPayload["eligible_failed_transactions"] != float64(3) {
					t.Fatalf("dataPayload[eligible_failed_transactions] = %v, want %v", dataPayload["eligible_failed_transactions"], float64(3))
				}
				if dataPayload["retried_transactions"] != float64(2) {
					t.Fatalf("dataPayload[retried_transactions] = %v, want %v", dataPayload["retried_transactions"], float64(2))
				}
				skippedPayload := dataPayload["skipped"].([]any)
				if len(skippedPayload) != 1 {
					t.Fatalf("len(skippedPayload) = %d, want %d", len(skippedPayload), 1)
				}
				skippedEntry := skippedPayload[0].(map[string]any)
				if skippedEntry["reason"] != "already_queued_or_not_failed" {
					t.Fatalf("skippedEntry[reason] = %v, want %q", skippedEntry["reason"], "already_queued_or_not_failed")
				}
			},
		},
		{
			name:     "maps duplicate upload error",
			method:   http.MethodPost,
			target:   "/api/v1/transaction-uploads",
			withFile: true,
			createStub: &createTransactionUploadPortStub{err: &usecases.ConflictError{
				Resource: "transaction upload",
				Reason:   domain.ErrDuplicateUpload.Message,
			}},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			wantStatus:    http.StatusConflict,
			wantErrorCode: domain.ErrDuplicateUpload.Code,
		},
		{
			name:          "maps forbidden download",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1 + "/download",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			downloadStub:  &downloadTransactionUploadPortStub{err: &usecases.ForbiddenError{Resource: "transaction upload", Reason: "actor group does not have access to this upload"}},
			wantStatus:    http.StatusForbidden,
			wantErrorCode: "forbidden",
		},
		{
			name:          "maps not found get",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1,
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{err: &usecases.NotFoundError{Resource: "transaction upload", ID: transactionUploadID1}},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "maps not found get preview",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1 + "/preview",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{err: &usecases.NotFoundError{Resource: "transaction upload", ID: transactionUploadID1}},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			downloadStub:  &downloadTransactionUploadPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "NOT_FOUND",
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["message"] != "Upload not found" {
					t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "Upload not found")
				}
			},
		},
		{
			name:          "maps forbidden get preview",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1 + "/preview",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{err: &usecases.ForbiddenError{Resource: "transaction upload", Reason: "upload belongs to a different group"}},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			downloadStub:  &downloadTransactionUploadPortStub{},
			wantStatus:    http.StatusForbidden,
			wantErrorCode: "forbidden",
		},
		{
			name:          "maps not found download",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1 + "/download",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			downloadStub:  &downloadTransactionUploadPortStub{err: &usecases.NotFoundError{Resource: "transaction upload", ID: transactionUploadID1}},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "maps malformed download upload id",
			method:        http.MethodGet,
			target:        "/api/v1/transaction-uploads/invalid/download",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			downloadStub:  &downloadTransactionUploadPortStub{err: domain.ErrInvalidUploadID},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
		{
			name:          "maps not found delete",
			method:        http.MethodDelete,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1,
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{err: &usecases.NotFoundError{Resource: "transaction upload", ID: transactionUploadID1}},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "maps conflict delete while processing",
			method:        http.MethodDelete,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1,
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{err: &usecases.ConflictError{Resource: "transaction upload", Reason: "cannot delete transaction upload while transactions are still processing"}},
			retryStub:     &retryTransactionUploadClassificationPortStub{},
			wantStatus:    http.StatusConflict,
			wantErrorCode: "UPLOAD_STILL_PROCESSING",
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				wantMessage := "cannot delete transaction upload: upload is still processing"
				if errorPayload["message"] != wantMessage {
					t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], wantMessage)
				}
			},
		},
		{
			name:          "maps not found retry classification",
			method:        http.MethodPost,
			target:        "/api/v1/transaction-uploads/" + transactionUploadID1 + "/retry-classification",
			createStub:    &createTransactionUploadPortStub{},
			streamStub:    &transactionUploadStreamExecutorStub{},
			getStub:       &getTransactionUploadPortStub{},
			previewStub:   &getTransactionUploadPreviewPortStub{},
			listStub:      &listTransactionUploadsPortStub{},
			deleteStub:    &deleteTransactionUploadPortStub{},
			retryStub:     &retryTransactionUploadClassificationPortStub{err: &usecases.NotFoundError{Resource: "transaction upload", ID: transactionUploadID1}},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:        "streams upload progress",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads/stream",
			withFile:    true,
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{updates: []ports.TransactionUploadProgressUpdate{{Status: "received", Message: "upload received", Progress: 5}, {Status: "completed", Message: "upload completed", Progress: 100, Upload: &ports.TransactionUploadResult{ID: transactionUploadID1, Status: "uploaded"}}}},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, streamStub *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
				t.Helper()
				if streamStub.command.FileName != "transactions.csv" {
					t.Fatalf("streamStub.command.FileName = %q, want %q", streamStub.command.FileName, "transactions.csv")
				}
				if streamStub.command.ClassificationTask != "" {
					t.Fatalf("streamStub.command.ClassificationTask = %q, want empty for default react path", streamStub.command.ClassificationTask)
				}
				if streamStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("streamStub.command.ActorUserID = %q, want %q", streamStub.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
			},
		},
		{
			name:        "streams upload progress with skipped rows",
			method:      http.MethodPost,
			target:      "/api/v1/transaction-uploads/stream",
			withFile:    true,
			createStub:  &createTransactionUploadPortStub{},
			streamStub:  &transactionUploadStreamExecutorStub{updates: []ports.TransactionUploadProgressUpdate{{Status: "received", Message: "upload received", Progress: 5}, {Status: "completed", Message: "upload completed", Progress: 100, Upload: &ports.TransactionUploadResult{ID: transactionUploadID1, Status: "uploaded"}, SkippedRows: []ports.TransactionUploadSkippedRow{{RowNumber: 3, Reason: ports.TransactionUploadSkippedRowReasonMalformed}}}}},
			getStub:     &getTransactionUploadPortStub{},
			previewStub: &getTransactionUploadPreviewPortStub{},
			listStub:    &listTransactionUploadsPortStub{},
			deleteStub:  &deleteTransactionUploadPortStub{},
			retryStub:   &retryTransactionUploadClassificationPortStub{},
			wantStatus:  http.StatusOK,
		},
		{
			name:               "streams upload progress with classification task",
			method:             http.MethodPost,
			target:             "/api/v1/transaction-uploads/stream",
			withFile:           true,
			classificationTask: "dual_use_review",
			createStub:         &createTransactionUploadPortStub{},
			streamStub:         &transactionUploadStreamExecutorStub{updates: []ports.TransactionUploadProgressUpdate{{Status: "received", Message: "upload received", Progress: 5}, {Status: "completed", Message: "upload completed", Progress: 100, Upload: &ports.TransactionUploadResult{ID: transactionUploadID2, Status: "uploaded"}}}},
			getStub:            &getTransactionUploadPortStub{},
			previewStub:        &getTransactionUploadPreviewPortStub{},
			listStub:           &listTransactionUploadsPortStub{},
			deleteStub:         &deleteTransactionUploadPortStub{},
			retryStub:          &retryTransactionUploadClassificationPortStub{},
			wantStatus:         http.StatusOK,
			assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, streamStub *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, _ map[string]any) {
				t.Helper()
				if streamStub.command.ClassificationTask != "dual_use_review" {
					t.Fatalf("streamStub.command.ClassificationTask = %q, want %q", streamStub.command.ClassificationTask, "dual_use_review")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpTransactionUploadAdapter(tt.createStub, tt.streamStub, tt.getStub, tt.previewStub, tt.listStub, tt.deleteStub, tt.retryStub, tt.downloadStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := newTransactionUploadRequest(t, tt.method, tt.target, tt.withFile, tt.classificationTask)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusNoContent {
				if tt.assert != nil {
					tt.assert(t, response, tt.createStub, tt.streamStub, tt.getStub, tt.previewStub, tt.listStub, tt.deleteStub, tt.retryStub, tt.downloadStub, nil)
				}
				return
			}

			if tt.target == "/api/v1/transaction-uploads/"+transactionUploadID1+"/download" && tt.wantStatus == http.StatusOK {
				if tt.assert != nil {
					tt.assert(t, response, tt.createStub, tt.streamStub, tt.getStub, tt.previewStub, tt.listStub, tt.deleteStub, tt.retryStub, tt.downloadStub, nil)
				}
				return
			}

			if tt.target == "/api/v1/transaction-uploads/stream" {
				body := response.Body.String()
				if got := response.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
					t.Fatalf("Content-Type = %q, want %q", got, "text/event-stream; charset=utf-8")
				}
				if !bytes.Contains([]byte(body), []byte("event: progress")) {
					t.Fatalf("response body = %q, want progress SSE event", body)
				}
				if !bytes.Contains([]byte(body), []byte(`"status":"received"`)) {
					t.Fatalf("response body = %q, want received progress update", body)
				}
				if !bytes.Contains([]byte(body), []byte(`"status":"completed"`)) {
					t.Fatalf("response body = %q, want completed status update", body)
				}
				if !bytes.Contains([]byte(body), []byte("event: completed")) {
					t.Fatalf("response body = %q, want completed SSE event", body)
				}
				if !bytes.Contains([]byte(body), []byte(`"progress":100`)) {
					t.Fatalf("response body = %q, want completed progress update", body)
				}
				if !bytes.Contains([]byte(body), []byte(`"upload":{"id":"`)) {
					t.Fatalf("response body = %q, want upload payload in completed SSE event", body)
				}
				if !bytes.Contains([]byte(body), []byte(`"status":"uploaded"`)) {
					t.Fatalf("response body = %q, want upload status in SSE payload", body)
				}
				if tt.assert != nil {
					tt.assert(t, response, tt.createStub, tt.streamStub, tt.getStub, tt.previewStub, tt.listStub, tt.deleteStub, tt.retryStub, tt.downloadStub, nil)
				}
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if tt.wantErrorCode != "" {
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != tt.wantErrorCode {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], tt.wantErrorCode)
				}
			}

			if tt.assert != nil {
				tt.assert(t, response, tt.createStub, tt.streamStub, tt.getStub, tt.previewStub, tt.listStub, tt.deleteStub, tt.retryStub, tt.downloadStub, payload)
			}
		})
	}
}

func TestNormalizedTransactionUploadValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errors []ports.TransactionFileValidationError
		want   *errorResponse
	}{
		{
			name:   "returns generic validation failure for empty errors",
			errors: nil,
			want:   &errorResponse{Code: "UPLOAD_VALIDATION_FAILED", Message: "Upload validation failed."},
		},
		{
			name:   "returns accurate missing column message",
			errors: []ports.TransactionFileValidationError{{Code: "missing_required_column", ColumnName: "Year"}},
			want:   &errorResponse{Code: "MISSING_FIELD", Message: "Required column is missing."},
		},
		{
			name:   "returns missing field for missing required value",
			errors: []ports.TransactionFileValidationError{{Code: "missing_required_value", RowNumber: 2, ColumnName: "Year"}},
			want:   &errorResponse{Code: "MISSING_FIELD", Message: "Required field is empty or null."},
		},
		{
			name: "combines and deduplicates missing field and invalid format categories",
			errors: []ports.TransactionFileValidationError{
				{Code: "type_mismatch", RowNumber: 2, ColumnName: "Year"},
				{Code: "missing_required_value", RowNumber: 3, ColumnName: "Month"},
				{Code: "missing_required_value", RowNumber: 4, ColumnName: "Month"},
			},
			want: &errorResponse{
				Code:    "MISSING_FIELD_AND_INVALID_FORMAT",
				Message: "The upload has missing required fields and invalidly formatted or unsupported values.",
			},
		},
		{
			name: "maps allowed-value invalid value errors to invalid reference",
			errors: []ports.TransactionFileValidationError{{
				Code:       "invalid_value",
				Message:    "Applicant must be one of CG, RPA, or RCF",
				RowNumber:  2,
				ColumnName: "Applicant",
			}},
			want: &errorResponse{
				Code:    "INVALID_REFERENCE",
				Message: "One or more values do not match the allowed reference values.",
			},
		},
		{
			name: "maps invalid value to invalid reference without relying on message wording",
			errors: []ports.TransactionFileValidationError{{
				Code:       "invalid_value",
				Message:    "unexpected value provided",
				RowNumber:  5,
				ColumnName: "Applicant",
			}},
			want: &errorResponse{
				Code:    "INVALID_REFERENCE",
				Message: "One or more values do not match the allowed reference values.",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizedTransactionUploadValidationError(tt.errors)
			if got == nil {
				t.Fatal("normalizedTransactionUploadValidationError() = nil, want non-nil")
			}
			if got.Code != tt.want.Code {
				t.Fatalf("normalizedTransactionUploadValidationError().Code = %q, want %q", got.Code, tt.want.Code)
			}
			if got.Message != tt.want.Message {
				t.Fatalf("normalizedTransactionUploadValidationError().Message = %q, want %q", got.Message, tt.want.Message)
			}
		})
	}
}

func newTransactionUploadRequest(t *testing.T, method, target string, withFile bool, classificationTask string) *http.Request {
	t.Helper()

	if !withFile {
		request := httptest.NewRequest(method, target, nil)
		request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
		request.Header.Set(actorGroupIDHeader, transactionUploadActorGroupID)
		return request
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "transactions.csv")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}

	if _, err := part.Write([]byte("Product,Year,Month,DMC:IB,DMC,Partner Bank,Reference Number,Value of Transactions,No. of Transactions,Goods Description,Goods Classification (Sector),Applicant (CG/RPA) or Sub-Borrower (RCF) Country,Beneficiary Country,Source,Destination,Tenor > 1 year,E&S Category,PA Alignment\nCG,2026,4,IB,DMC,Partner Bank,REF-1,698436.80,1,Goods,Classification,Philippines,Japan,Thailand,Philippines,N,,PA Aligned\n")); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}

	if classificationTask != "" {
		if err := writer.WriteField("classification_task", classificationTask); err != nil {
			t.Fatalf("writer.WriteField() error = %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	request := httptest.NewRequest(method, target, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
	request.Header.Set(actorGroupIDHeader, transactionUploadActorGroupID)
	return request
}

func TestSSETransactionUploadProgressReporterIncludesSkippedRows(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/transaction-uploads/stream", nil)

	reporter := newSSETransactionUploadProgressReporter(context)
	err := reporter.Report(context.Request.Context(), ports.TransactionUploadProgressUpdate{
		Status:   ports.TransactionUploadProgressStatusCompleted,
		Message:  "upload completed",
		Progress: 100,
		Upload:   &ports.TransactionUploadResult{ID: transactionUploadID1, Status: "uploaded"},
		SkippedRows: []ports.TransactionUploadSkippedRow{{
			RowNumber: 3,
			Reason:    ports.TransactionUploadSkippedRowReasonMalformed,
		}},
	})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	body := response.Body.String()
	if !bytes.Contains([]byte(body), []byte("event: completed")) {
		t.Fatalf("response body = %q, want completed event", body)
	}
	if !bytes.Contains([]byte(body), []byte(`"skipped_rows":[{"row_number":3,"reason":"malformed"}]`)) {
		t.Fatalf("response body = %q, want skipped row warning", body)
	}
	if !bytes.Contains([]byte(body), []byte(`"upload":{"id":"01962b8f-aeb2-7e03-a8ff-1edce1300201","file_name":"","file_format":"","content_md5":"","storage_provider":"","storage_key":"","schema_version":"","status":"uploaded","row_count":0,"uploaded_at":""}`)) {
		t.Fatalf("response body = %q, want upload payload with status", body)
	}
}

func TestSSETransactionUploadProgressReporterIncludesNormalizedValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		validationErrors []ports.TransactionFileValidationError
		wantErrorPayload string
	}{
		{
			name: "with validation errors",
			validationErrors: []ports.TransactionFileValidationError{
				{Code: "missing_required_value", RowNumber: 2, ColumnName: "Month"},
				{Code: "type_mismatch", RowNumber: 3, ColumnName: "Year", Value: "20XX"},
			},
			wantErrorPayload: `"error":{"code":"MISSING_FIELD_AND_INVALID_FORMAT","message":"The upload has missing required fields and invalidly formatted or unsupported values."}`,
		},
		{
			name:             "without validation errors",
			validationErrors: nil,
			wantErrorPayload: `"error":{"code":"UPLOAD_VALIDATION_FAILED","message":"Upload validation failed."}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/transaction-uploads/stream", nil)

			reporter := newSSETransactionUploadProgressReporter(context)
			err := reporter.Report(context.Request.Context(), ports.TransactionUploadProgressUpdate{
				Status:           ports.TransactionUploadProgressStatusValidationFailed,
				Message:          "upload validation failed",
				Progress:         100,
				ValidationErrors: tt.validationErrors,
			})
			if err != nil {
				t.Fatalf("Report() error = %v", err)
			}

			body := response.Body.String()
			if !bytes.Contains([]byte(body), []byte("event: validation_failed")) {
				t.Fatalf("response body = %q, want validation_failed event", body)
			}
			if !bytes.Contains([]byte(body), []byte(tt.wantErrorPayload)) {
				t.Fatalf("response body = %q, want normalized error payload %s", body, tt.wantErrorPayload)
			}
			if bytes.Contains([]byte(body), []byte(`"validation_errors":`)) {
				t.Fatalf("response body = %q, want validation_errors omitted", body)
			}
		})
	}
}
