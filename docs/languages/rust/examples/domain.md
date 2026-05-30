# Rust — Domain Layer Examples

Real Rust code illustrating `src/domain/` patterns.
All examples use the Order / Payment domain.

---

## Errors (`src/domain/errors.rs`)

```rust
use thiserror::Error;

/// All domain errors in one enum. Adapters match on variants to map to HTTP status codes.
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

    #[error("order not found: {0}")]
    OrderNotFound(String),
}
```

---

## Value Objects

### `src/domain/value_objects/order_id.rs`

```rust
use std::fmt;
use uuid::Uuid;
use crate::domain::errors::DomainError;

/// Typed wrapper around a UUID. Prevents mixing up different ID types.
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct OrderId(Uuid);

impl OrderId {
    /// Generate a new random OrderId.
    pub fn generate() -> Self {
        Self(Uuid::new_v4())
    }

    /// Wrap an existing UUID (used in reconstitute).
    pub fn from_uuid(uuid: Uuid) -> Self {
        Self(uuid)
    }

    /// Parse a UUID string into an OrderId.
    pub fn parse(s: &str) -> Result<Self, DomainError> {
        Uuid::parse_str(s)
            .map(Self)
            .map_err(|_| DomainError::InvalidId(s.to_string()))
    }

    pub fn as_uuid(&self) -> &Uuid {
        &self.0
    }
}

impl fmt::Display for OrderId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}
```

### `src/domain/value_objects/user_id.rs`

```rust
use std::fmt;
use uuid::Uuid;
use crate::domain::errors::DomainError;

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct UserId(Uuid);

impl UserId {
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

    pub fn as_uuid(&self) -> &Uuid {
        &self.0
    }
}

impl fmt::Display for UserId {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}
```

### `src/domain/value_objects/money.rs`

```rust
use std::fmt;
use rust_decimal::Decimal;
use crate::domain::errors::DomainError;

/// Monetary value — always carries currency. Immutable; all operations return new values.
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct Money {
    amount: Decimal,
    currency: String,
}

impl Money {
    /// Create Money after validating amount and currency.
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

    /// Zero-value money for the given currency — infallible.
    pub fn zero(currency: impl Into<String>) -> Self {
        Self {
            amount: Decimal::ZERO,
            currency: currency.into(),
        }
    }

    pub fn amount(&self) -> Decimal {
        self.amount
    }

    pub fn currency(&self) -> &str {
        &self.currency
    }

    /// Add two Money values — returns Err if currencies differ.
    pub fn add(&self, other: &Money) -> Result<Money, DomainError> {
        if self.currency != other.currency {
            return Err(DomainError::CurrencyMismatch {
                left: self.currency.clone(),
                right: other.currency.clone(),
            });
        }
        Ok(Self {
            amount: self.amount + other.amount,
            currency: self.currency.clone(),
        })
    }

    /// Multiply by an integer factor — currency is preserved.
    pub fn multiply(&self, factor: u32) -> Self {
        Self {
            amount: self.amount * Decimal::from(factor),
            currency: self.currency.clone(),
        }
    }
}

impl fmt::Display for Money {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{:.2} {}", self.amount, self.currency)
    }
}
```

---

## Domain Events (`src/domain/events.rs`)

```rust
use uuid::Uuid;
use chrono::{DateTime, Utc};
use crate::domain::value_objects::{OrderId, UserId, Money};

/// All domain events in one enum. Use cases drain these from entities after saving.
#[derive(Debug, Clone)]
pub enum DomainEvent {
    OrderPlaced(OrderPlaced),
    OrderConfirmed(OrderConfirmed),
}

impl DomainEvent {
    pub fn event_id(&self) -> &Uuid {
        match self {
            DomainEvent::OrderPlaced(e) => &e.event_id,
            DomainEvent::OrderConfirmed(e) => &e.event_id,
        }
    }

    pub fn event_type(&self) -> &'static str {
        match self {
            DomainEvent::OrderPlaced(_) => "order.placed",
            DomainEvent::OrderConfirmed(_) => "order.confirmed",
        }
    }

    pub fn occurred_at(&self) -> &DateTime<Utc> {
        match self {
            DomainEvent::OrderPlaced(e) => &e.occurred_at,
            DomainEvent::OrderConfirmed(e) => &e.occurred_at,
        }
    }
}

/// Emitted when a new order is successfully created.
#[derive(Debug, Clone)]
pub struct OrderPlaced {
    pub event_id: Uuid,
    pub occurred_at: DateTime<Utc>,
    pub order_id: OrderId,
    pub customer_id: UserId,
    pub total: Money,
}

impl OrderPlaced {
    pub fn new(order_id: OrderId, customer_id: UserId, total: Money) -> Self {
        Self {
            event_id: Uuid::new_v4(),
            occurred_at: Utc::now(),
            order_id,
            customer_id,
            total,
        }
    }
}

/// Emitted when a pending order transitions to CONFIRMED.
#[derive(Debug, Clone)]
pub struct OrderConfirmed {
    pub event_id: Uuid,
    pub occurred_at: DateTime<Utc>,
    pub order_id: OrderId,
}

impl OrderConfirmed {
    pub fn new(order_id: OrderId) -> Self {
        Self {
            event_id: Uuid::new_v4(),
            occurred_at: Utc::now(),
            order_id,
        }
    }
}
```

---

## Entity (`src/domain/entities/order.rs`)

```rust
use crate::domain::errors::DomainError;
use crate::domain::events::{DomainEvent, OrderConfirmed, OrderPlaced};
use crate::domain::value_objects::{Money, OrderId, UserId};

/// Lifecycle states of an order.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum OrderStatus {
    Pending,
    Confirmed,
    Cancelled,
}

impl std::fmt::Display for OrderStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            OrderStatus::Pending => write!(f, "PENDING"),
            OrderStatus::Confirmed => write!(f, "CONFIRMED"),
            OrderStatus::Cancelled => write!(f, "CANCELLED"),
        }
    }
}

/// A single line item within an order.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct OrderItem {
    pub product_id: String,
    pub quantity: u32,
    pub unit_price: Money,
}

impl OrderItem {
    pub fn subtotal(&self) -> Money {
        self.unit_price.multiply(self.quantity)
    }
}

/// The Order aggregate root.
///
/// All fields are private — accessed only via pub getters.
/// Use `Order::new()` for creation and `Order::reconstitute()` for DB hydration.
pub struct Order {
    id: OrderId,
    customer_id: UserId,
    items: Vec<OrderItem>,
    total: Money,
    status: OrderStatus,
    events: Vec<DomainEvent>,
}

impl Order {
    // --- Constructors ---

    /// Create a new order — enforces invariants and emits `OrderPlaced`.
    pub fn new(customer_id: UserId, items: Vec<OrderItem>) -> Result<Self, DomainError> {
        if items.is_empty() {
            return Err(DomainError::InvalidOrderItems);
        }

        let total = items
            .iter()
            .try_fold(Money::zero("USD"), |acc, item| acc.add(&item.subtotal()))?;

        let id = OrderId::generate();
        let event = DomainEvent::OrderPlaced(OrderPlaced::new(
            id.clone(),
            customer_id.clone(),
            total.clone(),
        ));

        Ok(Self {
            id,
            customer_id,
            items,
            total,
            status: OrderStatus::Pending,
            events: vec![event],
        })
    }

    /// Hydrate an Order from storage — skips invariant checks, emits no events.
    ///
    /// Called only in outbound adapters when mapping DB rows to domain types.
    pub fn reconstitute(
        id: OrderId,
        customer_id: UserId,
        items: Vec<OrderItem>,
        total: Money,
        status: OrderStatus,
    ) -> Self {
        Self {
            id,
            customer_id,
            items,
            total,
            status,
            events: vec![],
        }
    }

    // --- Getters ---

    pub fn id(&self) -> &OrderId {
        &self.id
    }

    pub fn customer_id(&self) -> &UserId {
        &self.customer_id
    }

    pub fn items(&self) -> &[OrderItem] {
        &self.items
    }

    pub fn total(&self) -> &Money {
        &self.total
    }

    pub fn status(&self) -> &OrderStatus {
        &self.status
    }

    // --- State transitions ---

    /// Transition a PENDING order to CONFIRMED.
    pub fn confirm(&mut self) -> Result<(), DomainError> {
        if self.status != OrderStatus::Pending {
            return Err(DomainError::InvalidOrderState {
                current: self.status.to_string(),
                attempted: "CONFIRMED".into(),
            });
        }
        self.status = OrderStatus::Confirmed;
        self.events
            .push(DomainEvent::OrderConfirmed(OrderConfirmed::new(self.id.clone())));
        Ok(())
    }

    /// Cancel an order — not allowed if already cancelled.
    pub fn cancel(&mut self, _reason: &str) -> Result<(), DomainError> {
        if self.status == OrderStatus::Cancelled {
            return Err(DomainError::InvalidOrderState {
                current: "CANCELLED".into(),
                attempted: "CANCELLED".into(),
            });
        }
        self.status = OrderStatus::Cancelled;
        Ok(())
    }

    // --- Event collection ---

    /// Return all pending domain events and clear the internal list.
    /// Call this after persisting the entity.
    pub fn drain_events(&mut self) -> Vec<DomainEvent> {
        std::mem::take(&mut self.events)
    }
}

/// Entities are equal when their IDs match.
impl PartialEq for Order {
    fn eq(&self, other: &Self) -> bool {
        self.id == other.id
    }
}
```

---

## Module wiring (`src/domain/mod.rs`)

```rust
pub mod entities;
pub mod errors;
pub mod events;
pub mod value_objects;

// Re-export the most commonly referenced types for convenience.
pub use entities::order::{Order, OrderItem, OrderStatus};
pub use errors::DomainError;
pub use events::{DomainEvent, OrderConfirmed, OrderPlaced};
pub use value_objects::{money::Money, order_id::OrderId, user_id::UserId};
```

---

## Key rules illustrated here

- `DomainError` is an `enum` derived with `thiserror::Error` — all variants in one file
- Value objects have private fields — all access via pub getters or `Display` impl
- `rust_decimal::Decimal` is used for monetary values — never `f64` or `f32`
- All value object constructors return `Result<Self, DomainError>` — never `unwrap()` or `panic!()`
- `Order::new()` enforces invariants and emits `OrderPlaced`; `Order::reconstitute()` does neither
- `drain_events()` uses `std::mem::take` — clean and idiomatic, avoids the borrow checker fighting a `split_at_mut`
- `impl PartialEq for Order` compares by ID only — not by all fields
- No import of any HTTP, DB, or framework crate anywhere in this module
