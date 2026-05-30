# AGENTS.md — Rust Language Supplement

Load this after `AGENTS.md` (root) and `src/{layer}/AGENTS.md`.
This file is authoritative for Rust idioms and tooling. It never overrides
architecture rules — if a conflict exists, the architecture rule wins.

**Stack**: Axum · sqlx · cargo test + tokio::test · manual DI only

---

## Cargo Workspace

Rust projects in this template use a Cargo workspace where each architecture layer is its own crate. This gives compile-time enforcement of the dependency rule — the `domain` crate physically cannot use `sqlx` unless it is listed in its own `Cargo.toml`.

```
Cargo.toml                     ← [workspace] members = ["crates/*"]
crates/
  domain/
    Cargo.toml                 ← deps: thiserror, rust_decimal, uuid
    src/lib.rs
  ports/
    Cargo.toml                 ← deps: domain, async-trait, tokio
    src/lib.rs
  application/
    Cargo.toml                 ← deps: domain, ports
    src/lib.rs
  adapters/
    Cargo.toml                 ← deps: ports, domain, axum, sqlx, serde
    src/lib.rs
  infrastructure/
    Cargo.toml                 ← deps: adapters, application, ports, domain, tokio
    src/lib.rs
    src/main.rs                ← binary entry point
```

If an agent places a `sqlx` import in the `domain` crate without adding it to that crate's `Cargo.toml`, `cargo check` fails immediately. The dependency graph enforces the architecture rule mechanically.

---

## Module System

Rust's module system requires explicit declaration — directories do not automatically become modules.

**`mod.rs` per directory**: within each crate, sub-directories must be declared as modules. Each directory needs a `mod.rs` that declares its sub-modules and re-exports the public API.

Example — the `domain` crate (`crates/domain/`):

```rust
// crates/domain/src/lib.rs — domain crate root
pub mod entities;
pub mod events;
pub mod value_objects;
pub mod services;
pub mod errors;

// Re-export the primary public surface
pub use errors::DomainError;
pub use value_objects::money::Money;
pub use value_objects::order_id::OrderId;
```

```rust
// crates/domain/src/entities/mod.rs
pub mod order;
pub use order::Order;
```

Typical directory structure inside one crate (e.g. `crates/domain/`):

```
crates/domain/
  Cargo.toml
  src/
    lib.rs                   ← crate root; pub mod entities; pub mod value_objects; ...
    entities/
      mod.rs                 ← pub mod order;
      order.rs
    value_objects/
      mod.rs                 ← pub mod money; pub mod order_id;
      money.rs
      order_id.rs
    errors.rs                ← single file; no sub-directory needed
```

The `infrastructure` crate contains the binary entry point:

```rust
// crates/infrastructure/src/main.rs
#[tokio::main]
async fn main() -> anyhow::Result<()> {
    infrastructure::run().await
}
```

Every `mod.rs` controls visibility upward. Only `pub use` what callers need — keep internal types unexported.

**Inline unit tests, not `tests/unit/`**

Rust unit tests live in `#[cfg(test)] mod tests { ... }` blocks inline within the same file as production code. The `tests/unit/` directory is not used for Rust. Cargo's `tests/` directory corresponds to this template's `tests/integration/`.

```rust
// crates/domain/src/value_objects/money.rs
impl Money { /* ... */ }

#[cfg(test)]
mod tests {
    use super::*;
    use rust_decimal_macros::dec;

    #[test]
    fn add_same_currency_returns_sum() {
        let a = Money::new(dec!(10.00), "USD").unwrap();
        let b = Money::new(dec!(5.00), "USD").unwrap();
        assert_eq!(a.add(&b).unwrap().amount(), dec!(15.00));
    }

    #[test]
    fn add_different_currencies_returns_error() {
        let a = Money::new(dec!(10.00), "USD").unwrap();
        let b = Money::new(dec!(5.00), "EUR").unwrap();
        assert!(a.add(&b).is_err());
    }
}
```

---

## Path Conventions

Code file paths in sections 1–11 are written relative to the containing crate's `src/` directory. Read `src/ports/outbound/order_repository.rs` as `crates/ports/src/outbound/order_repository.rs`, `src/domain/value_objects/money.rs` as `crates/domain/src/value_objects/money.rs`, and so on.

---

## 1. Port / Interface Mechanism

Rust uses `trait` for all port definitions. Async port methods require either
`#[async_trait]` (stable, broad compatibility) or RPITIT (Rust ≥ 1.75).

### Recommended: `#[async_trait]`

```rust
// src/ports/outbound/order_repository.rs
use async_trait::async_trait;

#[async_trait]
pub trait OrderRepository: Send + Sync {
    async fn save(&self, order: &Order) -> Result<(), RepositoryError>;
    async fn find_by_id(&self, id: &OrderId) -> Result<Option<Order>, RepositoryError>;
    async fn find_by_customer_id(&self, customer_id: &UserId) -> Result<Vec<Order>, RepositoryError>;
}
```

### Alternative: RPITIT (Rust ≥ 1.75, no macro needed)

```rust
pub trait OrderRepository: Send + Sync {
    fn save(&self, order: &Order) -> impl Future<Output = Result<(), RepositoryError>> + Send;
    fn find_by_id(&self, id: &OrderId) -> impl Future<Output = Result<Option<Order>, RepositoryError>> + Send;
}
```

**Rules:**
- All port traits must be `Send + Sync` — they are used across async task boundaries
- Use `#[async_trait]` unless the project targets Rust ≥ 1.75 and all team members understand RPITIT
- Port traits live in `src/ports/` only — never define them inline in adapters or use cases
- Implement port traits on concrete adapter structs with `impl OrderRepository for PostgresOrderRepository`
- Never expose Axum or sqlx types in port trait signatures

---

## 2. Value Object Immutability

Use `struct` with private fields and an `impl` block. Derive standard traits.

```rust
// src/domain/value_objects/money.rs
use rust_decimal::Decimal;
use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct Money {
    amount: Decimal,
    currency: String,
}

impl Money {
    pub fn new(amount: Decimal, currency: impl Into<String>) -> Result<Self, DomainError> {
        let currency = currency.into();
        if amount.is_sign_negative() {
            return Err(DomainError::InvalidMoney("amount cannot be negative".into()));
        }
        if currency.is_empty() {
            return Err(DomainError::InvalidMoney("currency is required".into()));
        }
        Ok(Self { amount, currency })
    }

    pub fn zero(currency: impl Into<String>) -> Self {
        Self { amount: Decimal::ZERO, currency: currency.into() }
    }

    pub fn amount(&self) -> Decimal { self.amount }
    pub fn currency(&self) -> &str  { &self.currency }

    pub fn add(&self, other: &Money) -> Result<Money, DomainError> {
        if self.currency != other.currency {
            return Err(DomainError::CurrencyMismatch {
                left: self.currency.clone(),
                right: other.currency.clone(),
            });
        }
        Ok(Self { amount: self.amount + other.amount, currency: self.currency.clone() })
    }
}
```

**Rules:**
- Private fields — accessed only via pub getters
- `derive(Debug, Clone, PartialEq, Eq, Hash)` — derive all that apply; `Hash` requires `Eq`
- Use `rust_decimal::Decimal` for monetary values — never `f64` or `f32`
- Constructor returns `Result<Self, DomainError>` — never `unwrap()` / `panic!()`
- Structs are immutable by default in Rust — no `mut` fields on value objects

---

## 3. Entity Identity and Equality

```rust
// src/domain/value_objects/order_id.rs
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct OrderId(Uuid);

impl OrderId {
    pub fn generate() -> Self {
        Self(Uuid::new_v4())
    }

    pub fn from_uuid(uuid: Uuid) -> Self {
        Self(uuid)
    }

    pub fn parse(s: &str) -> Result<Self, DomainError> {
        Uuid::parse_str(s)
            .map(Self)
            .map_err(|_| DomainError::InvalidId(s.to_string()))
    }

    pub fn as_uuid(&self) -> &Uuid { &self.0 }
}

impl fmt::Display for OrderId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}
```

Entity equality — since `Order` is a struct with an `OrderId`, derive or implement `PartialEq` comparing only the ID:

```rust
impl PartialEq for Order {
    fn eq(&self, other: &Self) -> bool {
        self.id == other.id
    }
}
```

---

## 4. Domain Error Types

Use `enum` with `thiserror::Error` derive. This is the idiomatic Rust approach.

```rust
// src/domain/errors.rs
use thiserror::Error;

#[derive(Debug, Error)]
pub enum DomainError {
    #[error("order must have at least one item")]
    InvalidOrderItems,

    #[error("cannot transition from {current} to {attempted}")]
    InvalidOrderState { current: String, attempted: String },

    #[error("currency mismatch: {left} vs {right}")]
    CurrencyMismatch { left: String, right: String },

    #[error("invalid money: {0}")]
    InvalidMoney(String),

    #[error("invalid id: {0}")]
    InvalidId(String),
}
```

Adapter error mapping — convert `DomainError` to Axum responses via `IntoResponse`:

```rust
// src/adapters/inbound/error_response.rs
impl IntoResponse for DomainError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            DomainError::InvalidOrderItems | DomainError::InvalidOrderState { .. } => {
                (StatusCode::UNPROCESSABLE_ENTITY, self.to_string())
            }
            DomainError::CurrencyMismatch { .. } => {
                (StatusCode::UNPROCESSABLE_ENTITY, self.to_string())
            }
            _ => (StatusCode::INTERNAL_SERVER_ERROR, "internal server error".to_string()),
        };
        (status, Json(json!({ "error": message }))).into_response()
    }
}
```

---

## 5. Null / Absence Handling

- Use `Option<T>` for all absent values — never raw null (Rust has no null)
- Repository methods that may not find a record return `Result<Option<T>, RepositoryError>`
- Use case checks `Option` and converts to `DomainError`:

```rust
let order = self.order_repo
    .find_by_id(&cmd.order_id)
    .await?
    .ok_or(DomainError::OrderNotFound(cmd.order_id.to_string()))?;
```

The `?` operator propagates `RepositoryError`; `.ok_or()` converts `None` to `DomainError`.

---

## 6. Error Propagation

Rust uses `Result<T, E>` and the `?` operator pervasively.

```
domain          → Result<T, DomainError>
application     → Result<T, ApplicationError> (wraps DomainError via From impl)
inbound adapter → impl IntoResponse for ApplicationError → Axum HTTP response
outbound adapter → Result<T, RepositoryError> (converted to ApplicationError via ?)
```

Define `From<DomainError> for ApplicationError` so `?` works across the boundary:

```rust
impl From<DomainError> for ApplicationError {
    fn from(e: DomainError) -> Self {
        ApplicationError::Domain(e)
    }
}
```

Never use `unwrap()` or `expect()` outside of tests.

---

## 7. Reconstitute Pattern

Two separate associated functions on the entity:

```rust
impl Order {
    /// Creates a new order — enforces invariants, emits OrderPlaced event.
    pub fn new(
        customer_id: UserId,
        items: Vec<OrderItem>,
    ) -> Result<Self, DomainError> {
        if items.is_empty() {
            return Err(DomainError::InvalidOrderItems);
        }
        let total = items.iter().try_fold(Money::zero("USD"), |acc, i| acc.add(&i.subtotal))?;
        let id = OrderId::generate();
        let mut order = Self {
            id: id.clone(),
            customer_id,
            items,
            total: total.clone(),
            status: OrderStatus::Pending,
            events: vec![],
        };
        order.events.push(DomainEvent::OrderPlaced(OrderPlaced { order_id: id, total }));
        Ok(order)
    }

    /// Hydrates an Order from storage — skips invariant checks, emits no events.
    pub fn reconstitute(
        id: OrderId,
        customer_id: UserId,
        items: Vec<OrderItem>,
        total: Money,
        status: OrderStatus,
    ) -> Self {
        Self { id, customer_id, items, total, status, events: vec![] }
    }
}
```

`reconstitute()` is called only in outbound adapters when mapping DB rows to domain types.

---

## 8. Async Conventions

Axum and sqlx are async-native. The entire adapter stack is async.

- All port trait methods are `async fn` (via `#[async_trait]` or RPITIT)
- All use case `execute()` methods are `async fn`
- All Axum handler functions are `async fn`
- All sqlx queries use `.await`
- Use `tokio` as the async runtime — configured in `main.rs` with `#[tokio::main]`
- Domain and application layer code is sync — `async` is introduced only at the adapter boundary

```rust
// Correct — domain method is sync
impl Order {
    pub fn confirm(&mut self) -> Result<(), DomainError> { ... }
}

// Correct — use case is async because it calls async ports
impl PlaceOrderUseCase {
    pub async fn execute(&self, cmd: PlaceOrderCommand) -> Result<PlaceOrderResult, ApplicationError> {
        let order = Order::new(cmd.customer_id, cmd.items)?;
        self.order_repo.save(&order).await?;
        self.event_publisher.publish_all(order.drain_events()).await?;
        Ok(PlaceOrderResult { order_id: order.id })
    }
}
```

---

## 9. Dependency Injection

Rust uses **manual constructor wiring only**. No DI container library is idiomatic.

Infrastructure resources that are shared (DB pool) are wrapped in `Arc<T>`:

```rust
// crates/infrastructure/src/di/wire.rs
pub async fn wire(cfg: &AppConfig) -> Result<Application, SetupError> {
    // 1. External connections
    let pool = PgPool::connect(&cfg.database.url).await?;
    let pool = Arc::new(pool);

    // 2. Outbound adapters (Arc the pool so it can be cloned into each adapter)
    let order_repo = Arc::new(PostgresOrderRepository::new(Arc::clone(&pool)));
    let payment_gw  = Arc::new(StripePaymentGateway::new(cfg.stripe.secret_key.clone()));
    let event_pub   = Arc::new(KafkaEventPublisher::new(cfg.kafka.brokers.clone()));

    // 3. Use cases
    let place_order = Arc::new(PlaceOrderUseCase::new(
        Arc::clone(&order_repo),
        Arc::clone(&payment_gw),
        Arc::clone(&event_pub),
    ));
    let get_order = Arc::new(GetOrderUseCase::new(Arc::clone(&order_repo)));

    // 4. Router
    let router = build_router(place_order, get_order);

    Ok(Application { router, pool })
}
```

**Rules:**
- `Arc<T>` wraps shared infrastructure resources — belongs in DI wiring, not in domain structs
- Domain and application types are NOT `Arc`-wrapped — they are owned or passed by reference
- Port trait objects in use cases are `Arc<dyn PortTrait>` — required for async + Send + Sync

---

## 10. File and Module Naming

| Concept | File name | Type name |
|---|---|---|
| Entity | `order.rs` | `Order` |
| Value object | `money.rs`, `order_id.rs` | `Money`, `OrderId` |
| Domain event | `order_placed.rs` | `OrderPlaced` |
| Domain error | `errors.rs` | `DomainError` (enum) |
| Use case | `place_order.rs` | `PlaceOrderUseCase` |
| Inbound port | `place_order_port.rs` | `PlaceOrderPort` |
| Outbound port | `order_repository.rs` | `OrderRepository` |
| Inbound adapter | `http_order_adapter.rs` | handler functions + `build_router()` |
| Outbound adapter | `postgres_order_repository.rs` | `PostgresOrderRepository` |
| Config | `config.rs` | `AppConfig`, `DatabaseConfig` |

**Rules:**
- `snake_case` for all file names, module names, function names, variable names
- `PascalCase` for all type names (structs, enums, traits)
- `SCREAMING_SNAKE_CASE` for constants
- Module structure mirrors directory structure — each `mod.rs` re-exports the public API
- Avoid deeply nested modules — keep the layer hierarchy flat

---

## 11. Axum Integration

Axum handler functions are **inbound adapters**. They live in the `adapters` crate (`crates/adapters/src/inbound/`).

```rust
// crates/adapters/src/inbound/http_order_adapter.rs
use axum::{extract::State, Json, http::StatusCode};

#[derive(Clone)]
pub struct OrderHandlerState {
    pub place_order: Arc<dyn PlaceOrderPort>,
    pub get_order:   Arc<dyn GetOrderPort>,
}

pub async fn place_order(
    State(state): State<OrderHandlerState>,
    Json(body): Json<PlaceOrderRequest>,
) -> Result<(StatusCode, Json<PlaceOrderResponse>), DomainError> {
    let command = body.try_into_command()?;
    let result = state.place_order.execute(command).await?;
    Ok((StatusCode::CREATED, Json(PlaceOrderResponse::from(result))))
}

pub fn order_router(state: OrderHandlerState) -> Router {
    Router::new()
        .route("/orders", post(place_order))
        .route("/orders/:id", get(get_order_handler))
        .with_state(state)
}
```

**Rules:**
- Handlers receive port interfaces via Axum `State` extractor — never use case structs directly
- Use serde `Deserialize` / `Serialize` for request/response types (in adapter layer only)
- Return `Result<T, DomainError>` where `DomainError: IntoResponse` — Axum calls `into_response()` on errors
- Do not put business logic in handler functions — call the port, map the result, return

---

## 12. Forbidden Patterns

| Pattern | Why forbidden |
|---|---|
| `f64` / `f32` for monetary values | Use `rust_decimal::Decimal` — floats lose precision |
| `unwrap()` / `expect()` outside tests | Use `?` or match — panics crash the process |
| `unsafe` in domain/application/ports | No unsafe outside infrastructure (if ever) |
| `static` mutable global state | Use `Arc<T>` passed via DI — globals break test isolation |
| Domain types without `Clone` | Async adapters need to own values — `Clone` prevents lifetime fights |
| `Arc<T>` in domain structs | `Arc` is an infrastructure concern — inject it via DI wiring |
| `Box<dyn Error>` in domain/application | Use typed error enums — `Box<dyn Error>` loses type information |
| `#[allow(dead_code)]` in production code | Fix the actual issue — don't suppress warnings |
| Skipping `clippy` | `clippy` catches real bugs — run it in CI |

---

## 13. Tooling

| Tool | Purpose | Minimum config |
|---|---|---|
| `cargo check` | Fast compilation check | Run on every save |
| `clippy` | Linting | `cargo clippy -- -D warnings` |
| `rustfmt` | Formatting | `rustfmt.toml` with `edition = "2021"` |
| `cargo test` | Test runner | Built-in |
| `tokio::test` | Async test support | `#[tokio::test]` macro on async tests |
| `testcontainers` | Real DB in integration tests | `testcontainers` crate |
| `cargo deny` | Dependency auditing | `deny.toml` |

Minimum `rustfmt.toml`:
```toml
edition = "2021"
max_width = 100
```

Minimum CI check:
```sh
cargo fmt --check
cargo clippy -- -D warnings
cargo test
```

---

## 14. Self-Audit Checklist

Before submitting any Rust code:

- [ ] All port traits are `Send + Sync` with `#[async_trait]` or RPITIT
- [ ] Value objects have private fields and derive `Debug, Clone, PartialEq`
- [ ] `rust_decimal::Decimal` used for monetary values, never `f64`
- [ ] All constructors return `Result<Self, DomainError>`, never panic
- [ ] Entities have both `Order::new()` and `Order::reconstitute()` associated functions
- [ ] `Order::reconstitute()` emits no domain events
- [ ] Domain errors use `thiserror` enum with descriptive variants
- [ ] `DomainError: IntoResponse` implemented in the inbound adapter layer
- [ ] `Arc<dyn PortTrait>` used for port injection in use cases
- [ ] No `unwrap()` or `expect()` outside test code
- [ ] No `unsafe` in domain, application, or ports layers
- [ ] Unit tests are inline `#[cfg(test)] mod tests` blocks, not in `tests/unit/`
- [ ] Integration tests are in Cargo's `tests/` directory, not `tests/integration/`
- [ ] `cargo clippy -- -D warnings` passes
- [ ] `cargo fmt --check` passes
- [ ] `cargo test` passes

---

## See also

- [`examples/domain.md`](examples/domain.md)
- [`examples/application.md`](examples/application.md)
- [`examples/ports.md`](examples/ports.md)
- [`examples/adapters.md`](examples/adapters.md)
- [`examples/infrastructure.md`](examples/infrastructure.md)
- [`examples/tests.md`](examples/tests.md)
- Root [`AGENTS.md`](../../../AGENTS.md)
