# Transaction Upload Preview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `GET /api/v1/transaction-uploads/:id/preview` so same-group callers can retrieve persisted upload preview rows and schema validation errors for both successful and failed transaction uploads.

**Architecture:** Persist upload ownership and preview artifacts at create time, then serve them through a dedicated preview query use case. Keep authorization and event recording in the application layer, store preview artifacts behind a dedicated outbound repository, and keep HTTP-specific error/JSON shaping inside the transaction upload adapter.

**Tech Stack:** Go, Gin, pgx/pgxmock, PostgreSQL migrations, hexagonal architecture, standard `testing` package

---

## File structure

### Create
- `internal/ports/transaction_upload_preview_get_port.go` — inbound port for preview queries.
- `internal/ports/transaction_upload_preview_repository.go` — outbound port for persisted preview artifacts.
- `internal/application/use-cases/transaction_upload_preview_get_use_case.go` — preview query orchestration and same-group authorization.
- `internal/application/use-cases/transaction_upload_preview_get_use_case_test.go` — unit tests for preview use case behavior.
- `internal/adapters/transaction_upload_preview_postgres_repository.go` — Postgres implementation of preview persistence.
- `internal/adapters/transaction_upload_preview_postgres_repository_test.go` — repository tests for ordered preview persistence and hydration.
- `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.up.sql` — schema changes for `group_id` and preview tables.
- `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.down.sql` — rollback for the preview migration.

### Modify
- `internal/domain/entities/transaction_upload.go` — add `groupID`, `GroupID()`, and `RecordPreviewed()`.
- `internal/domain/entities/transaction_upload_test.go` — verify group ownership and preview events.
- `internal/domain/events/admin_event_types.go` — add `PreviewTransactionUploadEventType`.
- `internal/ports/transaction_upload_create_port.go` — extend upload DTOs with `GroupID`.
- `internal/ports/transaction_uploads_list_port.go` — reuse updated upload DTO in list/detail results.
- `internal/ports/ports_test.go` — contract assertions for new preview ports/repositories.
- `internal/application/use-cases/errors.go` — add `ForbiddenError`.
- `internal/application/use-cases/upload_result_mapper.go` — include `GroupID` in upload DTO mapping.
- `internal/application/use-cases/transaction_upload_create_use_case.go` — persist preview artifacts during upload creation.
- `internal/application/use-cases/transaction_upload_create_use_case_test.go` — verify preview persistence for success/failure paths.
- `internal/domain/services/transaction_upload_factory.go` — pass `groupID` into uploaded upload construction.
- `internal/application/use-cases/transaction_upload_get_use_case_test.go` — update reconstitution helpers for `groupID`.
- `internal/adapters/transaction_upload_postgres_repository.go` — persist/hydrate `group_id`.
- `internal/adapters/transaction_upload_postgres_repository_test.go` — assert `group_id` SQL arguments and scans.
- `internal/adapters/transaction_upload_http_adapter.go` — register preview route and map preview-specific errors.
- `internal/adapters/transaction_upload_http_adapter_test.go` — endpoint and JSON contract tests.
- `internal/infrastructure/di/wire.go` — wire preview repository and use case into the HTTP adapter.
- `internal/infrastructure/db/migrate_test.go` — verify new migration files and required SQL fragments.

---

### Task 1: Add ownership + preview contracts in domain and ports

**Files:**
- Create: `internal/ports/transaction_upload_preview_get_port.go`
- Create: `internal/ports/transaction_upload_preview_repository.go`
- Modify: `internal/domain/entities/transaction_upload.go`
- Modify: `internal/domain/entities/transaction_upload_test.go`
- Modify: `internal/domain/events/admin_event_types.go`
- Modify: `internal/ports/transaction_upload_create_port.go`
- Modify: `internal/ports/ports_test.go`

- [ ] **Step 1: Write the failing domain/port tests first**

Add a new domain test case in `internal/domain/entities/transaction_upload_test.go` proving the upload stores a group ID and emits a preview event.

```go
func TestTransactionUploadStoresGroupAndRecordsPreview(t *testing.T) {
	t.Parallel()

	uploadID := mustUploadID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	upload, err := NewTransactionUpload(
		uploadID,
		groupID,
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.csv",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		1,
		now,
	)
	if err != nil {
		t.Fatalf("NewTransactionUpload() error = %v", err)
	}

	if upload.GroupID().String() != groupID.String() {
		t.Fatalf("upload.GroupID() = %q, want %q", upload.GroupID().String(), groupID.String())
	}

	if err := upload.RecordPreviewed(now, "01962b8f-aeb2-7e03-a8ff-1edce1300002", groupID.String()); err != nil {
		t.Fatalf("RecordPreviewed() error = %v", err)
	}

	events := upload.PullDomainEvents()
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
}
```

Add contract assertions in `internal/ports/ports_test.go`:

```go
type getTransactionUploadPreviewPortStub struct{}

func (getTransactionUploadPreviewPortStub) Execute(context.Context, GetTransactionUploadPreviewQuery) (GetTransactionUploadPreviewResult, error) {
	return GetTransactionUploadPreviewResult{}, nil
}

type transactionUploadPreviewRepositoryStub struct{}

func (transactionUploadPreviewRepositoryStub) Save(context.Context, TransactionUploadPreviewRecord) error {
	return nil
}

func (transactionUploadPreviewRepositoryStub) FindByUploadID(context.Context, valueobjects.UploadID) (*TransactionUploadPreviewRecord, error) {
	return nil, nil
}
```

- [ ] **Step 2: Run the focused tests to verify they fail**

Run:

```powershell
go test ./internal/domain/entities ./internal/ports -run 'TestTransactionUploadStoresGroupAndRecordsPreview|TestPorts' -count=1
```

Expected: FAIL with compile errors such as `too many arguments in call to NewTransactionUpload`, `upload.GroupID undefined`, `upload.RecordPreviewed undefined`, and missing preview port/repository types.

- [ ] **Step 3: Add the minimal domain and port contracts**

Create `internal/ports/transaction_upload_preview_get_port.go`:

```go
package ports

import "context"

type GetTransactionUploadPreviewQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

type GetTransactionUploadPreviewResult struct {
	FileID           string
	FileName         string
	Columns          []string
	Rows             [][]string
	TotalRows        int
	ValidationErrors []TransactionFileValidationError
}

type GetTransactionUploadPreviewPort interface {
	Execute(ctx context.Context, query GetTransactionUploadPreviewQuery) (GetTransactionUploadPreviewResult, error)
}
```

Create `internal/ports/transaction_upload_preview_repository.go`:

```go
package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type TransactionUploadPreviewRecord struct {
	UploadID         string
	Columns          []string
	Rows             [][]string
	TotalRows        int
	ValidationErrors []TransactionFileValidationError
}

type TransactionUploadPreviewRepository interface {
	Save(ctx context.Context, preview TransactionUploadPreviewRecord) error
	FindByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (*TransactionUploadPreviewRecord, error)
}
```

Update `internal/ports/transaction_upload_create_port.go`:

```go
type TransactionUploadResult struct {
	ID              string
	GroupID         string
	FileName        string
	FileFormat      string
	ContentMD5      string
	StorageProvider string
	StorageKey      string
	SchemaVersion   string
	Status          string
	RowCount        int
	UploadedAt      string
}
```

Update `internal/domain/events/admin_event_types.go`:

```go
const (
	GetTransactionUploadEventType     = "GetTransactionUpload"
	CreateTransactionUploadEventType  = "CreateTransactionUpload"
	DeleteTransactionUploadEventType  = "DeleteTransactionUpload"
	PreviewTransactionUploadEventType = "PreviewTransactionUpload"
)
```

Update `internal/domain/entities/transaction_upload.go` by adding the field, constructor parameter, accessor, and preview-event method.

```go
type TransactionUpload struct {
	aggregateRoot
	id              valueobjects.UploadID
	groupID         valueobjects.GroupID
	fileName        string
	// ...
}

func (u *TransactionUpload) GroupID() valueobjects.GroupID { return u.groupID }

func (u *TransactionUpload) RecordPreviewed(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.PreviewTransactionUploadEventType, map[string]any{
		"action":    "preview",
		"resource":  "transaction_upload",
		"upload_id": u.ID().String(),
		"file_name": u.FileName(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}
```

Also update `NewTransactionUpload` and `ReconstituteTransactionUpload` to accept `groupID valueobjects.GroupID` and validate it with `valueobjects.GroupIDFromString(groupID.String())`.

- [ ] **Step 4: Run the tests again to verify the new contracts pass**

Run:

```powershell
go test ./internal/domain/entities ./internal/ports -run 'TestTransactionUploadStoresGroupAndRecordsPreview|TestPorts' -count=1
```

Expected: PASS.

- [ ] **Step 5: Verification checkpoint**

Run:

```powershell
go test ./internal/domain/entities ./internal/ports -count=1
```

Expected: all tests in both packages pass. Do **not** commit unless the user explicitly asks.

---

### Task 2: Persist preview artifacts during upload creation

**Files:**
- Modify: `internal/domain/services/transaction_upload_factory.go`
- Modify: `internal/application/use-cases/upload_result_mapper.go`
- Modify: `internal/application/use-cases/transaction_upload_create_use_case.go`
- Modify: `internal/application/use-cases/transaction_upload_create_use_case_test.go`

- [ ] **Step 1: Write failing create-use-case tests for preview persistence**

In `internal/application/use-cases/transaction_upload_create_use_case_test.go`, add a preview repository mock and assert it receives create-time preview artifacts for success and validation-failure paths.

```go
type transactionUploadPreviewRepositoryMock struct {
	saved    []ports.TransactionUploadPreviewRecord
	saveErr  error
	findByID *ports.TransactionUploadPreviewRecord
	findErr  error
}

func (m *transactionUploadPreviewRepositoryMock) Save(_ context.Context, preview ports.TransactionUploadPreviewRecord) error {
	m.saved = append(m.saved, preview)
	return m.saveErr
}

func (m *transactionUploadPreviewRepositoryMock) FindByUploadID(context.Context, valueobjects.UploadID) (*ports.TransactionUploadPreviewRecord, error) {
	return m.findByID, m.findErr
}
```

Add assertions like:

```go
if len(previewRepo.saved) != 1 {
	t.Fatalf("len(previewRepo.saved) = %d, want 1", len(previewRepo.saved))
}
if diff := cmp.Diff(parsedResult.Headers, previewRepo.saved[0].Columns); diff != "" {
	t.Fatalf("preview columns mismatch (-want +got):\n%s", diff)
}
if previewRepo.saved[0].TotalRows != 1 {
	t.Fatalf("previewRepo.saved[0].TotalRows = %d, want 1", previewRepo.saved[0].TotalRows)
}
```

Add a failure-path assertion:

```go
if len(result.ValidationErrors) == 0 {
	t.Fatal("expected validation errors")
}
if len(previewRepo.saved[0].ValidationErrors) == 0 {
	t.Fatal("expected saved preview validation errors")
}
```

- [ ] **Step 2: Run the create-use-case tests to verify failure**

Run:

```powershell
go test ./internal/application/use-cases -run TestCreateTransactionUploadUseCaseExecute -count=1
```

Expected: FAIL with constructor/signature errors because the use case does not yet accept a preview repository and the factory does not yet accept a group ID.

- [ ] **Step 3: Implement the minimal upload-create changes**

Update `internal/domain/services/transaction_upload_factory.go` so `Build` accepts `groupID` and passes it into `entities.NewTransactionUpload`.

```go
func (f *TransactionUploadFactory) Build(
	uploadID valueobjects.UploadID,
	groupID valueobjects.GroupID,
	fileName string,
	fileFormat string,
	contentMD5 string,
	storageProvider string,
	storageKey string,
	schemaVersion string,
	headers []string,
	rows [][]string,
	uploadedAt time.Time,
	newTransactionID func() (valueobjects.TransactionID, error),
) (*entities.TransactionUpload, []*entities.Transaction, error) {
	// ...
	upload, err := entities.NewTransactionUpload(uploadID, groupID, fileName, fileFormat, contentMD5, storageProvider, storageKey, schemaVersion, valueobjects.UploadedTransactionUploadStatus(), len(transactions), uploadedAt)
```

Update `internal/application/use-cases/upload_result_mapper.go`:

```go
return ports.TransactionUploadResult{
	ID:              upload.ID().String(),
	GroupID:         upload.GroupID().String(),
	FileName:        upload.FileName(),
	// ...
}
```

Update `internal/application/use-cases/transaction_upload_create_use_case.go`:

```go
type CreateTransactionUploadUseCase struct {
	uploadRepository        ports.TransactionUploadRepository
	previewRepository       ports.TransactionUploadPreviewRepository
	transactionRepo         ports.TransactionRepository
	// ...
}

func NewCreateTransactionUploadUseCase(
	uploadRepository ports.TransactionUploadRepository,
	previewRepository ports.TransactionUploadPreviewRepository,
	transactionRepo ports.TransactionRepository,
	// ...
) *CreateTransactionUploadUseCase {
	return &CreateTransactionUploadUseCase{
		uploadRepository:  uploadRepository,
		previewRepository: previewRepository,
		transactionRepo:   transactionRepo,
		// ...
	}
}
```

Inside `Execute`, parse the actor group ID and save preview artifacts inside the transaction block.

```go
actorGroupID, err := valueobjects.GroupIDFromString(command.ActorGroupID)
if err != nil {
	return ports.CreateTransactionUploadResult{}, fmt.Errorf("parsing actor group id: %w", err)
}

previewRecord := ports.TransactionUploadPreviewRecord{
	UploadID:         uploadID.String(),
	Columns:          slices.Clone(parsedFile.Headers),
	Rows:             cloneStringMatrix(filteredRows.EligibleRows),
	TotalRows:        len(filteredRows.EligibleRows),
	ValidationErrors: toTransactionFileValidationErrors(validationReport.Errors()),
}
```

Save it in both the failed-upload and successful-upload transaction bodies:

```go
if err := uc.previewRepository.Save(txCtx, previewRecord); err != nil {
	return fmt.Errorf("saving transaction upload preview: %w", err)
}
```

Add a local helper in the same file for copying row matrices:

```go
func cloneStringMatrix(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}
	cloned := make([][]string, 0, len(rows))
	for _, row := range rows {
		cloned = append(cloned, slices.Clone(row))
	}
	return cloned
}
```

For the zero-eligible-row path, keep `Rows: nil` and `TotalRows: 0`.

- [ ] **Step 4: Run the create-use-case tests again**

Run:

```powershell
go test ./internal/application/use-cases -run TestCreateTransactionUploadUseCaseExecute -count=1
```

Expected: PASS.

- [ ] **Step 5: Verification checkpoint**

Run:

```powershell
go test ./internal/application/use-cases -run 'TestCreateTransactionUploadUseCaseExecute|TestGetTransactionUploadUseCaseExecute' -count=1
```

Expected: PASS, including updated helper calls that now provide `groupID`. Do **not** commit unless the user explicitly asks.

---

### Task 3: Add preview query use case and application error mapping

**Files:**
- Modify: `internal/application/use-cases/errors.go`
- Create: `internal/application/use-cases/transaction_upload_preview_get_use_case.go`
- Create: `internal/application/use-cases/transaction_upload_preview_get_use_case_test.go`

- [ ] **Step 1: Write the failing preview use-case tests**

Create `internal/application/use-cases/transaction_upload_preview_get_use_case_test.go` with cases for success, not found, wrong group, actor validation failure, preview-record missing, and event publication failure.

```go
func TestGetTransactionUploadPreviewUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID := mustUploadID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")

	upload := entities.ReconstituteTransactionUpload(
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

	preview := &ports.TransactionUploadPreviewRecord{
		UploadID:  uploadID.String(),
		Columns:   []string{"Product", "Year"},
		Rows:      [][]string{{"CG", "2026"}},
		TotalRows: 1,
		ValidationErrors: []ports.TransactionFileValidationError{{
			Code: "MISSING_FIELD", Message: "Year is required", RowNumber: 1, ColumnName: "Year", ColumnIndex: 2,
		}},
	}

	// assert success, NotFoundError, ForbiddenError, and wrapped publish error
}
```

- [ ] **Step 2: Run the preview use-case tests to verify failure**

Run:

```powershell
go test ./internal/application/use-cases -run TestGetTransactionUploadPreviewUseCaseExecute -count=1
```

Expected: FAIL because `ForbiddenError` and `GetTransactionUploadPreviewUseCase` do not yet exist.

- [ ] **Step 3: Implement the minimal preview query path**

Update `internal/application/use-cases/errors.go`:

```go
type ForbiddenError struct {
	Resource string
	Reason   string
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("%s forbidden: %s", e.Resource, e.Reason)
}
```

Create `internal/application/use-cases/transaction_upload_preview_get_use_case.go`:

```go
package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports"
)

type GetTransactionUploadPreviewUseCase struct {
	uploadRepository  ports.TransactionUploadRepository
	previewRepository ports.TransactionUploadPreviewRepository
	eventRecorder     adminEventRecorder
	actorDirectory    ports.ActorDirectory
	now               func() time.Time
}

func NewGetTransactionUploadPreviewUseCase(uploadRepository ports.TransactionUploadRepository, previewRepository ports.TransactionUploadPreviewRepository, eventRecorder adminEventRecorder, actorDirectory ports.ActorDirectory) *GetTransactionUploadPreviewUseCase {
	return &GetTransactionUploadPreviewUseCase{uploadRepository: uploadRepository, previewRepository: previewRepository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

func (uc *GetTransactionUploadPreviewUseCase) Execute(ctx context.Context, query ports.GetTransactionUploadPreviewQuery) (ports.GetTransactionUploadPreviewResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return ports.GetTransactionUploadPreviewResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(query.ID)
	if err != nil {
		return ports.GetTransactionUploadPreviewResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return ports.GetTransactionUploadPreviewResult{}, fmt.Errorf("finding upload by id: %w", err)
	}
	if upload == nil {
		return ports.GetTransactionUploadPreviewResult{}, &NotFoundError{Resource: "transaction upload", ID: query.ID}
	}
	if upload.GroupID().String() != query.ActorGroupID {
		return ports.GetTransactionUploadPreviewResult{}, &ForbiddenError{Resource: "transaction upload", Reason: "upload belongs to a different group"}
	}

	preview, err := uc.previewRepository.FindByUploadID(ctx, uploadID)
	if err != nil {
		return ports.GetTransactionUploadPreviewResult{}, fmt.Errorf("finding upload preview by id: %w", err)
	}
	if preview == nil {
		return ports.GetTransactionUploadPreviewResult{}, &NotFoundError{Resource: "transaction upload preview", ID: query.ID}
	}

	if err := upload.RecordPreviewed(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return ports.GetTransactionUploadPreviewResult{}, fmt.Errorf("recording transaction upload preview event: %w", err)
	}
	if err := publishDomainEvents(ctx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
		return ports.GetTransactionUploadPreviewResult{}, fmt.Errorf("publishing transaction upload preview events: %w", err)
	}

	return ports.GetTransactionUploadPreviewResult{
		FileID:           upload.ID().String(),
		FileName:         upload.FileName(),
		Columns:          slices.Clone(preview.Columns),
		Rows:             cloneStringMatrix(preview.Rows),
		TotalRows:        preview.TotalRows,
		ValidationErrors: slices.Clone(preview.ValidationErrors),
	}, nil
	}
```

- [ ] **Step 4: Run the preview-use-case tests again**

Run:

```powershell
go test ./internal/application/use-cases -run TestGetTransactionUploadPreviewUseCaseExecute -count=1
```

Expected: PASS.

- [ ] **Step 5: Verification checkpoint**

Run:

```powershell
go test ./internal/application/use-cases -run 'TestGetTransactionUploadPreviewUseCaseExecute|TestCreateTransactionUploadUseCaseExecute' -count=1
```

Expected: PASS. Do **not** commit unless the user explicitly asks.

---

### Task 4: Add Postgres schema, migration, and repositories

**Files:**
- Create: `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.up.sql`
- Create: `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.down.sql`
- Modify: `internal/adapters/transaction_upload_postgres_repository.go`
- Modify: `internal/adapters/transaction_upload_postgres_repository_test.go`
- Create: `internal/adapters/transaction_upload_preview_postgres_repository.go`
- Create: `internal/adapters/transaction_upload_preview_postgres_repository_test.go`
- Modify: `internal/infrastructure/db/migrate_test.go`

- [ ] **Step 1: Write the failing repository and migration tests first**

Update `internal/adapters/transaction_upload_postgres_repository_test.go` to require `group_id` in insert/select/list SQL expectations.

```go
groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
upload, err := entities.NewTransactionUpload(uploadID, groupID, "transactions.csv", "csv", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file.csv", "transaction-file-v1", valueobjects.UploadedTransactionUploadStatus(), 1, now)

mock.ExpectExec(regexp.QuoteMeta(createTransactionUploadQuery)).
	WithArgs(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt()).
	WillReturnResult(pgxmock.NewResult("INSERT", 1))
```

Create `internal/adapters/transaction_upload_preview_postgres_repository_test.go`:

```go
func TestPostgresTransactionUploadPreviewRepositorySaveAndFindByUploadID(t *testing.T) {
	t.Parallel()

	uploadID := mustUploadID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	preview := ports.TransactionUploadPreviewRecord{
		UploadID:  uploadID.String(),
		Columns:   []string{"Product", "Year"},
		Rows:      [][]string{{"CG", "2026"}},
		TotalRows: 1,
		ValidationErrors: []ports.TransactionFileValidationError{{Code: "MISSING_FIELD", Message: "Year is required", RowNumber: 1, ColumnName: "Year", ColumnIndex: 2}},
	}

	// expect preview row insert, error table insert, then ordered read back
}
```

Add a migration-file existence test to `internal/infrastructure/db/migrate_test.go` similar to the existing `000206` test.

```go
func TestTransactionUploadPreviewMigrationFilesExist(t *testing.T) {
	t.Parallel()

	// read 000208 up/down files and assert presence of:
	// ALTER TABLE transaction_upload ADD COLUMN group_id TEXT;
	// UPDATE transaction_upload SET group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001';
	// CREATE TABLE transaction_upload_preview
	// CREATE TABLE transaction_upload_preview_validation_error
	// DROP TABLE IF EXISTS transaction_upload_preview_validation_error;
	// DROP TABLE IF EXISTS transaction_upload_preview;
}
```

- [ ] **Step 2: Run the adapter/db tests to verify failure**

Run:

```powershell
go test ./internal/adapters ./internal/infrastructure/db -run 'TestPostgresTransactionUploadRepositoryCreateAndQueries|TestPostgresTransactionUploadPreviewRepositorySaveAndFindByUploadID|TestTransactionUploadPreviewMigrationFilesExist' -count=1
```

Expected: FAIL because the SQL, repository type, and migration files do not exist yet.

- [ ] **Step 3: Implement the migration and repository changes**

Create `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.up.sql`:

```sql
ALTER TABLE transaction_upload ADD COLUMN group_id TEXT;

UPDATE transaction_upload
SET group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001'
WHERE group_id IS NULL;

ALTER TABLE transaction_upload ALTER COLUMN group_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transaction_upload_group_id ON transaction_upload(group_id);

CREATE TABLE IF NOT EXISTS transaction_upload_preview (
    upload_id TEXT PRIMARY KEY REFERENCES transaction_upload(id) ON DELETE CASCADE,
    columns_json JSONB NOT NULL,
    rows_json JSONB NOT NULL,
    total_rows INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS transaction_upload_preview_validation_error (
    upload_id TEXT NOT NULL REFERENCES transaction_upload(id) ON DELETE CASCADE,
    ordinal INTEGER NOT NULL,
    code TEXT NOT NULL,
    message TEXT NOT NULL,
    row_number INTEGER NOT NULL,
    column_name TEXT NOT NULL,
    column_index INTEGER NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (upload_id, ordinal)
);
```

Create `internal/infrastructure/db/migrations/000208_add_transaction_upload_preview.down.sql`:

```sql
DROP TABLE IF EXISTS transaction_upload_preview_validation_error;
DROP TABLE IF EXISTS transaction_upload_preview;
DROP INDEX IF EXISTS idx_transaction_upload_group_id;
ALTER TABLE transaction_upload DROP COLUMN IF EXISTS group_id;
```

Update `internal/adapters/transaction_upload_postgres_repository.go` SQL and scanners.

```go
const createTransactionUploadQuery = `
	INSERT INTO transaction_upload (id, group_id, file_name, file_format, content_md5, storage_provider, storage_key, schema_version, status, row_count, uploaded_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`
```

Adjust `scanTransactionUpload`:

```go
var groupID string
if err := row.Scan(&uploadID, &groupID, &fileName, &fileFormat, &contentMD5, &storageProvider, &storageKey, &schemaVersion, &status, &rowCount, &uploadedAt); err != nil {
	// ...
}

parsedGroupID, err := valueobjects.GroupIDFromString(groupID)
if err != nil {
	return nil, fmt.Errorf("parsing upload group id: %w", err)
}

return entities.ReconstituteTransactionUpload(parsedUploadID, parsedGroupID, fileName, fileFormat, contentMD5, storageProvider, storageKey, schemaVersion, parsedStatus, rowCount, uploadedAt), nil
```

Create `internal/adapters/transaction_upload_preview_postgres_repository.go` with JSONB marshal/unmarshal helpers.

```go
type PostgresTransactionUploadPreviewRepository struct {
	pool pgxQuerier
}

func NewPostgresTransactionUploadPreviewRepository(pool pgxQuerier) *PostgresTransactionUploadPreviewRepository {
	return &PostgresTransactionUploadPreviewRepository{pool: pool}
}

func (r *PostgresTransactionUploadPreviewRepository) Save(ctx context.Context, preview ports.TransactionUploadPreviewRecord) error {
	querier := txQuerierFromContext(ctx, r.pool)
	columnsJSON, err := json.Marshal(preview.Columns)
	if err != nil { return fmt.Errorf("marshaling preview columns: %w", err) }
	rowsJSON, err := json.Marshal(preview.Rows)
	if err != nil { return fmt.Errorf("marshaling preview rows: %w", err) }

	if _, err := querier.Exec(ctx, saveTransactionUploadPreviewQuery, preview.UploadID, columnsJSON, rowsJSON, preview.TotalRows); err != nil {
		return fmt.Errorf("saving transaction upload preview: %w", err)
	}
	if _, err := querier.Exec(ctx, deleteTransactionUploadPreviewValidationErrorsQuery, preview.UploadID); err != nil {
		return fmt.Errorf("clearing preview validation errors: %w", err)
	}
	for index, validationError := range preview.ValidationErrors {
		if _, err := querier.Exec(ctx, insertTransactionUploadPreviewValidationErrorQuery, preview.UploadID, index, validationError.Code, validationError.Message, validationError.RowNumber, validationError.ColumnName, validationError.ColumnIndex, validationError.Value); err != nil {
			return fmt.Errorf("saving preview validation error %d: %w", index, err)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run the repository/migration tests again**

Run:

```powershell
go test ./internal/adapters ./internal/infrastructure/db -run 'TestPostgresTransactionUploadRepositoryCreateAndQueries|TestPostgresTransactionUploadPreviewRepositorySaveAndFindByUploadID|TestTransactionUploadPreviewMigrationFilesExist' -count=1
```

Expected: PASS.

- [ ] **Step 5: Verification checkpoint**

Run:

```powershell
go test ./internal/adapters ./internal/infrastructure/db -count=1
```

Expected: PASS. Do **not** commit unless the user explicitly asks.

---

### Task 5: Expose the preview endpoint over HTTP and wire it up

**Files:**
- Modify: `internal/adapters/transaction_upload_http_adapter.go`
- Modify: `internal/adapters/transaction_upload_http_adapter_test.go`
- Modify: `internal/infrastructure/di/wire.go`

- [ ] **Step 1: Write the failing HTTP adapter tests first**

Add a preview port stub to `internal/adapters/transaction_upload_http_adapter_test.go`:

```go
type getTransactionUploadPreviewPortStub struct {
	result ports.GetTransactionUploadPreviewResult
	err    error
	query  ports.GetTransactionUploadPreviewQuery
}

func (s *getTransactionUploadPreviewPortStub) Execute(_ context.Context, query ports.GetTransactionUploadPreviewQuery) (ports.GetTransactionUploadPreviewResult, error) {
	s.query = query
	return s.result, s.err
}
```

Add route cases for:

```go
{
	name:       "gets upload preview",
	method:     http.MethodGet,
	target:     "/api/v1/transaction-uploads/" + transactionUploadID1 + "/preview",
	previewStub: &getTransactionUploadPreviewPortStub{result: ports.GetTransactionUploadPreviewResult{
		FileID:    transactionUploadID1,
		FileName:  "transactions.csv",
		Columns:   []string{"Product", "Year"},
		Rows:      [][]string{{"CG", "2026"}},
		TotalRows: 1,
		ValidationErrors: []ports.TransactionFileValidationError{{Code: "MISSING_FIELD", Message: "Year is required", RowNumber: 1, ColumnName: "Year", ColumnIndex: 2}},
	}},
	wantStatus: http.StatusOK,
}
```

And a preview-specific 404 assertion:

```go
if errorPayload["code"] != "NOT_FOUND" {
	t.Fatalf("error.code = %v, want %q", errorPayload["code"], "NOT_FOUND")
}
if errorPayload["message"] != "Upload not found" {
	t.Fatalf("error.message = %v, want %q", errorPayload["message"], "Upload not found")
}
```

- [ ] **Step 2: Run the HTTP adapter tests to verify failure**

Run:

```powershell
go test ./internal/adapters -run TestHttpTransactionUploadAdapterRoutes -count=1
```

Expected: FAIL because the HTTP adapter constructor/route set does not yet include a preview port or preview handler.

- [ ] **Step 3: Implement the minimal HTTP + DI changes**

Update `internal/adapters/transaction_upload_http_adapter.go`:

```go
type HttpTransactionUploadAdapter struct {
	createUpload              ports.CreateTransactionUploadPort
	streamUpload              transactionUploadStreamExecutor
	getUpload                 ports.GetTransactionUploadPort
	getUploadPreview          ports.GetTransactionUploadPreviewPort
	listUploads               ports.ListTransactionUploadsPort
	deleteUpload              ports.DeleteTransactionUploadPort
	retryUploadClassification ports.RetryTransactionUploadClassificationPort
}
```

Update the constructor and route registration:

```go
func NewHttpTransactionUploadAdapter(createUpload ports.CreateTransactionUploadPort, streamUpload transactionUploadStreamExecutor, getUpload ports.GetTransactionUploadPort, getUploadPreview ports.GetTransactionUploadPreviewPort, listUploads ports.ListTransactionUploadsPort, deleteUpload ports.DeleteTransactionUploadPort, retryUploadClassification ports.RetryTransactionUploadClassificationPort) *HttpTransactionUploadAdapter {
	return &HttpTransactionUploadAdapter{createUpload: createUpload, streamUpload: streamUpload, getUpload: getUpload, getUploadPreview: getUploadPreview, listUploads: listUploads, deleteUpload: deleteUpload, retryUploadClassification: retryUploadClassification}
}

func (a *HttpTransactionUploadAdapter) RegisterRoutes(r *gin.Engine) {
	uploads := r.Group("/api/v1/transaction-uploads")
	uploads.POST("", a.CreateUpload)
	uploads.POST("/stream", a.CreateUploadStream)
	uploads.POST("/:id/retry-classification", a.RetryClassification)
	uploads.GET("", a.ListUploads)
	uploads.GET("/:id", a.GetUpload)
	uploads.GET("/:id/preview", a.GetUploadPreview)
	uploads.DELETE("/:id", a.DeleteUpload)
}
```

Add the preview response/handler:

```go
type transactionUploadPreviewResponse struct {
	FileID           string                           `json:"file_id"`
	FileName         string                           `json:"file_name"`
	Columns          []string                         `json:"columns"`
	Rows             [][]string                       `json:"rows"`
	TotalRows        int                              `json:"total_rows"`
	ValidationErrors []transactionFileValidationError `json:"validation_errors"`
}

func (a *HttpTransactionUploadAdapter) GetUploadPreview(c *gin.Context) {
	result, err := a.getUploadPreview.Execute(c.Request.Context(), ports.GetTransactionUploadPreviewQuery{
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
		Rows:             result.Rows,
		TotalRows:        result.TotalRows,
		ValidationErrors: toTransactionFileValidationErrors(result.ValidationErrors),
	}})
}
```

Add preview-specific error mapping:

```go
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
```

Update `internal/infrastructure/di/wire.go`:

```go
transactionUploadPreviewRepository := adapters.NewPostgresTransactionUploadPreviewRepository(pool)

createTransactionUploadUseCase := usecases.NewCreateTransactionUploadUseCase(transactionUploadRepository, transactionUploadPreviewRepository, transactionRepository, transactionProcessingQueue, rawFileStore, transactionFileParser, transactionManager, eventRecorderService, actorDirectory, transactionFileValidator)
getTransactionUploadPreviewUseCase := usecases.NewGetTransactionUploadPreviewUseCase(transactionUploadRepository, transactionUploadPreviewRepository, eventRecorderService, actorDirectory)

httpTransactionUploadAdapter := adapters.NewHttpTransactionUploadAdapter(createTransactionUploadUseCase, transactionUploadProgressService, getTransactionUploadUseCase, getTransactionUploadPreviewUseCase, listTransactionUploadsUseCase, deleteTransactionUploadUseCase, retryTransactionUploadClassificationUseCase)
```

- [ ] **Step 4: Run the HTTP adapter tests again**

Run:

```powershell
go test ./internal/adapters -run TestHttpTransactionUploadAdapterRoutes -count=1
```

Expected: PASS.

- [ ] **Step 5: Verification checkpoint**

Run:

```powershell
go test ./internal/adapters ./internal/infrastructure/di -count=1
```

Expected: PASS. Do **not** commit unless the user explicitly asks.

---

### Task 6: End-to-end verification for the preview slice

**Files:**
- Modify: any touched files from Tasks 1-5 as needed to fix integration breakages discovered by full-package runs.

- [ ] **Step 1: Run the full transaction-upload focused test set**

Run:

```powershell
go test ./internal/domain/entities ./internal/application/use-cases ./internal/adapters ./internal/infrastructure/db ./internal/infrastructure/di ./internal/ports -count=1
```

Expected: PASS. If any package fails, fix the smallest possible issue in the file named by the test output and re-run this exact command.

- [ ] **Step 2: Run whole-repo verification for regressions**

Run:

```powershell
go test ./... -count=1
```

Expected: PASS across the repo.

- [ ] **Step 3: Confirm the acceptance criteria against concrete outputs**

Check these items against tests/code before declaring completion:

```text
- GET /api/v1/transaction-uploads/:id/preview exists
- same-group authorization enforced with persisted transaction_upload.group_id
- uploaded and failed uploads both have preview data
- response payload includes file_id, file_name, columns, rows, total_rows, validation_errors
- validation_errors preserve code/message/row_number/column_name/column_index/value
- unknown upload returns 404 with code NOT_FOUND and message Upload not found
```

- [ ] **Step 4: Final verification checkpoint**

Run:

```powershell
git status --short
```

Expected: only the intended issue-115 files are modified/untracked. Do **not** commit unless the user explicitly asks.

---

## Self-review checklist

- Spec coverage: Tasks 1-5 cover `group_id`, persisted preview artifacts, preview use case, HTTP endpoint, DI wiring, repository changes, and migration work. Task 6 covers repo-wide verification.
- Placeholder scan: no `TODO`, `TBD`, or “similar to above” references remain.
- Type consistency: `GetTransactionUploadPreviewQuery`, `GetTransactionUploadPreviewResult`, `TransactionUploadPreviewRecord`, `ForbiddenError`, and `PreviewTransactionUploadEventType` use the same names throughout all tasks.
