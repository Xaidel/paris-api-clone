# Issue 118 Inbound/Outbound Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reorganize `internal/ports` and `internal/adapters` into `inbound/` and `outbound/` subdirectories without changing behavior, while delivering the refactor as two directional commits that stay buildable.

**Architecture:** Keep Go package names unchanged (`package ports`, `package adapters`) while moving files into directional subdirectories. Because directory moves change import paths even when package names stay the same, update application and infrastructure imports to alias the new inbound and outbound paths explicitly, and place shared files in whichever directional commit is needed to keep that commit buildable.

**Tech Stack:** Go 1.25, standard `go test`, git, hexagonal architecture under `internal/`

---

## File Structure Map

### Ports files that move to `internal/ports/inbound/`

- `internal/ports/audit_event_get_port.go`
- `internal/ports/audit_events_list_port.go`
- `internal/ports/bug_report_create_port.go`
- `internal/ports/bug_report_delete_port.go`
- `internal/ports/bug_report_get_port.go`
- `internal/ports/bug_report_update_port.go`
- `internal/ports/bug_reports_list_port.go`
- `internal/ports/exclusion_list_create_port.go`
- `internal/ports/exclusion_list_delete_port.go`
- `internal/ports/exclusion_list_get_port.go`
- `internal/ports/exclusion_list_list_port.go`
- `internal/ports/exclusion_list_update_port.go`
- `internal/ports/group_create_port.go`
- `internal/ports/group_delete_port.go`
- `internal/ports/group_get_port.go`
- `internal/ports/group_update_port.go`
- `internal/ports/groups_list_port.go`
- `internal/ports/sector_create_port.go`
- `internal/ports/sector_delete_port.go`
- `internal/ports/sector_get_port.go`
- `internal/ports/sector_update_port.go`
- `internal/ports/sectors_list_port.go`
- `internal/ports/transaction_create_port.go`
- `internal/ports/transaction_delete_port.go`
- `internal/ports/transaction_feedback_delete_port.go`
- `internal/ports/transaction_feedback_get_port.go`
- `internal/ports/transaction_feedback_upsert_port.go`
- `internal/ports/transaction_file_validate_port.go`
- `internal/ports/transaction_get_port.go`
- `internal/ports/transaction_navigation_get_port.go`
- `internal/ports/transaction_step_4_create_port.go`
- `internal/ports/transaction_step_5_create_port.go`
- `internal/ports/transaction_upload_create_port.go`
- `internal/ports/transaction_upload_delete_port.go`
- `internal/ports/transaction_upload_download_port.go`
- `internal/ports/transaction_upload_get_port.go`
- `internal/ports/transaction_upload_preview_get_port.go`
- `internal/ports/transaction_upload_retry_classification_port.go`
- `internal/ports/transaction_uploads_list_port.go`
- `internal/ports/transactions_list_port.go`
- `internal/ports/u1_list_create_port.go`
- `internal/ports/u1_list_delete_port.go`
- `internal/ports/u1_list_get_port.go`
- `internal/ports/u1_list_list_port.go`
- `internal/ports/u1_list_update_port.go`
- `internal/ports/user_create_port.go`
- `internal/ports/user_delete_port.go`
- `internal/ports/user_get_port.go`
- `internal/ports/user_update_port.go`
- `internal/ports/users_list_port.go`
- matching inbound-only tests: `internal/ports/transaction_upload_create_port_test.go`, `internal/ports/transaction_upload_download_port_test.go`

### Ports files that move to `internal/ports/outbound/`

- `internal/ports/actor_directory.go`
- `internal/ports/admin_event_repository.go`
- `internal/ports/bug_report_repository.go`
- `internal/ports/classification_job_queue.go`
- `internal/ports/classification_list_repository.go`
- `internal/ports/classification_metrics.go`
- `internal/ports/event_publisher.go`
- `internal/ports/exclusion_list_repository.go`
- `internal/ports/feedback_repository.go`
- `internal/ports/group_repository.go`
- `internal/ports/password_hasher.go`
- `internal/ports/query_embedding_store.go`
- `internal/ports/raw_file_store.go`
- `internal/ports/sector_repository.go`
- `internal/ports/transaction_classification_gateway.go`
- `internal/ports/transaction_classification_retry_repository.go`
- `internal/ports/transaction_file_parser.go`
- `internal/ports/transaction_manager.go`
- `internal/ports/transaction_processing_queue.go`
- `internal/ports/transaction_repository.go`
- `internal/ports/transaction_step_4_repository.go`
- `internal/ports/transaction_step_5_repository.go`
- `internal/ports/transaction_upload_preview_repository.go`
- `internal/ports/transaction_upload_repository.go`
- `internal/ports/u1_list_repository.go`
- `internal/ports/user_repository.go`

### Ports shared files that may move with either commit to keep it buildable

- `internal/ports/doc.go`
- `internal/ports/audit_event_result.go`
- `internal/ports/bug_report_result.go`
- `internal/ports/classification_job.go`
- `internal/ports/classification_job_test.go`
- `internal/ports/exclusion_list_result.go`
- `internal/ports/feedback_result.go`
- `internal/ports/group_result.go`
- `internal/ports/sector_result.go`
- `internal/ports/transaction_result.go`
- `internal/ports/transaction_step_4_result.go`
- `internal/ports/transaction_step_5_result.go`
- `internal/ports/transaction_upload_progress.go`
- `internal/ports/u1_list_result.go`
- `internal/ports/user_result.go`
- `internal/ports/ports_test.go`

### Adapter files that move to `internal/adapters/inbound/`

- `internal/adapters/audit_event_http_adapter.go`
- `internal/adapters/audit_event_http_adapter_test.go`
- `internal/adapters/bug_report_http_adapter.go`
- `internal/adapters/bug_report_http_adapter_test.go`
- `internal/adapters/exclusion_list_http_adapter.go`
- `internal/adapters/exclusion_list_http_adapter_test.go`
- `internal/adapters/group_http_adapter.go`
- `internal/adapters/group_http_adapter_test.go`
- `internal/adapters/http_common.go`
- `internal/adapters/sector_http_adapter.go`
- `internal/adapters/sector_http_adapter_test.go`
- `internal/adapters/transaction_feedback_http_adapter.go`
- `internal/adapters/transaction_feedback_http_adapter_test.go`
- `internal/adapters/transaction_http_adapter.go`
- `internal/adapters/transaction_http_adapter_test.go`
- `internal/adapters/transaction_step_4_http_adapter.go`
- `internal/adapters/transaction_step_4_http_adapter_test.go`
- `internal/adapters/transaction_step_5_http_adapter.go`
- `internal/adapters/transaction_step_5_http_adapter_test.go`
- `internal/adapters/transaction_upload_http_adapter.go`
- `internal/adapters/transaction_upload_http_adapter_test.go`
- `internal/adapters/u1_list_http_adapter.go`
- `internal/adapters/u1_list_http_adapter_test.go`
- `internal/adapters/user_http_adapter.go`
- `internal/adapters/user_http_adapter_test.go`

### Adapter files that move to `internal/adapters/outbound/`

- `internal/adapters/actor_directory_postgres.go`
- `internal/adapters/admin_event_outbox_postgres_repository.go`
- `internal/adapters/admin_event_postgres_repository.go`
- `internal/adapters/admin_event_postgres_repository_test.go`
- `internal/adapters/bug_report_postgres_repository.go`
- `internal/adapters/bug_report_postgres_repository_test.go`
- `internal/adapters/classification_job_queue_postgres.go`
- `internal/adapters/classification_job_queue_postgres_test.go`
- `internal/adapters/classification_list_postgres_repository.go`
- `internal/adapters/classification_list_postgres_repository_test.go`
- `internal/adapters/exclusion_list_postgres_repository.go`
- `internal/adapters/exclusion_list_postgres_repository_test.go`
- `internal/adapters/feedback_postgres_repository.go`
- `internal/adapters/feedback_postgres_repository_test.go`
- `internal/adapters/group_postgres_repository.go`
- `internal/adapters/group_postgres_repository_test.go`
- `internal/adapters/password_hasher_bcrypt.go`
- `internal/adapters/password_hasher_bcrypt_test.go`
- `internal/adapters/postgres_transaction_classification_retry_repository.go`
- `internal/adapters/postgres_transaction_classification_retry_repository_test.go`
- `internal/adapters/raw_file_azure_blob_store.go`
- `internal/adapters/raw_file_azure_blob_store_test.go`
- `internal/adapters/raw_file_local_store.go`
- `internal/adapters/raw_file_local_store_test.go`
- `internal/adapters/sector_postgres_repository.go`
- `internal/adapters/sector_postgres_repository_test.go`
- `internal/adapters/transaction_classification_react_gateway.go`
- `internal/adapters/transaction_classification_react_gateway_test.go`
- `internal/adapters/transaction_file_parser.go`
- `internal/adapters/transaction_file_parser_test.go`
- `internal/adapters/transaction_manager_pgx.go`
- `internal/adapters/transaction_manager_pgx_test.go`
- `internal/adapters/transaction_postgres_repository.go`
- `internal/adapters/transaction_postgres_repository_test.go`
- `internal/adapters/transaction_processing_queue_postgres.go`
- `internal/adapters/transaction_processing_queue_postgres_test.go`
- `internal/adapters/transaction_step_4_postgres_repository.go`
- `internal/adapters/transaction_step_4_postgres_repository_test.go`
- `internal/adapters/transaction_step_5_postgres_repository.go`
- `internal/adapters/transaction_step_5_postgres_repository_test.go`
- `internal/adapters/transaction_upload_postgres_repository.go`
- `internal/adapters/transaction_upload_postgres_repository_test.go`
- `internal/adapters/transaction_upload_preview_postgres_repository.go`
- `internal/adapters/transaction_upload_preview_postgres_repository_test.go`
- `internal/adapters/u1_list_postgres_repository.go`
- `internal/adapters/u1_list_postgres_repository_test.go`
- `internal/adapters/user_postgres_repository.go`
- `internal/adapters/user_postgres_repository_test.go`

### Adapter shared files that may move with either commit to keep it buildable

- `internal/adapters/doc.go`
- `internal/adapters/pgvector.go`
- `internal/adapters/transaction_classification_prompt_react.go`
- `internal/adapters/transaction_classification_prompt_react_test.go`

### Non-move files that will need import updates

- `internal/application/use-cases/*.go` and `internal/application/use-cases/*_test.go` importing `internal/ports`
- `internal/application/services/*.go` and `internal/application/services/*_test.go` importing `internal/ports`
- `internal/infrastructure/di/wire.go`
- `internal/infrastructure/di/raw_file_store_test.go`
- `internal/infrastructure/observability/metrics.go`

### Import strategy

- Files using only inbound port types should import `github.com/gyud-adb/paris-api/internal/ports/inbound` with alias `inboundports`
- During the inbound commit, files that still need unmoved outbound port types should keep importing `github.com/gyud-adb/paris-api/internal/ports` with alias `ports`
- After outbound files move, files using only outbound port types should import `github.com/gyud-adb/paris-api/internal/ports/outbound` with alias `outboundports`
- After outbound files move, files using both should import both aliases explicitly
- During the inbound commit, `internal/infrastructure/di/wire.go` should import `internal/adapters/inbound` as `inboundadapters` and keep `internal/adapters` as `adapters` for unmoved outbound constructors
- After outbound files move, `internal/infrastructure/di/wire.go` should import `internal/adapters/outbound` as `outboundadapters`
- Shared files such as `internal/infrastructure/observability/metrics.go` should import the path that owns the needed interface after the split

### Known baseline before this issue

- `go test ./...` is already failing in `internal/domain/entities/transaction_upload_test.go` because `ReconstituteTransactionUpload` now returns two values. This plan treats that as pre-existing and out of scope. Verification for this issue should confirm the refactor does not introduce additional failures.

### Task 1: Create directories and capture failing import-path baseline

**Files:**
- Create: `internal/ports/inbound/`
- Create: `internal/ports/outbound/`
- Create: `internal/adapters/inbound/`
- Create: `internal/adapters/outbound/`
- Modify: no source files yet

- [ ] **Step 1: Create the four directional directories**

```bash
mkdir internal\ports\inbound
mkdir internal\ports\outbound
mkdir internal\adapters\inbound
mkdir internal\adapters\outbound
```

- [ ] **Step 2: Run the existing test suite to confirm the baseline failure**

Run: `go test ./...`
Expected: FAIL only because `internal/domain/entities/transaction_upload_test.go` has the pre-existing `assignment mismatch` error for `ReconstituteTransactionUpload`

- [ ] **Step 3: Record the baseline in your notes before changing code**

```text
Baseline failure is outside issue #118:
- internal/domain/entities/transaction_upload_test.go:311
- internal/domain/entities/transaction_upload_test.go:343
No ports/adapters import-path failures yet because no files have moved.
```

- [ ] **Step 4: Commit directory scaffolding only if the repo policy allows empty directory tracking; otherwise skip commit**

```bash
# Skip commit if directories are empty and Git does not track them yet.
```

### Task 2: Move inbound port files and fix inbound port imports

**Files:**
- Modify: all inbound port files listed above
- Modify: `internal/application/use-cases/*.go`
- Modify: `internal/application/use-cases/*_test.go`
- Modify: `internal/application/services/*.go`
- Modify: `internal/application/services/*_test.go`
- Modify: `internal/infrastructure/di/wire.go`
- Test: `internal/application/use-cases/*_test.go`, `internal/application/services/*_test.go`

- [ ] **Step 1: Move inbound port files into `internal/ports/inbound/`**

```bash
git mv internal/ports/audit_event_get_port.go internal/ports/inbound/audit_event_get_port.go
git mv internal/ports/audit_events_list_port.go internal/ports/inbound/audit_events_list_port.go
git mv internal/ports/bug_report_create_port.go internal/ports/inbound/bug_report_create_port.go
git mv internal/ports/bug_report_delete_port.go internal/ports/inbound/bug_report_delete_port.go
git mv internal/ports/bug_report_get_port.go internal/ports/inbound/bug_report_get_port.go
git mv internal/ports/bug_report_update_port.go internal/ports/inbound/bug_report_update_port.go
git mv internal/ports/bug_reports_list_port.go internal/ports/inbound/bug_reports_list_port.go
git mv internal/ports/exclusion_list_create_port.go internal/ports/inbound/exclusion_list_create_port.go
git mv internal/ports/exclusion_list_delete_port.go internal/ports/inbound/exclusion_list_delete_port.go
git mv internal/ports/exclusion_list_get_port.go internal/ports/inbound/exclusion_list_get_port.go
git mv internal/ports/exclusion_list_list_port.go internal/ports/inbound/exclusion_list_list_port.go
git mv internal/ports/exclusion_list_update_port.go internal/ports/inbound/exclusion_list_update_port.go
git mv internal/ports/group_create_port.go internal/ports/inbound/group_create_port.go
git mv internal/ports/group_delete_port.go internal/ports/inbound/group_delete_port.go
git mv internal/ports/group_get_port.go internal/ports/inbound/group_get_port.go
git mv internal/ports/group_update_port.go internal/ports/inbound/group_update_port.go
git mv internal/ports/groups_list_port.go internal/ports/inbound/groups_list_port.go
git mv internal/ports/sector_create_port.go internal/ports/inbound/sector_create_port.go
git mv internal/ports/sector_delete_port.go internal/ports/inbound/sector_delete_port.go
git mv internal/ports/sector_get_port.go internal/ports/inbound/sector_get_port.go
git mv internal/ports/sector_update_port.go internal/ports/inbound/sector_update_port.go
git mv internal/ports/sectors_list_port.go internal/ports/inbound/sectors_list_port.go
git mv internal/ports/transaction_create_port.go internal/ports/inbound/transaction_create_port.go
git mv internal/ports/transaction_delete_port.go internal/ports/inbound/transaction_delete_port.go
git mv internal/ports/transaction_feedback_delete_port.go internal/ports/inbound/transaction_feedback_delete_port.go
git mv internal/ports/transaction_feedback_get_port.go internal/ports/inbound/transaction_feedback_get_port.go
git mv internal/ports/transaction_feedback_upsert_port.go internal/ports/inbound/transaction_feedback_upsert_port.go
git mv internal/ports/transaction_file_validate_port.go internal/ports/inbound/transaction_file_validate_port.go
git mv internal/ports/transaction_get_port.go internal/ports/inbound/transaction_get_port.go
git mv internal/ports/transaction_navigation_get_port.go internal/ports/inbound/transaction_navigation_get_port.go
git mv internal/ports/transaction_step_4_create_port.go internal/ports/inbound/transaction_step_4_create_port.go
git mv internal/ports/transaction_step_5_create_port.go internal/ports/inbound/transaction_step_5_create_port.go
git mv internal/ports/transaction_upload_create_port.go internal/ports/inbound/transaction_upload_create_port.go
git mv internal/ports/transaction_upload_delete_port.go internal/ports/inbound/transaction_upload_delete_port.go
git mv internal/ports/transaction_upload_download_port.go internal/ports/inbound/transaction_upload_download_port.go
git mv internal/ports/transaction_upload_get_port.go internal/ports/inbound/transaction_upload_get_port.go
git mv internal/ports/transaction_upload_preview_get_port.go internal/ports/inbound/transaction_upload_preview_get_port.go
git mv internal/ports/transaction_upload_retry_classification_port.go internal/ports/inbound/transaction_upload_retry_classification_port.go
git mv internal/ports/transaction_uploads_list_port.go internal/ports/inbound/transaction_uploads_list_port.go
git mv internal/ports/transactions_list_port.go internal/ports/inbound/transactions_list_port.go
git mv internal/ports/u1_list_create_port.go internal/ports/inbound/u1_list_create_port.go
git mv internal/ports/u1_list_delete_port.go internal/ports/inbound/u1_list_delete_port.go
git mv internal/ports/u1_list_get_port.go internal/ports/inbound/u1_list_get_port.go
git mv internal/ports/u1_list_list_port.go internal/ports/inbound/u1_list_list_port.go
git mv internal/ports/u1_list_update_port.go internal/ports/inbound/u1_list_update_port.go
git mv internal/ports/user_create_port.go internal/ports/inbound/user_create_port.go
git mv internal/ports/user_delete_port.go internal/ports/inbound/user_delete_port.go
git mv internal/ports/user_get_port.go internal/ports/inbound/user_get_port.go
git mv internal/ports/user_update_port.go internal/ports/inbound/user_update_port.go
git mv internal/ports/users_list_port.go internal/ports/inbound/users_list_port.go
git mv internal/ports/transaction_upload_create_port_test.go internal/ports/inbound/transaction_upload_create_port_test.go
git mv internal/ports/transaction_upload_download_port_test.go internal/ports/inbound/transaction_upload_download_port_test.go
```

- [ ] **Step 2: Update imports in application code to use the inbound alias where only inbound port types are referenced**

```go
import (
    "context"

    inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type GetUserUseCase struct{}

func (uc *GetUserUseCase) Execute(ctx context.Context, query inboundports.GetUserQuery) (inboundports.UserResult, error) {
    return inboundports.UserResult{}, nil
}
```

- [ ] **Step 3: Update mixed-use files to import inbound aliases while keeping unmoved outbound types on the existing `internal/ports` import path**

```go
import (
    "context"

    inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
    ports "github.com/gyud-adb/paris-api/internal/ports"
)

type CreateTransactionUploadUseCase struct {
    uploadRepository ports.TransactionUploadRepository
    previewRepository ports.TransactionUploadPreviewRepository
    rawFileStore ports.RawFileStore
}

func (uc *CreateTransactionUploadUseCase) Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand) (inboundports.CreateTransactionUploadResult, error) {
    _ = ctx
    _ = command
    return inboundports.CreateTransactionUploadResult{}, nil
}
```

- [ ] **Step 4: Update `internal/infrastructure/di/wire.go` to use the inbound port alias for application capability fields while keeping unmoved outbound contracts on the existing port import**

```go
import (
    inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
    ports "github.com/gyud-adb/paris-api/internal/ports"
)

type Application struct {
    ClassificationListRepository ports.ClassificationListRepository
    ValidateTransactionFile      inboundports.ValidateTransactionFilePort
    CreateTransaction            inboundports.CreateTransactionPort
}
```

- [ ] **Step 5: Run targeted tests for representative inbound consumers**

Run: `go test ./internal/application/use-cases ./internal/application/services ./internal/infrastructure/di ./internal/infrastructure/observability`
Expected: package import-path errors related to inbound ports are gone; only the pre-existing domain baseline may remain outside these packages

- [ ] **Step 6: Move any shared ports files required to keep the inbound commit buildable**

```text
Candidates:
- internal/ports/doc.go
- internal/ports/*_result.go
- internal/ports/transaction_upload_progress.go

Place them in inbound if compilation of moved inbound files still depends on them
being in the same import path package for this commit boundary.
```

- [ ] **Step 7: Stage the inbound ports refactor but do not commit yet because the first required commit must include both inbound ports and inbound adapters**

```bash
git add internal/ports/inbound internal/application internal/infrastructure/di/wire.go internal/infrastructure/observability/metrics.go
```

### Task 3: Move inbound adapter files and fix inbound adapter imports

**Files:**
- Modify: inbound adapter files listed above
- Modify: `internal/infrastructure/di/wire.go`
- Modify: `internal/infrastructure/di/raw_file_store_test.go`
- Test: `internal/adapters/inbound/*_test.go`, `internal/infrastructure/di/*`

- [ ] **Step 1: Move inbound adapter files into `internal/adapters/inbound/`**

```bash
git mv internal/adapters/audit_event_http_adapter.go internal/adapters/inbound/audit_event_http_adapter.go
git mv internal/adapters/audit_event_http_adapter_test.go internal/adapters/inbound/audit_event_http_adapter_test.go
git mv internal/adapters/bug_report_http_adapter.go internal/adapters/inbound/bug_report_http_adapter.go
git mv internal/adapters/bug_report_http_adapter_test.go internal/adapters/inbound/bug_report_http_adapter_test.go
git mv internal/adapters/exclusion_list_http_adapter.go internal/adapters/inbound/exclusion_list_http_adapter.go
git mv internal/adapters/exclusion_list_http_adapter_test.go internal/adapters/inbound/exclusion_list_http_adapter_test.go
git mv internal/adapters/group_http_adapter.go internal/adapters/inbound/group_http_adapter.go
git mv internal/adapters/group_http_adapter_test.go internal/adapters/inbound/group_http_adapter_test.go
git mv internal/adapters/http_common.go internal/adapters/inbound/http_common.go
git mv internal/adapters/sector_http_adapter.go internal/adapters/inbound/sector_http_adapter.go
git mv internal/adapters/sector_http_adapter_test.go internal/adapters/inbound/sector_http_adapter_test.go
git mv internal/adapters/transaction_feedback_http_adapter.go internal/adapters/inbound/transaction_feedback_http_adapter.go
git mv internal/adapters/transaction_feedback_http_adapter_test.go internal/adapters/inbound/transaction_feedback_http_adapter_test.go
git mv internal/adapters/transaction_http_adapter.go internal/adapters/inbound/transaction_http_adapter.go
git mv internal/adapters/transaction_http_adapter_test.go internal/adapters/inbound/transaction_http_adapter_test.go
git mv internal/adapters/transaction_step_4_http_adapter.go internal/adapters/inbound/transaction_step_4_http_adapter.go
git mv internal/adapters/transaction_step_4_http_adapter_test.go internal/adapters/inbound/transaction_step_4_http_adapter_test.go
git mv internal/adapters/transaction_step_5_http_adapter.go internal/adapters/inbound/transaction_step_5_http_adapter.go
git mv internal/adapters/transaction_step_5_http_adapter_test.go internal/adapters/inbound/transaction_step_5_http_adapter_test.go
git mv internal/adapters/transaction_upload_http_adapter.go internal/adapters/inbound/transaction_upload_http_adapter.go
git mv internal/adapters/transaction_upload_http_adapter_test.go internal/adapters/inbound/transaction_upload_http_adapter_test.go
git mv internal/adapters/u1_list_http_adapter.go internal/adapters/inbound/u1_list_http_adapter.go
git mv internal/adapters/u1_list_http_adapter_test.go internal/adapters/inbound/u1_list_http_adapter_test.go
git mv internal/adapters/user_http_adapter.go internal/adapters/inbound/user_http_adapter.go
git mv internal/adapters/user_http_adapter_test.go internal/adapters/inbound/user_http_adapter_test.go
```

- [ ] **Step 2: Update `internal/infrastructure/di/wire.go` to import inbound adapters separately while keeping unmoved outbound adapters on the existing import path**

```go
import (
    inboundadapters "github.com/gyud-adb/paris-api/internal/adapters/inbound"
    adapters "github.com/gyud-adb/paris-api/internal/adapters"
)

var newOpenAIReActChatModel = adapters.NewOpenAIReActChatModel

httpUserAdapter := inboundadapters.NewHttpUserAdapter(createUserUseCase, getUserUseCase, listUsersUseCase, updateUserUseCase, deleteUserUseCase)
userRepository := adapters.NewPostgresUserRepository(pool)
```

- [ ] **Step 3: Leave `internal/infrastructure/di/raw_file_store_test.go` unchanged during the inbound commit because its outbound adapter imports are not moving yet**

```text
No source edit is needed in this step.
```

- [ ] **Step 4: Run targeted inbound adapter tests**

Run: `go test ./internal/adapters/inbound ./internal/infrastructure/di`
Expected: inbound adapter packages compile and tests pass; no new import-path failures

- [ ] **Step 5: Move any shared adapter files required to keep the inbound commit buildable**

```text
Candidates:
- internal/adapters/doc.go

Do not move outbound-only helpers such as pgvector or ReAct prompt builder into
the inbound commit unless compilation proves they are required.
```

- [ ] **Step 6: Create the first required commit containing both inbound ports and inbound adapters, with the shared-file justification in the message**

```bash
git add internal/ports/inbound internal/adapters/inbound internal/application internal/infrastructure/di/wire.go internal/infrastructure/observability/metrics.go
git commit -m "refactor(hex): move inbound ports and adapters into inbound subdirectories and include shared package files needed to keep the commit buildable"
```

### Task 4: Move outbound port files and finish outbound import split

**Files:**
- Modify: outbound port files listed above
- Modify: remaining `internal/application/use-cases/*.go`
- Modify: remaining `internal/application/services/*.go`
- Modify: `internal/infrastructure/di/wire.go`
- Modify: `internal/infrastructure/observability/metrics.go`
- Test: `internal/application/use-cases`, `internal/application/services`, `internal/infrastructure/*`

- [ ] **Step 1: Move outbound port files into `internal/ports/outbound/`**

```bash
git mv internal/ports/actor_directory.go internal/ports/outbound/actor_directory.go
git mv internal/ports/admin_event_repository.go internal/ports/outbound/admin_event_repository.go
git mv internal/ports/bug_report_repository.go internal/ports/outbound/bug_report_repository.go
git mv internal/ports/classification_job_queue.go internal/ports/outbound/classification_job_queue.go
git mv internal/ports/classification_list_repository.go internal/ports/outbound/classification_list_repository.go
git mv internal/ports/classification_metrics.go internal/ports/outbound/classification_metrics.go
git mv internal/ports/event_publisher.go internal/ports/outbound/event_publisher.go
git mv internal/ports/exclusion_list_repository.go internal/ports/outbound/exclusion_list_repository.go
git mv internal/ports/feedback_repository.go internal/ports/outbound/feedback_repository.go
git mv internal/ports/group_repository.go internal/ports/outbound/group_repository.go
git mv internal/ports/password_hasher.go internal/ports/outbound/password_hasher.go
git mv internal/ports/query_embedding_store.go internal/ports/outbound/query_embedding_store.go
git mv internal/ports/raw_file_store.go internal/ports/outbound/raw_file_store.go
git mv internal/ports/sector_repository.go internal/ports/outbound/sector_repository.go
git mv internal/ports/transaction_classification_gateway.go internal/ports/outbound/transaction_classification_gateway.go
git mv internal/ports/transaction_classification_retry_repository.go internal/ports/outbound/transaction_classification_retry_repository.go
git mv internal/ports/transaction_file_parser.go internal/ports/outbound/transaction_file_parser.go
git mv internal/ports/transaction_manager.go internal/ports/outbound/transaction_manager.go
git mv internal/ports/transaction_processing_queue.go internal/ports/outbound/transaction_processing_queue.go
git mv internal/ports/transaction_repository.go internal/ports/outbound/transaction_repository.go
git mv internal/ports/transaction_step_4_repository.go internal/ports/outbound/transaction_step_4_repository.go
git mv internal/ports/transaction_step_5_repository.go internal/ports/outbound/transaction_step_5_repository.go
git mv internal/ports/transaction_upload_preview_repository.go internal/ports/outbound/transaction_upload_preview_repository.go
git mv internal/ports/transaction_upload_repository.go internal/ports/outbound/transaction_upload_repository.go
git mv internal/ports/u1_list_repository.go internal/ports/outbound/u1_list_repository.go
git mv internal/ports/user_repository.go internal/ports/outbound/user_repository.go
```

- [ ] **Step 2: Move the remaining shared ports files into the path that keeps the package model coherent**

```bash
git mv internal/ports/doc.go internal/ports/outbound/doc.go
git mv internal/ports/audit_event_result.go internal/ports/outbound/audit_event_result.go
git mv internal/ports/bug_report_result.go internal/ports/outbound/bug_report_result.go
git mv internal/ports/classification_job.go internal/ports/outbound/classification_job.go
git mv internal/ports/classification_job_test.go internal/ports/outbound/classification_job_test.go
git mv internal/ports/exclusion_list_result.go internal/ports/outbound/exclusion_list_result.go
git mv internal/ports/feedback_result.go internal/ports/outbound/feedback_result.go
git mv internal/ports/group_result.go internal/ports/outbound/group_result.go
git mv internal/ports/sector_result.go internal/ports/outbound/sector_result.go
git mv internal/ports/transaction_result.go internal/ports/outbound/transaction_result.go
git mv internal/ports/transaction_step_4_result.go internal/ports/outbound/transaction_step_4_result.go
git mv internal/ports/transaction_step_5_result.go internal/ports/outbound/transaction_step_5_result.go
git mv internal/ports/transaction_upload_progress.go internal/ports/outbound/transaction_upload_progress.go
git mv internal/ports/u1_list_result.go internal/ports/outbound/u1_list_result.go
git mv internal/ports/user_result.go internal/ports/outbound/user_result.go
git mv internal/ports/ports_test.go internal/ports/outbound/ports_test.go
```

- [ ] **Step 3: Update mixed application files to use both aliases everywhere**

```go
import (
    inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
    outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func NewListTransactionsUseCase(
    transactionRepository outboundports.TransactionRepository,
    transactionStep4Repository outboundports.TransactionStep4Repository,
    transactionStep5Repository outboundports.TransactionStep5Repository,
    sectorRepository outboundports.SectorRepository,
) *ListTransactionsUseCase {
    return &ListTransactionsUseCase{}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, query inboundports.ListTransactionsQuery) (inboundports.ListTransactionsResult, error) {
    return inboundports.ListTransactionsResult{}, nil
}
```

- [ ] **Step 4: Update `internal/infrastructure/observability/metrics.go` to import the outbound port path**

```go
import outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"

func NewClassificationMetrics() outboundports.ClassificationMetrics {
    return &NoopClassificationMetrics{}
}
```

- [ ] **Step 5: Run targeted tests for outbound port consumers**

Run: `go test ./internal/application/use-cases ./internal/application/services ./internal/infrastructure/observability ./internal/infrastructure/di`
Expected: no unresolved `internal/ports` imports remain in touched packages

- [ ] **Step 6: Stage the outbound ports changes**

```bash
git add internal/ports/outbound internal/application internal/infrastructure/observability/metrics.go internal/infrastructure/di/wire.go
```

### Task 5: Move outbound adapter files and finish wiring updates

**Files:**
- Modify: outbound adapter files listed above
- Modify: `internal/infrastructure/di/wire.go`
- Modify: `internal/infrastructure/di/raw_file_store_test.go`
- Test: `internal/adapters/outbound`, `internal/infrastructure/di`

- [ ] **Step 1: Move outbound adapter files into `internal/adapters/outbound/`**

```bash
git mv internal/adapters/actor_directory_postgres.go internal/adapters/outbound/actor_directory_postgres.go
git mv internal/adapters/admin_event_outbox_postgres_repository.go internal/adapters/outbound/admin_event_outbox_postgres_repository.go
git mv internal/adapters/admin_event_postgres_repository.go internal/adapters/outbound/admin_event_postgres_repository.go
git mv internal/adapters/admin_event_postgres_repository_test.go internal/adapters/outbound/admin_event_postgres_repository_test.go
git mv internal/adapters/bug_report_postgres_repository.go internal/adapters/outbound/bug_report_postgres_repository.go
git mv internal/adapters/bug_report_postgres_repository_test.go internal/adapters/outbound/bug_report_postgres_repository_test.go
git mv internal/adapters/classification_job_queue_postgres.go internal/adapters/outbound/classification_job_queue_postgres.go
git mv internal/adapters/classification_job_queue_postgres_test.go internal/adapters/outbound/classification_job_queue_postgres_test.go
git mv internal/adapters/classification_list_postgres_repository.go internal/adapters/outbound/classification_list_postgres_repository.go
git mv internal/adapters/classification_list_postgres_repository_test.go internal/adapters/outbound/classification_list_postgres_repository_test.go
git mv internal/adapters/exclusion_list_postgres_repository.go internal/adapters/outbound/exclusion_list_postgres_repository.go
git mv internal/adapters/exclusion_list_postgres_repository_test.go internal/adapters/outbound/exclusion_list_postgres_repository_test.go
git mv internal/adapters/feedback_postgres_repository.go internal/adapters/outbound/feedback_postgres_repository.go
git mv internal/adapters/feedback_postgres_repository_test.go internal/adapters/outbound/feedback_postgres_repository_test.go
git mv internal/adapters/group_postgres_repository.go internal/adapters/outbound/group_postgres_repository.go
git mv internal/adapters/group_postgres_repository_test.go internal/adapters/outbound/group_postgres_repository_test.go
git mv internal/adapters/password_hasher_bcrypt.go internal/adapters/outbound/password_hasher_bcrypt.go
git mv internal/adapters/password_hasher_bcrypt_test.go internal/adapters/outbound/password_hasher_bcrypt_test.go
git mv internal/adapters/postgres_transaction_classification_retry_repository.go internal/adapters/outbound/postgres_transaction_classification_retry_repository.go
git mv internal/adapters/postgres_transaction_classification_retry_repository_test.go internal/adapters/outbound/postgres_transaction_classification_retry_repository_test.go
git mv internal/adapters/raw_file_azure_blob_store.go internal/adapters/outbound/raw_file_azure_blob_store.go
git mv internal/adapters/raw_file_azure_blob_store_test.go internal/adapters/outbound/raw_file_azure_blob_store_test.go
git mv internal/adapters/raw_file_local_store.go internal/adapters/outbound/raw_file_local_store.go
git mv internal/adapters/raw_file_local_store_test.go internal/adapters/outbound/raw_file_local_store_test.go
git mv internal/adapters/sector_postgres_repository.go internal/adapters/outbound/sector_postgres_repository.go
git mv internal/adapters/sector_postgres_repository_test.go internal/adapters/outbound/sector_postgres_repository_test.go
git mv internal/adapters/transaction_classification_react_gateway.go internal/adapters/outbound/transaction_classification_react_gateway.go
git mv internal/adapters/transaction_classification_react_gateway_test.go internal/adapters/outbound/transaction_classification_react_gateway_test.go
git mv internal/adapters/transaction_file_parser.go internal/adapters/outbound/transaction_file_parser.go
git mv internal/adapters/transaction_file_parser_test.go internal/adapters/outbound/transaction_file_parser_test.go
git mv internal/adapters/transaction_manager_pgx.go internal/adapters/outbound/transaction_manager_pgx.go
git mv internal/adapters/transaction_manager_pgx_test.go internal/adapters/outbound/transaction_manager_pgx_test.go
git mv internal/adapters/transaction_postgres_repository.go internal/adapters/outbound/transaction_postgres_repository.go
git mv internal/adapters/transaction_postgres_repository_test.go internal/adapters/outbound/transaction_postgres_repository_test.go
git mv internal/adapters/transaction_processing_queue_postgres.go internal/adapters/outbound/transaction_processing_queue_postgres.go
git mv internal/adapters/transaction_processing_queue_postgres_test.go internal/adapters/outbound/transaction_processing_queue_postgres_test.go
git mv internal/adapters/transaction_step_4_postgres_repository.go internal/adapters/outbound/transaction_step_4_postgres_repository.go
git mv internal/adapters/transaction_step_4_postgres_repository_test.go internal/adapters/outbound/transaction_step_4_postgres_repository_test.go
git mv internal/adapters/transaction_step_5_postgres_repository.go internal/adapters/outbound/transaction_step_5_postgres_repository.go
git mv internal/adapters/transaction_step_5_postgres_repository_test.go internal/adapters/outbound/transaction_step_5_postgres_repository_test.go
git mv internal/adapters/transaction_upload_postgres_repository.go internal/adapters/outbound/transaction_upload_postgres_repository.go
git mv internal/adapters/transaction_upload_postgres_repository_test.go internal/adapters/outbound/transaction_upload_postgres_repository_test.go
git mv internal/adapters/transaction_upload_preview_postgres_repository.go internal/adapters/outbound/transaction_upload_preview_postgres_repository.go
git mv internal/adapters/transaction_upload_preview_postgres_repository_test.go internal/adapters/outbound/transaction_upload_preview_postgres_repository_test.go
git mv internal/adapters/u1_list_postgres_repository.go internal/adapters/outbound/u1_list_postgres_repository.go
git mv internal/adapters/u1_list_postgres_repository_test.go internal/adapters/outbound/u1_list_postgres_repository_test.go
git mv internal/adapters/user_postgres_repository.go internal/adapters/outbound/user_postgres_repository.go
git mv internal/adapters/user_postgres_repository_test.go internal/adapters/outbound/user_postgres_repository_test.go
git mv internal/adapters/doc.go internal/adapters/outbound/doc.go
git mv internal/adapters/pgvector.go internal/adapters/outbound/pgvector.go
git mv internal/adapters/transaction_classification_prompt_react.go internal/adapters/outbound/transaction_classification_prompt_react.go
git mv internal/adapters/transaction_classification_prompt_react_test.go internal/adapters/outbound/transaction_classification_prompt_react_test.go
```

- [ ] **Step 2: Finish `internal/infrastructure/di/wire.go` so every constructor call uses the correct adapter alias**

```go
userRepository := outboundadapters.NewPostgresUserRepository(pool)
reactSystemPromptBuilder := outboundadapters.NewReActTransactionClassificationSystemPromptBuilder(
    cfg.Classification.ReactSystemPrompt,
    u1ListRepository,
    exclusionListRepository,
)
httpTransactionAdapter := inboundadapters.NewHttpTransactionAdapter(
    createTransactionUseCase,
    getTransactionUseCase,
    getTransactionNavigationUseCase,
    listTransactionsUseCase,
    deleteTransactionUseCase,
)
```

- [ ] **Step 3: Run targeted outbound adapter and wiring tests**

Run: `go test ./internal/adapters/outbound ./internal/infrastructure/di`
Expected: outbound constructors and tests compile; `wire.go` resolves all adapter references through aliases

- [ ] **Step 4: Update `internal/infrastructure/di/raw_file_store_test.go` to import the outbound adapter path after outbound files move**

```go
import outboundadapters "github.com/gyud-adb/paris-api/internal/adapters/outbound"
```

- [ ] **Step 5: Commit the outbound refactor with the justification in the message**

```bash
git add internal/ports/outbound internal/adapters/outbound internal/infrastructure/di/wire.go internal/infrastructure/di/raw_file_store_test.go internal/infrastructure/observability/metrics.go internal/application
git commit -m "refactor(hex): move outbound ports and adapters into outbound subdirectories and include shared package files needed to keep the commit buildable"
```

### Task 6: Final verification and review of the two-commit history

**Files:**
- Modify: none unless fixes are needed
- Test: whole repository

- [ ] **Step 1: Confirm the working tree is clean except for unrelated pre-existing files**

Run: `git status --short`
Expected: no unstaged changes from the refactor remain

- [ ] **Step 2: Run the full test suite again**

Run: `go test ./...`
Expected: same known baseline failure in `internal/domain/entities/transaction_upload_test.go`; no new failures from moved ports or adapters

- [ ] **Step 3: Review the last two commit messages to ensure they document the shared-file justification**

Run: `git log --oneline -2`
Expected: two refactor commits, one inbound-oriented and one outbound-oriented, each mentioning that shared package files were included to keep the commit buildable

- [ ] **Step 4: If the full suite shows any new failures caused by the refactor, fix them before declaring success**

```text
Only fix regressions introduced by the directory refactor.
Do not fold unrelated domain test repairs into issue #118.
```
