# Request Flow: Transaction Upload

This walkthrough traces the richer ingestion path used for transaction uploads.

The request is more complex than user creation because one HTTP call fans out into several outbound responsibilities:

- raw file storage,
- upload metadata persistence,
- transaction row persistence,
- background processing queue persistence,
- domain event publication.

That makes it the best example in this repo for showing how hexagonal architecture avoids collapsing everything into one large handler.

## 1. Composition Root Wires Multiple Outbound Responsibilities

`internal/infrastructure/di/wire.go` assembles the upload path from several concrete pieces.

The important constructors are:

1. `transactionUploadRepository := adapters.NewPostgresTransactionUploadRepository(pool)`
2. `transactionRepository := adapters.NewPostgresTransactionRepository(pool)`
3. `transactionProcessingQueue := adapters.NewPostgresTransactionProcessingQueue(pool)`
4. `transactionManager := adapters.NewPgxTransactionManager(pool)`
5. `transactionFileParser := adapters.NewTransactionFileParser()`
6. `rawFileStore, err := buildRawFileStore(ctx, cfg.Storage)`
7. `transactionFileValidator := domainservices.NewTransactionFileValidator(...)`
8. `createTransactionUploadUseCase := usecases.NewCreateTransactionUploadUseCase(...)`
9. `transactionUploadProgressService := services.NewTransactionUploadProgressService(createTransactionUploadUseCase)`
10. `downloadTransactionUploadUseCase := usecases.NewDownloadTransactionUploadUseCase(...)`
11. `httpTransactionUploadAdapter := adapters.NewHttpTransactionUploadAdapter(...)`

The use case is therefore free to think in ports:

- `TransactionUploadRepository`
- `TransactionRepository`
- `TransactionProcessingQueue`
- `RawFileStore`
- `TransactionFileParser`
- `TransactionManager`
- `EventPublisher`
- `ActorDirectory`

## 2. One Adapter Exposes Multiple HTTP Behaviors

`internal/adapters/transaction_upload_http_adapter.go` registers:

- `POST /api/v1/transaction-uploads` for regular multipart upload,
- `POST /api/v1/transaction-uploads/stream` for streaming progress updates,
- `GET /api/v1/transaction-uploads/:id/download` for attachment download,
- plus read and delete endpoints for uploads.

This is a good example of transport logic staying in the adapter.

`CreateUpload` handles classic HTTP multipart form submission:

1. read the file from `FormFile("file")`,
2. load the bytes,
3. build `ports.CreateTransactionUploadCommand`,
4. call `a.createUpload.Execute(...)`,
5. map validation failures to HTTP 422 or success to HTTP 201.

`CreateUploadStream` handles server-sent events:

1. configure the response as `text/event-stream`,
2. create a progress reporter bound to the Gin context,
3. read the stream request payload,
4. call `a.streamUpload.Execute(...)` with that reporter,
5. emit progress events back to the client.

The adapter owns multipart parsing, SSE mechanics, and HTTP response shaping. It does not own file validation or persistence decisions.

`DownloadUpload` handles raw-file download:

1. read the upload ID from the route,
2. read actor headers,
3. build `ports.DownloadTransactionUploadQuery`,
4. call `a.downloadUpload.Execute(...)`,
5. return the file bytes as an attachment with explicit content headers.

## 3. The Use Case Owns The Ingestion Workflow

`internal/application/use-cases/transaction_upload_create_use_case.go` contains the orchestration in `CreateTransactionUploadUseCase.Execute`.

The high-level sequence is:

1. Validate the actor.
2. Parse the actor user ID into a domain value object.
3. Parse the uploaded file through `fileParser.Parse`.
4. Validate the parsed file through the domain `TransactionFileValidator`.
5. If validation fails, return a structured result instead of persisting anything.
6. Generate a new upload ID.
7. Compute the file content MD5.
8. Check for duplicate uploads through `uploadRepository.FindByContentMD5`.
9. Store the raw file through `rawFileStore.Store`.
10. Build the upload aggregate and transaction entities through `TransactionUploadFactory.Build`.
11. Mark each transaction with the creating user.
12. Open `transactionManager.WithinTransaction`.
13. Inside the transaction, persist the upload metadata.
14. Record the upload creation event.
15. Persist the transaction rows.
16. Enqueue each transaction for background processing.
17. Mark each transaction as processing and persist that status update.
18. Publish the upload's pulled domain events.
19. Return the upload result and report completion progress.

That is application orchestration in the intended hexagonal form: the use case coordinates several ports, but does not know how any of them are implemented.

## 4. Raw File Storage Is A Separate Port

The raw file is not stored by the upload repository.

Instead, `CreateTransactionUploadUseCase.Execute` calls `rawFileStore.Store`, and `DownloadTransactionUploadUseCase.Execute` calls `rawFileStore.Read`. The concrete implementation is chosen in `buildRawFileStore` inside `internal/infrastructure/di/wire.go`.

That runtime switch is one of the clearest examples of the architecture working in practice:

- `local` selects `NewLocalRawFileStore`,
- `azure_blob` selects `NewAzureBlobRawFileStore`.

The use cases receive the same `RawFileStore` port either way for both write and read responsibilities.

## 5. SQL Persistence Is Split Across Adapters

The upload path does not have one repository that does everything.

Instead, different adapters own different persistence concerns.

### Upload metadata repository

`internal/adapters/transaction_upload_postgres_repository.go` owns the `transaction_upload` table.

`PostgresTransactionUploadRepository.Create` inserts:

- owning group ID,
- upload ID,
- file name and format,
- content MD5,
- storage provider and key,
- schema version,
- row count,
- upload timestamp.

It also supports duplicate detection through `FindByContentMD5`.

For legacy rows introduced before `group_id` existed, migration `000208` makes the ownership assumption explicit: if no source record can resolve the upload owner, the row is assigned to the seeded superadmin group `01962b8f-aeb2-7e03-a8ff-1edce1300001` because deployed legacy data uses that group only.

### Transaction row repository

`internal/adapters/transaction_postgres_repository.go` owns the `transactions` table.

For this request path, the important methods are:

- `CreateMany` for initial row insertion,
- `Update` after each transaction is marked processing.

The repository translates each `Transaction` entity into SQL insert or update arguments. It does not know why those rows are being created or queued.

### Processing queue adapter

`internal/adapters/transaction_processing_queue_postgres.go` owns the `transaction_processing_queue` table.

`Enqueue` inserts `(task_name, transaction_id)` and uses `ON CONFLICT DO NOTHING` so repeated enqueue attempts do not create duplicate queue entries.

This is a useful architectural detail: queueing work is modeled as its own outbound port, not hidden inside the transaction repository.

## 6. The Transaction Boundary Spans Several Ports

The use case wraps the database-side workflow in `transactionManager.WithinTransaction`.

Inside that transaction it coordinates multiple outbound ports:

1. `uploadRepository.Create`
2. `transactionRepo.CreateMany`
3. `processingQueue.Enqueue`
4. `transactionRepo.Update`
5. `publishDomainEvents`

This is exactly the kind of logic that often gets buried in a giant service or handler in tightly coupled systems. Here it is explicit, testable, and still independent of concrete infrastructure classes.

## 7. Validation And Progress Reporting Are Separated

The upload path also shows a useful split between business workflow and transport feedback.

- The use case decides when parsing, validation, storage, persistence, and completion milestones happen.
- The adapter decides how those milestones are represented to the client.

For the streaming endpoint, the adapter converts progress updates into SSE events. For the regular endpoint, it converts the result into JSON and HTTP status codes.

The workflow stays the same.

## 8. Cleanup Logic Stays Close To The Workflow

The use case includes compensating cleanup when later steps fail after raw file storage succeeds.

If building the upload aggregate or running the database transaction fails, it attempts `rawFileStore.Delete` using the stored key.

This is another reason the orchestration belongs in the use case. The repository layer would be the wrong place to coordinate file-store cleanup for failures that occur across multiple ports.

## 9. Why This Example Matters

This path demonstrates several boundaries at once:

- HTTP parsing and SSE setup stay in the adapter.
- Validation and workflow sequencing stay in the use case.
- File validation rules rely on domain services.
- Storage, SQL persistence, and queue persistence are split into separate outbound adapters.
- Infrastructure still decides which concrete implementations are active.

If someone is used to tightly integrating logic and implementation, this example shows the practical payoff of the extra indirection: each decision sits in the layer that can change independently.

## 10. What Would Change With In-Memory Persistence

If you wanted a teaching or unit-test version of this flow using in-memory persistence, the use case would still look almost identical.

The replaced pieces would be the outbound ports:

- an in-memory `TransactionUploadRepository`,
- an in-memory `TransactionRepository`,
- an in-memory `TransactionProcessingQueue`,
- a test `RawFileStore`,
- possibly a simplified `TransactionManager`.

The HTTP adapter could still call the same port, and the use case would still coordinate the same workflow. That is the main architectural lesson this request path exposes.

## Related Reading

- `docs/internal-wiring-overview.md`
- `docs/internal-request-flow-create-user.md`
- `docs/architecture.md`
