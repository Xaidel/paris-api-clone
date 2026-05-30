# Rust — Test Examples

Real Rust code illustrating test patterns with `cargo test`, `tokio::test`, and `testcontainers`.

---

## In-Memory Port Implementations (`tests/utils/in_memory_order_repository.rs`)

```rust
use async_trait::async_trait;
use std::collections::HashMap;
use std::sync::Mutex;

use crate::domain::entities::order::Order;
use crate::domain::value_objects::{OrderId, UserId};
use crate::ports::outbound::order_repository::{OrderRepository, RepositoryError};

/// In-memory implementation of OrderRepository.
/// Used in unit tests — implements the same trait the real adapter does.
/// No mocking library required.
pub struct InMemoryOrderRepository {
    store: Mutex<HashMap<String, Order>>,
}

impl InMemoryOrderRepository {
    pub fn new() -> Self {
        Self {
            store: Mutex::new(HashMap::new()),
        }
    }

    /// Test helper — not part of the port contract.
    pub fn all_orders(&self) -> Vec<String> {
        self.store.lock().unwrap().keys().cloned().collect()
    }

    pub fn count(&self) -> usize {
        self.store.lock().unwrap().len()
    }
}

#[async_trait]
impl OrderRepository for InMemoryOrderRepository {
    async fn save(&self, order: &Order) -> Result<(), RepositoryError> {
        // Order must be Clone to store it — derive Clone on Order for tests.
        self.store
            .lock()
            .unwrap()
            .insert(order.id().to_string(), order.clone());
        Ok(())
    }

    async fn find_by_id(&self, id: &OrderId) -> Result<Option<Order>, RepositoryError> {
        Ok(self.store.lock().unwrap().get(&id.to_string()).cloned())
    }

    async fn find_by_customer_id(&self, customer_id: &UserId) -> Result<Vec<Order>, RepositoryError> {
        let store = self.store.lock().unwrap();
        Ok(store
            .values()
            .filter(|o| o.customer_id() == customer_id)
            .cloned()
            .collect())
    }
}
```

```rust
// tests/utils/in_memory_event_publisher.rs
use async_trait::async_trait;
use std::sync::Mutex;

use crate::domain::events::DomainEvent;
use crate::ports::outbound::event_publisher::{EventPublisher, PublisherError};

/// In-memory event publisher — stores events so tests can assert on them.
pub struct InMemoryEventPublisher {
    pub published: Mutex<Vec<DomainEvent>>,
}

impl InMemoryEventPublisher {
    pub fn new() -> Self {
        Self {
            published: Mutex::new(Vec::new()),
        }
    }

    pub fn events(&self) -> Vec<DomainEvent> {
        self.published.lock().unwrap().clone()
    }
}

#[async_trait]
impl EventPublisher for InMemoryEventPublisher {
    async fn publish(&self, event: DomainEvent) -> Result<(), PublisherError> {
        self.published.lock().unwrap().push(event);
        Ok(())
    }

    async fn publish_all(&self, events: Vec<DomainEvent>) -> Result<(), PublisherError> {
        self.published.lock().unwrap().extend(events);
        Ok(())
    }
}
```

---

## Unit Tests — Domain Entity (`tests/unit/test_order.rs`)

```rust
#[cfg(test)]
mod tests {
    use rust_decimal_macros::dec;

    use crate::domain::entities::order::{Order, OrderItem, OrderStatus};
    use crate::domain::errors::DomainError;
    use crate::domain::events::DomainEvent;
    use crate::domain::value_objects::{Money, UserId};

    fn make_item(price: &str) -> OrderItem {
        OrderItem {
            product_id: "prod-1".to_string(),
            quantity: 2,
            unit_price: Money::new(price.parse().unwrap(), "USD").unwrap(),
        }
    }

    fn make_order() -> Order {
        Order::new(UserId::generate(), vec![make_item("9.99")]).unwrap()
    }

    // --- Creation ---

    #[test]
    fn emits_order_placed_event_on_creation() {
        let mut order = make_order();
        let events = order.drain_events();
        assert_eq!(events.len(), 1);
        assert!(matches!(events[0], DomainEvent::OrderPlaced(_)));
    }

    #[test]
    fn status_is_pending_after_creation() {
        let order = make_order();
        assert_eq!(order.status(), &OrderStatus::Pending);
    }

    #[test]
    fn returns_error_when_items_is_empty() {
        let result = Order::new(UserId::generate(), vec![]);
        assert!(matches!(result, Err(DomainError::InvalidOrderItems)));
    }

    #[test]
    fn calculates_total_correctly() {
        let items = vec![
            OrderItem {
                product_id: "prod-1".to_string(),
                quantity: 2,
                unit_price: Money::new(dec!(10.00), "USD").unwrap(),
            },
            OrderItem {
                product_id: "prod-2".to_string(),
                quantity: 1,
                unit_price: Money::new(dec!(5.00), "USD").unwrap(),
            },
        ];
        let order = Order::new(UserId::generate(), items).unwrap();
        assert_eq!(order.total().amount(), dec!(25.00));
    }

    // --- State transitions ---

    #[test]
    fn confirms_a_pending_order() {
        let mut order = make_order();
        order.drain_events(); // clear OrderPlaced
        order.confirm().unwrap();
        let events = order.drain_events();
        assert_eq!(order.status(), &OrderStatus::Confirmed);
        assert_eq!(events.len(), 1);
        assert!(matches!(events[0], DomainEvent::OrderConfirmed(_)));
    }

    #[test]
    fn returns_error_when_confirming_cancelled_order() {
        let mut order = make_order();
        order.cancel("test").unwrap();
        let result = order.confirm();
        assert!(matches!(result, Err(DomainError::InvalidOrderState { .. })));
    }

    #[test]
    fn drain_events_clears_the_list() {
        let mut order = make_order();
        let first = order.drain_events();
        let second = order.drain_events();
        assert_eq!(first.len(), 1);
        assert_eq!(second.len(), 0);
    }

    // --- Reconstitute ---

    #[test]
    fn reconstitute_emits_no_events() {
        let original = make_order();
        let mut reconstituted = Order::reconstitute(
            original.id().clone(),
            original.customer_id().clone(),
            original.items().to_vec(),
            original.total().clone(),
            original.status().clone(),
        );
        assert_eq!(reconstituted.drain_events().len(), 0);
    }

    #[test]
    fn reconstitute_preserves_all_fields() {
        let original = make_order();
        let reconstituted = Order::reconstitute(
            original.id().clone(),
            original.customer_id().clone(),
            original.items().to_vec(),
            original.total().clone(),
            OrderStatus::Confirmed,
        );
        assert_eq!(reconstituted.id(), original.id());
        assert_eq!(reconstituted.status(), &OrderStatus::Confirmed);
    }
}
```

---

## Unit Tests — Use Case (`tests/unit/test_place_order_use_case.rs`)

```rust
#[cfg(test)]
mod tests {
    use std::sync::Arc;
    use rust_decimal_macros::dec;

    use crate::application::use_cases::place_order::PlaceOrderUseCase;
    use crate::domain::errors::DomainError;
    use crate::domain::events::DomainEvent;
    use crate::ports::inbound::place_order_port::{PlaceOrderCommand, PlaceOrderItemInput, PlaceOrderPort};
    use super::super::utils::{InMemoryEventPublisher, InMemoryOrderRepository};

    fn valid_command() -> PlaceOrderCommand {
        PlaceOrderCommand {
            customer_id: "00000000-0000-0000-0000-000000000001".to_string(),
            items: vec![PlaceOrderItemInput {
                product_id: "prod-1".to_string(),
                quantity: 2,
                unit_price_amount: "9.99".to_string(),
                unit_price_currency: "USD".to_string(),
            }],
        }
    }

    fn make_use_case() -> (PlaceOrderUseCase, Arc<InMemoryOrderRepository>, Arc<InMemoryEventPublisher>) {
        let repo = Arc::new(InMemoryOrderRepository::new());
        let publisher = Arc::new(InMemoryEventPublisher::new());
        let uc = PlaceOrderUseCase::new(
            Arc::clone(&repo) as Arc<dyn crate::ports::outbound::OrderRepository>,
            Arc::clone(&publisher) as Arc<dyn crate::ports::outbound::EventPublisher>,
        );
        (uc, repo, publisher)
    }

    #[tokio::test]
    async fn places_order_and_returns_result() {
        let (uc, repo, _) = make_use_case();

        let result = uc.execute(valid_command()).await.unwrap();

        assert!(!result.order_id.is_empty());
        assert_eq!(result.status, "PENDING");
        assert_eq!(result.total_amount, "19.98");
        assert_eq!(repo.count(), 1);
    }

    #[tokio::test]
    async fn publishes_order_placed_event() {
        let (uc, _, publisher) = make_use_case();

        uc.execute(valid_command()).await.unwrap();

        let events = publisher.events();
        assert_eq!(events.len(), 1);
        assert!(matches!(events[0], DomainEvent::OrderPlaced(_)));
    }

    #[tokio::test]
    async fn returns_error_when_items_is_empty() {
        let (uc, _, _) = make_use_case();
        let cmd = PlaceOrderCommand {
            customer_id: "00000000-0000-0000-0000-000000000001".to_string(),
            items: vec![],
        };

        let result = uc.execute(cmd).await;
        assert!(matches!(result, Err(DomainError::InvalidOrderItems)));
    }

    #[tokio::test]
    async fn returns_error_for_invalid_customer_id() {
        let (uc, _, _) = make_use_case();
        let mut cmd = valid_command();
        cmd.customer_id = "not-a-uuid".to_string();

        let result = uc.execute(cmd).await;
        assert!(matches!(result, Err(DomainError::InvalidId(_))));
    }
}
```

---

## Integration Test — Repository (`tests/integration/test_postgres_order_repository.rs`)

```rust
#[cfg(test)]
mod tests {
    use std::sync::Arc;
    use testcontainers::{core::WaitFor, runners::AsyncRunner, GenericImage};
    use sqlx::PgPool;

    use crate::adapters::outbound::postgres_order_repository::PostgresOrderRepository;
    use crate::domain::entities::order::{Order, OrderItem, OrderStatus};
    use crate::domain::value_objects::{Money, OrderId, UserId};
    use crate::ports::outbound::OrderRepository;

    const MIGRATIONS: &str = r#"
        CREATE TABLE IF NOT EXISTS orders (
            id              UUID        PRIMARY KEY,
            customer_id     UUID        NOT NULL,
            status          TEXT        NOT NULL,
            total_amount    NUMERIC     NOT NULL,
            total_currency  TEXT        NOT NULL
        );
        CREATE TABLE IF NOT EXISTS order_items (
            id                    BIGSERIAL   PRIMARY KEY,
            order_id              UUID        NOT NULL REFERENCES orders(id),
            product_id            TEXT        NOT NULL,
            quantity              INTEGER     NOT NULL,
            unit_price_amount     NUMERIC     NOT NULL,
            unit_price_currency   TEXT        NOT NULL
        );
    "#;

    async fn setup_pool() -> PgPool {
        let pg = GenericImage::new("postgres", "16")
            .with_env_var("POSTGRES_PASSWORD", "test")
            .with_env_var("POSTGRES_DB", "test")
            .with_wait_for(WaitFor::message_on_stderr("database system is ready"))
            .start()
            .await
            .expect("Failed to start Postgres container");

        let port = pg.get_host_port_ipv4(5432).await.unwrap();
        let url = format!("postgres://postgres:test@127.0.0.1:{port}/test");

        let pool = PgPool::connect(&url).await.expect("Failed to connect to test DB");
        sqlx::query(MIGRATIONS).execute(&pool).await.unwrap();
        pool
    }

    fn make_order() -> Order {
        Order::new(
            UserId::generate(),
            vec![OrderItem {
                product_id: "prod-1".to_string(),
                quantity: 2,
                unit_price: Money::new("10.00".parse().unwrap(), "USD").unwrap(),
            }],
        )
        .unwrap()
    }

    #[tokio::test]
    async fn saves_and_retrieves_order_by_id() {
        let pool = setup_pool().await;
        let repo = PostgresOrderRepository::new(pool);

        let order = make_order();
        repo.save(&order).await.unwrap();

        let found = repo.find_by_id(order.id()).await.unwrap();
        assert!(found.is_some());
        let found = found.unwrap();
        assert_eq!(found.id(), order.id());
        assert_eq!(found.status(), &OrderStatus::Pending);
        assert_eq!(found.items().len(), 1);
    }

    #[tokio::test]
    async fn returns_none_for_unknown_id() {
        let pool = setup_pool().await;
        let repo = PostgresOrderRepository::new(pool);

        let result = repo.find_by_id(&OrderId::generate()).await.unwrap();
        assert!(result.is_none());
    }

    #[tokio::test]
    async fn updates_status_on_re_save() {
        let pool = setup_pool().await;
        let repo = PostgresOrderRepository::new(pool);

        let mut order = make_order();
        repo.save(&order).await.unwrap();
        order.confirm().unwrap();
        repo.save(&order).await.unwrap();

        let found = repo.find_by_id(order.id()).await.unwrap().unwrap();
        assert_eq!(found.status(), &OrderStatus::Confirmed);
    }

    #[tokio::test]
    async fn finds_orders_by_customer_id() {
        let pool = setup_pool().await;
        let repo = PostgresOrderRepository::new(pool);

        let customer = UserId::generate();
        let other = UserId::generate();

        let order1 = Order::new(customer.clone(), vec![OrderItem {
            product_id: "p1".into(), quantity: 1, unit_price: Money::new("5.00".parse().unwrap(), "USD").unwrap(),
        }]).unwrap();
        let order2 = Order::new(customer.clone(), vec![OrderItem {
            product_id: "p2".into(), quantity: 1, unit_price: Money::new("5.00".parse().unwrap(), "USD").unwrap(),
        }]).unwrap();
        let order3 = Order::new(other.clone(), vec![OrderItem {
            product_id: "p3".into(), quantity: 1, unit_price: Money::new("5.00".parse().unwrap(), "USD").unwrap(),
        }]).unwrap();

        for o in [&order1, &order2, &order3] {
            repo.save(o).await.unwrap();
        }

        let results = repo.find_by_customer_id(&customer).await.unwrap();
        assert_eq!(results.len(), 2);
    }
}
```

---

## Integration Test — Inbound Adapter (`tests/integration/test_http_order_adapter.rs`)

```rust
#[cfg(test)]
mod tests {
    use std::sync::Arc;
    use axum::{body::Body, http::{Request, StatusCode}};
    use serde_json::{json, Value};
    use tower::ServiceExt;

    use crate::adapters::inbound::http_order_adapter::{order_router, OrderState};
    use crate::application::use_cases::{get_order::GetOrderUseCase, place_order::PlaceOrderUseCase};
    use super::super::utils::{InMemoryEventPublisher, InMemoryOrderRepository};

    fn make_app() -> axum::Router {
        let repo = Arc::new(InMemoryOrderRepository::new());
        let publisher = Arc::new(InMemoryEventPublisher::new());

        let place_order = Arc::new(PlaceOrderUseCase::new(
            Arc::clone(&repo) as Arc<dyn crate::ports::outbound::OrderRepository>,
            Arc::clone(&publisher) as Arc<dyn crate::ports::outbound::EventPublisher>,
        ));
        let get_order = Arc::new(GetOrderUseCase::new(
            Arc::clone(&repo) as Arc<dyn crate::ports::outbound::OrderRepository>,
        ));

        order_router(OrderState {
            place_order: place_order as Arc<dyn crate::ports::inbound::PlaceOrderPort>,
            get_order: get_order as Arc<dyn crate::ports::inbound::GetOrderPort>,
        })
    }

    #[tokio::test]
    async fn place_order_returns_201() {
        let app = make_app();
        let body = json!({
            "customer_id": "00000000-0000-0000-0000-000000000001",
            "items": [{"product_id": "prod-1", "quantity": 2, "unit_price_amount": "9.99", "unit_price_currency": "USD"}]
        });

        let response = app
            .oneshot(
                Request::builder()
                    .method("POST")
                    .uri("/orders")
                    .header("content-type", "application/json")
                    .body(Body::from(body.to_string()))
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::CREATED);
        let bytes = axum::body::to_bytes(response.into_body(), usize::MAX).await.unwrap();
        let json: Value = serde_json::from_slice(&bytes).unwrap();
        assert!(json["order_id"].as_str().is_some());
        assert_eq!(json["status"], "PENDING");
    }

    #[tokio::test]
    async fn place_order_returns_422_for_empty_items() {
        let app = make_app();
        let body = json!({
            "customer_id": "00000000-0000-0000-0000-000000000001",
            "items": []
        });

        let response = app
            .oneshot(
                Request::builder()
                    .method("POST")
                    .uri("/orders")
                    .header("content-type", "application/json")
                    .body(Body::from(body.to_string()))
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::UNPROCESSABLE_ENTITY);
    }

    #[tokio::test]
    async fn get_order_returns_404_for_unknown_id() {
        let app = make_app();

        let response = app
            .oneshot(
                Request::builder()
                    .method("GET")
                    .uri("/orders/00000000-0000-0000-0000-000000000099?requesting_user_id=00000000-0000-0000-0000-000000000001")
                    .body(Body::empty())
                    .unwrap(),
            )
            .await
            .unwrap();

        assert_eq!(response.status(), StatusCode::NOT_FOUND);
    }
}
```

---

## Cargo.toml test dependencies

```toml
[dev-dependencies]
# Async test runtime
tokio = { version = "1", features = ["full", "test-util"] }

# Integration DB tests
testcontainers = { version = "0.21", features = ["async"] }

# HTTP integration tests (Axum tower::ServiceExt)
tower = { version = "0.4", features = ["util"] }

# Decimal literals in tests
rust_decimal_macros = "1"

# JSON assertion
serde_json = "1"
```

---

## Key rules illustrated here

- In-memory repositories implement the same trait as the real adapter — no mocking library needed
- `#[tokio::test]` is required on every `async fn` test — the macro sets up the Tokio runtime
- Unit tests are in `#[cfg(test)]` modules within the source tree — standard Rust convention
- Integration tests use `testcontainers` for a real Postgres instance — no mocking of SQL
- Inbound adapter integration tests use `tower::ServiceExt::oneshot` — a real HTTP dispatch against a real (but in-memory-backed) Axum router
- `Mutex<HashMap>` in `InMemoryOrderRepository` makes it `Send + Sync` — required for `Arc<dyn OrderRepository>`
- `Order` must derive or impl `Clone` to be stored and retrieved from the in-memory store
- `std::mem::take` in `drain_events()` works with `#[cfg(test)]` as-is — no test-specific changes needed
