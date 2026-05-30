# Python — Application Layer Examples

Real Python code illustrating `src/application/` patterns.

---

## Use Case: Place Order (`src/application/use_cases/place_order_use_case.py`)

```python
from __future__ import annotations
from src.domain.entities.order import Order, OrderItem
from src.domain.services.pricing_service import PricingService
from src.domain.value_objects.money import Money
from src.domain.value_objects.user_id import UserId
from src.ports.inbound.place_order_port import PlaceOrderCommand, PlaceOrderResult
from src.ports.outbound.order_repository import OrderRepository
from src.ports.outbound.event_publisher import EventPublisher
from decimal import Decimal


class PlaceOrderUseCase:
    """
    Command use case: create and persist a new Order.

    Implements PlaceOrderPort via structural subtyping (Protocol).
    No explicit 'implements' declaration needed.
    """

    def __init__(
        self,
        order_repo: OrderRepository,
        event_publisher: EventPublisher,
        pricing_service: PricingService,
    ) -> None:
        self._order_repo = order_repo
        self._event_publisher = event_publisher
        self._pricing_service = pricing_service

    async def execute(self, command: PlaceOrderCommand) -> PlaceOrderResult:
        # 1. Map primitives → domain types
        customer_id = UserId.from_string(command.customer_id)
        items = [
            OrderItem(
                product_id=item.product_id,
                quantity=item.quantity,
                unit_price=Money.of(item.unit_price_amount, item.unit_price_currency),
            )
            for item in command.items
        ]

        # 2. Create domain entity (invariants enforced inside Order.__init__)
        order = Order.create(customer_id=customer_id, items=items)

        # 3. Persist (port call — no knowledge of Postgres)
        await self._order_repo.save(order)

        # 4. Publish events (port call — no knowledge of Kafka)
        await self._event_publisher.publish_all(order.drain_events())

        # 5. Return primitive result DTO
        return PlaceOrderResult(
            order_id=order.id,
            status=order.status.value,
            total_amount=order.total.amount,
            total_currency=order.total.currency,
        )
```

---

## Use Case: Get Order (`src/application/use_cases/get_order_use_case.py`)

```python
from __future__ import annotations
from src.domain.errors import OrderNotFoundError
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId
from src.ports.inbound.get_order_port import GetOrderQuery, GetOrderResult, OrderItemView
from src.ports.outbound.order_repository import OrderRepository


class GetOrderUseCase:
    """
    Query use case: retrieve a single Order by ID.

    Implements GetOrderPort via structural subtyping (Protocol).
    """

    def __init__(self, order_repo: OrderRepository) -> None:
        self._order_repo = order_repo

    async def execute(self, query: GetOrderQuery) -> GetOrderResult:
        # 1. Parse identifiers
        order_id = OrderId.from_string(query.order_id)

        # 2. Load from repository
        order = await self._order_repo.find_by_id(order_id)
        if order is None:
            raise OrderNotFoundError(query.order_id)

        # 3. Map domain entity → result DTO (primitives only)
        return GetOrderResult(
            order_id=str(order.id),
            customer_id=str(order.customer_id),
            status=order.status.value,
            total_amount=order.total.amount,
            total_currency=order.total.currency,
            items=[
                OrderItemView(
                    product_id=item.product_id,
                    quantity=item.quantity,
                    unit_price_amount=item.unit_price.amount,
                    unit_price_currency=item.unit_price.currency,
                    subtotal_amount=item.subtotal.amount,
                )
                for item in order.items
            ],
        )
```

---

## Application Service (`src/application/services/notification_service.py`)

```python
from __future__ import annotations
from src.ports.outbound.notification_gateway import NotificationGateway, NotificationMessage
from src.ports.outbound.order_repository import OrderRepository
from src.domain.value_objects.order_id import OrderId
from src.domain.errors import OrderNotFoundError


class NotificationService:
    """
    Application service — shared logic used by multiple use cases.
    Coordinates ports; contains no business rules.
    """

    def __init__(
        self,
        notification_gateway: NotificationGateway,
        order_repo: OrderRepository,
    ) -> None:
        self._notification_gateway = notification_gateway
        self._order_repo = order_repo

    async def notify_order_placed(self, order_id: OrderId) -> None:
        order = await self._order_repo.find_by_id(order_id)
        if order is None:
            raise OrderNotFoundError(str(order_id))

        await self._notification_gateway.send(
            NotificationMessage(
                recipient_id=str(order.customer_id),
                subject="Your order has been placed",
                body=f"Order {order_id} for {order.total} is confirmed.",
            )
        )

    async def notify_order_cancelled(self, order_id: OrderId, reason: str) -> None:
        order = await self._order_repo.find_by_id(order_id)
        if order is None:
            raise OrderNotFoundError(str(order_id))

        await self._notification_gateway.send(
            NotificationMessage(
                recipient_id=str(order.customer_id),
                subject="Your order has been cancelled",
                body=f"Order {order_id} was cancelled: {reason}",
            )
        )
```

---

## Key rules illustrated here

- Use cases receive port interfaces in `__init__` — never concrete adapter classes
- `execute()` takes a command/query dataclass and returns a result dataclass (primitives only)
- Domain entity construction happens inside the use case, not in the adapter
- Events are drained from the entity and published after the entity is persisted
- Use cases do NOT catch exceptions — they let `DomainError` propagate to the adapter
- Application services depend only on ports — never on adapters or infrastructure
