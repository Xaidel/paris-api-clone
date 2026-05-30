# Rust — Application Layer Examples

Real Rust code illustrating `src/application/` patterns with use cases and application services.

---

## Application Error (`src/application/errors.rs`)

```rust
use thiserror::Error;
use crate::domain::errors::DomainError;
use crate::ports::outbound::{GatewayError, PublisherError, RepositoryError};

/// Application-layer error — wraps domain and infrastructure errors.
/// Adapters match on variants to determine HTTP status codes.
#[derive(Debug, Error)]
pub enum ApplicationError {
    #[error(transparent)]
    Domain(#[from] DomainError),

    #[error("repository error: {0}")]
    Repository(#[from] RepositoryError),

    #[error("payment gateway error: {0}")]
    Gateway(#[from] GatewayError),

    #[error("event publisher error: {0}")]
    Publisher(#[from] PublisherError),

    #[error("invalid input: {0}")]
    InvalidInput(String),
}
```

---

## Use Case: Place Order (`src/application/use_cases/place_order.rs`)

```rust
use std::sync::Arc;
use async_trait::async_trait;
use rust_decimal::Decimal;
use std::str::FromStr;

use crate::application::errors::ApplicationError;
use crate::domain::entities::order::{Order, OrderItem};
use crate::domain::value_objects::{Money, UserId};
use crate::ports::inbound::place_order_port::{
    PlaceOrderCommand, PlaceOrderPort, PlaceOrderResult,
};
use crate::ports::outbound::{EventPublisher, OrderRepository};

/// Orchestrates order placement.
/// Implements `PlaceOrderPort` — verified by the compiler at the DI wiring site.
pub struct PlaceOrderUseCase {
    order_repo: Arc<dyn OrderRepository>,
    event_publisher: Arc<dyn EventPublisher>,
}

impl PlaceOrderUseCase {
    pub fn new(
        order_repo: Arc<dyn OrderRepository>,
        event_publisher: Arc<dyn EventPublisher>,
    ) -> Self {
        Self {
            order_repo,
            event_publisher,
        }
    }
}

#[async_trait]
impl PlaceOrderPort for PlaceOrderUseCase {
    async fn execute(&self, cmd: PlaceOrderCommand) -> Result<PlaceOrderResult, crate::domain::errors::DomainError> {
        // 1. Parse customer ID
        let customer_id = UserId::parse(&cmd.customer_id)?;

        // 2. Map item DTOs → domain OrderItem values
        let mut items = Vec::with_capacity(cmd.items.len());
        for input in cmd.items {
            let amount = Decimal::from_str(&input.unit_price_amount)
                .map_err(|_| crate::domain::errors::DomainError::InvalidMoney(
                    format!("cannot parse amount: {}", input.unit_price_amount),
                ))?;
            let unit_price = Money::new(amount, input.unit_price_currency)?;
            items.push(OrderItem {
                product_id: input.product_id,
                quantity: input.quantity,
                unit_price,
            });
        }

        // 3. Create the domain aggregate (invariants enforced inside Order::new)
        let mut order = Order::new(customer_id, items)?;

        // 4. Persist (port call — no knowledge of Postgres)
        self.order_repo
            .save(&order)
            .await
            .map_err(|e| crate::domain::errors::DomainError::InvalidMoney(e.to_string()))?;

        // 5. Publish events (port call — no knowledge of Kafka)
        let events = order.drain_events();
        self.event_publisher
            .publish_all(events)
            .await
            .map_err(|e| crate::domain::errors::DomainError::InvalidMoney(e.to_string()))?;

        // 6. Return primitive result DTO
        Ok(PlaceOrderResult {
            order_id: order.id().to_string(),
            status: order.status().to_string(),
            total_amount: format!("{:.2}", order.total().amount()),
            total_currency: order.total().currency().to_string(),
        })
    }
}
```

> **Note:** The `map_err` calls above convert `RepositoryError` / `PublisherError` through
> `ApplicationError` when the use case is wired to return `ApplicationError` instead of
> `DomainError`. For use cases that return `ApplicationError`, define the method as
> `async fn execute(...) -> Result<PlaceOrderResult, ApplicationError>` and let `?` do the
> conversion via the `From` impls in `application/errors.rs`.

---

## Use Case: Get Order (`src/application/use_cases/get_order.rs`)

```rust
use std::sync::Arc;
use async_trait::async_trait;

use crate::domain::errors::DomainError;
use crate::domain::value_objects::OrderId;
use crate::ports::inbound::get_order_port::{
    GetOrderPort, GetOrderQuery, GetOrderResult, OrderItemView,
};
use crate::ports::outbound::OrderRepository;

/// Retrieves a single order by ID.
pub struct GetOrderUseCase {
    order_repo: Arc<dyn OrderRepository>,
}

impl GetOrderUseCase {
    pub fn new(order_repo: Arc<dyn OrderRepository>) -> Self {
        Self { order_repo }
    }
}

#[async_trait]
impl GetOrderPort for GetOrderUseCase {
    async fn execute(&self, query: GetOrderQuery) -> Result<GetOrderResult, DomainError> {
        // 1. Parse the order ID
        let order_id = OrderId::parse(&query.order_id)?;

        // 2. Load from repository
        let order = self
            .order_repo
            .find_by_id(&order_id)
            .await
            .map_err(|e| DomainError::OrderNotFound(e.to_string()))?
            .ok_or_else(|| DomainError::OrderNotFound(query.order_id.clone()))?;

        // 3. Map domain entity → result DTO
        let items = order
            .items()
            .iter()
            .map(|item| OrderItemView {
                product_id: item.product_id.clone(),
                quantity: item.quantity,
                unit_price_amount: format!("{:.2}", item.unit_price.amount()),
                unit_price_currency: item.unit_price.currency().to_string(),
                subtotal_amount: format!("{:.2}", item.subtotal().amount()),
            })
            .collect();

        Ok(GetOrderResult {
            order_id: order.id().to_string(),
            customer_id: order.customer_id().to_string(),
            status: order.status().to_string(),
            total_amount: format!("{:.2}", order.total().amount()),
            total_currency: order.total().currency().to_string(),
            items,
        })
    }
}
```

---

## Application Service (`src/application/services/notification_service.rs`)

```rust
use std::sync::Arc;

use crate::domain::errors::DomainError;
use crate::domain::value_objects::OrderId;
use crate::ports::outbound::{NotificationGateway, NotificationMessage, OrderRepository};

/// Orchestrates cross-cutting notification logic.
/// Not a use case — called from multiple use cases, not directly from an inbound adapter.
pub struct NotificationService {
    order_repo: Arc<dyn OrderRepository>,
    notification_gateway: Arc<dyn NotificationGateway>,
}

impl NotificationService {
    pub fn new(
        order_repo: Arc<dyn OrderRepository>,
        notification_gateway: Arc<dyn NotificationGateway>,
    ) -> Self {
        Self {
            order_repo,
            notification_gateway,
        }
    }

    /// Send a placement confirmation to the customer.
    pub async fn notify_order_placed(&self, order_id: &OrderId) -> Result<(), DomainError> {
        let order = self
            .order_repo
            .find_by_id(order_id)
            .await
            .map_err(|e| DomainError::OrderNotFound(e.to_string()))?
            .ok_or_else(|| DomainError::OrderNotFound(order_id.to_string()))?;

        self.notification_gateway
            .send(NotificationMessage {
                recipient_id: order.customer_id().to_string(),
                subject: "Your order has been placed".to_string(),
                body: format!(
                    "Order {} for {} is confirmed.",
                    order_id,
                    order.total()
                ),
            })
            .await
            .map_err(|e| DomainError::InvalidMoney(e.to_string()))?;

        Ok(())
    }
}
```

---

## Module wiring (`src/application/mod.rs`)

```rust
pub mod errors;
pub mod services;
pub mod use_cases;

pub use errors::ApplicationError;
pub use use_cases::get_order::GetOrderUseCase;
pub use use_cases::place_order::PlaceOrderUseCase;
```

---

## Key rules illustrated here

- Use cases receive `Arc<dyn PortTrait>` in their constructor — never concrete adapter types
- `PlaceOrderUseCase` implements `PlaceOrderPort` — verified by the compiler at the DI wiring site
- `execute()` takes command/query structs and returns result structs with primitives only
- Domain entity construction (`Order::new()`) happens inside the use case — not in the inbound adapter
- Events are drained from the entity and published after the entity is persisted
- Use cases do NOT match on `DomainError` variants — they let errors propagate to the adapter
- `ApplicationError` wraps all downstream error types via `#[from]` — `?` works transparently across error boundaries
- No Axum, sqlx, or other framework imports anywhere in this module
