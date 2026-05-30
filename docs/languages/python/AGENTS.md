# AGENTS.md — Python Language Supplement

Load this after `AGENTS.md` (root) and `src/{layer}/AGENTS.md`.
This file is authoritative for Python idioms and tooling. It never overrides
architecture rules — if a conflict exists, the architecture rule wins.

**Stack**: FastAPI · asyncpg · pytest · python-dependency-injector (complex DI only)

---

## Package Structure

**`src/` is not a Python package**

`src/` is a root directory added to `PYTHONPATH` — it does **not** have an `__init__.py`. This is the standard [src-layout](https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/). Every layer directory *does* have an `__init__.py`.

```
src/                          ← no __init__.py; added to PYTHONPATH
  domain/
    __init__.py               ← re-exports primary public types
    entities/
      __init__.py
      order.py
    value_objects/
      __init__.py
      money.py
      order_id.py
    events/
      __init__.py
      order_placed.py
    errors.py
  ports/
    __init__.py
    inbound/
      __init__.py
      place_order_port.py
    outbound/
      __init__.py
      order_repository.py
  adapters/
    __init__.py
    inbound/
      __init__.py
    outbound/
      __init__.py
  application/
    __init__.py
    use_cases/
      __init__.py
  infrastructure/
    __init__.py
    di/
      __init__.py
```

**Imports use the layer name directly — never prefix with `src`**:

```python
# Correct
from domain.entities.order import Order
from ports.outbound.order_repository import OrderRepository
from adapters.outbound.postgres_order_repository import PostgresOrderRepository

# WRONG — never prefix with src
from src.domain.entities.order import Order
```

**`__init__.py` re-export pattern** — each `__init__.py` re-exports the public API so callers import from the package, not the internal module:

```python
# src/domain/__init__.py
from domain.entities.order import Order
from domain.value_objects.money import Money
from domain.value_objects.order_id import OrderId
from domain.errors import DomainError
```

**Minimum configuration** — `pyproject.toml`:

```toml
[tool.setuptools.packages.find]
where = ["src"]
```

**Minimum configuration** — `pytest.ini` (or `pyproject.toml [tool.pytest.ini_options]`):

```ini
[pytest]
pythonpath = src
asyncio_mode = auto
```

**`Protocol` structural conformance — adapters do not inherit**

Concrete adapter classes do NOT inherit from the Protocol class. Python's structural subtyping means conformance is checked by method signature alone. An adapter that implements the matching methods satisfies the Protocol implicitly:

```python
# src/ports/outbound/order_repository.py
class OrderRepository(Protocol):
    async def save(self, order: Order) -> None: ...
    async def find_by_id(self, order_id: OrderId) -> Order | None: ...

# src/adapters/outbound/postgres_order_repository.py
class PostgresOrderRepository:          # no "(OrderRepository)" — intentional
    async def save(self, order: Order) -> None:
        ...                             # real implementation
    async def find_by_id(self, order_id: OrderId) -> Order | None:
        ...                             # real implementation
# PostgresOrderRepository satisfies OrderRepository implicitly via @runtime_checkable
```

---

## 1. Port / Interface Mechanism

Python has no `interface` keyword. Use `typing.Protocol` for all port definitions.

```python
from typing import Protocol, runtime_checkable

@runtime_checkable
class OrderRepository(Protocol):
    async def save(self, order: Order) -> None: ...
    async def find_by_id(self, order_id: OrderId) -> Order | None: ...
```

**Rules:**
- ALL ports (inbound and outbound) MUST use `Protocol`, never `ABC`
- Add `@runtime_checkable` so `isinstance()` checks work in tests
- Protocol methods must have `...` bodies — never `pass`, never `raise NotImplementedError`
- Concrete adapter classes do NOT declare `implements` or inherit from the Protocol — structural subtyping is implicit
- Do not import `Protocol` in domain or application — protocols live in `src/ports/` only

**Forbidden:**
```python
# WRONG — ABC is not a port; it leaks implementation detail
from abc import ABC, abstractmethod
class OrderRepository(ABC):
    @abstractmethod
    async def save(self, order: Order) -> None: ...
```

---

## 2. Value Object Immutability

Use `@dataclass(frozen=True)` for all value objects.

```python
from dataclasses import dataclass
from decimal import Decimal

@dataclass(frozen=True)
class Money:
    amount: Decimal
    currency: str

    def __post_init__(self) -> None:
        if self.amount < 0:
            raise DomainError("Money amount cannot be negative")
        if not self.currency:
            raise DomainError("Currency is required")

    def add(self, other: "Money") -> "Money":
        if self.currency != other.currency:
            raise DomainError(f"Currency mismatch: {self.currency} vs {other.currency}")
        return Money(amount=self.amount + other.amount, currency=self.currency)

    @classmethod
    def zero(cls, currency: str) -> "Money":
        return cls(amount=Decimal("0"), currency=currency)
```

**Rules:**
- `frozen=True` is mandatory — generates `__hash__` and prevents mutation
- Use `__post_init__` for validation — never validate in `__init__` directly
- Use `Decimal` for monetary values — never `float`
- Provide factory classmethods (`zero()`, `of()`) for common construction patterns
- Value objects are compared by value — `frozen=True` handles `__eq__` and `__hash__` automatically

**Forbidden:**
```python
@dataclass          # WRONG — mutable dataclass is not a value object
class Money:
    amount: float   # WRONG — float loses precision for money
```

---

## 3. Entity Identity and Equality

Entities have identity via a typed ID value object. Equality is by ID only.

```python
import uuid
from dataclasses import dataclass, field

@dataclass(frozen=True)
class OrderId:
    value: uuid.UUID

    @classmethod
    def generate(cls) -> "OrderId":
        return cls(value=uuid.uuid4())

    @classmethod
    def from_string(cls, raw: str) -> "OrderId":
        return cls(value=uuid.UUID(raw))

    def __str__(self) -> str:
        return str(self.value)
```

Entity classes are NOT frozen — they have mutable state (status, events list).
Override `__eq__` and `__hash__` to compare by ID only:

```python
class Order:
    def __init__(self, order_id: OrderId, ...) -> None:
        self._id = order_id
        ...

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, Order):
            return NotImplemented
        return self._id == other._id

    def __hash__(self) -> int:
        return hash(self._id)
```

---

## 4. Domain Error Types

Define a hierarchy rooted at `DomainError(Exception)`.

```python
class DomainError(Exception):
    """Base class for all domain errors."""

class InvalidOrderItemsError(DomainError):
    """Raised when an order has no items or invalid items."""

class OrderNotFoundError(DomainError):
    """Raised when an order cannot be found."""

class InvalidOrderStateError(DomainError):
    """Raised when a state transition is not permitted."""

class CurrencyMismatchError(DomainError):
    """Raised when money operations cross currencies."""
```

**Rules:**
- All domain errors inherit from `DomainError`, not directly from `Exception`
- Use specific subclasses — never `raise DomainError("order not found")` directly
- Application errors (not domain rules) use a separate `ApplicationError` hierarchy
- Adapters catch `DomainError` subclasses and map to HTTP status codes

---

## 5. Null / Absence Handling

Use `T | None` (Python 3.10+ union syntax). Never use `Optional[T]` in new code.

```python
async def find_by_id(self, order_id: OrderId) -> Order | None: ...
```

- Return `None` for "not found" — do not raise an exception from repository methods
- The use case raises `OrderNotFoundError` after receiving `None` from the repository
- Never use `Optional` from `typing` in new code — `T | None` is cleaner and preferred

---

## 6. Error Propagation

Python uses `raise` / `except`. Do not introduce `Result` types in domain or application.

```
domain          → raises DomainError subclasses
application     → raises DomainError (re-raised) or ApplicationError
inbound adapter → except DomainError → HTTP 422 / 404 / etc.
                  except ApplicationError → HTTP 400 / 403 / etc.
                  except Exception → HTTP 500
```

Adapters are the only place where `except` maps errors to transport codes.
Use case methods do not catch exceptions — they let them propagate.

---

## 7. Reconstitute Pattern

Entities need two constructors:
- `__init__` (or a `create()` classmethod) — enforces all invariants, emits events
- `reconstitute()` classmethod — hydrates from storage, skips invariant re-validation, emits NO events

```python
@classmethod
def reconstitute(
    cls,
    order_id: OrderId,
    customer_id: UserId,
    items: list[OrderItem],
    total: Money,
    status: OrderStatus,
) -> "Order":
    """Hydrate an Order from storage without re-validating invariants."""
    order = cls.__new__(cls)
    order._id = order_id
    order._customer_id = customer_id
    order._items = items
    order._total = total
    order._status = status
    order._events: list[DomainEvent] = []
    return order
```

Use `cls.__new__(cls)` to bypass `__init__` entirely. The `reconstitute()` method
is called only in outbound adapters (`PostgresOrderRepository.find_by_id()`).

---

## 8. Async Conventions

FastAPI and asyncpg are both async-native. The entire stack is async.

- ALL port methods are `async def`
- ALL use case `execute()` methods are `async def`
- ALL adapter methods are `async def`
- ALL FastAPI route handlers are `async def`
- Do NOT mix sync and async in the same call chain — use `asyncio.to_thread()` only for blocking third-party SDKs

**Test async code** with `pytest-asyncio`:
```ini
# pytest.ini
[pytest]
asyncio_mode = auto
```

Mark async tests with `@pytest.mark.asyncio` only when `asyncio_mode` is not `auto`.

---

## 9. Dependency Injection

### Simple case — manual constructor wiring

Wire everything manually in `src/infrastructure/di/`. No library needed for simple apps.

```python
# src/infrastructure/di/wire.py
async def create_app() -> FastAPI:
    config = load_config()
    pool = await asyncpg.create_pool(config.database.dsn)

    order_repo = PostgresOrderRepository(pool)
    payment_gateway = StripePaymentGateway(config.stripe.api_key)
    event_publisher = KafkaEventPublisher(config.kafka.bootstrap_servers)

    place_order_use_case = PlaceOrderUseCase(order_repo, payment_gateway, event_publisher)
    get_order_use_case = GetOrderUseCase(order_repo)

    app = FastAPI()
    register_order_routes(app, place_order_use_case, get_order_use_case)
    return app
```

### Complex case — python-dependency-injector

Use `dependency_injector.containers.DeclarativeContainer` when the manual wiring
file exceeds ~100 lines or when lifecycle management (singletons, scoped) is needed.

```python
from dependency_injector import containers, providers

class Container(containers.DeclarativeContainer):
    config = providers.Configuration()

    db_pool = providers.Resource(asyncpg.create_pool, dsn=config.database.dsn)

    order_repo = providers.Factory(PostgresOrderRepository, pool=db_pool)
    payment_gateway = providers.Factory(StripePaymentGateway, api_key=config.stripe.api_key)

    place_order_use_case = providers.Factory(
        PlaceOrderUseCase,
        order_repo=order_repo,
        payment_gateway=payment_gateway,
    )
```

Do NOT use `@inject` decorators in domain or application layer classes.
Injection is the infrastructure's job.

---

## 10. File and Package Naming

| Concept | File name | Class name |
|---|---|---|
| Entity | `order.py` | `Order` |
| Value object | `money.py`, `order_id.py` | `Money`, `OrderId` |
| Domain event | `order_placed.py` | `OrderPlaced` |
| Domain service | `pricing_service.py` | `PricingService` |
| Domain error | `errors.py` (all errors in one file per layer) | `DomainError`, `InvalidOrderItemsError` |
| Use case | `place_order_use_case.py` | `PlaceOrderUseCase` |
| Inbound port | `place_order_port.py` | `PlaceOrderPort` |
| Outbound port | `order_repository.py` | `OrderRepository` |
| Inbound adapter | `http_order_adapter.py` | `HttpOrderAdapter` (or a FastAPI router module) |
| Outbound adapter | `postgres_order_repository.py` | `PostgresOrderRepository` |
| Config | `config.py` | `AppConfig`, `DatabaseConfig` |

**Rules:**
- `snake_case` for all file names
- `PascalCase` for all class names
- `SCREAMING_SNAKE_CASE` for module-level constants
- One primary class per file (helpers and small supporting types may coexist)
- Package `__init__.py` files export the public API of that package — never import from sub-modules directly in other layers

---

## 11. FastAPI Integration

FastAPI route handlers are **inbound adapters**. They live in `src/adapters/inbound/`.

```python
# src/adapters/inbound/http_order_adapter.py
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from ports.inbound.place_order_port import PlaceOrderPort
from application.use_cases.place_order_use_case import PlaceOrderCommand
from domain.errors import DomainError

router = APIRouter(prefix="/orders", tags=["orders"])

class PlaceOrderRequest(BaseModel):
    customer_id: str
    items: list[OrderItemRequest]

def make_order_router(place_order_port: PlaceOrderPort) -> APIRouter:
    @router.post("/", status_code=201)
    async def place_order(body: PlaceOrderRequest) -> PlaceOrderResponse:
        try:
            command = PlaceOrderCommand(
                customer_id=body.customer_id,
                items=[...],
            )
            result = await place_order_port.execute(command)
            return PlaceOrderResponse(order_id=str(result.order_id), status=result.status)
        except DomainError as exc:
            raise HTTPException(status_code=422, detail=str(exc))
    return router
```

**Rules:**
- Route handlers receive port interfaces via closure or dependency injection — never use case classes directly
- Use Pydantic `BaseModel` for all request/response schemas — these live in the adapter, not in domain or application
- Map `DomainError` → 422, `OrderNotFoundError` → 404, authorization errors → 403
- Do not put business logic in route handlers — call the port, map the result, return

---

## 12. Forbidden Patterns

| Pattern | Why forbidden |
|---|---|
| `ABC` for port definitions | Use `Protocol` — ABC forces inheritance, Protocol is structural |
| Mutable `@dataclass` as value object | Value objects must not change after creation — use `frozen=True` |
| `float` for monetary values | Use `Decimal` — floats lose precision |
| `Optional[T]` in new code | Use `T \| None` — cleaner syntax, same semantics |
| `raise DomainError("message")` at call sites | Raise specific subclasses for precise error handling |
| `except Exception` in domain/application | Only adapters catch and map exceptions |
| Business logic in FastAPI route handlers | Route handlers translate only — no `if/else` business rules |
| `from src.infrastructure import ...` in any other layer | Infrastructure is never imported by other layers |
| `from src.domain import ...` anywhere | Use `from domain import ...` — `src` is on `PYTHONPATH`, not a package |
| `import *` from any module | Explicit imports only |
| Global mutable state outside `src/infrastructure/` | Shared state belongs in infrastructure only |

---

## 13. Tooling

| Tool | Purpose | Minimum config |
|---|---|---|
| `mypy` | Static type checking | `strict = true` in `mypy.ini` |
| `ruff` | Linting + formatting | `target-version = "py312"`, `line-length = 88` |
| `pytest` | Test runner | `asyncio_mode = auto` in `pytest.ini` |
| `pytest-asyncio` | Async test support | Required for all async tests |
| `testcontainers` | Real DB in integration tests | `testcontainers[postgresql]` |

Minimum `mypy.ini`:
```ini
[mypy]
strict = true
python_version = 3.12
```

Minimum `ruff.toml`:
```toml
target-version = "py312"
line-length = 88

[lint]
select = ["E", "F", "I", "UP", "B", "SIM"]
```

---

## 14. Self-Audit Checklist

Before submitting any Python code:

- [ ] All ports use `Protocol`, not `ABC`
- [ ] All value objects use `@dataclass(frozen=True)`
- [ ] `Decimal` used for all monetary values, never `float`
- [ ] `T | None` used for optional return types, not `Optional[T]`
- [ ] All port methods and use case `execute()` are `async def`
- [ ] Entities have both a `create()` / `__init__` constructor and a `reconstitute()` classmethod
- [ ] `reconstitute()` uses `cls.__new__(cls)` and emits no domain events
- [ ] Domain errors use specific subclasses of `DomainError`
- [ ] Only inbound adapters contain `except` blocks that map to HTTP status codes
- [ ] FastAPI route handlers receive port interfaces, not use case classes
- [ ] Pydantic schemas are in the adapter layer, not in domain or application
- [ ] `src/` has no `__init__.py`; every layer sub-directory does
- [ ] Imports use `from domain.x import Y`, never `from src.domain.x import Y`
- [ ] `pytest.ini` (or equivalent) sets `pythonpath = src`
- [ ] `mypy --strict` passes with no errors
- [ ] `ruff check` passes with no errors

---

## See also

- [`examples/domain.md`](examples/domain.md)
- [`examples/application.md`](examples/application.md)
- [`examples/ports.md`](examples/ports.md)
- [`examples/adapters.md`](examples/adapters.md)
- [`examples/infrastructure.md`](examples/infrastructure.md)
- [`examples/tests.md`](examples/tests.md)
- Root [`AGENTS.md`](../../../AGENTS.md)
