# Issue 113 Download Transaction Upload Design

## Summary

Add `GET /api/v1/transaction-uploads/:id/download` so clients can download a previously uploaded raw transaction file as an attachment.

The endpoint must:

- allow downloads for both `uploaded` and `failed` transaction uploads,
- restrict access to callers in the same group as the upload,
- return attachment metadata using the original filename and inferred content type,
- return clear not-found and forbidden responses,
- record a dedicated successful download admin event.

## Why

The upload flow already stores raw files and persists storage references, but the API currently exposes only upload metadata and accepted transaction rows. Users need a reliable server-side way to retrieve uploaded content after the initial upload flow.

## Constraints and decisions

- Authorization model: **same-group access**.
- Route shape: **`GET /api/v1/transaction-uploads/:id/download`**.
- Eligible upload statuses: **both `uploaded` and `failed`**.
- Successful downloads must emit a dedicated admin event.
- MIME persistence is out of scope. The response will infer `Content-Type` from the original filename or file format and fall back to `application/octet-stream`.
- Work stays on the current branch.

## Existing codebase observations

The current transaction upload slice already has the major foundations needed for this feature:

- `CreateTransactionUploadUseCase` stores raw files through `ports.RawFileStore`.
- `TransactionUpload` persists `storage_provider` and `storage_key`.
- Local and Azure Blob adapters already support storing and deleting raw files.
- The HTTP upload adapter already owns error mapping and route registration for the upload slice.

However, the current model does **not** persist the upload's owning group, so same-group authorization cannot be enforced reliably for all uploads from existing data alone, especially failed uploads with zero accepted transaction rows.

## Recommended architecture

### 1. Persist upload group ownership explicitly

Add `group_id` to `transaction_upload` and make it the authorization source for download access.

This keeps authorization explicit and allows the core application to determine access without guessing from transaction rows or audit history.

### 2. Add a download use case and inbound port

Introduce a new download use case dedicated to returning a raw file for an existing upload:

- validate actor membership through `ActorDirectory`,
- load upload metadata,
- check `upload.GroupID()` against the caller's group,
- read bytes from raw storage,
- record a dedicated admin event,
- return a transport-agnostic download result.

This keeps the application layer responsible for orchestration and authorization while keeping the HTTP adapter focused on response shaping.

### 3. Extend the raw file boundary with read support

The current `RawFileStore` boundary is write/delete only. Downloading requires a core-facing read contract.

The simplest design is to extend the existing raw file storage port with a read method because the same adapters already own provider-specific storage translation.

The read contract should return:

- raw bytes,
- storage-level content type when available,
- not-found behavior that the use case can map cleanly.

### 4. Keep HTTP-specific download mechanics in the adapter

The HTTP adapter should:

- register the new route,
- build the query from route params and actor headers,
- call the download port,
- set `Content-Disposition`, `Content-Type`, and `Content-Length`,
- write the file body,
- map application errors to 403/404/500.

## Detailed design

### Domain layer

#### `TransactionUpload`

Extend the entity to carry `groupID`.

Changes:

- add a `groupID valueobjects.GroupID` field,
- update constructor and reconstitution signatures,
- validate group ID in `NewTransactionUpload`,
- expose `GroupID() valueobjects.GroupID` accessor,
- add a `RecordDownloaded(now, actorUserID, actorGroupID string)` method that records a dedicated admin action event.

This event payload should include:

- `action: "download"`,
- `resource: "transaction_upload"`,
- `upload_id`,
- `file_name`.

#### Admin event types

Add a new event type constant:

- `DownloadTransactionUploadEventType = "DownloadTransactionUpload"`.

### Ports layer

#### New inbound port

Add a new inbound port file for the download operation.

Suggested contract:

- `DownloadTransactionUploadQuery` with `ID`, `ActorUserID`, `ActorGroupID`
- `DownloadTransactionUploadResult` with `FileName`, `ContentType`, `ContentLength`, `FileBytes`
- `DownloadTransactionUploadPort`

#### Raw file storage contract

Extend `ports.RawFileStore` with read support.

Suggested additions:

- `ReadRawFileCommand { Key string }`
- `ReadRawFileResult { FileBytes []byte; ContentType string }`
- `Read(ctx, command ReadRawFileCommand) (ReadRawFileResult, error)`

This keeps the boundary transport-agnostic and provider-neutral.

#### Transaction upload result contract

Extend `ports.TransactionUploadResult` with `GroupID string` so repository hydration and test assertions can move through application-facing DTOs when needed.

This field is internal-facing API output rather than public HTTP output unless the adapter deliberately exposes it. The current download feature does not require returning it to clients, but carrying it in port-layer DTOs simplifies consistency across use cases and tests.

### Application layer

#### New application error

Add `ForbiddenError` to `internal/application/use-cases/errors.go`.

Suggested shape:

- `Resource string`
- `Reason string`

This keeps 403 mapping in the adapter while giving the use case a transport-agnostic way to report same-group access denial.

#### `DownloadTransactionUploadUseCase`

Dependencies:

- `ports.TransactionUploadRepository`
- `ports.RawFileStore`
- `adminEventRecorder`
- `ports.ActorDirectory`
- `now func() time.Time`

Workflow:

1. validate actor via `validateActor`
2. parse upload ID
3. load upload by ID
4. return `NotFoundError` if missing
5. compare `upload.GroupID().String()` to `query.ActorGroupID`
6. return `ForbiddenError` on mismatch
7. call `rawFileStore.Read` using the persisted storage key
8. map missing underlying file to `NotFoundError`
9. record `RecordDownloaded`
10. publish the resulting domain event
11. derive outward content type:
   - prefer storage-returned type if present,
   - otherwise infer from filename or known file format,
   - otherwise `application/octet-stream`
12. return `DownloadTransactionUploadResult`

The use case remains framework-free and does not know anything about HTTP attachment semantics.

### Adapters layer

#### Postgres upload repository

Extend SQL statements to read and write `group_id`.

Changes:

- add `group_id` to insert statement,
- add `group_id` to `SELECT` projections,
- update scan/reconstitution logic,
- update list-query tests and repository tests accordingly.

#### Raw file adapters

Extend both raw file adapters to implement `Read`.

##### Local storage

- read the file from the stored key path,
- treat missing files as a clean not-found condition the use case can interpret,
- derive content type using filename extension or `http.DetectContentType`.

##### Azure Blob storage

- download the blob by container + key,
- surface blob not found distinctly enough for the use case to map to not found,
- return any provider-reported content type when available.

#### HTTP transaction upload adapter

Extend the adapter constructor to accept the new download port and register the route:

- `GET /api/v1/transaction-uploads/:id/download`

Handler behavior:

1. build `DownloadTransactionUploadQuery` from route + actor headers,
2. call the download port,
3. set attachment headers,
4. write bytes with status 200.

Error mapping additions:

- `ForbiddenError` -> HTTP 403 with `{ code: "forbidden", message: ... }`

The current adapter already handles `NotFoundError`, `ConflictError`, and domain errors, so this is a small extension of the existing pattern.

### Infrastructure layer

#### Dependency injection

Wire the new use case into `internal/infrastructure/di/wire.go` and pass it into `NewHttpTransactionUploadAdapter`.

No new runtime configuration is needed because the feature reuses the existing raw storage provider selection.

### Database migration

Add a new migration after `000207`.

#### Up migration responsibilities

1. add nullable `group_id` column to `transaction_upload`
2. backfill from related transaction creators where inferable
3. backfill any remaining nulls with the seeded `superadmin` group ID:
   - `01962b8f-aeb2-7e03-a8ff-1edce1300001`
4. set `group_id` to `NOT NULL`
5. add a foreign key to `user_group(id)` if compatible with existing column type
6. add an index on `group_id`

#### Backfill strategy

Because the current deployed data uses the seeded group only, the explicit fallback to the seeded `superadmin` group is acceptable for legacy uploads that cannot infer ownership from other records.

This fallback is intentional and deterministic, not best-effort: unresolved legacy uploads are assigned to `01962b8f-aeb2-7e03-a8ff-1edce1300001` by design assumption so every legacy row becomes auditable under the only group present in deployed data.

This avoids leaving old failed uploads inaccessible while still establishing explicit ownership for all rows going forward.

#### Down migration responsibilities

- drop index / foreign key if added,
- drop `group_id` column.

### HTTP response behavior

On success:

- status `200 OK`
- `Content-Disposition: attachment; filename="<original file name>"`
- `Content-Type: <derived or provider content type>`
- `Content-Length: <byte count>`

Errors:

- missing upload -> 404
- missing raw file -> 404
- wrong group -> 403
- unexpected storage or event publication failure -> 500

### Documentation updates

Update the concrete runtime docs for the transaction-upload slice to mention the download endpoint and the added read responsibility on the raw file storage port.

Targets:

- `docs/internal-request-flow-transaction-upload.md`
- `docs/internal-wiring-overview.md`

## Testing strategy

### Unit tests

#### Domain

- `TransactionUpload` constructor validates and exposes `groupID`
- `RecordDownloaded` emits the expected admin event

#### Application

- download succeeds for same-group actor
- missing upload returns `NotFoundError`
- group mismatch returns `ForbiddenError`
- missing raw file returns `NotFoundError`
- actor validation failure is propagated
- download event publication failure is wrapped and returned

#### Ports contract tests

- add the new inbound port and the expanded raw file port to `internal/ports/ports_test.go`

#### Adapters

- HTTP route registration and success path
- HTTP 403 mapping for forbidden download
- HTTP 404 mapping for missing upload/file
- attachment headers are set correctly
- local raw file adapter read success + missing file
- Azure blob adapter read success + missing blob
- Postgres repository persists and hydrates `group_id`

### Migration tests

Extend `internal/infrastructure/db/migrate_test.go` with a new migration-file existence/content test for the `group_id` migration, following the existing migration test pattern.

That test should assert both the seeded-group fallback SQL and nearby explanatory comments so the legacy ownership assumption remains visible and reviewable on this branch.

## Out of scope

- changing the upload request shape,
- owner-only authorization,
- MIME-type schema persistence,
- streaming download support,
- multi-file or bulk download,
- frontend changes.

## Implementation notes

- Follow the existing transaction-upload naming patterns.
- Keep authorization in the use case, not the adapter.
- Keep provider-specific file retrieval in outbound adapters.
- Treat legacy backfill as a migration concern, not a runtime conditional.
