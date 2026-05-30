# Transaction Upload Download Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a same-group-protected `GET /api/v1/transaction-uploads/:id/download` endpoint that returns uploaded raw files as attachments and records a download audit event.

**Architecture:** Extend the transaction upload aggregate and persistence model with explicit `group_id` ownership, add a dedicated download use case plus inbound port, and extend the raw file storage boundary with read support. Keep authorization and event publication in the application layer, keep storage reads in outbound adapters, and keep attachment/header handling in the HTTP adapter.

**Tech Stack:** Go, Gin, pgx, golang-migrate SQL migrations, local filesystem raw storage, Azure Blob raw storage, standard testing package.

---

## File map

### Domain

- Modify: `internal/domain/entities/transaction_upload.go` — add `groupID`, constructor/reconstitution changes, and download event recording.
- Modify: `internal/domain/entities/transaction_upload_test.go` — cover `groupID` invariants and download event emission.
- Modify: `internal/domain/events/admin_event_types.go` — add `DownloadTransactionUploadEventType`.

### Ports

- Modify: `internal/ports/raw_file_store.go` — add read command/result and `Read` to `RawFileStore`.
- Add: `internal/ports/transaction_upload_download_port.go` — new query/result/port contract.
- Modify: `internal/ports/transaction_upload_create_port.go` — add `GroupID` to `TransactionUploadResult`.
- Modify: `internal/ports/ports_test.go` — add stubs/contracts for the new port and expanded raw-file contract.

### Application

- Modify: `internal/application/use-cases/errors.go` — add `ForbiddenError`.
- Add: `internal/application/use-cases/transaction_upload_download_use_case.go` — download orchestration.
- Add: `internal/application/use-cases/transaction_upload_download_use_case_test.go` — TDD coverage for authorization, storage read, and audit event behavior.
- Modify: `internal/application/use-cases/upload_result_mapper.go` — map `GroupID` from entity to result DTO.

### Adapters

- Modify: `internal/adapters/raw_file_local_store.go` — implement `Read`.
- Modify: `internal/adapters/raw_file_local_store_test.go` — add read coverage.
- Modify: `internal/adapters/raw_file_azure_blob_store.go` — implement `Read` and Azure blob download client support.
- Modify: `internal/adapters/raw_file_azure_blob_store_test.go` — add read coverage.
- Modify: `internal/adapters/transaction_upload_postgres_repository.go` — persist and hydrate `group_id`.
- Modify: `internal/adapters/transaction_upload_postgres_repository_test.go` — update SQL expectations and hydration coverage.
- Modify: `internal/adapters/transaction_upload_http_adapter.go` — inject new port, add route/handler, map forbidden errors.
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go` — add download route/header/error tests.

### Infrastructure and docs

- Modify: `internal/infrastructure/di/wire.go` — wire the new use case into the HTTP adapter.
- Add: `internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.up.sql`
- Add: `internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.down.sql`
- Modify: `internal/infrastructure/db/migrate_test.go` — verify migration file contents.
- Modify: `docs/internal-request-flow-transaction-upload.md` — document the new download path.
- Modify: `docs/internal-wiring-overview.md` — mention download in the upload slice wiring map.

---

### Task 1: Add the core domain and port contracts

**Files:**
- Modify: `internal/domain/entities/transaction_upload.go`
- Modify: `internal/domain/entities/transaction_upload_test.go`
- Modify: `internal/domain/events/admin_event_types.go`
- Modify: `internal/ports/raw_file_store.go`
- Add: `internal/ports/transaction_upload_download_port.go`
- Modify: `internal/ports/transaction_upload_create_port.go`
- Modify: `internal/ports/ports_test.go`

- [ ] **Step 1: Write the failing tests**

Add table-driven assertions in `internal/domain/entities/transaction_upload_test.go` for:

```go
func TestNewTransactionUploadStoresGroupID(t *testing.T) {
    t.Parallel()

    uploadID, err := valueobjects.UploadIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300201")
    if err != nil {
        t.Fatalf("UploadIDFromString() error = %v", err)
    }
    groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
    if err != nil {
        t.Fatalf("GroupIDFromString() error = %v", err)
    }
    uploadedAt := time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC)

    upload, err := NewTransactionUpload(
        uploadID,
        groupID,
        "transactions.csv",
        "csv",
        "0123456789abcdef0123456789abcdef",
        "local",
        "01962b8f-aeb2-7e03-a8ff-1edce1300201/transactions.csv",
        "2026-05",
        valueobjects.UploadedTransactionUploadStatus(),
        2,
        uploadedAt,
    )
    if err != nil {
        t.Fatalf("NewTransactionUpload() error = %v", err)
    }
    if got := upload.GroupID().String(); got != groupID.String() {
        t.Fatalf("upload.GroupID() = %q, want %q", got, groupID.String())
    }
}

func TestTransactionUploadRecordDownloaded(t *testing.T) {
    t.Parallel()

    uploadID, err := valueobjects.UploadIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300201")
    if err != nil {
        t.Fatalf("UploadIDFromString() error = %v", err)
    }
    groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
    if err != nil {
        t.Fatalf("GroupIDFromString() error = %v", err)
    }

    upload := ReconstituteTransactionUpload(
        uploadID,
        groupID,
        "transactions.csv",
        "csv",
        "0123456789abcdef0123456789abcdef",
        "local",
        "01962b8f-aeb2-7e03-a8ff-1edce1300201/transactions.csv",
        "2026-05",
        valueobjects.UploadedTransactionUploadStatus(),
        2,
        time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
    )

    now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
    if err := upload.RecordDownloaded(now, "01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300001"); err != nil {
        t.Fatalf("RecordDownloaded() error = %v", err)
    }

    events := upload.PullDomainEvents()
    if len(events) != 1 {
        t.Fatalf("len(events) = %d, want %d", len(events), 1)
    }
}
```

Add failing contract assertions in `internal/ports/ports_test.go` for the new port and expanded `RawFileStore`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/entities ./internal/ports -count=1`

Expected: FAIL because `GroupID`, `RecordDownloaded`, the new download port, and `RawFileStore.Read` do not exist yet.

- [ ] **Step 3: Write the minimal implementation**

Implement:

- `groupID` field/accessor/validation in `TransactionUpload`
- `RecordDownloaded`
- `DownloadTransactionUploadEventType`
- read contract in `raw_file_store.go`
- `transaction_upload_download_port.go`
- `GroupID string` on `TransactionUploadResult`
- port test stubs updated to satisfy the new interfaces

Key signatures to add:

```go
type ReadRawFileCommand struct {
    Key string
}

type ReadRawFileResult struct {
    FileBytes   []byte
    ContentType string
}

type RawFileStore interface {
    Store(ctx context.Context, command StoreRawFileCommand) (StoreRawFileResult, error)
    Read(ctx context.Context, command ReadRawFileCommand) (ReadRawFileResult, error)
    Delete(ctx context.Context, command DeleteRawFileCommand) error
}
```

and:

```go
type DownloadTransactionUploadQuery struct {
    ID           string
    ActorUserID  string
    ActorGroupID string
}

type DownloadTransactionUploadResult struct {
    FileName      string
    ContentType   string
    ContentLength int
    FileBytes     []byte
}

type DownloadTransactionUploadPort interface {
    Execute(ctx context.Context, query DownloadTransactionUploadQuery) (DownloadTransactionUploadResult, error)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/domain/entities ./internal/ports -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/entities/transaction_upload.go internal/domain/entities/transaction_upload_test.go internal/domain/events/admin_event_types.go internal/ports/raw_file_store.go internal/ports/transaction_upload_download_port.go internal/ports/transaction_upload_create_port.go internal/ports/ports_test.go
git commit -m "feat(transaction-upload): add download domain and port contracts"
```

### Task 2: Implement the download use case with TDD

**Files:**
- Modify: `internal/application/use-cases/errors.go`
- Add: `internal/application/use-cases/transaction_upload_download_use_case.go`
- Add: `internal/application/use-cases/transaction_upload_download_use_case_test.go`
- Modify: `internal/application/use-cases/upload_result_mapper.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/application/use-cases/transaction_upload_download_use_case_test.go` with cases for:

```go
func TestDownloadTransactionUploadUseCaseExecute(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name        string
        query       ports.DownloadTransactionUploadQuery
        upload      *entities.TransactionUpload
        uploadErr   error
        readResult  ports.ReadRawFileResult
        readErr     error
        actorErr    error
        publishErr  error
        wantErr     string
        wantType    string
        wantLength  int
    }{
        {
            name: "downloads upload for same group",
            query: ports.DownloadTransactionUploadQuery{
                ID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001",
            },
            upload: existingUpload(t, "01962b8f-aeb2-7e03-a8ff-1edce1300201", "01962b8f-aeb2-7e03-a8ff-1edce1300001"),
            readResult: ports.ReadRawFileResult{FileBytes: []byte("a,b\n1,2\n"), ContentType: "text/csv"},
            wantType: "text/csv",
            wantLength: len([]byte("a,b\n1,2\n")),
        },
        {
            name: "returns forbidden for different group",
            query: ports.DownloadTransactionUploadQuery{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300003"},
            upload: existingUpload(t, "01962b8f-aeb2-7e03-a8ff-1edce1300201", "01962b8f-aeb2-7e03-a8ff-1edce1300001"),
            wantErr: "forbidden",
        },
        {
            name: "returns not found when upload missing",
            query: ports.DownloadTransactionUploadQuery{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
            wantErr: "was not found",
        },
        {
            name: "returns not found when raw file missing",
            query: ports.DownloadTransactionUploadQuery{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
            upload: existingUpload(t, "01962b8f-aeb2-7e03-a8ff-1edce1300201", "01962b8f-aeb2-7e03-a8ff-1edce1300001"),
            readErr: os.ErrNotExist,
            wantErr: "was not found",
        },
    }
}

func existingUpload(t *testing.T, uploadIDValue string, groupIDValue string) *entities.TransactionUpload {
    t.Helper()

    uploadID, err := valueobjects.UploadIDFromString(uploadIDValue)
    if err != nil {
        t.Fatalf("UploadIDFromString(%q) error = %v", uploadIDValue, err)
    }
    groupID, err := valueobjects.GroupIDFromString(groupIDValue)
    if err != nil {
        t.Fatalf("GroupIDFromString(%q) error = %v", groupIDValue, err)
    }

    return entities.ReconstituteTransactionUpload(
        uploadID,
        groupID,
        "transactions.csv",
        "csv",
        "0123456789abcdef0123456789abcdef",
        "local",
        uploadIDValue+"/transactions.csv",
        "2026-05",
        valueobjects.UploadedTransactionUploadStatus(),
        2,
        time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
    )
}
```

Also add a test that verifies filename-based fallback content type becomes `application/octet-stream` when inference fails.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/application/use-cases -run TestDownloadTransactionUploadUseCaseExecute -count=1`

Expected: FAIL because the use case and `ForbiddenError` do not exist yet.

- [ ] **Step 3: Write the minimal implementation**

Implement:

- `ForbiddenError` in `errors.go`
- `DownloadTransactionUploadUseCase`
- filename/content-type inference helper in the same file or a focused helper
- mapping of `GroupID` in `upload_result_mapper.go`

Use this structure:

```go
type DownloadTransactionUploadUseCase struct {
    uploadRepository ports.TransactionUploadRepository
    rawFileStore     ports.RawFileStore
    eventRecorder    adminEventRecorder
    actorDirectory   ports.ActorDirectory
    now              func() time.Time
}
```

Core behavior:

- validate actor with `validateActor`
- parse upload ID with `valueobjects.UploadIDFromString`
- load upload and 404 if nil
- compare upload group to actor group and return `&ForbiddenError{Resource: "transaction upload", Reason: "actor is not allowed to download this upload"}` on mismatch
- call `rawFileStore.Read`
- convert missing underlying file to `NotFoundError`
- call `upload.RecordDownloaded(...)`
- publish events with `publishDomainEvents`
- return bytes + content type + length

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/application/use-cases -run TestDownloadTransactionUploadUseCaseExecute -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/application/use-cases/errors.go internal/application/use-cases/transaction_upload_download_use_case.go internal/application/use-cases/transaction_upload_download_use_case_test.go internal/application/use-cases/upload_result_mapper.go
git commit -m "feat(transaction-upload): add download use case"
```

### Task 3: Add storage read support to local and Azure adapters

**Files:**
- Modify: `internal/adapters/raw_file_local_store.go`
- Modify: `internal/adapters/raw_file_local_store_test.go`
- Modify: `internal/adapters/raw_file_azure_blob_store.go`
- Modify: `internal/adapters/raw_file_azure_blob_store_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests for local store:

```go
func TestLocalRawFileStoreRead(t *testing.T) {
    t.Parallel()

    dir := t.TempDir()
    store := NewLocalRawFileStore(dir)

    _, err := store.Store(context.Background(), ports.StoreRawFileCommand{
        UploadID: "01962b8f-aeb2-7e03-a8ff-1edce1300201",
        FileName: "transactions.csv",
        FileBytes: []byte("a,b\n1,2\n"),
    })
    if err != nil {
        t.Fatalf("Store() error = %v", err)
    }

    got, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "01962b8f-aeb2-7e03-a8ff-1edce1300201/transactions.csv"})
    if err != nil {
        t.Fatalf("Read() error = %v", err)
    }
    if string(got.FileBytes) != "a,b\n1,2\n" {
        t.Fatalf("Read().FileBytes = %q, want %q", string(got.FileBytes), "a,b\n1,2\n")
    }
}
```

and missing-file coverage:

```go
func TestLocalRawFileStoreReadMissingFile(t *testing.T) {
    t.Parallel()
    store := NewLocalRawFileStore(t.TempDir())
    _, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "missing.csv"})
    if err == nil {
        t.Fatal("Read() error = nil, want error")
    }
}
```

For Azure, add a stubbed download test that verifies key/container usage and missing-blob behavior.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapters -run "Test(LocalRawFileStoreRead|AzureBlobRawFileStoreRead)" -count=1`

Expected: FAIL because `Read` does not exist.

- [ ] **Step 3: Write the minimal implementation**

Implement:

- `Read` on `LocalRawFileStore`
- `DownloadBuffer` or equivalent read method on the Azure client abstraction
- `Read` on `AzureBlobRawFileStore`

Suggested local pattern:

```go
func (s *LocalRawFileStore) Read(_ context.Context, command ports.ReadRawFileCommand) (ports.ReadRawFileResult, error) {
    targetPath := filepath.Join(s.basePath, filepath.FromSlash(command.Key))
    fileBytes, err := os.ReadFile(targetPath)
    if err != nil {
        return ports.ReadRawFileResult{}, fmt.Errorf("reading local raw file: %w", err)
    }
    return ports.ReadRawFileResult{
        FileBytes: fileBytes,
        ContentType: http.DetectContentType(fileBytes),
    }, nil
}
```

For Azure, return provider content type when available; otherwise leave empty and let the use case infer/fallback.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapters -run "Test(LocalRawFileStoreRead|AzureBlobRawFileStoreRead)" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/raw_file_local_store.go internal/adapters/raw_file_local_store_test.go internal/adapters/raw_file_azure_blob_store.go internal/adapters/raw_file_azure_blob_store_test.go
git commit -m "feat(storage): add raw file read support"
```

### Task 4: Persist upload group ownership and add migration coverage

**Files:**
- Modify: `internal/adapters/transaction_upload_postgres_repository.go`
- Modify: `internal/adapters/transaction_upload_postgres_repository_test.go`
- Add: `internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.up.sql`
- Add: `internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.down.sql`
- Modify: `internal/infrastructure/db/migrate_test.go`

- [ ] **Step 1: Write the failing tests**

Update repository tests so they expect `group_id` in inserts/selects.

Example expectation change:

```go
rows := pgxmock.NewRows([]string{"id", "group_id", "file_name", "file_format", "content_md5", "storage_provider", "storage_key", "schema_version", "status", "row_count", "uploaded_at"}).
    AddRow(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt())
```

Add a new migration-file test in `migrate_test.go` modeled on `TestTransactionUploadStatusMigrationFilesExist` that checks for:

- `ALTER TABLE transaction_upload ADD COLUMN group_id`
- backfill from existing records
- intentional fallback to `01962b8f-aeb2-7e03-a8ff-1edce1300001` with an adjacent audit comment explaining that current deployed legacy data uses the seeded superadmin group only
- `ALTER TABLE transaction_upload ALTER COLUMN group_id SET NOT NULL`
- index creation

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapters ./internal/infrastructure/db -run "Test(PostgresTransactionUploadRepository|TransactionUploadGroupMigrationFilesExist)" -count=1`

Expected: FAIL because repository SQL and migration files have not been updated yet.

- [ ] **Step 3: Write the minimal implementation**

Implement:

- repository SQL updates for `group_id`
- scan/reconstitution changes
- migration SQL files
- migration-file assertions in `migrate_test.go`

Migration up file should include a deterministic fallback:

```sql
-- Legacy uploads that still cannot infer ownership are intentionally assigned
-- to the seeded superadmin group because deployed legacy data uses that group only.
UPDATE transaction_upload
SET group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001'
WHERE group_id IS NULL;
```

If you backfill from transaction creators first, keep that statement before the fallback update.

Do not change that fallback target on this branch unless the underlying data assumption changes and the user explicitly approves the new migration semantics.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapters ./internal/infrastructure/db -run "Test(PostgresTransactionUploadRepository|TransactionUploadGroupMigrationFilesExist)" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/transaction_upload_postgres_repository.go internal/adapters/transaction_upload_postgres_repository_test.go internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.up.sql internal/infrastructure/db/migrations/000208_add_transaction_upload_group_and_download_support.down.sql internal/infrastructure/db/migrate_test.go
git commit -m "feat(transaction-upload): persist upload group ownership"
```

### Task 5: Add the HTTP download endpoint, DI wiring, and docs

**Files:**
- Modify: `internal/adapters/transaction_upload_http_adapter.go`
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go`
- Modify: `internal/infrastructure/di/wire.go`
- Modify: `docs/internal-request-flow-transaction-upload.md`
- Modify: `docs/internal-wiring-overview.md`

- [ ] **Step 1: Write the failing tests**

Extend `internal/adapters/transaction_upload_http_adapter_test.go` with cases for:

```go
{
    name: "downloads upload",
    method: http.MethodGet,
    target: "/api/v1/transaction-uploads/" + transactionUploadID1 + "/download",
    downloadStub: &downloadTransactionUploadPortStub{result: ports.DownloadTransactionUploadResult{
        FileName: "transactions.csv",
        ContentType: "text/csv",
        ContentLength: len([]byte("a,b\n1,2\n")),
        FileBytes: []byte("a,b\n1,2\n"),
    }},
    wantStatus: http.StatusOK,
}
```

and:

- forbidden download -> 403 with `code = "forbidden"`
- missing download -> 404

Assert response headers:

- `Content-Disposition` contains `attachment; filename="transactions.csv"`
- `Content-Type` contains `text/csv`
- body equals the bytes returned by the port

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/adapters -run TestHttpTransactionUploadAdapterRoutes -count=1`

Expected: FAIL because the adapter does not have a download port, route, or forbidden mapping.

- [ ] **Step 3: Write the minimal implementation**

Implement:

- add `downloadUpload ports.DownloadTransactionUploadPort` field to the adapter
- update constructor signature and DI wiring
- register `uploads.GET(":id/download", a.DownloadUpload)`
- implement `DownloadUpload`
- extend `handleError` to map `ForbiddenError` to 403
- update docs to mention the new route and read-capable raw file storage boundary

Suggested handler shape:

```go
func (a *HttpTransactionUploadAdapter) DownloadUpload(c *gin.Context) {
    result, err := a.downloadUpload.Execute(c.Request.Context(), ports.DownloadTransactionUploadQuery{
        ID: c.Param("id"),
        ActorUserID: c.GetHeader(actorUserIDHeader),
        ActorGroupID: c.GetHeader(actorGroupIDHeader),
    })
    if err != nil {
        a.handleError(c, err)
        return
    }

    c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", result.FileName))
    c.Data(http.StatusOK, result.ContentType, result.FileBytes)
}
```

If `c.Data` does not preserve the explicit content length you want, set the `Content-Length` header before writing.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapters ./internal/infrastructure/di -count=1`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/transaction_upload_http_adapter.go internal/adapters/transaction_upload_http_adapter_test.go internal/infrastructure/di/wire.go docs/internal-request-flow-transaction-upload.md docs/internal-wiring-overview.md
git commit -m "feat(transaction-upload): add download endpoint"
```

### Task 6: Final verification pass

**Files:**
- Verify only

- [ ] **Step 1: Run targeted package verification**

Run: `go test ./internal/domain/entities ./internal/ports ./internal/application/use-cases ./internal/adapters ./internal/infrastructure/db ./internal/infrastructure/di -count=1`

Expected: PASS.

- [ ] **Step 2: Run repository-wide tests**

Run: `go test ./... -count=1`

Expected: PASS.

- [ ] **Step 3: Run build verification**

Run: `go build ./...`

Expected: PASS.

- [ ] **Step 4: Review working tree**

Run: `git status --short`

Expected: only intended tracked changes, or clean if each task commit was created successfully.

- [ ] **Step 5: Commit any final fixups if needed**

```bash
git add -A
git commit -m "test(transaction-upload): finalize download coverage"
```

Only do this if verification required a final code change. If no fixups were needed, skip creating an extra commit.
