# Rust — Infrastructure Layer Examples

Real Rust code illustrating `src/infrastructure/` patterns with Axum and sqlx.

---

## Config (`src/infrastructure/config/mod.rs`)

```rust
use std::env;

/// Read a required environment variable — panics at startup if missing.
fn require(key: &str) -> String {
    env::var(key).unwrap_or_else(|_| panic!("Required environment variable {key:?} is not set"))
}

fn optional(key: &str, default: &str) -> String {
    env::var(key).unwrap_or_else(|_| default.to_string())
}

#[derive(Debug, Clone)]
pub struct DatabaseConfig {
    pub url: String,
    pub max_connections: u32,
}

impl DatabaseConfig {
    pub fn from_env() -> Self {
        Self {
            url: require("DATABASE_URL"),
            max_connections: optional("DB_MAX_CONNECTIONS", "10")
                .parse()
                .expect("DB_MAX_CONNECTIONS must be a number"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct StripeConfig {
    pub secret_key: String,
    pub webhook_secret: String,
}

impl StripeConfig {
    pub fn from_env() -> Self {
        Self {
            secret_key: require("STRIPE_SECRET_KEY"),
            webhook_secret: require("STRIPE_WEBHOOK_SECRET"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct KafkaConfig {
    pub brokers: String,
    pub orders_topic: String,
}

impl KafkaConfig {
    pub fn from_env() -> Self {
        Self {
            brokers: require("KAFKA_BROKERS"),
            orders_topic: optional("KAFKA_ORDERS_TOPIC", "orders.events"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct ServerConfig {
    pub host: String,
    pub port: u16,
}

impl ServerConfig {
    pub fn from_env() -> Self {
        Self {
            host: optional("SERVER_HOST", "0.0.0.0"),
            port: optional("SERVER_PORT", "3000")
                .parse()
                .expect("SERVER_PORT must be a number"),
        }
    }

    pub fn bind_address(&self) -> String {
        format!("{}:{}", self.host, self.port)
    }
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub database: DatabaseConfig,
    pub stripe: StripeConfig,
    pub kafka: KafkaConfig,
    pub server: ServerConfig,
}

impl AppConfig {
    /// Load all config from environment variables. Called once at startup.
    pub fn from_env() -> Self {
        Self {
            database: DatabaseConfig::from_env(),
            stripe: StripeConfig::from_env(),
            kafka: KafkaConfig::from_env(),
            server: ServerConfig::from_env(),
        }
    }
}
```

---

## DI Wiring (`src/infrastructure/di/wire.rs`)

```rust
use std::sync::Arc;
use sqlx::postgres::PgPoolOptions;
use thiserror::Error;

use crate::adapters::inbound::http_order_adapter::{order_router, OrderState};
use crate::adapters::outbound::{
    postgres_order_repository::PostgresOrderRepository,
    stripe_payment_gateway::StripePaymentGateway,
};
use crate::application::use_cases::{
    get_order::GetOrderUseCase,
    place_order::PlaceOrderUseCase,
};
use crate::infrastructure::config::AppConfig;

/// Errors that can occur during application startup.
#[derive(Debug, Error)]
pub enum SetupError {
    #[error("database connection failed: {0}")]
    Database(#[from] sqlx::Error),
}

/// The fully-wired application — owns the router and the DB pool.
pub struct Application {
    pub router: axum::Router,
}

/// Composition root — wires all layers together and returns a runnable Application.
/// Called once at startup.
pub async fn wire(cfg: &AppConfig) -> Result<Application, SetupError> {
    // 1. External connections
    let pool = PgPoolOptions::new()
        .max_connections(cfg.database.max_connections)
        .connect(&cfg.database.url)
        .await?;

    // 2. Outbound adapters (wrap in Arc — cloned cheaply into each use case)
    let order_repo = Arc::new(PostgresOrderRepository::new(pool.clone()));
    let payment_gw = Arc::new(StripePaymentGateway::new(cfg.stripe.secret_key.clone()));

    // 3. Use cases
    let place_order = Arc::new(PlaceOrderUseCase::new(
        Arc::clone(&order_repo) as Arc<dyn crate::ports::outbound::OrderRepository>,
        Arc::clone(&payment_gw) as Arc<dyn crate::ports::outbound::PaymentGateway>,
    ));
    let get_order = Arc::new(GetOrderUseCase::new(
        Arc::clone(&order_repo) as Arc<dyn crate::ports::outbound::OrderRepository>,
    ));

    // 4. Inbound adapter state
    let order_state = OrderState {
        place_order: Arc::clone(&place_order) as Arc<dyn crate::ports::inbound::PlaceOrderPort>,
        get_order: Arc::clone(&get_order) as Arc<dyn crate::ports::inbound::GetOrderPort>,
    };

    // 5. Router
    let router = order_router(order_state);

    Ok(Application { router })
}
```

---

## Entry Point (`src/main.rs`)

```rust
use infrastructure::config::AppConfig;
use infrastructure::di::wire;
use tracing_subscriber::EnvFilter;

mod adapters;
mod application;
mod domain;
mod infrastructure;
mod ports;

#[tokio::main]
async fn main() {
    // 1. Observability — init before anything else
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    // 2. Config — load all env vars; panics on missing required vars
    let config = AppConfig::from_env();

    tracing::info!("Starting order service on {}", config.server.bind_address());

    // 3. Wire — connect to DB, build adapters, use cases, and router
    let app = wire::wire(&config)
        .await
        .expect("Failed to wire application");

    // 4. Serve
    let listener = tokio::net::TcpListener::bind(config.server.bind_address())
        .await
        .expect("Failed to bind TCP listener");

    axum::serve(listener, app.router)
        .await
        .expect("Server error");
}
```

---

## Database Migrations (`src/infrastructure/db/migrations.rs`)

```rust
use sqlx::PgPool;

/// Run all pending migrations at startup.
/// In production, prefer `sqlx migrate run` as a pre-deploy step.
pub async fn run_migrations(pool: &PgPool) -> Result<(), sqlx::Error> {
    sqlx::migrate!("./migrations").run(pool).await?;
    Ok(())
}
```

Example migration file at `migrations/001_create_orders.sql`:

```sql
CREATE TABLE IF NOT EXISTS orders (
    id              UUID        PRIMARY KEY,
    customer_id     UUID        NOT NULL,
    status          TEXT        NOT NULL,
    total_amount    NUMERIC     NOT NULL,
    total_currency  TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id                    BIGSERIAL   PRIMARY KEY,
    order_id              UUID        NOT NULL REFERENCES orders(id),
    product_id            TEXT        NOT NULL,
    quantity              INTEGER     NOT NULL,
    unit_price_amount     NUMERIC     NOT NULL,
    unit_price_currency   TEXT        NOT NULL
);
```

---

## Observability (`src/infrastructure/observability/tracing.rs`)

```rust
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt, EnvFilter};

/// Configure structured tracing. Call once at startup before any other code.
pub fn init_tracing() {
    tracing_subscriber::registry()
        .with(EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()))
        .with(tracing_subscriber::fmt::layer().json())
        .init();
}
```

---

## Key rules illustrated here

- `AppConfig::from_env()` reads ALL env vars — no `env::var()` calls anywhere else in the codebase
- Missing required vars panic at startup — fail fast before accepting traffic
- `PgPool` is not `Arc`-wrapped — `sqlx::PgPool` is already internally ref-counted and `Clone`
- `Arc<dyn PortTrait>` is used for all port references passed into use cases — required for `Send + Sync` across async task boundaries
- The `wire()` function is the **only** place that imports from every layer — it is the true composition root
- `main.rs` is minimal: init observability, load config, wire, serve — no business logic
- Infrastructure is never imported by domain, application, ports, or adapters
