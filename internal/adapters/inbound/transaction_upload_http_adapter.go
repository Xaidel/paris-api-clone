package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// HttpTransactionUploadAdapter exposes transaction upload ingestion over HTTP.
type HttpTransactionUploadAdapter struct {
	createUpload              inboundports.CreateTransactionUploadPort
	streamUpload              transactionUploadStreamExecutor
	getUpload                 inboundports.GetTransactionUploadPort
	getUploadPreview          inboundports.GetTransactionUploadPreviewPort
	listUploads               inboundports.ListTransactionUploadsPort
	deleteUpload              inboundports.DeleteTransactionUploadPort
	retryUploadClassification inboundports.RetryTransactionUploadClassificationPort
	downloadUpload            inboundports.DownloadTransactionUploadPort
}

type transactionUploadStreamExecutor interface {
	Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand, reporter ports.TransactionUploadProgressReporter) (inboundports.CreateTransactionUploadResult, error)
}

type transactionUploadResponse struct {
	ID              string `json:"id"`
	FileName        string `json:"file_name"`
	FileFormat      string `json:"file_format"`
	ContentMD5      string `json:"content_md5"`
	StorageProvider string `json:"storage_provider"`
	StorageKey      string `json:"storage_key"`
	SchemaVersion   string `json:"schema_version"`
	Status          string `json:"status"`
	RowCount        int    `json:"row_count"`
	UploadedAt      string `json:"uploaded_at"`
}

type transactionUploadDetailsResponse struct {
	transactionUploadResponse
	Transactions []transactionResponse `json:"transactions"`
}

type transactionUploadPreviewResponse struct {
	FileID           string                                    `json:"file_id"`
	FileName         string                                    `json:"file_name"`
	Columns          []string                                  `json:"columns"`
	Rows             []map[string]string                       `json:"rows"`
	TotalRows        int                                       `json:"total_rows"`
	ValidationErrors []transactionUploadPreviewValidationError `json:"validation_errors"`
}

type createTransactionUploadResponse struct {
	Upload           transactionUploadResponse        `json:"upload"`
	ValidationErrors []transactionFileValidationError `json:"validation_errors,omitempty"`
	SkippedRows      []transactionUploadSkippedRow    `json:"skipped_rows,omitempty"`
}

type listTransactionUploadsResponse struct {
	Uploads []transactionUploadDetailsResponse `json:"uploads"`
}

type transactionFileValidationError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	RowNumber   int    `json:"row_number"`
	ColumnName  string `json:"column_name"`
	ColumnIndex int    `json:"column_index"`
	Value       string `json:"value"`
}

type transactionUploadPreviewValidationError struct {
	Code        string `json:"Code"`
	Message     string `json:"Message"`
	RowNumber   int    `json:"RowNumber"`
	ColumnName  string `json:"ColumnName"`
	ColumnIndex int    `json:"ColumnIndex"`
	Value       string `json:"Value"`
}

type transactionUploadSkippedRow struct {
	RowNumber int    `json:"row_number"`
	Reason    string `json:"reason"`
}

const (
	uploadStillProcessingErrorCode     = "UPLOAD_STILL_PROCESSING"
	transactionUploadConflictErrorCode = "TRANSACTION_UPLOAD_CONFLICT"
	uploadStillProcessingReason        = "cannot delete transaction upload while transactions are still processing"
	uploadStillProcessingMessage       = "cannot delete transaction upload: upload is still processing"
)

// NewHttpTransactionUploadAdapter builds an HttpTransactionUploadAdapter.
func NewHttpTransactionUploadAdapter(createUpload inboundports.CreateTransactionUploadPort, streamUpload transactionUploadStreamExecutor, getUpload inboundports.GetTransactionUploadPort, getUploadPreview inboundports.GetTransactionUploadPreviewPort, listUploads inboundports.ListTransactionUploadsPort, deleteUpload inboundports.DeleteTransactionUploadPort, retryUploadClassification inboundports.RetryTransactionUploadClassificationPort, downloadUpload inboundports.DownloadTransactionUploadPort) *HttpTransactionUploadAdapter {
	return &HttpTransactionUploadAdapter{createUpload: createUpload, streamUpload: streamUpload, getUpload: getUpload, getUploadPreview: getUploadPreview, listUploads: listUploads, deleteUpload: deleteUpload, retryUploadClassification: retryUploadClassification, downloadUpload: downloadUpload}
}

// RegisterRoutes attaches transaction upload handlers to the Gin engine.
func (a *HttpTransactionUploadAdapter) RegisterRoutes(r *gin.Engine) {
	uploads := r.Group("/api/v1/transaction-uploads")
	uploads.POST("", a.CreateUpload)
	uploads.POST("/stream", a.CreateUploadStream)
	uploads.POST("/:id/retry-classification", a.RetryClassification)
	uploads.GET("", a.ListUploads)
	uploads.GET("/:id", a.GetUpload)
	uploads.GET("/:id/preview", a.GetUploadPreview)
	uploads.GET("/:id/download", a.DownloadUpload)
	uploads.DELETE("/:id", a.DeleteUpload)
}

type retryTransactionUploadClassificationResponse struct {
	UploadID                   string                                                                `json:"upload_id"`
	EligibleFailedTransactions int                                                                   `json:"eligible_failed_transactions"`
	RetriedTransactions        int                                                                   `json:"retried_transactions"`
	SkippedTransactions        int                                                                   `json:"skipped_transactions"`
	Skipped                    []inboundports.RetryTransactionUploadClassificationSkippedTransaction `json:"skipped,omitempty"`
	FailedRetryCreations       int                                                                   `json:"failed_retry_creations"`
	Failures                   []inboundports.RetryTransactionUploadClassificationFailure            `json:"failures,omitempty"`
}

// CreateUpload handles POST /api/v1/transaction-uploads.
func (a *HttpTransactionUploadAdapter) CreateUpload(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: "file is required"}})
		return
	}

	fileBytes, err := readMultipartFile(fileHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createUpload.Execute(c.Request.Context(), inboundports.CreateTransactionUploadCommand{
		ClassificationTask: c.PostForm("classification_task"),
		FileName:           fileHeader.Filename,
		FileBytes:          fileBytes,
		ActorUserID:        c.GetHeader(actorUserIDHeader),
		ActorGroupID:       c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	if result.Upload.Status == "failed" {
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: normalizedTransactionUploadValidationError(result.ValidationErrors)})
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: toCreateTransactionUploadResponse(result)})
}

// CreateUploadStream handles POST /api/v1/transaction-uploads/stream.
func (a *HttpTransactionUploadAdapter) CreateUploadStream(c *gin.Context) {
	responseController := http.NewResponseController(c.Writer)
	_ = responseController.EnableFullDuplex()
	_ = responseController.SetWriteDeadline(time.Time{})

	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	c.Writer.WriteHeaderNow()
	responseController.Flush()

	reporter := newSSETransactionUploadProgressReporter(c)
	reportTransactionUploadProgress(c.Request.Context(), reporter, "accepted", "upload request accepted", 1, nil, nil, nil)

	request, err := readTransactionUploadStreamRequest(c, reporter)
	if err != nil {
		_ = reporter.Report(c.Request.Context(), ports.TransactionUploadProgressUpdate{Status: "failed", Message: err.Error(), Progress: 100})
		return
	}

	_, err = a.streamUpload.Execute(c.Request.Context(), inboundports.CreateTransactionUploadCommand{
		ClassificationTask: request.classificationTask,
		FileName:           request.fileName,
		FileBytes:          request.fileBytes,
		ActorUserID:        c.GetHeader(actorUserIDHeader),
		ActorGroupID:       c.GetHeader(actorGroupIDHeader),
	}, reporter)
	if err != nil {
		_ = reporter.Report(c.Request.Context(), ports.TransactionUploadProgressUpdate{Status: "failed", Message: err.Error(), Progress: 100})
	}
}

type transactionUploadStreamRequest struct {
	classificationTask string
	fileName           string
	fileBytes          []byte
}

// ListUploads handles GET /api/v1/transaction-uploads.
func (a *HttpTransactionUploadAdapter) ListUploads(c *gin.Context) {
	query, err := buildListTransactionUploadsHTTPQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.listUploads.Execute(c.Request.Context(), query)
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toListTransactionUploadsResponse(result)})
}

// GetUpload handles GET /api/v1/transaction-uploads/:id.
func (a *HttpTransactionUploadAdapter) GetUpload(c *gin.Context) {
	result, err := a.getUpload.Execute(c.Request.Context(), inboundports.GetTransactionUploadQuery{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toTransactionUploadDetailsResponse(result)})
}

// GetUploadPreview handles GET /api/v1/transaction-uploads/:id/preview.
func (a *HttpTransactionUploadAdapter) GetUploadPreview(c *gin.Context) {
	result, err := a.getUploadPreview.Execute(c.Request.Context(), inboundports.GetTransactionUploadPreviewQuery{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handlePreviewError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: transactionUploadPreviewResponse{
		FileID:           result.FileID,
		FileName:         result.FileName,
		Columns:          result.Columns,
		Rows:             toTransactionUploadPreviewRows(result.Columns, result.Rows),
		TotalRows:        result.TotalRows,
		ValidationErrors: toTransactionUploadPreviewValidationErrors(result.ValidationErrors),
	}})
}

// DownloadUpload handles GET /api/v1/transaction-uploads/:id/download.
func (a *HttpTransactionUploadAdapter) DownloadUpload(c *gin.Context) {
	result, err := a.downloadUpload.Execute(c.Request.Context(), inboundports.DownloadTransactionUploadQuery{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", result.FileName))
	c.Header("Content-Length", fmt.Sprintf("%d", result.ContentLength))
	c.Data(http.StatusOK, result.ContentType, result.FileBytes)
}

type sseTransactionUploadProgressReporter struct {
	context *gin.Context
}

// uploadProgressResponse is the SSE payload shape streamed back to upload
// clients while ingestion progresses.
type uploadProgressResponse struct {
	Status      string                        `json:"status"`
	Message     string                        `json:"message"`
	Progress    int                           `json:"progress"`
	Upload      *transactionUploadResponse    `json:"upload,omitempty"`
	Error       *errorResponse                `json:"error,omitempty"`
	SkippedRows []transactionUploadSkippedRow `json:"skipped_rows,omitempty"`
}

func newSSETransactionUploadProgressReporter(context *gin.Context) *sseTransactionUploadProgressReporter {
	return &sseTransactionUploadProgressReporter{context: context}
}

// Report sends one progress update as an SSE event.
func (r *sseTransactionUploadProgressReporter) Report(_ context.Context, update ports.TransactionUploadProgressUpdate) error {
	// Marshal once so the event body and final flush operate on a consistent
	// snapshot of the update payload.
	var normalizedError *errorResponse
	if update.Status == ports.TransactionUploadProgressStatusValidationFailed {
		normalizedError = normalizedTransactionUploadValidationError(update.ValidationErrors)
	}

	payload, err := json.Marshal(uploadProgressResponse{
		Status:      update.Status,
		Message:     update.Message,
		Progress:    update.Progress,
		Upload:      toOptionalTransactionUploadResponse(update.Upload),
		Error:       normalizedError,
		SkippedRows: toTransactionUploadSkippedRows(update.SkippedRows),
	})
	if err != nil {
		return fmt.Errorf("marshaling upload progress: %w", err)
	}

	if _, err := fmt.Fprintf(r.context.Writer, "event: %s\n", sseEventName(update.Status)); err != nil {
		return fmt.Errorf("writing upload progress event: %w", err)
	}

	if _, err := fmt.Fprintf(r.context.Writer, "data: %s\n\n", payload); err != nil {
		return fmt.Errorf("writing upload progress: %w", err)
	}

	r.context.Writer.Flush()
	return nil
}

func sseEventName(status string) string {
	// Keep the SSE event taxonomy intentionally small so browser clients can treat
	// unknown statuses as generic progress updates.
	switch status {
	case "completed":
		return "completed"
	case "validation_failed":
		return "validation_failed"
	case "failed":
		return "failed"
	default:
		return "progress"
	}
}

func readTransactionUploadStreamRequest(c *gin.Context, reporter ports.TransactionUploadProgressReporter) (transactionUploadStreamRequest, error) {
	multipartReader, err := c.Request.MultipartReader()
	if err != nil {
		return transactionUploadStreamRequest{}, fmt.Errorf("reading multipart upload: %w", err)
	}

	var (
		request                transactionUploadStreamRequest
		classificationTaskRead bool
		fileFound              bool
	)

	for {
		part, err := multipartReader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return transactionUploadStreamRequest{}, fmt.Errorf("reading multipart upload part: %w", err)
		}

		switch {
		case part.FormName() == "classification_task" && !classificationTaskRead:
			classificationTaskBytes, err := io.ReadAll(part)
			_ = part.Close()
			if err != nil {
				return transactionUploadStreamRequest{}, fmt.Errorf("reading multipart upload classification task: %w", err)
			}

			request.classificationTask = string(classificationTaskBytes)
			classificationTaskRead = true
		case part.FormName() == "file" && !fileFound:
			request.fileName = part.FileName()
			request.fileBytes, err = copyTransactionUploadStreamPart(c.Request.Context(), part, c.Request.ContentLength, reporter)
			_ = part.Close()
			if err != nil {
				return transactionUploadStreamRequest{}, err
			}

			fileFound = true
		default:
			if _, err := io.Copy(io.Discard, part); err != nil {
				_ = part.Close()
				return transactionUploadStreamRequest{}, fmt.Errorf("discarding multipart upload part: %w", err)
			}
			_ = part.Close()
		}
	}

	if !fileFound {
		return transactionUploadStreamRequest{}, errors.New("file is required")
	}

	return request, nil
}

func copyTransactionUploadStreamPart(ctx context.Context, reader io.Reader, contentLength int64, reporter ports.TransactionUploadProgressReporter) ([]byte, error) {
	reportTransactionUploadProgress(ctx, reporter, "uploading", "receiving upload file", 2, nil, nil, nil)

	var fileBuffer bytes.Buffer
	buffer := make([]byte, 32*1024)
	var totalRead int64
	lastProgress := 2

	for {
		bytesRead, err := reader.Read(buffer)
		if bytesRead > 0 {
			totalRead += int64(bytesRead)
			if _, writeErr := fileBuffer.Write(buffer[:bytesRead]); writeErr != nil {
				return nil, fmt.Errorf("buffering uploaded file: %w", writeErr)
			}

			progress := estimateUploadStreamProgress(totalRead, contentLength)
			if progress > lastProgress {
				reportTransactionUploadProgress(ctx, reporter, "uploading", "receiving upload file", progress, nil, nil, nil)
				lastProgress = progress
			}
		}

		if err == nil {
			continue
		}
		if err == io.EOF {
			break
		}

		return nil, fmt.Errorf("reading uploaded file: %w", err)
	}

	if lastProgress < 4 {
		reportTransactionUploadProgress(ctx, reporter, "uploading", "upload file received", 4, nil, nil, nil)
	}

	return fileBuffer.Bytes(), nil
}

func reportTransactionUploadProgress(ctx context.Context, reporter ports.TransactionUploadProgressReporter, status, message string, progress int, upload *ports.TransactionUploadResult, validationErrors []ports.TransactionFileValidationError, skippedRows []ports.TransactionUploadSkippedRow) {
	if reporter == nil {
		return
	}

	_ = reporter.Report(ctx, ports.TransactionUploadProgressUpdate{
		Status:           status,
		Message:          message,
		Progress:         progress,
		Upload:           upload,
		ValidationErrors: validationErrors,
		SkippedRows:      skippedRows,
	})
}

func estimateUploadStreamProgress(totalRead, contentLength int64) int {
	if contentLength <= 0 {
		return 3
	}

	progress := 2 + int((totalRead*2)/contentLength)
	if progress < 2 {
		return 2
	}
	if progress > 4 {
		return 4
	}

	return progress
}

// DeleteUpload handles DELETE /api/v1/transaction-uploads/:id.
func (a *HttpTransactionUploadAdapter) DeleteUpload(c *gin.Context) {
	_, err := a.deleteUpload.Execute(c.Request.Context(), inboundports.DeleteTransactionUploadCommand{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// RetryClassification handles POST /api/v1/transaction-uploads/:id/retry-classification.
func (a *HttpTransactionUploadAdapter) RetryClassification(c *gin.Context) {
	result, err := a.retryUploadClassification.Execute(c.Request.Context(), inboundports.RetryTransactionUploadClassificationCommand{
		UploadID:     c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: retryTransactionUploadClassificationResponse{
		UploadID:                   result.UploadID,
		EligibleFailedTransactions: result.EligibleFailedTransactions,
		RetriedTransactions:        result.RetriedTransactions,
		SkippedTransactions:        result.SkippedTransactions,
		Skipped:                    result.Skipped,
		FailedRetryCreations:       result.FailedRetryCreations,
		Failures:                   result.Failures,
	}})
}

func (a *HttpTransactionUploadAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var conflictErr *usecases.ConflictError
	var forbiddenErr *usecases.ForbiddenError
	var domainErr *domain.DomainError

	switch {
	case errors.Is(err, domain.ErrInvalidUploadID):
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &conflictErr):
		c.JSON(http.StatusConflict, apiResponse{Error: transactionUploadConflictResponse(conflictErr)})
	case errors.As(err, &forbiddenErr):
		c.JSON(http.StatusForbidden, apiResponse{Error: &errorResponse{Code: "forbidden", Message: err.Error()}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		message := "internal server error"
		if gin.Mode() == gin.DebugMode {
			message = err.Error()
		}

		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: message}})
	}
}

func (a *HttpTransactionUploadAdapter) handlePreviewError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var forbiddenErr *usecases.ForbiddenError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "NOT_FOUND", Message: "Upload not found"}})
	case errors.As(err, &forbiddenErr):
		c.JSON(http.StatusForbidden, apiResponse{Error: &errorResponse{Code: "forbidden", Message: forbiddenErr.Error()}})
	default:
		a.handleError(c, err)
	}
}

func transactionUploadConflictResponse(conflictErr *usecases.ConflictError) *errorResponse {
	if conflictErr == nil {
		return &errorResponse{Code: transactionUploadConflictErrorCode, Message: "transaction upload conflict"}
	}

	switch conflictErr.Reason {
	case domain.ErrDuplicateUpload.Message:
		return &errorResponse{Code: domain.ErrDuplicateUpload.Code, Message: conflictErr.Error()}
	case uploadStillProcessingReason:
		return &errorResponse{Code: uploadStillProcessingErrorCode, Message: uploadStillProcessingMessage}
	default:
		return &errorResponse{Code: transactionUploadConflictErrorCode, Message: conflictErr.Error()}
	}
}

func buildListTransactionUploadsHTTPQuery(c *gin.Context) (inboundports.ListTransactionUploadsQuery, error) {
	startedAt, err := parseOptionalTime(c.Query("start_date"))
	if err != nil {
		return inboundports.ListTransactionUploadsQuery{}, err
	}

	endedAt, err := parseOptionalTime(c.Query("end_date"))
	if err != nil {
		return inboundports.ListTransactionUploadsQuery{}, err
	}

	return inboundports.ListTransactionUploadsQuery{
		FileName:     c.Query("file_name"),
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	}, nil
}

func toListTransactionUploadsResponse(result inboundports.ListTransactionUploadsResult) listTransactionUploadsResponse {
	var uploads []transactionUploadDetailsResponse
	if len(result.Uploads) > 0 {
		uploads = make([]transactionUploadDetailsResponse, 0, len(result.Uploads))
		for _, upload := range result.Uploads {
			uploads = append(uploads, toTransactionUploadDetailsResponse(upload))
		}
	}

	return listTransactionUploadsResponse{Uploads: uploads}
}

func toTransactionUploadDetailsResponse(result ports.TransactionUploadDetailsResult) transactionUploadDetailsResponse {
	var transactions []transactionResponse
	if len(result.Transactions) > 0 {
		transactions = make([]transactionResponse, 0, len(result.Transactions))
		for _, transaction := range result.Transactions {
			transactions = append(transactions, toTransactionResponse(transaction))
		}
	}

	return transactionUploadDetailsResponse{
		transactionUploadResponse: toTransactionUploadResponse(result.TransactionUploadResult),
		Transactions:              transactions,
	}
}

func toTransactionUploadResponse(result ports.TransactionUploadResult) transactionUploadResponse {
	return transactionUploadResponse{
		ID:              result.ID,
		FileName:        result.FileName,
		FileFormat:      result.FileFormat,
		ContentMD5:      result.ContentMD5,
		StorageProvider: result.StorageProvider,
		StorageKey:      result.StorageKey,
		SchemaVersion:   result.SchemaVersion,
		Status:          result.Status,
		RowCount:        result.RowCount,
		UploadedAt:      result.UploadedAt,
	}
}

func toOptionalTransactionUploadResponse(result *ports.TransactionUploadResult) *transactionUploadResponse {
	if result == nil {
		return nil
	}

	response := toTransactionUploadResponse(*result)
	return &response
}

func toCreateTransactionUploadResponse(result inboundports.CreateTransactionUploadResult) createTransactionUploadResponse {
	return createTransactionUploadResponse{
		Upload:           toTransactionUploadResponse(result.Upload),
		ValidationErrors: toTransactionFileValidationErrors(result.ValidationErrors),
		SkippedRows:      toTransactionUploadSkippedRows(result.SkippedRows),
	}
}

func normalizedTransactionUploadValidationError(errors []ports.TransactionFileValidationError) *errorResponse {
	if len(errors) == 0 {
		return &errorResponse{Code: "UPLOAD_VALIDATION_FAILED", Message: "Upload validation failed."}
	}

	categories, hasMissingRequiredColumn := normalizedTransactionUploadValidationCategories(errors)
	code := strings.Join(categories, "_AND_")

	return &errorResponse{
		Code:    code,
		Message: normalizedTransactionUploadValidationMessage(code, hasMissingRequiredColumn),
	}
}

func normalizedTransactionUploadValidationCategories(errors []ports.TransactionFileValidationError) ([]string, bool) {
	seen := make(map[string]struct{}, 3)
	hasMissingRequiredColumn := false

	for _, validationError := range errors {
		if validationError.Code == "missing_required_column" {
			hasMissingRequiredColumn = true
		}

		category := normalizedTransactionUploadValidationCategory(validationError)
		seen[category] = struct{}{}
	}

	orderedCategories := make([]string, 0, len(seen))
	for _, category := range []string{"MISSING_FIELD", "INVALID_FORMAT", "INVALID_REFERENCE"} {
		if _, ok := seen[category]; ok {
			orderedCategories = append(orderedCategories, category)
		}
	}

	return orderedCategories, hasMissingRequiredColumn
}

func normalizedTransactionUploadValidationCategory(validationError ports.TransactionFileValidationError) string {
	switch validationError.Code {
	case "missing_required_column", "missing_required_value":
		return "MISSING_FIELD"
	case "invalid_value":
		return "INVALID_REFERENCE"
	case "type_mismatch":
		return "INVALID_FORMAT"
	default:
		return "INVALID_FORMAT"
	}
}

func normalizedTransactionUploadValidationMessage(code string, hasMissingRequiredColumn bool) string {
	switch code {
	case "MISSING_FIELD":
		if hasMissingRequiredColumn {
			return "Required column is missing."
		}
		return "Required field is empty or null."
	case "INVALID_FORMAT":
		return "One or more values do not match the expected format."
	case "INVALID_REFERENCE":
		return "One or more values do not match the allowed reference values."
	case "MISSING_FIELD_AND_INVALID_FORMAT":
		return "The upload has missing required fields and invalidly formatted or unsupported values."
	case "MISSING_FIELD_AND_INVALID_REFERENCE":
		return "The upload has missing required fields and values that do not match the allowed reference values."
	case "INVALID_FORMAT_AND_INVALID_REFERENCE":
		return "The upload has invalidly formatted or unsupported values and values that do not match the allowed reference values."
	case "MISSING_FIELD_AND_INVALID_FORMAT_AND_INVALID_REFERENCE":
		return "The upload has missing required fields, invalidly formatted or unsupported values, and values that do not match the allowed reference values."
	default:
		return "Upload validation failed."
	}
}

func toTransactionFileValidationErrors(errors []ports.TransactionFileValidationError) []transactionFileValidationError {
	if len(errors) == 0 {
		return nil
	}

	result := make([]transactionFileValidationError, 0, len(errors))
	for _, validationError := range errors {
		result = append(result, transactionFileValidationError{
			Code:        validationError.Code,
			Message:     validationError.Message,
			RowNumber:   validationError.RowNumber,
			ColumnName:  validationError.ColumnName,
			ColumnIndex: validationError.ColumnIndex,
			Value:       validationError.Value,
		})
	}

	return result
}

func toTransactionUploadPreviewValidationErrors(errors []ports.TransactionFileValidationError) []transactionUploadPreviewValidationError {
	if len(errors) == 0 {
		return nil
	}

	result := make([]transactionUploadPreviewValidationError, 0, len(errors))
	for _, validationError := range errors {
		result = append(result, transactionUploadPreviewValidationError{
			Code:        mapTransactionUploadPreviewValidationCode(validationError.Code),
			Message:     validationError.Message,
			RowNumber:   validationError.RowNumber,
			ColumnName:  validationError.ColumnName,
			ColumnIndex: validationError.ColumnIndex,
			Value:       validationError.Value,
		})
	}

	return result
}

func mapTransactionUploadPreviewValidationCode(code string) string {
	switch code {
	case "missing_required_value":
		return "MISSING_FIELD"
	case "type_mismatch":
		return "INVALID_FORMAT"
	default:
		return code
	}
}

func toTransactionUploadPreviewRows(columns []string, rows [][]string) []map[string]string {
	if len(rows) == 0 {
		return nil
	}

	mappedRows := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		mappedRow := make(map[string]string, len(columns))
		for index, column := range columns {
			value := ""
			if index < len(row) {
				value = row[index]
			}
			mappedRow[column] = value
		}
		mappedRows = append(mappedRows, mappedRow)
	}

	return mappedRows
}

func toTransactionUploadSkippedRows(rows []ports.TransactionUploadSkippedRow) []transactionUploadSkippedRow {
	if len(rows) == 0 {
		return nil
	}

	result := make([]transactionUploadSkippedRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, transactionUploadSkippedRow{
			RowNumber: row.RowNumber,
			Reason:    row.Reason,
		})
	}

	return result
}
