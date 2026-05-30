# Python — Domain Layer Examples

Real Python code illustrating `src/domain/` patterns.
All examples use the Order / Payment domain.

---

## Errors (`src/domain/errors.py`)

```python
class DomainError(Exception):
    """Base class for all domain errors. Never raise this directly."""

class InvalidOrderItemsError(DomainError):
    def __init__(self) -> None:
        super().__init__("Order must have at least one item")

class InvalidOrderStateError(DomainError):
    def __init__(self, current: str, attempted: str) -> None:
        super().__init__(f"Cannot transition from {current} to {attempted}")

class OrderNotFoundError(DomainError):
    def __init__(self, order_id: str) -> None:
        super().__init__(f"Order {order_id} not found")

class CurrencyMismatchError(DomainError):
    def __init__(self, left: str, right: str) -> None:
        super().__init__(f"Currency mismatch: {left} vs {right}")

class InsufficientStockError(DomainError):
    def __init__(self, product_id: str) -> None:
        super().__init__(f"Insufficient stock for product {product_id}")
```

---

## Value Objects

### `src/domain/value_objects/order_id.py`

```python
from __future__ import annotations
import uuid
from dataclasses import dataclass

@dataclass(frozen=True)
class OrderId:
    value: uuid.UUID

    @classmethod
    def generate(cls) -> OrderId:
        return cls(value=uuid.uuid4())

    @classmethod
    def from_string(cls, raw: str) -> OrderId:
        try:
            return cls(value=uuid.UUID(raw))
        except ValueError:
            from src.domain.errors import DomainError
            raise DomainError(f"Invalid order id: {raw!r}")

    def __str__(self) -> str:
        return str(self.value)
```

### `src/domain/value_objects/money.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from decimal import Decimal
from src.domain.errors import CurrencyMismatchError, DomainError

@dataclass(frozen=True)
class Money:
    amount: Decimal
    currency: str

    def __post_init__(self) -> None:
        if self.amount < 0:
            raise DomainError("Money amount cannot be negative")
        if not self.currency:
            raise DomainError("Currency is required")

    @classmethod
    def of(cls, amount: str | int | Decimal, currency: str) -> Money:
        return cls(amount=Decimal(str(amount)), currency=currency)

    @classmethod
    def zero(cls, currency: str) -> Money:
        return cls(amount=Decimal("0"), currency=currency)

    def add(self, other: Money) -> Money:
        if self.currency != other.currency:
            raise CurrencyMismatchError(self.currency, other.currency)
        return Money(amount=self.amount + other.amount, currency=self.currency)

    def multiply(self, factor: int) -> Money:
        return Money(amount=self.amount * factor, currency=self.currency)

    def __str__(self) -> str:
        return f"{self.amount:.2f} {self.currency}"
```

### `src/domain/value_objects/user_id.py`

```python
from __future__ import annotations
import uuid
from dataclasses import dataclass

@dataclass(frozen=True)
class UserId:
    value: uuid.UUID

    @classmethod
    def generate(cls) -> UserId:
        return cls(value=uuid.uuid4())

    @classmethod
    def from_string(cls, raw: str) -> UserId:
        try:
            return cls(value=uuid.UUID(raw))
        except ValueError:
            from src.domain.errors import DomainError
            raise DomainError(f"Invalid user id: {raw!r}")

    def __str__(self) -> str:
        return str(self.value)
```

---

## Domain Events (`src/domain/events/`)

### `src/domain/events/base.py`

```python
from __future__ import annotations
import uuid
from dataclasses import dataclass, field
from datetime import datetime, timezone

@dataclass(frozen=True)
class DomainEvent:
    event_id: uuid.UUID = field(default_factory=uuid.uuid4)
    occurred_at: datetime = field(default_factory=lambda: datetime.now(timezone.utc))

    @property
    def event_type(self) -> str:
        return self.__class__.__name__
```

### `src/domain/events/order_placed.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from src.domain.events.base import DomainEvent
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId
from src.domain.value_objects.money import Money

@dataclass(frozen=True)
class OrderPlaced(DomainEvent):
    order_id: OrderId
    customer_id: UserId
    total: Money
```

### `src/domain/events/order_confirmed.py`

```python
from __future__ import annotations
from dataclasses import dataclass
from src.domain.events.base import DomainEvent
from src.domain.value_objects.order_id import OrderId

@dataclass(frozen=True)
class OrderConfirmed(DomainEvent):
    order_id: OrderId
```

---

## Entity (`src/domain/entities/order.py`)

```python
from __future__ import annotations
from enum import Enum
from dataclasses import dataclass
from decimal import Decimal
from src.domain.errors import InvalidOrderItemsError, InvalidOrderStateError
from src.domain.events.base import DomainEvent
from src.domain.events.order_placed import OrderPlaced
from src.domain.events.order_confirmed import OrderConfirmed
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId
from src.domain.value_objects.money import Money


class OrderStatus(str, Enum):
    PENDING = "PENDING"
    CONFIRMED = "CONFIRMED"
    CANCELLED = "CANCELLED"


@dataclass(frozen=True)
class OrderItem:
    product_id: str
    quantity: int
    unit_price: Money

    @property
    def subtotal(self) -> Money:
        return self.unit_price.multiply(self.quantity)


class Order:
    """Aggregate root for the Order domain concept."""

    def __init__(
        self,
        order_id: OrderId,
        customer_id: UserId,
        items: list[OrderItem],
    ) -> None:
        if not items:
            raise InvalidOrderItemsError()
        self._id = order_id
        self._customer_id = customer_id
        self._items = list(items)
        self._total = self._calculate_total()
        self._status = OrderStatus.PENDING
        self._events: list[DomainEvent] = []
        self._events.append(OrderPlaced(
            order_id=self._id,
            customer_id=self._customer_id,
            total=self._total,
        ))

    # --- Factories ---

    @classmethod
    def create(cls, customer_id: UserId, items: list[OrderItem]) -> Order:
        """Create a new Order. Enforces all invariants and emits OrderPlaced."""
        return cls(order_id=OrderId.generate(), customer_id=customer_id, items=items)

    @classmethod
    def reconstitute(
        cls,
        order_id: OrderId,
        customer_id: UserId,
        items: list[OrderItem],
        total: Money,
        status: OrderStatus,
    ) -> Order:
        """Hydrate an Order from storage. Skips invariant checks, emits no events."""
        order = cls.__new__(cls)
        order._id = order_id
        order._customer_id = customer_id
        order._items = list(items)
        order._total = total
        order._status = status
        order._events = []
        return order

    # --- Properties ---

    @property
    def id(self) -> OrderId:
        return self._id

    @property
    def customer_id(self) -> UserId:
        return self._customer_id

    @property
    def items(self) -> list[OrderItem]:
        return list(self._items)

    @property
    def total(self) -> Money:
        return self._total

    @property
    def status(self) -> OrderStatus:
        return self._status

    # --- State transitions ---

    def confirm(self) -> None:
        if self._status != OrderStatus.PENDING:
            raise InvalidOrderStateError(self._status.value, OrderStatus.CONFIRMED.value)
        self._status = OrderStatus.CONFIRMED
        self._events.append(OrderConfirmed(order_id=self._id))

    def cancel(self, reason: str) -> None:
        if self._status == OrderStatus.CANCELLED:
            raise InvalidOrderStateError(self._status.value, OrderStatus.CANCELLED.value)
        self._status = OrderStatus.CANCELLED

    # --- Event collection ---

    def drain_events(self) -> list[DomainEvent]:
        """Return and clear the accumulated domain events."""
        events = list(self._events)
        self._events.clear()
        return events

    # --- Helpers ---

    def _calculate_total(self) -> Money:
        total = Money.zero("USD")
        for item in self._items:
            total = total.add(item.subtotal)
        return total

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Order):
            return NotImplemented
        return self._id == other._id

    def __hash__(self) -> int:
        return hash(self._id)

    def __repr__(self) -> str:
        return f"Order(id={self._id}, status={self._status.value})"
```

---

## Domain Service (`src/domain/services/pricing_service.py`)

```python
from src.domain.value_objects.money import Money
from src.domain.entities.order import OrderItem
from decimal import Decimal


class PricingService:
    """Stateless domain service — no I/O, no state."""

    def apply_discount(self, items: list[OrderItem], discount_pct: Decimal) -> Money:
        """Calculate total after applying a percentage discount."""
        subtotal = Money.zero("USD")
        for item in items:
            subtotal = subtotal.add(item.subtotal)
        discount = Money(
            amount=(subtotal.amount * discount_pct / 100).quantize(Decimal("0.01")),
            currency=subtotal.currency,
        )
        return Money(
            amount=subtotal.amount - discount.amount,
            currency=subtotal.currency,
        )
```
