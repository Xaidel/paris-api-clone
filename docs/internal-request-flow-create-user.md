# Request Flow: Create User

This walkthrough traces one of the clearest request paths in the codebase:

- request: `POST /api/v1/users`
- inbound adapter: `HttpUserAdapter`
- application use case: `CreateUserUseCase`
- persistence adapter: `PostgresUserRepository`

It is a good first example because the flow is short enough to follow without hiding the important boundaries.

## 1. Composition Root Chooses The Concrete Pieces

`internal/infrastructure/di/wire.go` assembles the request path before the server starts.

For user creation, the important constructor chain is:

1. `userRepository := adapters.NewPostgresUserRepository(pool)`
2. `groupRepository := adapters.NewPostgresGroupRepository(pool)`
3. `passwordHasher := adapters.NewBcryptPasswordHasher()`
4. `transactionManager := adapters.NewPgxTransactionManager(pool)`
5. `eventRecorderService := services.NewEventRecorderService(...)`
6. `actorDirectory := adapters.NewPostgresActorDirectory(pool)`
7. `createUserUseCase := usecases.NewCreateUserUseCase(...)`
8. `httpUserAdapter := adapters.NewHttpUserAdapter(...)`
9. `httpserver.NewRouter(logger, httpUserAdapter, ...)`

That is the first hexagonal lesson in this repo: the handler does not create the use case, and the use case does not create the repository.

## 2. The HTTP Adapter Owns HTTP

`internal/adapters/user_http_adapter.go` defines both route registration and request handling.

`RegisterRoutes` creates the `/api/v1/users` group and attaches:

- `POST ""` to `CreateUser`
- `GET ""` to `ListUsers`
- `GET "/:id"` to `GetUser`
- `PUT "/:id"` to `UpdateUser`
- `DELETE "/:id"` to `DeleteUser`

The `CreateUser` handler does only transport-facing work:

1. Decode the JSON body into `createUserRequest`.
2. Read actor headers from the request.
3. Build `ports.CreateUserCommand`.
4. Call `a.createUser.Execute(...)`.
5. Map success to HTTP 201 or map errors via `handleError`.

What it does not do is just as important:

- it does not hash passwords,
- it does not verify the group exists,
- it does not open a database transaction,
- it does not write SQL.

All of that is pushed behind ports.

## 3. The Use Case Owns Orchestration

`internal/application/use-cases/user_create_use_case.go` contains the application workflow in `CreateUserUseCase.Execute`.

Its dependencies are all ports:

- `UserRepository`
- `GroupRepository`
- `PasswordHasher`
- `EventPublisher`
- `TransactionManager`
- `ActorDirectory`

The execution sequence is:

1. Validate the actor with `validateActor` and `ActorDirectory`.
2. Validate the plaintext password by constructing a domain value object.
3. Parse the `GroupID` string into a domain value object.
4. Load the group through `groupRepository.FindByID`.
5. Fail with `NotFoundError` if the group does not exist.
6. Generate a new `UserID`.
7. Hash the password through `passwordHasher.Hash`.
8. Build `UserProfile` and `User` domain objects.
9. Record the creation domain event on the user entity.
10. Open `transactionManager.WithinTransaction`.
11. Inside the transaction, call `userRepository.Create`.
12. Still inside the transaction, publish the user's pulled domain events through `eventPublisher`.
13. Return `ports.UserResult`.

This file is the center of the request path. It knows the business sequence, but it has no knowledge of Gin, SQL, or connection pools.

## 4. Domain Types Own Invariants

The use case constructs domain value objects and entities before persistence:

- `valueobjects.NewPlaintextPassword`
- `valueobjects.GroupIDFromString`
- `valueobjects.NewUserProfile`
- `entities.NewUser`

That is where the request stops being “an HTTP payload” and becomes validated domain data. The adapter only forwards strings. The use case turns those strings into domain-safe values.

## 5. The Transaction Boundary Lives Above The Repository

One common tightly coupled pattern is to put transaction logic inside the handler or repository. This code does neither.

`CreateUserUseCase.Execute` uses `TransactionManager.WithinTransaction` because the application layer decides that these operations belong in one unit of work:

1. create the user record,
2. publish the resulting domain events.

That is an application-level orchestration decision, not an HTTP concern and not a SQL concern.

## 6. The Repository Owns SQL Translation

`internal/adapters/user_postgres_repository.go` is where the `UserRepository` port becomes SQL.

`PostgresUserRepository.Create`:

1. Resolves the active querier from context via `txQuerierFromContext`.
2. Executes `createUserQuery` to insert into the `user` table.
3. Extracts the optional middle name from the domain object.
4. Executes `createUserProfileQuery` to insert into `user_profile`.

This repository does not decide whether the user should be created. By the time it runs, the decision has already been made by the use case and the domain objects.

Its job is translation:

- domain entity fields to SQL parameters,
- transaction-aware context to `pgx` execution,
- database errors to wrapped Go errors.

## 7. Error Mapping Returns To HTTP At The Edge

Control returns to `HttpUserAdapter.CreateUser` after the use case finishes.

The adapter's `handleError` maps:

- `NotFoundError` to HTTP 404,
- `domain.DomainError` to HTTP 422,
- everything else to HTTP 500.

That keeps transport semantics at the edge. The use case returns Go errors, not HTTP responses.

## 8. Why This Example Matters

This request path shows the pattern the rest of the codebase repeats:

- infrastructure picks implementations,
- the HTTP adapter translates transport,
- the use case owns workflow,
- domain types enforce invariants,
- the repository translates to SQL,
- the adapter maps the result back to HTTP.

For a reader coming from a tightly integrated design, the main shift is that no single file owns the whole request. The composition root, adapter, use case, and repository each own one kind of decision.

## 9. What Would Change With In-Memory Persistence

If this flow were tested with an in-memory repository, the use case code would stay the same.

Only the composition root for the test would change:

- `UserRepository` would be backed by an in-memory implementation,
- `GroupRepository` could also be in-memory,
- `TransactionManager` might become a no-op test implementation,
- the HTTP adapter would still call the same `CreateUserPort`.

That is the architectural payoff. The use case does not care whether `UserRepository` is Postgres-backed or memory-backed.

## Related Reading

- `docs/internal-wiring-overview.md`
- `docs/internal-request-flow-transaction-upload.md`
- `docs/architecture.md`
