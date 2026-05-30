# AGENTS.md — Tests Contract

You are working in `tests/`. Read this before writing any test code here.

---

## What this directory is

Tests are organized by **scope**, not by the layer they exercise. Each scope
has different rules about what infrastructure is available and what may be
substituted with test doubles.

| Scope | Directory | Layers exercised | Infrastructure |
|---|---|---|---|
| Unit | `unit/` | `domain`, `application` | None — pure in-memory |
| Integration | `integration/` | `adapters` (via port contracts) | Real or containerized (DB, cache, queue) |
| End-to-end | `e2e/` | Full stack | Full app + all real infrastructure |

---

## Where does a new test go?

```
Does the test exercise business rules with no I/O?
  └─ YES → tests/unit/

Does the test verify that an adapter correctly implements a port contract?
  └─ YES → tests/integration/

Does the test validate a complete user scenario through the running application?
  └─ YES → tests/e2e/
```

---

## Universal rules (all scopes)

1. **Tests are not a layer** — tests may import from any src/ layer they need to
   exercise, but must not contain business logic themselves
2. **Tests never call infrastructure DI** — do not use the composition root in
   `src/infrastructure/di/`; wire test subjects manually
3. **One assertion focus per test** — each test verifies one observable outcome;
   use multiple tests for multiple behaviors
4. **Descriptive test names** — name tests as sentences: `"raises error when items list is empty"`
5. **No test depends on another test** — tests are independent; no shared mutable state

---

## Scope-specific rules

### `unit/`

- **No I/O** — no database, no network, no file system
- **No framework** — no HTTP server, no gRPC server
- **No use of real adapters** — substitute all outbound ports with in-memory
  implementations (not mocks — see below)
- **Fast** — every test completes in milliseconds

### `integration/`

- **Test the adapter, not the business logic** — business rules are unit-tested
- **Use real infrastructure** — containers (Postgres, Redis, Kafka) via Docker
  Compose or testcontainers
- **Test the full port surface** — every method on the port interface must be covered
- **Isolate test data** — run each test in a transaction that is rolled back, or
  use unique identifiers per test
- **No use cases** — call adapters directly through the port interface

### `e2e/`

- **Test scenarios, not implementation** — E2E tests have no knowledge of layers
- **Full stack** — real HTTP/gRPC client, real DB, real queue (all containerized)
- **Minimal set** — only critical user journeys; edge cases belong in unit tests
- **Unique identifiers** — each test generates its own IDs; tests do not share data

---

## Test double strategy

| Layer under test | What to substitute |
|---|---|
| Domain entities | Nothing — test them directly (no deps) |
| Domain services | Nothing — test them directly (no deps) |
| Use cases | Inject in-memory implementations of all outbound ports |
| Inbound adapters | Start a test server; call via real HTTP/gRPC/WS client |
| Outbound adapters | Test against real infrastructure in a container |

### Prefer in-memory implementations over mocks

Use **in-memory implementations** of outbound ports, not mock objects that just
record calls. In-memory implementations:

- Actually exercise the interface contract
- Are reusable across tests and test suites
- Can be initialized with fixture data
- Make test failures meaningful (they fail on wrong behavior, not wrong call order)

```
// Correct: in-memory implementation
InMemoryOrderRepository implements OrderRepository {
  store: Map<String, Order> = {}
  save(order)              → store[order.id] = order
  findById(id)             → store[id] or null
  findByCustomerId(custId) → store.values().filter(o → o.customerId == custId)
}

// Avoid: mock that records calls but doesn't behave like the real contract
MockOrderRepository {
  verify save was called with orderId "x"  // records calls, not behavior
}
```

---

## Naming conventions

| Concept | Pattern | Example |
|---|---|---|
| Unit test suite | `{Subject}Tests` | `OrderEntityTests`, `PlaceOrderUseCaseTests` |
| Integration test suite | `{AdapterClass}Tests` | `PostgresOrderRepositoryTests`, `HttpOrderAdapterTests` |
| E2E test suite | `{Scenario}E2ETests` | `OrderJourneyE2ETests`, `AuthFlowE2ETests` |
| In-memory test double | `InMemory{PortName}` | `InMemoryOrderRepository`, `InMemoryEventPublisher` |
| Test fixture builder | `build{Subject}` / `{Subject}Factory` | `buildTestOrder()`, `OrderFactory` |

---

## Self-audit checklist

Before submitting any test code:

- [ ] Test is in the correct scope directory (`unit/`, `integration/`, `e2e/`)
- [ ] Unit tests perform no I/O and use no real adapters
- [ ] Integration tests call adapters through port interfaces, not directly by class
- [ ] E2E tests use a real running application, not a wired-up partial stack
- [ ] Test doubles are in-memory implementations, not call-recording mocks
- [ ] Each test has a single assertion focus with a descriptive name
- [ ] No test depends on execution order or shared mutable state
- [ ] New in-memory implementations follow the `InMemory{PortName}` naming pattern
- [ ] Infrastructure (DB, queue) is containerized and torn down after the suite

---

## See also

- [`unit/README.md`](unit/README.md)
- [`integration/README.md`](integration/README.md)
- [`e2e/README.md`](e2e/README.md)
- Root [`AGENTS.md`](../AGENTS.md) for the full architecture contract

---

## Language-Specific Test Placement

The `tests/unit/`, `tests/integration/`, and `tests/e2e/` directories apply to **Python and TypeScript**. Go and Rust use language-idiomatic co-location instead.

### Go

| Scope | Location | Convention |
|---|---|---|
| Unit | Alongside production code in `internal/` | `*_test.go` files in the same package |
| Integration | Alongside adapter code in `internal/adapters/` | `*_test.go` files; use `testcontainers-go` |
| E2E | `tests/e2e/` | `*_test.go` files; start full app |

Go test files colocate with production code because `go test ./...` discovers `*_test.go` files by package. The `tests/unit/` and `tests/integration/` directories are **not used** for Go.

```
internal/
  domain/
    order.go
    order_test.go           ← unit test; same package
  adapters/
    postgres_order_repository.go
    postgres_order_repository_test.go  ← integration test; uses testcontainers
tests/
  e2e/                      ← full-stack E2E tests only
```

### Rust

| Scope | Location | Convention |
|---|---|---|
| Unit | Inline in production files | `#[cfg(test)] mod tests { ... }` block |
| Integration | Cargo's `tests/` directory | `.rs` files; use `testcontainers` crate |
| E2E | Cargo's `tests/` directory or `tests/e2e/` | Start full Axum app |

Rust unit tests are **inline** in the same file as the production code they test. Cargo's `tests/` directory is for integration tests (meaning the `tests/unit/` directory is **not used** for Rust).

```
src/
  domain/
    value_objects/
      money.rs              ← production code + #[cfg(test)] mod tests at bottom
tests/                      ← Cargo's integration test directory
  order_repository_test.rs  ← integration test; uses testcontainers
```

### Python and TypeScript

Use `tests/unit/`, `tests/integration/`, and `tests/e2e/` as defined in the main contract above.
