# Python — Test Examples

Real Python code illustrating test patterns with pytest, pytest-asyncio, and testcontainers.

---

## In-Memory Port Implementations (`tests/utils/in_memory_order_repository.py`)

```python
from __future__ import annotations
from src.domain.entities.order import Order
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId


class InMemoryOrderRepository:
    """
    In-memory implementation of OrderRepository.
    Used in unit tests — implements the same Protocol the real adapter does.
    No mocking library required.
    """

    def __init__(self) -> None:
        self._store: dict[str, Order] = {}

    async def save(self, order: Order) -> None:
        self._store[str(order.id)] = order

    async def find_by_id(self, order_id: OrderId) -> Order | None:
        return self._store.get(str(order_id))

    async def find_by_customer_id(self, customer_id: UserId) -> list[Order]:
        return [o for o in self._store.values() if o.customer_id == customer_id]

    # Test helper — not part of the port contract
    @property
    def all_orders(self) -> list[Order]:
        return list(self._store.values())
```

```python
# tests/utils/in_memory_event_publisher.py
from __future__ import annotations
from src.domain.events.base import DomainEvent


class InMemoryEventPublisher:
    """
    In-memory implementation of EventPublisher.
    Stores published events so tests can assert on them.
    """

    def __init__(self) -> None:
        self.published: list[DomainEvent] = []

    async def publish(self, event: DomainEvent) -> None:
        self.published.append(event)

    async def publish_all(self, events: list[DomainEvent]) -> None:
        self.published.extend(events)
```

---

## Unit Tests — Domain Entity (`tests/unit/test_order_entity.py`)

```python
import pytest
from decimal import Decimal
from src.domain.entities.order import Order, OrderItem, OrderStatus
from src.domain.errors import InvalidOrderItemsError, InvalidOrderStateError
from src.domain.events.order_placed import OrderPlaced
from src.domain.events.order_confirmed import OrderConfirmed
from src.domain.value_objects.money import Money
from src.domain.value_objects.user_id import UserId


def make_item(price: str = "9.99") -> OrderItem:
    return OrderItem(
        product_id="prod-1",
        quantity=2,
        unit_price=Money.of(price, "USD"),
    )


def make_order(items: list[OrderItem] | None = None) -> Order:
    return Order.create(
        customer_id=UserId.generate(),
        items=items or [make_item()],
    )


class TestOrderCreation:
    def test_emits_order_placed_event_on_creation(self) -> None:
        order = make_order()
        events = order.drain_events()
        assert len(events) == 1
        assert isinstance(events[0], OrderPlaced)
        assert events[0].order_id == order.id

    def test_status_is_pending_after_creation(self) -> None:
        order = make_order()
        assert order.status == OrderStatus.PENDING

    def test_raises_when_items_is_empty(self) -> None:
        with pytest.raises(InvalidOrderItemsError):
            Order.create(customer_id=UserId.generate(), items=[])

    def test_calculates_total_correctly(self) -> None:
        items = [
            OrderItem("prod-1", 2, Money.of("10.00", "USD")),
            OrderItem("prod-2", 1, Money.of("5.00", "USD")),
        ]
        order = make_order(items)
        assert order.total.amount == Decimal("25.00")


class TestOrderStateTransitions:
    def test_confirms_a_pending_order(self) -> None:
        order = make_order()
        order.drain_events()  # clear OrderPlaced
        order.confirm()
        events = order.drain_events()
        assert order.status == OrderStatus.CONFIRMED
        assert len(events) == 1
        assert isinstance(events[0], OrderConfirmed)

    def test_raises_when_confirming_a_cancelled_order(self) -> None:
        order = make_order()
        order.cancel("test cancellation")
        with pytest.raises(InvalidOrderStateError):
            order.confirm()

    def test_draining_events_clears_the_list(self) -> None:
        order = make_order()
        first_drain = order.drain_events()
        second_drain = order.drain_events()
        assert len(first_drain) == 1
        assert len(second_drain) == 0


class TestOrderReconstitute:
    def test_reconstitute_does_not_emit_events(self) -> None:
        original = make_order()
        order = Order.reconstitute(
            order_id=original.id,
            customer_id=original.customer_id,
            items=original.items,
            total=original.total,
            status=original.status,
        )
        assert order.drain_events() == []

    def test_reconstitute_preserves_all_fields(self) -> None:
        original = make_order()
        order = Order.reconstitute(
            order_id=original.id,
            customer_id=original.customer_id,
            items=original.items,
            total=original.total,
            status=OrderStatus.CONFIRMED,
        )
        assert order.id == original.id
        assert order.status == OrderStatus.CONFIRMED
```

---

## Unit Tests — Use Case (`tests/unit/test_place_order_use_case.py`)

```python
import pytest
from decimal import Decimal
from tests.utils.in_memory_order_repository import InMemoryOrderRepository
from tests.utils.in_memory_event_publisher import InMemoryEventPublisher
from src.application.use_cases.place_order_use_case import PlaceOrderUseCase
from src.domain.errors import InvalidOrderItemsError
from src.domain.events.order_placed import OrderPlaced
from src.domain.services.pricing_service import PricingService
from src.ports.inbound.place_order_port import PlaceOrderCommand, PlaceOrderItemInput


@pytest.fixture
def order_repo() -> InMemoryOrderRepository:
    return InMemoryOrderRepository()


@pytest.fixture
def event_publisher() -> InMemoryEventPublisher:
    return InMemoryEventPublisher()


@pytest.fixture
def use_case(
    order_repo: InMemoryOrderRepository,
    event_publisher: InMemoryEventPublisher,
) -> PlaceOrderUseCase:
    return PlaceOrderUseCase(
        order_repo=order_repo,
        event_publisher=event_publisher,
        pricing_service=PricingService(),
    )


def valid_command() -> PlaceOrderCommand:
    return PlaceOrderCommand(
        customer_id="00000000-0000-0000-0000-000000000001",
        items=[
            PlaceOrderItemInput(
                product_id="prod-1",
                quantity=2,
                unit_price_amount=Decimal("9.99"),
                unit_price_currency="USD",
            )
        ],
    )


class TestPlaceOrderUseCase:
    async def test_places_order_and_returns_result(
        self,
        use_case: PlaceOrderUseCase,
        order_repo: InMemoryOrderRepository,
    ) -> None:
        result = await use_case.execute(valid_command())

        assert result.order_id is not None
        assert result.status == "PENDING"
        assert result.total_amount == Decimal("19.98")
        assert len(order_repo.all_orders) == 1

    async def test_publishes_order_placed_event(
        self,
        use_case: PlaceOrderUseCase,
        event_publisher: InMemoryEventPublisher,
    ) -> None:
        await use_case.execute(valid_command())

        assert len(event_publisher.published) == 1
        assert isinstance(event_publisher.published[0], OrderPlaced)

    async def test_raises_when_items_is_empty(self, use_case: PlaceOrderUseCase) -> None:
        command = PlaceOrderCommand(
            customer_id="00000000-0000-0000-0000-000000000001",
            items=[],
        )
        with pytest.raises(InvalidOrderItemsError):
            await use_case.execute(command)
```

---

## Integration Test — Repository (`tests/integration/test_postgres_order_repository.py`)

```python
import pytest
import asyncpg
from decimal import Decimal
from testcontainers.postgres import PostgresContainer
from src.adapters.outbound.postgres_order_repository import PostgresOrderRepository
from src.domain.entities.order import Order, OrderItem, OrderStatus
from src.domain.value_objects.money import Money
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId

MIGRATIONS = """
CREATE TABLE IF NOT EXISTS orders (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    total_amount NUMERIC NOT NULL,
    total_currency TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES orders(id),
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit_price_amount NUMERIC NOT NULL,
    unit_price_currency TEXT NOT NULL
);
"""


@pytest.fixture(scope="module")
async def db_pool():
    with PostgresContainer("postgres:16") as pg:
        pool = await asyncpg.create_pool(pg.get_connection_url())
        await pool.execute(MIGRATIONS)
        yield pool
        await pool.close()


@pytest.fixture
async def repo(db_pool: asyncpg.Pool) -> PostgresOrderRepository:
    return PostgresOrderRepository(db_pool)


@pytest.fixture(autouse=True)
async def clean_tables(db_pool: asyncpg.Pool):
    yield
    await db_pool.execute("DELETE FROM order_items")
    await db_pool.execute("DELETE FROM orders")


def make_order() -> Order:
    return Order.create(
        customer_id=UserId.generate(),
        items=[OrderItem("prod-1", 2, Money.of("10.00", "USD"))],
    )


class TestPostgresOrderRepository:
    async def test_saves_and_retrieves_order_by_id(self, repo: PostgresOrderRepository) -> None:
        order = make_order()
        await repo.save(order)

        found = await repo.find_by_id(order.id)
        assert found is not None
        assert found.id == order.id
        assert found.status == OrderStatus.PENDING
        assert len(found.items) == 1

    async def test_returns_none_for_unknown_id(self, repo: PostgresOrderRepository) -> None:
        found = await repo.find_by_id(OrderId.generate())
        assert found is None

    async def test_updates_status_on_re_save(self, repo: PostgresOrderRepository) -> None:
        order = make_order()
        await repo.save(order)
        order.confirm()
        await repo.save(order)

        found = await repo.find_by_id(order.id)
        assert found is not None
        assert found.status == OrderStatus.CONFIRMED

    async def test_finds_orders_by_customer_id(self, repo: PostgresOrderRepository) -> None:
        customer = UserId.generate()
        other_customer = UserId.generate()
        order1 = Order.create(customer, [OrderItem("p1", 1, Money.of("5", "USD"))])
        order2 = Order.create(customer, [OrderItem("p2", 1, Money.of("5", "USD"))])
        order3 = Order.create(other_customer, [OrderItem("p3", 1, Money.of("5", "USD"))])
        for o in [order1, order2, order3]:
            await repo.save(o)

        results = await repo.find_by_customer_id(customer)
        assert len(results) == 2
```

---

## Integration Test — Inbound Adapter (`tests/integration/test_http_order_adapter.py`)

```python
import pytest
from decimal import Decimal
from httpx import AsyncClient, ASGITransport
from src.infrastructure.di.wire import create_app
from src.infrastructure.config.config import AppConfig, DatabaseConfig, StripeConfig, KafkaConfig, ServerConfig
from tests.utils.in_memory_order_repository import InMemoryOrderRepository
from tests.utils.in_memory_event_publisher import InMemoryEventPublisher
from src.adapters.inbound.http_order_adapter import make_order_router
from src.application.use_cases.place_order_use_case import PlaceOrderUseCase
from src.application.use_cases.get_order_use_case import GetOrderUseCase
from src.domain.services.pricing_service import PricingService
from fastapi import FastAPI


@pytest.fixture
def app() -> FastAPI:
    """Wire up a test app with in-memory repositories — no real DB needed."""
    order_repo = InMemoryOrderRepository()
    event_publisher = InMemoryEventPublisher()
    place_order = PlaceOrderUseCase(order_repo, event_publisher, PricingService())
    get_order = GetOrderUseCase(order_repo)
    app = FastAPI()
    app.include_router(make_order_router(place_order, get_order))
    return app


@pytest.fixture
async def client(app: FastAPI):
    async with AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as c:
        yield c


class TestHttpOrderAdapter:
    async def test_place_order_returns_201(self, client: AsyncClient) -> None:
        response = await client.post("/orders/", json={
            "customer_id": "00000000-0000-0000-0000-000000000001",
            "items": [{"product_id": "prod-1", "quantity": 2, "unit_price_amount": "9.99", "unit_price_currency": "USD"}],
        })
        assert response.status_code == 201
        body = response.json()
        assert body["order_id"] is not None
        assert body["status"] == "PENDING"

    async def test_place_order_returns_422_for_empty_items(self, client: AsyncClient) -> None:
        response = await client.post("/orders/", json={
            "customer_id": "00000000-0000-0000-0000-000000000001",
            "items": [],
        })
        assert response.status_code == 422

    async def test_get_order_returns_404_for_unknown_id(self, client: AsyncClient) -> None:
        response = await client.get(
            "/orders/00000000-0000-0000-0000-000000000099",
            params={"requesting_user_id": "00000000-0000-0000-0000-000000000001"},
        )
        assert response.status_code == 404
```

---

## pytest configuration (`pytest.ini`)

```ini
[pytest]
asyncio_mode = auto
testpaths = tests
python_files = test_*.py
python_classes = Test*
python_functions = test_*
```

---

## Key rules illustrated here

- In-memory repositories implement the same Protocol as the real adapter — structural subtyping, no inheritance
- Unit tests use `async def` throughout — `asyncio_mode = auto` removes the need for `@pytest.mark.asyncio`
- Integration tests use `testcontainers.postgres.PostgresContainer` for a real DB — no mocking of SQL
- Inbound adapter integration tests use `httpx.AsyncClient` with `ASGITransport` — a real HTTP client against a real (but in-memory-backed) app instance
- Fixtures clean up after themselves — `autouse=True` `clean_tables` fixture prevents state bleed between tests
