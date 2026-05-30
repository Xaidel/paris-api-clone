# Internal Wiring Overview

This guide explains how the code under `internal/` is wired today, from process startup to HTTP route handling to persistence. It is meant to complement `docs/architecture.md`: that document explains the architecture contract, while this one shows how the current Go implementation realizes it.

The key point is simple: runtime wiring is decided once in the composition root, then requests flow through adapters, ports, use cases, domain types, and outbound adapters. The business logic does not instantiate infrastructure for itself.

## Start At The Composition Root

The server entrypoint in `cmd/server/main.go` calls `di.Bootstrap(ctx)`, then starts the returned HTTP server and any configured workers.

`internal/infrastructure/di/wire.go` is the runtime source of truth.

`Bootstrap` does four jobs in order:

1. Load configuration and observability dependencies.
2. Establish infrastructure connections, including the Postgres pool.
3. Instantiate concrete outbound adapters, use cases, services, and inbound HTTP adapters.
4. Register all HTTP adapters with the router and return a runnable `Application`.

That file is intentionally the only place that knows all concrete types at once. Everywhere else depends on ports or domain types.

## Startup Flow

The runtime boot sequence is:

1. `cmd/server/main.go` creates a signal-aware context and calls `di.Bootstrap(ctx)`.
2. `Bootstrap` loads config via `internal/infrastructure/config` and initializes logging via `internal/infrastructure/observability`.
3. Database migrations run before the pool is opened.
4. The Postgres pool is created and passed into repository adapters such as `NewPostgresUserRepository`, `NewPostgresTransactionRepository`, and `NewPostgresTransactionUploadRepository`.
5. Infrastructure-only dependencies are built, including `NewPgxTransactionManager`, `NewBcryptPasswordHasher`, and the raw file store selected by `buildRawFileStore` for upload-file writes and reads.
6. Application services and use cases are constructed against ports, not against transport types.
7. HTTP adapters are constructed with those use cases.
8. `internal/infrastructure/httpserver/router.go` builds the Gin router and calls `RegisterRoutes` on every adapter.
9. `Bootstrap` returns an `Application` holding the router, server, workers, and selected ports.

## Router Pattern

The HTTP server code uses a narrow registration contract:

- `internal/infrastructure/httpserver/router.go` defines `RouteRegistrar` with one method: `RegisterRoutes(engine *gin.Engine)`.
- Every inbound adapter implements that method.
- `NewRouter` creates a Gin engine, adds middleware and `/healthz`, then loops over registrars and lets each adapter attach its own routes.

That means route ownership lives with the adapter, but transport assembly still belongs to infrastructure.

## Repeating Request Pattern

If you are used to a tightly coupled handler-service-repository stack, trace a request in this repo with this sequence:

1. Infrastructure decides which adapter instances exist.
2. The HTTP adapter decodes transport input, reads headers and params, and maps them to a port command or query.
3. The use case validates application-level inputs, constructs domain objects, and orchestrates outbound ports.
4. Outbound adapters translate those port calls into I/O such as SQL, queue inserts, or file storage operations.
5. The use case returns a port result.
6. The HTTP adapter maps that result or error back to HTTP.

The important boundary is that only the adapter knows about Gin and only the outbound adapter knows about SQL or storage SDKs.

## Runtime Implementation Choices

Today, the main runtime path is Postgres-backed.

- User, group, transaction, upload, feedback, bug report, audit, and classification repositories are instantiated as `Postgres...` adapters in `internal/infrastructure/di/wire.go`.
- The transaction manager is `NewPgxTransactionManager(pool)`, so application-level transactions run on Postgres transactions.
- Background classification work is enqueued through `NewPostgresTransactionProcessingQueue(pool)` and `NewPostgresClassificationJobQueue(pool)`.
- The raw file store is selected at startup by `buildRawFileStore`:
  - `local` uses `NewLocalRawFileStore`.
  - `azure_blob` uses `NewAzureBlobRawFileStore`.

This distinction matters for documentation: the codebase is architected so repositories could be swapped, but the current runnable server is not using in-memory repositories.

## Adapter To Use Case Wiring Map

The following examples are the most useful starting points when tracing requests:

| HTTP adapter | Route group | Primary use cases | Main outbound dependencies underneath |
| --- | --- | --- | --- |
| `HttpUserAdapter` | `/api/v1/users` | create/get/list/update/delete user | `UserRepository`, `GroupRepository`, `PasswordHasher`, `TransactionManager`, `EventPublisher`, `ActorDirectory` |
| `HttpTransactionAdapter` | `/api/v1/transactions` | create/get/list/delete transaction | `TransactionRepository`, `TransactionProcessingQueue`, `TransactionManager`, `EventPublisher`, `ActorDirectory` |
| `HttpTransactionUploadAdapter` | `/api/v1/transaction-uploads` | create/get/list/download/delete upload | `TransactionUploadRepository`, `TransactionRepository`, `TransactionProcessingQueue`, `RawFileStore`, `TransactionFileParser`, `TransactionManager`, `EventPublisher`, `ActorDirectory` |
| `HttpTransactionFeedbackAdapter` | `/api/v1/transactions/:id/feedback` | upsert/delete/get feedback | `FeedbackRepository`, `TransactionRepository`, `TransactionManager`, `EventPublisher` |
| `HttpBugReportAdapter` | `/api/v1/bug-reports` | create/get/list/update/delete bug report | `BugReportRepository`, `TransactionRepository`, `TransactionManager`, `EventPublisher` |

This is not a full inventory of every adapter in the system. It is the shortest map that reveals the pattern repeated across the codebase.

## Example 1: User Creation Wiring

The simple path is:

1. `HttpUserAdapter.RegisterRoutes` attaches `POST /api/v1/users`.
2. `HttpUserAdapter.CreateUser` decodes JSON and builds `ports.CreateUserCommand`.
3. `CreateUserUseCase.Execute` validates the actor, loads the target group, hashes the password, creates the user entity, records a domain event, and opens a transaction.
4. Inside the transaction, `UserRepository.Create` persists the user and `EventPublisher` publishes pulled domain events.
5. `PostgresUserRepository.Create` translates the entity into two SQL inserts: one for `user`, one for `user_profile`.
6. The adapter maps the result to HTTP 201 or maps errors to HTTP status codes.

See `docs/internal-request-flow-create-user.md` for the full trace.

## Example 2: Transaction Upload Wiring

The richer path is:

1. `HttpTransactionUploadAdapter.RegisterRoutes` attaches multipart upload, streaming upload, upload-read, and file-download endpoints.
2. The adapter reads the file, form fields, or stream payload and builds `ports.CreateTransactionUploadCommand`.
3. `CreateTransactionUploadUseCase.Execute` parses the file, validates it, checks for duplicates, stores the raw file, builds the upload aggregate plus transaction entities, and then opens a transaction.
4. Inside the transaction, it persists upload metadata, persists transaction rows, enqueues processing work, updates transaction status, and publishes domain events.
5. Outbound adapters split responsibilities across SQL persistence, queue persistence, and raw file storage.
6. The adapter returns either a normal JSON response, a binary attachment download, or server-sent events progress updates.

See `docs/internal-request-flow-transaction-upload.md` for the full trace.

## Where In-Memory Fits

This codebase's architecture docs and testing guidance still recommend the classic port-first progression:

1. define the port,
2. write the use case,
3. test it with an in-memory implementation,
4. implement the real adapter,
5. wire the real adapter in infrastructure.

That guidance appears in `docs/development-workflow.md` and `tests/AGENTS.md`.

The current code does not yet provide shared in-memory repository implementations for the main server paths. Instead:

- runtime wiring in `internal/infrastructure/di/wire.go` selects Postgres-backed repositories,
- adapter tests often use `pgxmock`,
- use case tests often use small mock structs defined in `*_test.go` files.

So the right mental model is:

- in-memory repositories are part of the intended testing style,
- Postgres repositories are the current production-style runtime implementation,
- this guide documents the real runtime path first.

## How To Read The Code Efficiently

For any request path, start in this order:

1. `cmd/server/main.go`
2. `internal/infrastructure/di/wire.go`
3. `internal/infrastructure/httpserver/router.go`
4. the HTTP adapter's `RegisterRoutes` and handler method
5. the use case's `Execute`
6. the outbound adapters instantiated for that use case

That sequence keeps the concrete wiring visible and avoids mistaking a port for an implementation.

## Related Documents

- `docs/architecture.md` explains the architectural contract.
- `docs/development-workflow.md` explains the intended port-first delivery loop.
- `docs/internal-request-flow-create-user.md` walks a simple CRUD request end to end.
- `docs/internal-request-flow-transaction-upload.md` walks a multi-port ingestion request end to end.
