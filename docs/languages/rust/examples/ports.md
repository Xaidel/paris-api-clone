# Rust — Ports Layer Examples

Real Rust code illustrating `src/ports/` patterns using traits.

---

## Inbound Ports (`src/ports/inbound/`)

### `src/ports/inbound/place_order_port.rs`

```rust
use async_trait::async_trait;
use crate::domain::errors::DomainError;

/// A single item in the place-order command.
/// Uses primitive types — no domain types leak through the inbound port.
#[derive(Debug, Clone)]
pub struct PlaceOrderItemInput {
    pub product_id: String,
    pub quantity: u32,
    /// Decimal string preserves precision: "9.99", not 9.99f64
    pub unit_price_amount: String,
    pub unit_price_currency: String,
}

/// Command sent by the inbound adapter to place a new order.
#[derive(Debug, Clone)]
pub struct PlaceOrderCommand {
    pub customer_id: String,
    pub items: Vec<PlaceOrderItemInput>,
}

/// Result returned to the inbound adapter on success.
#[derive(Debug, Clone)]
pub struct PlaceOrderResult {
    pub order_id: String,
    pub status: String,
    pub total_amount: String,
    pub total_currency: String,
}

/// Inbound port for placing a new order.
/// The use case implements this trait; the Axum handler calls it via `Arc<dyn PlaceOrderPort>`.
#[async_trait]
pub trait PlaceOrderPort: Send + Sync {
    async fn execute(&self, cmd: PlaceOrderCommand) -> Result<PlaceOrderResult, DomainError>;
}
```

### `src/ports/inbound/get_order_port.rs`

```rust
use async_trait::async_trait;
use crate::domain::errors::DomainError;

/// Query identifying which order to retrieve and who is asking.
#[derive(Debug, Clone)]
pub struct GetOrderQuery {
    pub order_id: String,
    pub requesting_user_id: String,
}

/// Read-model DTO for a single order line.
#[derive(Debug, Clone)]
pub struct OrderItemView {
    pub product_id: String,
    pub quantity: u32,
    pub unit_price_amount: String,
    pub unit_price_currency: String,
    pub subtotal_amount: String,
}

/// Read-model returned by GetOrderPort on success.
#[derive(Debug, Clone)]
pub struct GetOrderResult {
    pub order_id: String,
    pub customer_id: String,
    pub status: String,
    pub total_amount: String,
    pub total_currency: String,
    pub items: Vec<OrderItemView>,
}

/// Inbound port for retrieving an existing order.
#[async_trait]
pub trait GetOrderPort: Send + Sync {
    async fn execute(&self, query: GetOrderQuery) -> Result<GetOrderResult, DomainError>;
}
```

---

## Outbound Ports (`src/ports/outbound/`)

### `src/ports/outbound/order_repository.rs`

```rust
use async_trait::async_trait;
use crate::domain::entities::order::Order;
use crate::domain::value_objects::{OrderId, UserId};

/// Error type for repository operations — distinct from DomainError.
#[derive(Debug, thiserror::Error)]
pub enum RepositoryError {
    #[error("database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("serialization error: {0}")]
    Serialization(String),
}

/// Outbound port for persisting and querying orders.
/// The application layer depends on this trait; the Postgres adapter implements it.
#[async_trait]
pub trait OrderRepository: Send + Sync {
    async fn save(&self, order: &Order) -> Result<(), RepositoryError>;
    async fn find_by_id(&self, id: &OrderId) -> Result<Option<Order>, RepositoryError>;
    async fn find_by_customer_id(&self, customer_id: &UserId) -> Result<Vec<Order>, RepositoryError>;
}
```

### `src/ports/outbound/payment_gateway.rs`

```rust
use async_trait::async_trait;

/// Data required to charge a customer.
#[derive(Debug, Clone)]
pub struct PaymentRequest {
    pub order_id: String,
    pub amount: String,
    pub currency: String,
    pub payment_method_token: String,
}

/// Outcome of a charge or refund attempt.
#[derive(Debug, Clone)]
pub struct PaymentResult {
    pub success: bool,
    /// Empty string when not successful.
    pub transaction_id: String,
    /// Empty string on success.
    pub failure_reason: String,
}

/// Error type for gateway infrastructure failures.
#[derive(Debug, thiserror::Error)]
pub enum GatewayError {
    #[error("upstream payment provider error: {0}")]
    Upstream(String),

    #[error("network error: {0}")]
    Network(String),
}

/// Outbound port for processing payments.
#[async_trait]
pub trait PaymentGateway: Send + Sync {
    async fn charge(&self, req: PaymentRequest) -> Result<PaymentResult, GatewayError>;
    async fn refund(&self, transaction_id: &str, amount: &str) -> Result<PaymentResult, GatewayError>;
}
```

### `src/ports/outbound/event_publisher.rs`

```rust
use async_trait::async_trait;
use crate::domain::events::DomainEvent;

/// Error type for event publishing failures.
#[derive(Debug, thiserror::Error)]
pub enum PublisherError {
    #[error("publish failed: {0}")]
    Failed(String),
}

/// Outbound port for publishing domain events after an aggregate is persisted.
#[async_trait]
pub trait EventPublisher: Send + Sync {
    async fn publish(&self, event: DomainEvent) -> Result<(), PublisherError>;
    async fn publish_all(&self, events: Vec<DomainEvent>) -> Result<(), PublisherError>;
}
```

### `src/ports/outbound/notification_gateway.rs`

```rust
use async_trait::async_trait;

/// Message payload for a user notification.
#[derive(Debug, Clone)]
pub struct NotificationMessage {
    pub recipient_id: String,
    pub subject: String,
    pub body: String,
}

/// Error type for notification delivery failures.
#[derive(Debug, thiserror::Error)]
pub enum NotificationError {
    #[error("delivery failed: {0}")]
    Delivery(String),
}

/// Outbound port for sending notifications.
#[async_trait]
pub trait NotificationGateway: Send + Sync {
    async fn send(&self, msg: NotificationMessage) -> Result<(), NotificationError>;
}
```

---

## Module wiring (`src/ports/mod.rs`)

```rust
pub mod inbound;
pub mod outbound;

// Re-export the most commonly used port traits and DTOs.
pub use inbound::get_order_port::{GetOrderPort, GetOrderQuery, GetOrderResult, OrderItemView};
pub use inbound::place_order_port::{PlaceOrderCommand, PlaceOrderItemInput, PlaceOrderPort, PlaceOrderResult};
pub use outbound::event_publisher::{EventPublisher, PublisherError};
pub use outbound::notification_gateway::{NotificationGateway, NotificationError, NotificationMessage};
pub use outbound::order_repository::{OrderRepository, RepositoryError};
pub use outbound::payment_gateway::{GatewayError, PaymentGateway, PaymentRequest, PaymentResult};
```

---

## Key rules illustrated here

- All port traits are `Send + Sync` — required for use with `Arc<dyn Trait>` across async task boundaries
- `#[async_trait]` is applied to both the trait definition and all `impl` blocks
- Inbound port commands/results use primitive types (`String`, `u32`) — no domain types leak out
- Outbound port methods use domain types (`&Order`, `&OrderId`) — called from the application layer
- Each outbound port has its own error type (`RepositoryError`, `GatewayError`, etc.) — not `DomainError`
- No implementation code in any port file — interfaces/contracts only
- No Axum, sqlx, or other framework types in port signatures
