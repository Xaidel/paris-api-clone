# Python — Infrastructure Layer Examples

Real Python code illustrating `src/infrastructure/` patterns with FastAPI and asyncpg.

---

## Config (`src/infrastructure/config/config.py`)

```python
from __future__ import annotations
import os
from dataclasses import dataclass


def _require(key: str) -> str:
    value = os.getenv(key)
    if not value:
        raise RuntimeError(f"Required environment variable {key!r} is not set")
    return value


def _optional(key: str, default: str) -> str:
    return os.getenv(key, default)


@dataclass(frozen=True)
class DatabaseConfig:
    dsn: str
    min_connections: int
    max_connections: int

    @classmethod
    def from_env(cls) -> DatabaseConfig:
        return cls(
            dsn=_require("DATABASE_URL"),
            min_connections=int(_optional("DB_MIN_CONNECTIONS", "2")),
            max_connections=int(_optional("DB_MAX_CONNECTIONS", "10")),
        )


@dataclass(frozen=True)
class StripeConfig:
    api_key: str
    webhook_secret: str

    @classmethod
    def from_env(cls) -> StripeConfig:
        return cls(
            api_key=_require("STRIPE_API_KEY"),
            webhook_secret=_require("STRIPE_WEBHOOK_SECRET"),
        )


@dataclass(frozen=True)
class KafkaConfig:
    bootstrap_servers: str
    orders_topic: str

    @classmethod
    def from_env(cls) -> KafkaConfig:
        return cls(
            bootstrap_servers=_require("KAFKA_BOOTSTRAP_SERVERS"),
            orders_topic=_optional("KAFKA_ORDERS_TOPIC", "orders.events"),
        )


@dataclass(frozen=True)
class ServerConfig:
    host: str
    port: int
    debug: bool

    @classmethod
    def from_env(cls) -> ServerConfig:
        return cls(
            host=_optional("SERVER_HOST", "0.0.0.0"),
            port=int(_optional("SERVER_PORT", "8000")),
            debug=_optional("DEBUG", "false").lower() == "true",
        )


@dataclass(frozen=True)
class AppConfig:
    database: DatabaseConfig
    stripe: StripeConfig
    kafka: KafkaConfig
    server: ServerConfig

    @classmethod
    def from_env(cls) -> AppConfig:
        return cls(
            database=DatabaseConfig.from_env(),
            stripe=StripeConfig.from_env(),
            kafka=KafkaConfig.from_env(),
            server=ServerConfig.from_env(),
        )
```

---

## Simple DI — Manual Wiring (`src/infrastructure/di/wire.py`)

Use this for most projects. No library required.

```python
from __future__ import annotations
import asyncpg
import logging
from fastapi import FastAPI
from src.infrastructure.config.config import AppConfig
from src.adapters.outbound.postgres_order_repository import PostgresOrderRepository
from src.adapters.outbound.stripe_payment_gateway import StripePaymentGateway
from src.adapters.outbound.kafka_event_publisher import KafkaEventPublisher
from src.adapters.inbound.http_order_adapter import make_order_router
from src.application.use_cases.place_order_use_case import PlaceOrderUseCase
from src.application.use_cases.get_order_use_case import GetOrderUseCase
from src.domain.services.pricing_service import PricingService

logger = logging.getLogger(__name__)


async def create_app(config: AppConfig | None = None) -> FastAPI:
    """
    Composition root. Wires all layers together and returns a runnable FastAPI app.
    Called once at startup.
    """
    if config is None:
        config = AppConfig.from_env()

    # 1. External connections
    logger.info("Connecting to database...")
    pool = await asyncpg.create_pool(
        dsn=config.database.dsn,
        min_size=config.database.min_connections,
        max_size=config.database.max_connections,
    )

    # 2. Outbound adapters
    order_repo = PostgresOrderRepository(pool)
    payment_gateway = StripePaymentGateway(api_key=config.stripe.api_key)
    event_publisher = KafkaEventPublisher(bootstrap_servers=config.kafka.bootstrap_servers)

    # 3. Domain services (pure — no injection needed)
    pricing_service = PricingService()

    # 4. Use cases
    place_order_use_case = PlaceOrderUseCase(
        order_repo=order_repo,
        event_publisher=event_publisher,
        pricing_service=pricing_service,
    )
    get_order_use_case = GetOrderUseCase(order_repo=order_repo)

    # 5. FastAPI app + routes
    app = FastAPI(title="Order Service", debug=config.server.debug)
    app.include_router(make_order_router(place_order_use_case, get_order_use_case))

    # 6. Shutdown cleanup
    @app.on_event("shutdown")
    async def shutdown() -> None:
        logger.info("Closing database pool...")
        await pool.close()

    return app
```

### `src/infrastructure/di/main.py`

```python
import asyncio
import uvicorn
from src.infrastructure.config.config import AppConfig
from src.infrastructure.di.wire import create_app


async def main() -> None:
    config = AppConfig.from_env()
    app = await create_app(config)
    server = uvicorn.Server(
        uvicorn.Config(
            app=app,
            host=config.server.host,
            port=config.server.port,
            log_level="debug" if config.server.debug else "info",
        )
    )
    await server.serve()


if __name__ == "__main__":
    asyncio.run(main())
```

---

## Complex DI — python-dependency-injector (`src/infrastructure/di/container.py`)

Use this when the manual wiring file exceeds ~100 lines or lifecycle management is needed.

```python
from __future__ import annotations
import asyncpg
from dependency_injector import containers, providers
from src.infrastructure.config.config import AppConfig
from src.adapters.outbound.postgres_order_repository import PostgresOrderRepository
from src.adapters.outbound.stripe_payment_gateway import StripePaymentGateway
from src.adapters.outbound.kafka_event_publisher import KafkaEventPublisher
from src.application.use_cases.place_order_use_case import PlaceOrderUseCase
from src.application.use_cases.get_order_use_case import GetOrderUseCase
from src.domain.services.pricing_service import PricingService


class Container(containers.DeclarativeContainer):
    """Dependency injection container for the Order service."""

    config = providers.Singleton(AppConfig.from_env)

    # External connections
    db_pool = providers.Resource(
        asyncpg.create_pool,
        dsn=config.provided.database.dsn,
        min_size=config.provided.database.min_connections,
        max_size=config.provided.database.max_connections,
    )

    # Outbound adapters
    order_repo = providers.Factory(
        PostgresOrderRepository,
        pool=db_pool,
    )
    payment_gateway = providers.Factory(
        StripePaymentGateway,
        api_key=config.provided.stripe.api_key,
    )
    event_publisher = providers.Factory(
        KafkaEventPublisher,
        bootstrap_servers=config.provided.kafka.bootstrap_servers,
    )

    # Domain services
    pricing_service = providers.Factory(PricingService)

    # Use cases
    place_order_use_case = providers.Factory(
        PlaceOrderUseCase,
        order_repo=order_repo,
        event_publisher=event_publisher,
        pricing_service=pricing_service,
    )
    get_order_use_case = providers.Factory(
        GetOrderUseCase,
        order_repo=order_repo,
    )
```

### Wiring with the container

```python
# src/infrastructure/di/wire_container.py
from fastapi import FastAPI
from src.infrastructure.di.container import Container
from src.adapters.inbound.http_order_adapter import make_order_router


async def create_app_with_container() -> FastAPI:
    container = Container()
    await container.init_resources()   # initialises db_pool Resource

    app = FastAPI()
    app.include_router(
        make_order_router(
            place_order_port=container.place_order_use_case(),
            get_order_port=container.get_order_use_case(),
        )
    )

    @app.on_event("shutdown")
    async def shutdown() -> None:
        await container.shutdown_resources()

    return app
```

---

## Observability (`src/infrastructure/observability/logging.py`)

```python
import logging
import sys
from src.infrastructure.config.config import AppConfig


def configure_logging(config: AppConfig) -> None:
    """Configure structured logging. Call once at startup before creating the app."""
    level = logging.DEBUG if config.server.debug else logging.INFO
    handler = logging.StreamHandler(sys.stdout)

    if config.server.debug:
        # Human-readable in development
        formatter = logging.Formatter(
            "%(asctime)s [%(levelname)s] %(name)s: %(message)s"
        )
    else:
        # JSON-structured in production (use python-json-logger in real projects)
        formatter = logging.Formatter(
            '{"time":"%(asctime)s","level":"%(levelname)s","logger":"%(name)s","message":"%(message)s"}'
        )

    handler.setFormatter(formatter)
    logging.basicConfig(level=level, handlers=[handler])
    logging.getLogger("uvicorn.access").setLevel(logging.WARNING)
```

---

## Key rules illustrated here

- `AppConfig.from_env()` reads ALL env vars — no `os.getenv()` calls anywhere else in the codebase
- Missing required vars raise `RuntimeError` at startup — fail fast
- Manual `create_app()` covers most projects; `Container` is the complex-case alternative
- `@app.on_event("shutdown")` closes the DB pool — no leaked connections
- Domain and application layer classes are never imported into `container.py` for injection with `@inject` decorators — only infrastructure performs wiring
