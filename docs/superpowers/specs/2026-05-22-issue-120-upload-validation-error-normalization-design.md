# Issue 120 Upload Validation Error Normalization Design

## Summary

Issue #120 changes transaction upload validation failures from detailed row-level response payloads into a simplified frontend-facing error contract.

The change applies to:

- `POST /api/v1/transaction-uploads` when validation fails
- `POST /api/v1/transaction-uploads/stream` when a `validation_failed` SSE event is emitted

The change does **not** apply to successful upload responses, validation rules, or persisted preview validation detail.

## Goal

Return a single top-level `error` object with a normalized `code` and human-readable `message` so the frontend does not need to inspect or merge `validation_errors` entries.

## In Scope

- Replace failed create-upload HTTP response payloads with a top-level `error` object
- Replace streaming validation-failed payload `validation_errors` output with a top-level `error` object
- Normalize raw validator error codes into frontend-facing categories
- Combine multiple categories into a single specific code using `_AND_`
- Produce a natural-language summary message for one or more categories
- Add adapter tests covering the new HTTP and SSE contracts

## Out of Scope

- Changing transaction file validation rules
- Changing successful upload response payloads
- Changing preview persistence of row-level validation errors
- Changing `GET /api/v1/transaction-uploads/:id/preview` response shape
- Frontend UI implementation

## Current State

The upload create flow currently returns `apiResponse{Data: ...}` for both success and validation failure. Failed responses include `data.upload` and may include `data.validation_errors`.

The streaming upload flow currently emits `uploadProgressResponse` payloads that include `validation_errors` during `validation_failed` updates.

The preview endpoint already maps selected raw validation codes for display, but it still returns row-level validation entries. That behavior remains unchanged for this issue.

## Recommended Approach

Implement the normalization in `internal/adapters/transaction_upload_http_adapter.go` only.

This is a transport contract change rather than a domain or application rule change. The adapter already owns HTTP error translation, and keeping the normalization there preserves the existing use-case and preview storage contracts.

## Architecture Placement

### Adapter responsibilities

The HTTP adapter will:

- convert raw `[]ports.TransactionFileValidationError` into one `errorResponse`
- return `apiResponse{Error: ...}` for failed create-upload requests
- emit SSE validation-failed payloads with `error` instead of `validation_errors`

### Unchanged responsibilities

The following remain unchanged:

- `CreateTransactionUploadUseCase` still returns raw validation errors in `ports.CreateTransactionUploadResult`
- `TransactionUploadProgressUpdate` still carries raw validation errors internally
- preview persistence still stores row-level validation details
- preview reads still expose row-level validation details

This keeps internal detail available for diagnostics and preview UX while changing only the frontend-facing upload failure contract.

## Response Contract

### HTTP failed upload response

When validation fails, `POST /api/v1/transaction-uploads` returns:

```json
{
  "error": {
    "code": "MISSING_FIELD",
    "message": "Required field is empty or null."
  }
}
```

When multiple normalized categories are present, the code joins them with `_AND_`:

```json
{
  "error": {
    "code": "MISSING_FIELD_AND_INVALID_FORMAT",
    "message": "The upload has missing required fields and invalidly formatted or unsupported values."
  }
}
```

The response will no longer include `data`, `data.upload`, or `data.validation_errors` for validation failures.

### Streaming validation-failed SSE payload

For `POST /api/v1/transaction-uploads/stream`, normal progress payloads remain unchanged except for validation-failed updates.

During `validation_failed`, the SSE JSON payload will contain:

- `status`
- `message`
- `progress`
- `error`
- optional `upload` if present in the existing update
- optional `skipped_rows` if present in the existing update

It will no longer contain `validation_errors`.

Example SSE payload body:

```json
{
  "status": "validation_failed",
  "message": "validation failed",
  "progress": 100,
  "error": {
    "code": "MISSING_FIELD_AND_INVALID_FORMAT",
    "message": "The upload has missing required fields and invalidly formatted or unsupported values."
  }
}
```

## Normalization Taxonomy

The adapter will map raw validator codes into the following normalized categories for this issue:

- `missing_required_column` → `MISSING_FIELD`
- `missing_required_value` → `MISSING_FIELD`
- `type_mismatch` → `INVALID_FORMAT`
- `invalid_value` → `INVALID_REFERENCE`

This design intentionally implements only the mappings required by the currently observed upload validator outputs. In the current validator, `invalid_value` is emitted for schema columns with constrained allowed values, so it is normalized directly to `INVALID_REFERENCE`. Additional taxonomy entries such as `INVALID_TYPE`, `DUPLICATE`, and `OUT_OF_RANGE` can be added later if new raw validation codes are introduced.

## Combined Code Rules

When more than one normalized category is present:

1. Deduplicate categories
2. Order categories deterministically
3. Join categories with `_AND_`

Examples:

- `MISSING_FIELD`
- `INVALID_FORMAT`
- `MISSING_FIELD_AND_INVALID_FORMAT`
- `MISSING_FIELD_AND_INVALID_REFERENCE`

The ordering must be stable so tests and clients see consistent outputs.

Recommended category order for this issue:

1. `MISSING_FIELD`
2. `INVALID_FORMAT`
3. `INVALID_REFERENCE`

## Message Rules

The adapter will generate a summary message from the normalized categories instead of exposing row-level messages.

### Single-category messages

- `MISSING_FIELD` → `Required field is empty or null.` for missing required values
- `MISSING_FIELD` → `Required column is missing.` when the normalized error comes from `missing_required_column`
- `INVALID_FORMAT` → `One or more values do not match the expected format.`
- `INVALID_REFERENCE` → `One or more values do not match the allowed reference values.`

### Fallback message

When a validation failure is being surfaced but no row-level validation details are present, the adapter falls back to:

- `UPLOAD_VALIDATION_FAILED` → `Upload validation failed.`

This fallback is used to keep the transport contract stable for validation-failed responses, including SSE `validation_failed` events, even when the upstream validation error slice is empty.

### Multi-category messages

For multiple categories, the adapter will build one natural-language sentence that reads as a summary of the overall validation outcome.

Examples:

- `MISSING_FIELD_AND_INVALID_FORMAT` → `The upload has missing required fields and invalidly formatted or unsupported values.`
- `MISSING_FIELD_AND_INVALID_REFERENCE` → `The upload has missing required fields and values that do not match the allowed reference values.`
- `INVALID_FORMAT_AND_INVALID_REFERENCE` → `The upload has invalidly formatted or unsupported values and values that do not match the allowed reference values.`
- `MISSING_FIELD_AND_INVALID_FORMAT_AND_INVALID_REFERENCE` → `The upload has missing required fields, invalidly formatted or unsupported values, and values that do not match the allowed reference values.`

The message is intentionally category-based rather than row-based so frontend consumers receive one presentation-ready summary.

## Implementation Notes

The adapter will add a small shared helper set in `transaction_upload_http_adapter.go` to:

- extract normalized categories from raw validation errors
- build a combined normalized code
- build a summary message
- return `*errorResponse`

The helper will be reused by both the standard HTTP create-upload response path and the SSE progress reporter.

No port, use-case, validator, preview repository, or infrastructure wiring changes are required.

## Testing Strategy

Adapter tests will drive the implementation.

### HTTP tests

Add or update tests to verify:

- a failed upload with one raw validation error returns top-level `error.code` and `error.message`
- a failed upload with mixed raw categories returns a combined `_AND_` code
- failed upload responses omit `data`, `data.upload`, and `data.validation_errors`
- successful upload responses remain unchanged

### SSE tests

Add or update tests to verify:

- `validation_failed` SSE payload includes top-level `error`
- `validation_failed` SSE payload omits `validation_errors`
- the SSE error code and message match the HTTP normalization rules

### Regression boundaries

Keep preview tests unchanged because preview remains out of scope.

## Risks and Mitigations

### Risk: inconsistent normalization between HTTP and SSE

Mitigation: use one shared adapter helper for both paths.

### Risk: overfitting message text to current validator output

Mitigation: summarize by normalized category rather than raw row messages.

### Risk: accidental change to successful responses

Mitigation: retain the existing success path and add regression assertions for successful uploads.

## Acceptance Mapping

- Top-level `error` object with `code` and `message` → addressed by the new failed HTTP and SSE response contract
- Frontend no longer combines row-level entries → addressed by adapter-side normalization helper
- Existing validation failures map to normalized taxonomy → addressed by raw-to-normalized mapping rules
- Same-category and mixed-category failures collapse into one summarized error → addressed by deduplication, ordering, `_AND_` code generation, and summary message generation
- Successful upload responses unchanged → protected by unchanged success path and regression tests
