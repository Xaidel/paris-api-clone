# Issue 120 Upload Validation Error Normalization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace transaction upload validation failure payloads with a single normalized top-level error object for both standard HTTP uploads and streamed validation-failed SSE updates.

**Architecture:** Keep normalization in the HTTP adapter. The application and ports continue to carry raw validation errors internally; the adapter translates them into frontend-facing `error.code` and `error.message` values at the transport boundary.

**Tech Stack:** Go 1.x, Gin, standard `testing` package, existing adapter route tests in `internal/adapters`

---

### Task 1: Normalize failed create-upload HTTP responses

**Files:**
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go`
- Modify: `internal/adapters/transaction_upload_http_adapter.go`
- Test: `internal/adapters/transaction_upload_http_adapter_test.go`

- [ ] **Step 1: Write the failing route tests for top-level error responses**

Update the existing failed-upload subtests in `internal/adapters/transaction_upload_http_adapter_test.go` so they assert the new contract instead of `data.validation_errors`.

```go
{
	name:     "returns validation errors",
	method:   http.MethodPost,
	target:   "/api/v1/transaction-uploads",
	withFile: true,
	createStub: &createTransactionUploadPortStub{result: ports.CreateTransactionUploadResult{
		Upload: ports.TransactionUploadResult{ID: transactionUploadID1, Status: "failed"},
		ValidationErrors: []ports.TransactionFileValidationError{{
			Code:        "type_mismatch",
			Message:     "Year must be an integer",
			RowNumber:   2,
			ColumnName:  "Year",
			ColumnIndex: 2,
			Value:       "20XX",
		}},
	}},
	streamStub: &transactionUploadStreamExecutorStub{},
	getStub:    &getTransactionUploadPortStub{},
	previewStub: &getTransactionUploadPreviewPortStub{},
	listStub:   &listTransactionUploadsPortStub{},
	deleteStub: &deleteTransactionUploadPortStub{},
	retryStub:  &retryTransactionUploadClassificationPortStub{},
	wantStatus: http.StatusUnprocessableEntity,
	wantErrorCode: "INVALID_FORMAT",
	assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
		t.Helper()

		errorPayload := payload["error"].(map[string]any)
		if errorPayload["message"] != "One or more values do not match the expected format." {
			t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "One or more values do not match the expected format.")
		}
		if _, ok := payload["data"]; ok {
			t.Fatalf("payload[data] present = %v, want omitted", payload["data"])
		}
	},
},
{
	name:     "returns failed upload payload as top-level error only",
	method:   http.MethodPost,
	target:   "/api/v1/transaction-uploads",
	withFile: true,
	createStub: &createTransactionUploadPortStub{result: ports.CreateTransactionUploadResult{
		Upload: ports.TransactionUploadResult{ID: transactionUploadID1, FileName: "transactions.csv", FileFormat: "csv", Status: "failed", RowCount: 0},
		ValidationErrors: []ports.TransactionFileValidationError{{
			Code:        "missing_required_value",
			Message:     `row 4 column "Value of Transactions" is required`,
			RowNumber:   4,
			ColumnName:  "Value of Transactions",
			ColumnIndex: 8,
		}},
	}},
	streamStub: &transactionUploadStreamExecutorStub{},
	getStub:    &getTransactionUploadPortStub{},
	previewStub: &getTransactionUploadPreviewPortStub{},
	listStub:   &listTransactionUploadsPortStub{},
	deleteStub: &deleteTransactionUploadPortStub{},
	retryStub:  &retryTransactionUploadClassificationPortStub{},
	wantStatus: http.StatusUnprocessableEntity,
	wantErrorCode: "MISSING_FIELD",
	assert: func(t *testing.T, _ *httptest.ResponseRecorder, _ *createTransactionUploadPortStub, _ *transactionUploadStreamExecutorStub, _ *getTransactionUploadPortStub, _ *getTransactionUploadPreviewPortStub, _ *listTransactionUploadsPortStub, _ *deleteTransactionUploadPortStub, _ *retryTransactionUploadClassificationPortStub, _ *downloadTransactionUploadPortStub, payload map[string]any) {
		t.Helper()

		errorPayload := payload["error"].(map[string]any)
		if errorPayload["message"] != "Required field is empty or null." {
			t.Fatalf("errorPayload[message] = %v, want %q", errorPayload["message"], "Required field is empty or null.")
		}
		if _, ok := payload["data"]; ok {
			t.Fatalf("payload[data] present = %v, want omitted", payload["data"])
		}
	},
},
```

- [ ] **Step 2: Run the adapter route test to verify it fails**

Run:

```bash
go test ./internal/adapters -run TestHttpTransactionUploadAdapterRoutes -count=1
```

Expected: `FAIL` because the handler still returns `apiResponse{Data: ...}` with `validation_errors` instead of a top-level `error`.

- [ ] **Step 3: Write the minimal HTTP adapter implementation**

Update `internal/adapters/transaction_upload_http_adapter.go` so failed create-upload responses return `apiResponse{Error: ...}` and add a minimal normalization helper that covers the currently failing tests.

```go
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

	result, err := a.createUpload.Execute(c.Request.Context(), ports.CreateTransactionUploadCommand{
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
		c.JSON(http.StatusUnprocessableEntity, apiResponse{
			Error: normalizedTransactionUploadValidationError(result.ValidationErrors),
		})
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: toCreateTransactionUploadResponse(result)})
}

func normalizedTransactionUploadValidationError(errors []ports.TransactionFileValidationError) *errorResponse {
	for _, validationError := range errors {
		switch validationError.Code {
		case "missing_required_column", "missing_required_value":
			return &errorResponse{Code: "MISSING_FIELD", Message: "Required field is empty or null."}
		case "type_mismatch":
			return &errorResponse{Code: "INVALID_FORMAT", Message: "One or more values do not match the expected format."}
		}
	}

	return &errorResponse{Code: "INVALID_FORMAT", Message: "One or more values do not match the expected format."}
}
```

- [ ] **Step 4: Run the route test again to verify it passes**

Run:

```bash
go test ./internal/adapters -run TestHttpTransactionUploadAdapterRoutes -count=1
```

Expected: `ok` with the failed-upload route cases now asserting top-level `error` payloads and no `data` object.

- [ ] **Step 5: Commit the HTTP normalization change**

Run:

```bash
git add internal/adapters/transaction_upload_http_adapter.go internal/adapters/transaction_upload_http_adapter_test.go
git commit -m "feat(transaction-upload): normalize failed upload validation responses"
```

Expected: one commit containing only the HTTP failed-response contract change and its tests.

### Task 2: Add combined category normalization for upload validation errors

**Files:**
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go`
- Modify: `internal/adapters/transaction_upload_http_adapter.go`
- Test: `internal/adapters/transaction_upload_http_adapter_test.go`

- [ ] **Step 1: Write failing tests for combined codes and summary messages**

Add a focused helper test in `internal/adapters/transaction_upload_http_adapter_test.go` that locks the normalization rules before refactoring the helper.

```go
func TestNormalizedTransactionUploadValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errors       []ports.TransactionFileValidationError
		wantCode    string
		wantMessage string
	}{
		{
			name: "single missing field",
			errors: []ports.TransactionFileValidationError{{
				Code:    "missing_required_value",
				Message: `row 4 column "Value of Transactions" is required`,
			}},
			wantCode:    "MISSING_FIELD",
			wantMessage: "Required field is empty or null.",
		},
		{
			name: "combined missing field and invalid format",
			errors: []ports.TransactionFileValidationError{
				{Code: "type_mismatch", Message: `row value for column "Year" must be an integer`},
				{Code: "missing_required_value", Message: `row 4 column "Value of Transactions" is required`},
				{Code: "missing_required_value", Message: `row 6 column "Month" is required`},
			},
			wantCode:    "MISSING_FIELD_AND_INVALID_FORMAT",
			wantMessage: "The upload has missing required fields and invalidly formatted or unsupported values.",
		},
		{
			name: "invalid value maps to invalid reference",
			errors: []ports.TransactionFileValidationError{{
				Code:    "invalid_value",
				Message: `row value for column "PA Alignment" must be one of PA Aligned, Not PA Aligned`,
			}},
			wantCode:    "INVALID_REFERENCE",
			wantMessage: "One or more values do not match the allowed reference values.",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := normalizedTransactionUploadValidationError(tt.errors)
			if got.Code != tt.wantCode {
				t.Fatalf("normalizedTransactionUploadValidationError(...).Code = %q, want %q", got.Code, tt.wantCode)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("normalizedTransactionUploadValidationError(...).Message = %q, want %q", got.Message, tt.wantMessage)
			}
		})
	}
}
```

- [ ] **Step 2: Run the focused normalization test to verify it fails**

Run:

```bash
go test ./internal/adapters -run TestNormalizedTransactionUploadValidationError -count=1
```

Expected: `FAIL` because the current helper returns the first matching category only and does not combine or summarize categories.

- [ ] **Step 3: Refactor the helper into deterministic category normalization**

Replace the minimal helper in `internal/adapters/transaction_upload_http_adapter.go` with category extraction, deterministic ordering, `_AND_` code composition, and category-based summary messages.

```go
var transactionUploadValidationCategoryOrder = []string{
	"MISSING_FIELD",
	"INVALID_FORMAT",
	"INVALID_REFERENCE",
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
	seen := make(map[string]struct{}, len(errors))
	hasMissingRequiredColumn := false
	for _, validationError := range errors {
		if validationError.Code == "missing_required_column" {
			hasMissingRequiredColumn = true
		}
		category := normalizedTransactionUploadValidationCategory(validationError)
		if category == "" {
			continue
		}
		seen[category] = struct{}{}
	}

	result := make([]string, 0, len(seen))
	for _, category := range transactionUploadValidationCategoryOrder {
		if _, ok := seen[category]; ok {
			result = append(result, category)
		}
	}

	return result, hasMissingRequiredColumn
}

func normalizedTransactionUploadValidationCategory(validationError ports.TransactionFileValidationError) string {
	switch validationError.Code {
	case "missing_required_column", "missing_required_value":
		return "MISSING_FIELD"
	case "type_mismatch":
		return "INVALID_FORMAT"
	case "invalid_value":
		return "INVALID_REFERENCE"
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
```

Also update the import block in `internal/adapters/transaction_upload_http_adapter.go` to include `strings`.

- [ ] **Step 4: Run the focused normalization test again to verify it passes**

Run:

```bash
go test ./internal/adapters -run TestNormalizedTransactionUploadValidationError -count=1
```

Expected: `ok` with the helper returning stable combined codes and category-based summary messages.

- [ ] **Step 5: Commit the combined-category normalization helper**

Run:

```bash
git add internal/adapters/transaction_upload_http_adapter.go internal/adapters/transaction_upload_http_adapter_test.go
git commit -m "feat(transaction-upload): combine upload validation categories"
```

Expected: one commit containing the deterministic normalization helper and its direct unit tests.

### Task 3: Normalize validation-failed SSE payloads

**Files:**
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go`
- Modify: `internal/adapters/transaction_upload_http_adapter.go`
- Test: `internal/adapters/transaction_upload_http_adapter_test.go`

- [ ] **Step 1: Write the failing SSE reporter test**

Add a new reporter-focused test next to `TestSSETransactionUploadProgressReporterIncludesSkippedRows` in `internal/adapters/transaction_upload_http_adapter_test.go`.

```go
func TestSSETransactionUploadProgressReporterIncludesNormalizedValidationError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/transaction-uploads/stream", nil)

	reporter := newSSETransactionUploadProgressReporter(context)
	err := reporter.Report(context.Request.Context(), ports.TransactionUploadProgressUpdate{
		Status:   ports.TransactionUploadProgressStatusValidationFailed,
		Message:  "validation failed",
		Progress: 100,
		ValidationErrors: []ports.TransactionFileValidationError{
			{Code: "missing_required_value", Message: `row 4 column "Value of Transactions" is required`},
			{Code: "type_mismatch", Message: `row value for column "Year" must be an integer`, Value: "20XX"},
		},
	})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	body := response.Body.String()
	if !bytes.Contains([]byte(body), []byte("event: validation_failed")) {
		t.Fatalf("response body = %q, want validation_failed event", body)
	}
	if !bytes.Contains([]byte(body), []byte(`"error":{"code":"MISSING_FIELD_AND_INVALID_FORMAT","message":"The upload has missing required fields and invalidly formatted or unsupported values."}`)) {
		t.Fatalf("response body = %q, want normalized validation error payload", body)
	}
	if bytes.Contains([]byte(body), []byte(`"validation_errors"`)) {
		t.Fatalf("response body = %q, want validation_errors omitted", body)
	}
}
```

- [ ] **Step 2: Run the focused SSE test to verify it fails**

Run:

```bash
go test ./internal/adapters -run TestSSETransactionUploadProgressReporterIncludesNormalizedValidationError -count=1
```

Expected: `FAIL` because the SSE payload still serializes `validation_errors` and does not include a top-level `error` object.

- [ ] **Step 3: Update the SSE payload shape and reporter implementation**

Modify `internal/adapters/transaction_upload_http_adapter.go` so `uploadProgressResponse` carries `Error` instead of `ValidationErrors`, and populate it from the shared normalization helper.

```go
type uploadProgressResponse struct {
	Status     string                 `json:"status"`
	Message    string                 `json:"message"`
	Progress   int                    `json:"progress"`
	Upload     *transactionUploadResponse `json:"upload,omitempty"`
	Error      *errorResponse         `json:"error,omitempty"`
	SkippedRows []transactionUploadSkippedRow `json:"skipped_rows,omitempty"`
}

func (r *sseTransactionUploadProgressReporter) Report(_ context.Context, update ports.TransactionUploadProgressUpdate) error {
	payload, err := json.Marshal(uploadProgressResponse{
		Status:      update.Status,
		Message:     update.Message,
		Progress:    update.Progress,
		Upload:      toOptionalTransactionUploadResponse(update.Upload),
		Error:       normalizedTransactionUploadValidationError(update.ValidationErrors),
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
```

Keep the helper shared between standard HTTP upload failures and SSE validation-failed updates.

- [ ] **Step 4: Run the focused SSE test and the full adapter package to verify it passes**

Run:

```bash
go test ./internal/adapters -run TestSSETransactionUploadProgressReporterIncludesNormalizedValidationError -count=1
go test ./internal/adapters -count=1
```

Expected:

- first command: `ok`
- second command: `ok` confirming successful upload, preview, and skipped-row adapter tests still pass after the SSE payload change

- [ ] **Step 5: Commit the streamed validation payload change**

Run:

```bash
git add internal/adapters/transaction_upload_http_adapter.go internal/adapters/transaction_upload_http_adapter_test.go
git commit -m "feat(transaction-upload): normalize streamed validation failures"
```

Expected: one commit containing the SSE payload contract update and its test coverage.
