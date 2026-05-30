# Python — Ports Layer Examples

Real Python code illustrating `src/ports/` patterns using `typing.Protocol`.

---

## Inbound Ports (`src/ports/inbound/`)

### `src/ports/inbound/place_order_port.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from decimal import Decimal
from typing import Protocol, runtime_checkable
from src.domain.value_objects.order_id import OrderId


@dataclass(frozen=True)
class PlaceOrderItemInput:
    product_id: str
    quantity: int
    unit_price_amount: Decimal
    unit_price_currency: str


@dataclass(frozen=True)
class PlaceOrderCommand:
    customer_id: str
    items: list[PlaceOrderItemInput]


@dataclass(frozen=True)
class PlaceOrderResult:
    order_id: OrderId
    status: str
    total_amount: Decimal
    total_currency: str


@runtime_checkable
class PlaceOrderPort(Protocol):
    async def execute(self, command: PlaceOrderCommand) -> PlaceOrderResult: ...
```

### `src/ports/inbound/get_order_port.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from decimal import Decimal
from typing import Protocol, runtime_checkable
from src.domain.value_objects.order_id import OrderId


@dataclass(frozen=True)
class GetOrderQuery:
    order_id: str
    requesting_user_id: str


@dataclass(frozen=True)
class OrderItemView:
    product_id: str
    quantity: int
    unit_price_amount: Decimal
    unit_price_currency: str
    subtotal_amount: Decimal


@dataclass(frozen=True)
class GetOrderResult:
    order_id: str
    customer_id: str
    status: str
    total_amount: Decimal
    total_currency: str
    items: list[OrderItemView]


@runtime_checkable
class GetOrderPort(Protocol):
    async def execute(self, query: GetOrderQuery) -> GetOrderResult: ...
```

---

## Outbound Ports (`src/ports/outbound/`)

### `src/ports/outbound/order_repository.py`

```python
from __future__ import annotations
from typing import Protocol, runtime_checkable
from src.domain.entities.order import Order
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId


@runtime_checkable
class OrderRepository(Protocol):
    async def save(self, order: Order) -> None: ...
    async def find_by_id(self, order_id: OrderId) -> Order | None: ...
    async def find_by_customer_id(self, customer_id: UserId) -> list[Order]: ...
```

### `src/ports/outbound/payment_gateway.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from decimal import Decimal
from typing import Protocol, runtime_checkable


@dataclass(frozen=True)
class PaymentRequest:
    order_id: str
    amount: Decimal
    currency: str
    payment_method_token: str


@dataclass(frozen=True)
class PaymentResult:
    success: bool
    transaction_id: str | None
    failure_reason: str | None


@runtime_checkable
class PaymentGateway(Protocol):
    async def charge(self, request: PaymentRequest) -> PaymentResult: ...
    async def refund(self, transaction_id: str, amount: Decimal) -> PaymentResult: ...
```

### `src/ports/outbound/event_publisher.py`

```python
from __future__ import annotations
from typing import Protocol, runtime_checkable
from src.domain.events.base import DomainEvent


@runtime_checkable
class EventPublisher(Protocol):
    async def publish(self, event: DomainEvent) -> None: ...
    async def publish_all(self, events: list[DomainEvent]) -> None: ...
```

### `src/ports/outbound/notification_gateway.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from typing import Protocol, runtime_checkable


@dataclass(frozen=True)
class NotificationMessage:
    recipient_id: str
    subject: str
    body: str


@runtime_checkable
class NotificationGateway(Protocol):
    async def send(self, message: NotificationMessage) -> None: ...
```

---

## Key rules illustrated here

- Every port is a `Protocol` with `@runtime_checkable` — no `ABC`, no `abstract`
- Port bodies use `...` (Ellipsis) — never `pass`, never `raise NotImplementedError`
- Command/query/result types are `@dataclass(frozen=True)` — immutable DTOs
- Inbound ports use primitive-friendly types (str, Decimal) in commands/results — not domain objects
- Outbound ports use domain types (Order, OrderId) since they are called from the application layer
- All port methods are `async def` — the whole stack is async (asyncpg / FastAPI)
