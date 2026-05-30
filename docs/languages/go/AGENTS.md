# AGENTS.md — Go Language Supplement

Load this after `AGENTS.md` (root) and `src/{layer}/AGENTS.md`.
This file is authoritative for Go idioms and tooling. It never overrides
architecture rules — if a conflict exists, the architecture rule wins.

**Stack**: Gin · pgx · testing + testify · manual DI only

---

## Directory Structure

Go overrides two defaults from the language-neutral architecture contract.

**`internal/` replaces `src/`**

Go's `internal/` directory is enforced by the toolchain — no code outside this module can import from `internal/`. All layer paths shift accordingly:

| Language-neutral path | Go path |
|---|---|
| `src/domain/` | `internal/domain/` |
| `src/ports/` | `internal/ports/` |
| `src/adapters/` | `internal/adapters/` |
| `src/application/` | `internal/application/` |
| `src/infrastructure/` | `internal/infrastructure/` |

**`internal/ports/` is a single flat package**

`internal/ports/inbound/` and `internal/ports/outbound/` do not exist as separate sub-packages. All port definitions live in a single `package ports` at `internal/ports/`. The inbound/outbound distinction is preserved through naming conventions within the package (`PlaceOrderPort` = inbound; `OrderRepository`, `PaymentGateway` = outbound).

Splitting `internal/adapters/` into sub-packages is optional. When the project grows, prefer splitting by technology over splitting by direction: `internal/adapters/gin/`, `internal/adapters/postgres/`, `internal/adapters/kafka/`.

**Entry point lives outside `internal/`**

The application entry point is `cmd/myapp/main.go`. It imports `internal/infrastructure/di` and calls the wiring function. Nothing else provides a `main()`.

```
cmd/
  myapp/
    main.go             ← binary entry point; calls di.Wire()
internal/
  domain/
  ports/                ← single flat package; all port interfaces here
  adapters/
  application/
  infrastructure/
    di/
      wire.go
```

**Unit tests colocate with production code**

Go unit tests are `*_test.go` files that live alongside production code inside `internal/`. The `tests/unit/` directory is not used for Go. Integration tests are also `*_test.go` files within the adapter package they exercise.

---

## 1. Port / Interface Mechanism

Go uses **implicit interface satisfaction**. A type implements an interface by having
all the methods — no `implements` keyword, no inheritance.

```go
// internal/ports/order_repository.go
package ports

import "context"

type OrderRepository interface {
    Save(ctx context.Context, order *domain.Order) error
    FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
    FindByCustomerID(ctx context.Context, customerID domain.UserID) ([]*domain.Order, error)
}
```

**Rules:**
- Ports are Go `interface` types — one file per port, one interface per file
- The compiler checks interface satisfaction **at the assignment site** (in DI wiring), not at struct definition
- Concrete adapters do NOT declare `var _ ports.OrderRepository = (*PostgresOrderRepository)(nil)` (compile-time check) — but this pattern is acceptable if the team wants early error detection
- Pass `context.Context` as the first parameter to all port methods — this is idiomatic Go
- Accept interfaces, return structs: function parameters use the interface type; constructors return concrete types

**Forbidden:**
```go
// WRONG — Go has no implements keyword; this doesn't exist
type PostgresOrderRepository implements ports.OrderRepository struct { ... }
```

---

## 2. Value Object Immutability

Go has no `const` structs. Simulate immutability with unexported fields and constructor functions.

```go
// internal/domain/value_objects/money.go
package domain

import (
    "errors"
    "github.com/shopspring/decimal"
)

type Money struct {
    amount   decimal.Decimal
    currency string
}

func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
    if amount.IsNegative() {
        return Money{}, errors.New("money amount cannot be negative")
    }
    if currency == "" {
        return Money{}, errors.New("currency is required")
    }
    return Money{amount: amount, currency: currency}, nil
}

func ZeroMoney(currency string) Money {
    return Money{amount: decimal.Zero, currency: currency}
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() string        { return m.currency }

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.currency, other.currency)
    }
    return Money{amount: m.amount.Add(other.amount), currency: m.currency}, nil
}

func (m Money) Equal(other Money) bool {
    return m.amount.Equal(other.amount) && m.currency == other.currency
}
```

**Rules:**
- Unexported fields (`amount`, `currency`) — accessed only via exported getters
- Value receivers (not pointer receivers) for all methods — value types are copied, not mutated
- Use `github.com/shopspring/decimal` for monetary values — never `float64`
- Constructor returns `(T, error)` — never panics

---

## 3. Entity Identity and Equality

```go
// internal/domain/value_objects/order_id.go
package domain

import "github.com/google/uuid"

type OrderID struct {
    value uuid.UUID
}

func NewOrderID() OrderID {
    return OrderID{value: uuid.New()}
}

func OrderIDFromString(s string) (OrderID, error) {
    id, err := uuid.Parse(s)
    if err != nil {
        return OrderID{}, fmt.Errorf("invalid order id: %w", err)
    }
    return OrderID{value: id}, nil
}

func (id OrderID) String() string    { return id.value.String() }
func (id OrderID) Equal(other OrderID) bool { return id.value == other.value }
```

Entity equality — implement an `Equal()` method; never rely on struct equality (`==`) for entities with pointer receivers:

```go
func (o *Order) Equal(other *Order) bool {
    return o.id.Equal(other.id)
}
```

---

## 4. Domain Error Types

Use typed error structs implementing the `error` interface. Use `errors.New` sentinels for simple leaf errors.

```go
// internal/domain/errors.go
package domain

import "fmt"

// DomainError is the base marker interface for all domain errors.
// Use type assertions in adapters: if errors.As(err, &DomainError{}) { ... }
type DomainError struct {
    Code    string
    Message string
}

func (e *DomainError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Specific errors
var ErrInvalidOrderItems = &DomainError{Code: "INVALID_ORDER_ITEMS", Message: "order must have at least one item"}
var ErrInvalidOrderState = &DomainError{Code: "INVALID_ORDER_STATE", Message: "state transition not permitted"}

type OrderNotFoundError struct {
    OrderID string
}

func (e *OrderNotFoundError) Error() string {
    return fmt.Sprintf("order %s not found", e.OrderID)
}
```

Adapter error mapping:

```go
var domainErr *domain.DomainError
var notFoundErr *domain.OrderNotFoundError
switch {
case errors.As(err, &notFoundErr):
    c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
case errors.As(err, &domainErr):
    c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
default:
    c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
```

---

## 5. Null / Absence Handling

- Return `(*T, error)` from repository methods — `nil` for not found, `error` for actual failures
- Distinguish "not found" (return `nil, nil`) from "error" (return `nil, err`) in repositories
- Use case checks for `nil` and returns a typed error:

```go
order, err := r.orderRepo.FindByID(ctx, cmd.OrderID)
if err != nil {
    return nil, fmt.Errorf("finding order: %w", err)
}
if order == nil {
    return nil, &domain.OrderNotFoundError{OrderID: cmd.OrderID.String()}
}
```

---

## 6. Error Propagation

Go uses `(T, error)` return pairs. Always wrap errors with context using `fmt.Errorf("doing X: %w", err)`.

```
domain          → returns typed DomainError values
application     → wraps and re-returns errors; never swallows
inbound adapter → errors.As() checks → gin.JSON(status, body)
```

Never use `panic` for domain or application errors. `panic` is reserved for programming errors (nil dereference, etc.) and infrastructure startup failures only.

---

## 7. Reconstitute Pattern

Two separate constructor functions — never overload in Go:

```go
// NewOrder — enforces invariants, emits OrderPlaced event
func NewOrder(id OrderID, customerID UserID, items []OrderItem) (*Order, error) {
    if len(items) == 0 {
        return nil, ErrInvalidOrderItems
    }
    total := calculateTotal(items)
    o := &Order{
        id:         id,
        customerID: customerID,
        items:      items,
        total:      total,
        status:     OrderStatusPending,
        events:     []DomainEvent{},
    }
    o.events = append(o.events, NewOrderPlaced(id, customerID, total))
    return o, nil
}

// ReconstitueOrder — hydrates from storage, skips invariant checks, emits no events
func ReconstitueOrder(id OrderID, customerID UserID, items []OrderItem, total Money, status OrderStatus) *Order {
    return &Order{
        id:         id,
        customerID: customerID,
        items:      items,
        total:      total,
        status:     status,
        events:     []DomainEvent{},
    }
}
```

`ReconstitueOrder` is called only in outbound adapters when hydrating from the DB.

---

## 8. Async Conventions

Go uses goroutines, not async/await. Gin handlers are synchronous blocking calls.

- **Do not** spawn goroutines in domain or application layer code
- **Do not** use channels in domain or application layer code
- Gin handlers block while the use case executes — this is correct and expected
- pgx connection pool handles concurrent queries internally; no extra goroutine management needed
- Background work (e.g. event publishing) belongs in infrastructure, not in use cases

```go
// Correct — synchronous handler
func (a *HttpOrderAdapter) PlaceOrder(c *gin.Context) {
    var req PlaceOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    result, err := a.placeOrderPort.Execute(c.Request.Context(), toCommand(req))
    if err != nil {
        a.handleError(c, err)
        return
    }
    c.JSON(http.StatusCreated, toResponse(result))
}
```

---

## 9. Dependency Injection

Go uses **manual constructor wiring only**. There is no idiomatic DI container library.

```go
// internal/infrastructure/di/wire.go
package di

func Wire(cfg *config.AppConfig) (*Application, error) {
    // 1. External connections
    pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
    if err != nil {
        return nil, fmt.Errorf("creating db pool: %w", err)
    }

    // 2. Outbound adapters
    orderRepo := adapters.NewPostgresOrderRepository(pool)
    paymentGateway := adapters.NewStripePaymentGateway(cfg.Stripe.APIKey)

    // 3. Use cases
    placeOrderUC := usecases.NewPlaceOrderUseCase(orderRepo, paymentGateway)
    getOrderUC := usecases.NewGetOrderUseCase(orderRepo)

    // 4. Inbound adapters
    orderAdapter := adapters.NewHttpOrderAdapter(placeOrderUC, getOrderUC)

    // 5. Router
    router := gin.New()
    orderAdapter.RegisterRoutes(router)

    return &Application{router: router, pool: pool}, nil
}
```

---

## 10. File and Package Naming

| Concept | File name | Type name |
|---|---|---|
| Entity | `order.go` | `Order` |
| Value object | `money.go`, `order_id.go` | `Money`, `OrderID` |
| Domain event | `order_placed.go` | `OrderPlaced` |
| Domain error | `errors.go` | `DomainError`, `OrderNotFoundError` |
| Use case | `order_place_use_case.go` | `PlaceOrderUseCase` |
| Inbound port | `order_place_port.go` | `PlaceOrderPort` |
| Outbound port | `order_repository.go` | `OrderRepository` |
| Inbound adapter | `order_http_adapter.go` | `HttpOrderAdapter` |
| Outbound adapter | `order_postgres_repository.go` | `PostgresOrderRepository` |
| Config | `config.go` | `AppConfig`, `DatabaseConfig` |

**Rules:**
- `snake_case` for all file names
- Prefer noun-first file stems so related feature files group together in the editor
- Keep exported type names concept-first per the architecture naming table, even when the file stem is noun-first
- Mirror production file stems in tests, for example `order_place_use_case_test.go` and `order_http_adapter_test.go`
- `PascalCase` for all exported identifiers
- `camelCase` for unexported identifiers
- Package names: short, lowercase, no underscores (`domain`, `ports`, `adapters`, `usecases`, `di`)
- One package per layer sub-directory — do not create deep package hierarchies
- Avoid `init()` functions for wiring; wire explicitly in `di/wire.go`

---

## 11. Gin Integration

Gin handlers are **inbound adapters**. They live in `internal/adapters/` (or `internal/adapters/gin/` in larger projects).

```go
// internal/adapters/order_http_adapter.go
package adapters

type HttpOrderAdapter struct {
    placeOrder ports.PlaceOrderPort
    getOrder   ports.GetOrderPort
}

func NewHttpOrderAdapter(placeOrder ports.PlaceOrderPort, getOrder ports.GetOrderPort) *HttpOrderAdapter {
    return &HttpOrderAdapter{placeOrder: placeOrder, getOrder: getOrder}
}

func (a *HttpOrderAdapter) RegisterRoutes(r *gin.Engine) {
    orders := r.Group("/orders")
    orders.POST("/", a.PlaceOrder)
    orders.GET("/:id", a.GetOrder)
}
```

**Rules:**
- Handlers receive port interfaces in their adapter struct — never use case structs directly
- Use `c.ShouldBindJSON()` for request parsing (returns an error, does not abort)
- Use `c.JSON()` for all responses — never write raw `c.Writer`
- All error mapping lives in a shared `handleError(c *gin.Context, err error)` method on the adapter

---

## 12. Forbidden Patterns

| Pattern | Why forbidden |
|---|---|
| `float64` for monetary values | Use `shopspring/decimal` — floats lose precision |
| `panic` in domain/application | Use `(T, error)` returns — panic only for programming errors |
| `interface{}` / `any` in domain layer | Use concrete types — generic containers leak into domain |
| `init()` for wiring | Wire explicitly in `di/wire.go` — init() is order-dependent and hard to test |
| Global mutable vars outside infrastructure | Package-level vars in domain/application break test isolation |
| Goroutines in domain/application | Concurrency is an infrastructure concern |
| Pointer receivers on value objects | Value objects use value receivers — they are copied, not mutated |
| Direct struct field access across packages | Use exported getter methods on domain types |

---

## 13. Tooling

| Tool | Purpose | Minimum config |
|---|---|---|
| `go vet` | Static analysis | Run on every build |
| `staticcheck` | Extended static analysis | `staticcheck ./...` |
| `golangci-lint` | Linting | `.golangci.yml` with `govet`, `staticcheck`, `errcheck` |
| `gofmt` / `goimports` | Formatting | Run on save; enforced in CI |
| `go test` + `testify` | Test runner | `testify/assert` and `testify/require` |
| `testcontainers-go` | Real DB in integration tests | — |

Minimum `.golangci.yml`:
```yaml
linters:
  enable:
    - govet
    - staticcheck
    - errcheck
    - gosimple
    - ineffassign
    - unused
```

---

## 14. Self-Audit Checklist

Before submitting any Go code:

- [ ] All ports are `interface` types with `context.Context` as first parameter
- [ ] Value objects use unexported fields with value receivers
- [ ] `shopspring/decimal` used for monetary values, never `float64`
- [ ] All constructor functions return `(T, error)`, never panic
- [ ] Entities have both `NewOrder()` and `ReconstitueOrder()` constructors
- [ ] `ReconstitueOrder()` emits no domain events
- [ ] Errors are wrapped with `fmt.Errorf("...: %w", err)` for context
- [ ] Adapter error mapping uses `errors.As()` for typed domain errors
- [ ] No goroutines in domain or application layer
- [ ] No `init()` functions used for wiring
- [ ] Entry point is in `cmd/`, not in `internal/`
- [ ] Port definitions are in `internal/ports/` (flat, not `inbound/`/`outbound/` sub-packages)
- [ ] Unit tests are `*_test.go` files colocated with production code, not in `tests/unit/`
- [ ] `go vet ./...` passes
- [ ] `golangci-lint run` passes

---

## See also

- [`examples/domain.md`](examples/domain.md)
- [`examples/application.md`](examples/application.md)
- [`examples/ports.md`](examples/ports.md)
- [`examples/adapters.md`](examples/adapters.md)
- [`examples/infrastructure.md`](examples/infrastructure.md)
- [`examples/tests.md`](examples/tests.md)
- Root [`AGENTS.md`](../../../AGENTS.md)
