# Issue 115 Transaction Upload Preview Design

## Summary

Add `GET /api/v1/transaction-uploads/:id/preview` so the upload detail page can fetch previewable upload rows together with stored schema validation errors.

The endpoint must:

- allow preview access for both `uploaded` and `failed` transaction uploads,
- restrict access to callers in the same group as the upload,
- return stable preview data captured at upload time rather than re-parsing the raw file on demand,
- expose columns, rows, total row count, and structured validation errors in a frontend-friendly shape,
- return the exact unknown-upload contract required by issue 115.

## Why

The current upload flow validates files during ingestion, but it only returns validation errors in the create response and does not persist the parsed preview artifacts needed by the upload detail page. The frontend needs a durable backend source for rendering row data and highlighting invalid or missing cells after the initial upload request has completed.

## Constraints and decisions

- Authorization model: **same-group access**.
- Route shape: **`GET /api/v1/transaction-uploads/:id/preview`**.
- Eligible upload statuses: **both `uploaded` and `failed`**.
- Preview data must reflect **create-time parsing and validation results**, not a later re-computation.
- `rows` will be returned as **`[][]string` aligned by index with `columns`**.
- Preview rows will contain the rows that survive the existing transaction-count row filter. Rows skipped as `malformed` or `not_valid_transaction` remain outside the preview payload for this issue.
- `total_rows` equals `len(rows)` and does not count the header row.
- The endpoint will keep the codebase's existing response envelope convention; the preview payload fields live inside `data`.
- Work stays on the current branch.

## Existing codebase observations

The current transaction-upload slice already has several useful building blocks:

- `CreateTransactionUploadUseCase` parses the uploaded file, filters invalid transaction-count rows, validates the remaining rows against the schema, stores the raw file, and persists upload metadata.
- `TransactionFileParser` already produces the exact structural data the preview page needs: `Headers` and `Rows`.
- `TransactionFileValidator` already produces structured validation errors with `Code`, `Message`, `RowNumber`, `ColumnName`, `ColumnIndex`, and `Value`.
- `HttpTransactionUploadAdapter` already owns route registration, actor-header extraction, and application-error mapping for the upload slice.

However, two gaps prevent the preview endpoint from being implemented cleanly today:

1. `transaction_upload` does **not** persist the upload's owning group, so same-group authorization cannot be enforced reliably for failed uploads with zero accepted transactions.
2. The system does **not** persist preview artifacts (`columns`, preview rows, validation errors), so the application cannot reproduce the upload detail page state after the create request completes without re-parsing the raw file.

## Recommended architecture

### 1. Persist upload ownership explicitly

Add `group_id` to `transaction_upload` and make it the authorization source for preview access.

This keeps authorization explicit in the application layer and avoids inferring ownership from accepted transactions, admin events, or other indirect signals.

### 2. Persist preview artifacts at upload time

Store preview data during `CreateTransactionUploadUseCase` execution, using the already parsed and validated in-memory data.

Persisting preview artifacts at create time is the recommended approach because it:

- keeps preview responses deterministic,
- avoids re-reading and re-parsing raw files later,
- avoids schema drift when validation rules evolve,
- works uniformly for both successful and failed uploads.

### 3. Add a dedicated preview query use case and inbound port

Introduce a new preview query flow that:

- validates the actor through `ActorDirectory`,
- loads the upload metadata,
- enforces same-group access using persisted `group_id`,
- loads stored preview artifacts,
- records a preview admin event,
- returns a transport-agnostic preview result.

This keeps authorization and orchestration in the application layer while keeping HTTP-specific JSON shaping inside the adapter.

### 4. Keep preview transport formatting in the HTTP adapter

The HTTP adapter should:

- register `GET /api/v1/transaction-uploads/:id/preview`,
- construct the query from route params and actor headers,
- call the preview port,
- map the use-case result into the response payload,
- translate application errors into the required HTTP contract.

## Detailed design

### Domain layer

#### `TransactionUpload`

Extend the entity to carry `groupID`.

Changes:

- add a `groupID valueobjects.GroupID` field,
- update constructor and reconstitution signatures,
- validate group ID in `NewTransactionUpload`,
- expose `GroupID() valueobjects.GroupID`,
- add a `RecordPreviewed(now, actorUserID, actorGroupID string)` method.

`RecordPreviewed` should emit an admin action event with:

- event type: `PreviewTransactionUploadEventType`,
- `action: "preview"`,
- `resource: "transaction_upload"`,
- `upload_id`,
- `file_name`.

#### Admin event types

Add a new event type constant:

- `PreviewTransactionUploadEventType = "PreviewTransactionUpload"`.

### Ports layer

#### New inbound port

Add a new inbound port file for preview retrieval.

Suggested contract:

- `GetTransactionUploadPreviewQuery { ID string; ActorUserID string; ActorGroupID string }`
- `GetTransactionUploadPreviewResult { FileID string; FileName string; Columns []string; Rows [][]string; TotalRows int; ValidationErrors []TransactionFileValidationError }`
- `GetTransactionUploadPreviewPort`

#### New outbound preview repository

Add a dedicated outbound port for persisted preview artifacts.

Suggested contract:

- `TransactionUploadPreviewRecord { UploadID string; Columns []string; Rows [][]string; TotalRows int; ValidationErrors []TransactionFileValidationError }`
- `TransactionUploadPreviewRepository`
  - `Save(ctx context.Context, preview TransactionUploadPreviewRecord) error`
  - `FindByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (*TransactionUploadPreviewRecord, error)`

This keeps preview persistence independent from the upload metadata repository and avoids overloading `TransactionUploadRepository` with non-metadata concerns.

#### Existing upload result contract

Extend `ports.TransactionUploadResult` with `GroupID string`.

This is not for public API exposure. It keeps application-layer DTOs consistent with the new entity shape and simplifies repository hydration tests.

### Application layer

#### New application error

Add `ForbiddenError` to `internal/application/use-cases/errors.go`.

Suggested shape:

- `Resource string`
- `Reason string`

This gives the preview use case a transport-agnostic way to report same-group access denial.

#### `CreateTransactionUploadUseCase`

Extend the existing upload flow to persist preview artifacts.

Changes:

- inject `ports.TransactionUploadPreviewRepository`,
- pass `ActorGroupID` into `TransactionUpload` construction via persisted `group_id`,
- build preview data from:
  - `columns = parsedFile.Headers`,
  - `rows = filteredRows.EligibleRows`,
  - `total_rows = len(filteredRows.EligibleRows)`,
  - `validation_errors = toTransactionFileValidationErrors(validationReport.Errors())`.

Persistence rules:

- **successful upload**: save preview artifacts with empty `validation_errors`,
- **failed upload caused by schema validation**: save preview artifacts with validation errors,
- **failed upload with zero eligible rows**: save preview artifacts with the parsed headers, empty rows, `total_rows = 0`, and empty validation errors.

The preview artifact save must happen inside the same transaction boundary as upload creation so metadata and preview state cannot diverge.

#### `GetTransactionUploadPreviewUseCase`

Add a new query use case.

Dependencies:

- `ports.TransactionUploadRepository`
- `ports.TransactionUploadPreviewRepository`
- `adminEventRecorder`
- `ports.ActorDirectory`
- `now func() time.Time`

Workflow:

1. validate actor through `validateActor`,
2. parse upload ID,
3. load upload by ID,
4. return `NotFoundError` if missing,
5. compare `upload.GroupID().String()` to `query.ActorGroupID`,
6. return `ForbiddenError` on mismatch,
7. load preview artifacts by upload ID,
8. return `NotFoundError` if preview artifacts are unexpectedly missing,
9. call `upload.RecordPreviewed(...)`,
10. publish the resulting domain event,
11. return `GetTransactionUploadPreviewResult`.

The use case remains framework-free and does not know about HTTP response envelopes.

### Adapters layer

#### Postgres upload repository

Extend SQL statements to read and write `group_id`.

Changes:

- add `group_id` to insert statements,
- add `group_id` to `SELECT` projections,
- update scan and reconstitution logic,
- update repository tests accordingly.

#### New Postgres preview repository

Add a concrete adapter implementing `TransactionUploadPreviewRepository`.

Recommended schema mapping:

- `transaction_upload_preview`
  - `upload_id TEXT PRIMARY KEY REFERENCES transaction_upload(id) ON DELETE CASCADE`
  - `columns_json JSONB NOT NULL`
  - `rows_json JSONB NOT NULL`
  - `total_rows INTEGER NOT NULL`
- `transaction_upload_preview_validation_error`
  - `upload_id TEXT NOT NULL REFERENCES transaction_upload(id) ON DELETE CASCADE`
  - `ordinal INTEGER NOT NULL`
  - `code TEXT NOT NULL`
  - `message TEXT NOT NULL`
  - `row_number INTEGER NOT NULL`
  - `column_name TEXT NOT NULL`
  - `column_index INTEGER NOT NULL`
  - `value TEXT NOT NULL DEFAULT ''`
  - primary key `(upload_id, ordinal)`

Rationale:

- `columns_json` and `rows_json` preserve variable-width table structure cleanly,
- the child error table keeps validation errors queryable and ordered,
- `ordinal` preserves stable frontend ordering.

#### HTTP transaction upload adapter

Extend the adapter constructor to accept the new preview port and register:

- `GET /api/v1/transaction-uploads/:id/preview`

Handler behavior:

1. build `GetTransactionUploadPreviewQuery` from route + actor headers,
2. call the preview port,
3. return `200 OK` with the preview payload inside `data`.

Preview payload shape:

- `file_id`
- `file_name`
- `columns`
- `rows`
- `total_rows`
- `validation_errors`

Error mapping additions:

- `ForbiddenError` -> HTTP 403,
- unknown upload -> HTTP 404 with:
  - `error.code = "NOT_FOUND"`
  - `error.message = "Upload not found"`.

Because issue 115 requires an exact not-found contract that differs from the adapter's current generic lowercase `not_found` mapping, the preview route should use a preview-specific not-found response shape rather than silently changing every existing transaction-upload endpoint.

### Infrastructure layer

#### Dependency injection

Wire the new preview repository and preview use case into `internal/infrastructure/di/wire.go`, then pass the preview port into `NewHttpTransactionUploadAdapter`.

No new runtime configuration is required.

### Database migration

Add a new migration after `000207`.

#### Up migration responsibilities

1. add nullable `group_id` to `transaction_upload`,
2. backfill existing rows to the seeded `superadmin` group ID:
   - `01962b8f-aeb2-7e03-a8ff-1edce1300001`,
3. set `group_id` to `NOT NULL`,
4. add an index on `group_id`,
5. create `transaction_upload_preview`,
6. create `transaction_upload_preview_validation_error`.

#### Backfill strategy

This branch does not yet have persisted upload ownership. The deployed data assumption for this codebase is that the seeded `superadmin` group is the effective owner of existing transaction uploads, so the migration should backfill all legacy rows to that group deterministically.

This is a deliberate migration-time compatibility decision, not a runtime fallback.

#### Down migration responsibilities

- drop preview tables,
- drop the `group_id` index,
- drop `group_id` from `transaction_upload`.

## HTTP response behavior

On success:

- status `200 OK`,
- response body `data` contains the preview fields listed above,
- `validation_errors` may be an empty array for successful uploads.

Errors:

- missing upload -> `404 Not Found` with exact issue contract,
- wrong group -> `403 Forbidden`,
- unexpected repository or event publication failure -> `500 Internal Server Error`.

## Testing strategy

### Unit tests

#### Domain

- `TransactionUpload` constructor validates and exposes `groupID`,
- `RecordPreviewed` emits the expected admin event.

#### Application

- create use case persists preview artifacts for valid uploads,
- create use case persists preview artifacts for failed uploads with validation errors,
- create use case persists empty preview rows for uploads with zero eligible rows,
- preview succeeds for same-group actor,
- preview returns empty validation errors for successful uploads,
- preview returns `NotFoundError` for unknown upload ID,
- preview returns `ForbiddenError` for group mismatch,
- actor validation failure is propagated,
- preview event publication failures are wrapped and returned.

#### Ports contract tests

- add the new inbound preview port and outbound preview repository to `internal/ports/ports_test.go`.

#### Adapters

- HTTP route registration and success path for preview,
- HTTP 403 mapping for forbidden preview,
- HTTP 404 mapping for missing upload using the exact issue contract,
- preview JSON contains the required fields and validation-error members,
- Postgres upload repository persists and hydrates `group_id`,
- Postgres preview repository persists and loads columns, rows, total row count, and ordered validation errors.

### Migration tests

Extend `internal/infrastructure/db/migrate_test.go` with a migration-file existence/content test covering:

- `group_id` backfill to the seeded group,
- creation of the preview tables,
- teardown in the down migration.

## Out of scope

- download endpoints,
- Excel generation or highlighted file exports,
- preview pagination,
- adding skipped-row details to the preview payload,
- changing the upload request shape,
- frontend changes.

## Implementation notes

- Follow the existing transaction-upload naming patterns.
- Keep authorization in the use case, not the adapter.
- Keep preview persistence behind a dedicated outbound port.
- Keep preview rows aligned to the original parsed header order.
- Treat legacy ownership backfill as a migration concern, not a runtime conditional.
