# Rust — Adapters Layer Examples

Real Rust code illustrating `src/adapters/` patterns with Axum and sqlx.

---

## Inbound Adapter — Axum (`src/adapters/inbound/http_order_adapter.rs`)

```rust
use std::sync::Arc;
use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::str::FromStr;

use crate::domain::errors::DomainError;
use crate::ports::inbound::{
    get_order_port::{GetOrderPort, GetOrderQuery},
    place_order_port::{PlaceOrderCommand, PlaceOrderItemInput, PlaceOrderPort},
};

// --- Axum State ---

/// Shared handler state — cloned cheaply into every request via `Arc`.
#[derive(Clone)]
pub struct OrderState {
    pub place_order: Arc<dyn PlaceOrderPort>,
    pub get_order: Arc<dyn GetOrderPort>,
}

// --- Request / Response schemas (serde — adapter layer only) ---

#[derive(Debug, Deserialize)]
pub struct OrderItemRequest {
    pub product_id: String,
    pub quantity: u32,
    /// Decimal string — "9.99" not 9.99
    pub unit_price_amount: String,
    pub unit_price_currency: String,
}

#[derive(Debug, Deserialize)]
pub struct PlaceOrderRequest {
    pub customer_id: String,
    pub items: Vec<OrderItemRequest>,
}

#[derive(Debug, Serialize)]
pub struct PlaceOrderResponse {
    pub order_id: String,
    pub status: String,
    pub total_amount: String,
    pub total_currency: String,
}

#[derive(Debug, Serialize)]
pub struct OrderItemResponse {
    pub product_id: String,
    pub quantity: u32,
    pub unit_price_amount: String,
    pub unit_price_currency: String,
    pub subtotal_amount: String,
}

#[derive(Debug, Serialize)]
pub struct GetOrderResponse {
    pub order_id: String,
    pub customer_id: String,
    pub status: String,
    pub total_amount: String,
    pub total_currency: String,
    pub items: Vec<OrderItemResponse>,
}

#[derive(Debug, Deserialize)]
pub struct GetOrderParams {
    pub requesting_user_id: String,
}

// --- Handlers ---

/// POST /orders
pub async fn place_order(
    State(state): State<OrderState>,
    Json(body): Json<PlaceOrderRequest>,
) -> Result<(StatusCode, Json<PlaceOrderResponse>), DomainError> {
    let command = PlaceOrderCommand {
        customer_id: body.customer_id,
        items: body
            .items
            .into_iter()
            .map(|i| PlaceOrderItemInput {
                product_id: i.product_id,
                quantity: i.quantity,
                unit_price_amount: i.unit_price_amount,
                unit_price_currency: i.unit_price_currency,
            })
            .collect(),
    };

    let result = state.place_order.execute(command).await?;

    Ok((
        StatusCode::CREATED,
        Json(PlaceOrderResponse {
            order_id: result.order_id,
            status: result.status,
            total_amount: result.total_amount,
            total_currency: result.total_currency,
        }),
    ))
}

/// GET /orders/:id?requesting_user_id=...
pub async fn get_order(
    State(state): State<OrderState>,
    Path(order_id): Path<String>,
    Query(params): Query<GetOrderParams>,
) -> Result<Json<GetOrderResponse>, DomainError> {
    let result = state
        .get_order
        .execute(GetOrderQuery {
            order_id,
            requesting_user_id: params.requesting_user_id,
        })
        .await?;

    Ok(Json(GetOrderResponse {
        order_id: result.order_id,
        customer_id: result.customer_id,
        status: result.status,
        total_amount: result.total_amount,
        total_currency: result.total_currency,
        items: result
            .items
            .into_iter()
            .map(|i| OrderItemResponse {
                product_id: i.product_id,
                quantity: i.quantity,
                unit_price_amount: i.unit_price_amount,
                unit_price_currency: i.unit_price_currency,
                subtotal_amount: i.subtotal_amount,
            })
            .collect(),
    }))
}

/// Build the Axum router for order endpoints.
pub fn order_router(state: OrderState) -> Router {
    Router::new()
        .route("/orders", post(place_order))
        .route("/orders/:id", get(get_order))
        .with_state(state)
}
```

---

## Error Response Mapping (`src/adapters/inbound/error_response.rs`)

```rust
use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;

use crate::domain::errors::DomainError;

/// Map domain errors to Axum HTTP responses.
/// This is the ONLY place where DomainError variants are matched and converted to HTTP status codes.
impl IntoResponse for DomainError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            DomainError::OrderNotFound(_) => (StatusCode::NOT_FOUND, self.to_string()),

            DomainError::InvalidOrderItems
            | DomainError::InvalidOrderState { .. }
            | DomainError::CurrencyMismatch { .. }
            | DomainError::InvalidMoney(_)
            | DomainError::InvalidId(_) => (StatusCode::UNPROCESSABLE_ENTITY, self.to_string()),
        };

        (status, Json(json!({ "error": message }))).into_response()
    }
}
```

---

## Outbound Adapter — sqlx (`src/adapters/outbound/postgres_order_repository.rs`)

```rust
use async_trait::async_trait;
use rust_decimal::Decimal;
use sqlx::PgPool;
use uuid::Uuid;
use std::str::FromStr;

use crate::domain::entities::order::{Order, OrderItem, OrderStatus};
use crate::domain::value_objects::{Money, OrderId, UserId};
use crate::ports::outbound::order_repository::{OrderRepository, RepositoryError};

pub struct PostgresOrderRepository {
    pool: PgPool,
}

impl PostgresOrderRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl OrderRepository for PostgresOrderRepository {
    async fn save(&self, order: &Order) -> Result<(), RepositoryError> {
        let mut tx = self.pool.begin().await?;

        sqlx::query!(
            r#"
            INSERT INTO orders (id, customer_id, status, total_amount, total_currency)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                status = EXCLUDED.status,
                total_amount = EXCLUDED.total_amount
            "#,
            order.id().as_uuid(),
            order.customer_id().as_uuid(),
            order.status().to_string(),
            order.total().amount(),
            order.total().currency(),
        )
        .execute(&mut *tx)
        .await?;

        // Delete existing items and re-insert (simple upsert strategy).
        sqlx::query!("DELETE FROM order_items WHERE order_id = $1", order.id().as_uuid())
            .execute(&mut *tx)
            .await?;

        for item in order.items() {
            sqlx::query!(
                r#"
                INSERT INTO order_items (order_id, product_id, quantity, unit_price_amount, unit_price_currency)
                VALUES ($1, $2, $3, $4, $5)
                "#,
                order.id().as_uuid(),
                item.product_id,
                item.quantity as i32,
                item.unit_price.amount(),
                item.unit_price.currency(),
            )
            .execute(&mut *tx)
            .await?;
        }

        tx.commit().await?;
        Ok(())
    }

    async fn find_by_id(&self, id: &OrderId) -> Result<Option<Order>, RepositoryError> {
        let row = sqlx::query!(
            "SELECT id, customer_id, status, total_amount, total_currency FROM orders WHERE id = $1",
            id.as_uuid(),
        )
        .fetch_optional(&self.pool)
        .await?;

        let Some(row) = row else {
            return Ok(None);
        };

        let item_rows = sqlx::query!(
            "SELECT product_id, quantity, unit_price_amount, unit_price_currency FROM order_items WHERE order_id = $1",
            id.as_uuid(),
        )
        .fetch_all(&self.pool)
        .await?;

        Ok(Some(self.to_domain(row, item_rows)?))
    }

    async fn find_by_customer_id(&self, customer_id: &UserId) -> Result<Vec<Order>, RepositoryError> {
        let rows = sqlx::query!(
            "SELECT id, customer_id, status, total_amount, total_currency FROM orders WHERE customer_id = $1",
            customer_id.as_uuid(),
        )
        .fetch_all(&self.pool)
        .await?;

        let mut orders = Vec::with_capacity(rows.len());
        for row in rows {
            let item_rows = sqlx::query!(
                "SELECT product_id, quantity, unit_price_amount, unit_price_currency FROM order_items WHERE order_id = $1",
                row.id,
            )
            .fetch_all(&self.pool)
            .await?;
            orders.push(self.to_domain(row, item_rows)?);
        }
        Ok(orders)
    }
}

// Bring anonymous record types into scope for the helper below.
// In practice sqlx macros generate concrete types; this signature is illustrative.
impl PostgresOrderRepository {
    /// Map DB rows → Order domain object via `Order::reconstitute()` — skips invariant re-validation.
    fn to_domain(
        &self,
        row: impl OrderRow,
        item_rows: Vec<impl ItemRow>,
    ) -> Result<Order, RepositoryError> {
        let items = item_rows
            .into_iter()
            .map(|r| {
                let amount = r.unit_price_amount();
                let unit_price = Money::new(amount, r.unit_price_currency())
                    .map_err(|e| RepositoryError::Serialization(e.to_string()))?;
                Ok(OrderItem {
                    product_id: r.product_id(),
                    quantity: r.quantity() as u32,
                    unit_price,
                })
            })
            .collect::<Result<Vec<_>, RepositoryError>>()?;

        let status = match row.status().as_str() {
            "PENDING" => OrderStatus::Pending,
            "CONFIRMED" => OrderStatus::Confirmed,
            "CANCELLED" => OrderStatus::Cancelled,
            other => {
                return Err(RepositoryError::Serialization(format!(
                    "unknown status: {other}"
                )))
            }
        };

        let total = Money::new(row.total_amount(), row.total_currency())
            .map_err(|e| RepositoryError::Serialization(e.to_string()))?;

        Ok(Order::reconstitute(
            OrderId::from_uuid(row.id()),
            UserId::from_uuid(row.customer_id()),
            items,
            total,
            status,
        ))
    }
}
```

> **Note on sqlx query! macros**: `sqlx::query!` returns anonymous structs with typed fields.
> In real code the `to_domain` helper would accept those concrete types directly. The `impl OrderRow`
> / `impl ItemRow` traits above are placeholders to keep the example concise — replace them with
> the concrete `query!` return types when wiring the real implementation.

---

## Outbound Adapter — Stripe (`src/adapters/outbound/stripe_payment_gateway.rs`)

```rust
use async_trait::async_trait;
use rust_decimal::Decimal;
use std::str::FromStr;

use crate::ports::outbound::payment_gateway::{GatewayError, PaymentGateway, PaymentRequest, PaymentResult};

pub struct StripePaymentGateway {
    api_key: String,
}

impl StripePaymentGateway {
    pub fn new(api_key: impl Into<String>) -> Self {
        Self { api_key: api_key.into() }
    }
}

#[async_trait]
impl PaymentGateway for StripePaymentGateway {
    async fn charge(&self, req: PaymentRequest) -> Result<PaymentResult, GatewayError> {
        // In production: use the `stripe` crate or make an HTTP call to the Stripe API.
        // Here we show the translation pattern — parsing, calling, mapping result.
        let amount_str = req.amount.clone();
        let amount = Decimal::from_str(&amount_str)
            .map_err(|_| GatewayError::Upstream(format!("bad amount: {}", amount_str)))?;
        let amount_cents = (amount * Decimal::from(100))
            .to_u64()
            .ok_or_else(|| GatewayError::Upstream("amount overflow".into()))?;

        // --- Call to Stripe SDK omitted for brevity ---
        // let intent = stripe::PaymentIntent::create(&client, CreatePaymentIntent { ... }).await
        //     .map_err(|e| GatewayError::Network(e.to_string()))?;

        // Simulate a successful charge for illustration purposes.
        Ok(PaymentResult {
            success: true,
            transaction_id: format!("pi_test_{}", req.order_id),
            failure_reason: String::new(),
        })
    }

    async fn refund(&self, transaction_id: &str, amount: &str) -> Result<PaymentResult, GatewayError> {
        // --- Stripe refund call omitted for brevity ---
        Ok(PaymentResult {
            success: true,
            transaction_id: format!("re_test_{}", transaction_id),
            failure_reason: String::new(),
        })
    }
}
```

---

## Key rules illustrated here

- Handlers accept ports via Axum `State` extractor — never use case structs directly
- `impl IntoResponse for DomainError` lives in the inbound adapter layer — NOT in the domain layer
- `Order::reconstitute()` is called in `to_domain()` — never `Order::new()` when hydrating from DB
- Each sqlx adapter receives an injected `PgPool` — never creates its own connection
- `StripePaymentGateway` maps infrastructure errors to `GatewayError` — not to `DomainError`
- No business logic anywhere in this file — only translation between protocols and domain types
- Serde `Deserialize`/`Serialize` is used only in adapter request/response types — never in domain types
