# Python — Adapters Layer Examples

Real Python code illustrating `src/adapters/` patterns with FastAPI and asyncpg.

---

## Inbound Adapter — FastAPI (`src/adapters/inbound/http_order_adapter.py`)

```python
from __future__ import annotations
from decimal import Decimal
from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, field_validator
from src.domain.errors import DomainError, OrderNotFoundError
from src.ports.inbound.place_order_port import (
    PlaceOrderCommand,
    PlaceOrderItemInput,
    PlaceOrderPort,
    PlaceOrderResult,
)
from src.ports.inbound.get_order_port import GetOrderPort, GetOrderQuery


# --- Request / Response schemas (Pydantic — adapter layer only) ---

class OrderItemRequest(BaseModel):
    product_id: str
    quantity: int
    unit_price_amount: Decimal
    unit_price_currency: str

    @field_validator("quantity")
    @classmethod
    def quantity_must_be_positive(cls, v: int) -> int:
        if v <= 0:
            raise ValueError("quantity must be positive")
        return v


class PlaceOrderRequest(BaseModel):
    customer_id: str
    items: list[OrderItemRequest]


class PlaceOrderResponse(BaseModel):
    order_id: str
    status: str
    total_amount: Decimal
    total_currency: str


class OrderItemResponse(BaseModel):
    product_id: str
    quantity: int
    unit_price_amount: Decimal
    subtotal_amount: Decimal


class GetOrderResponse(BaseModel):
    order_id: str
    customer_id: str
    status: str
    total_amount: Decimal
    total_currency: str
    items: list[OrderItemResponse]


# --- Router factory ---

def make_order_router(
    place_order_port: PlaceOrderPort,
    get_order_port: GetOrderPort,
) -> APIRouter:
    """
    Returns a configured APIRouter.
    Ports are injected via closure — no global state, no DI container needed here.
    """
    router = APIRouter(prefix="/orders", tags=["orders"])

    @router.post("/", status_code=status.HTTP_201_CREATED, response_model=PlaceOrderResponse)
    async def place_order(body: PlaceOrderRequest) -> PlaceOrderResponse:
        try:
            command = PlaceOrderCommand(
                customer_id=body.customer_id,
                items=[
                    PlaceOrderItemInput(
                        product_id=item.product_id,
                        quantity=item.quantity,
                        unit_price_amount=item.unit_price_amount,
                        unit_price_currency=item.unit_price_currency,
                    )
                    for item in body.items
                ],
            )
            result: PlaceOrderResult = await place_order_port.execute(command)
            return PlaceOrderResponse(
                order_id=str(result.order_id),
                status=result.status,
                total_amount=result.total_amount,
                total_currency=result.total_currency,
            )
        except OrderNotFoundError as exc:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc))
        except DomainError as exc:
            raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))

    @router.get("/{order_id}", response_model=GetOrderResponse)
    async def get_order(order_id: str, requesting_user_id: str) -> GetOrderResponse:
        try:
            query = GetOrderQuery(order_id=order_id, requesting_user_id=requesting_user_id)
            result = await get_order_port.execute(query)
            return GetOrderResponse(
                order_id=result.order_id,
                customer_id=result.customer_id,
                status=result.status,
                total_amount=result.total_amount,
                total_currency=result.total_currency,
                items=[
                    OrderItemResponse(
                        product_id=i.product_id,
                        quantity=i.quantity,
                        unit_price_amount=i.unit_price_amount,
                        subtotal_amount=i.subtotal_amount,
                    )
                    for i in result.items
                ],
            )
        except OrderNotFoundError as exc:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc))
        except DomainError as exc:
            raise HTTPException(status_code=status.HTTP_422_UNPROCESSABLE_ENTITY, detail=str(exc))

    return router
```

---

## Outbound Adapter — asyncpg (`src/adapters/outbound/postgres_order_repository.py`)

```python
from __future__ import annotations
from decimal import Decimal
import asyncpg
from src.domain.entities.order import Order, OrderItem, OrderStatus
from src.domain.value_objects.money import Money
from src.domain.value_objects.order_id import OrderId
from src.domain.value_objects.user_id import UserId


class PostgresOrderRepository:
    """Implements OrderRepository using asyncpg connection pool."""

    def __init__(self, pool: asyncpg.Pool) -> None:
        self._pool = pool

    async def save(self, order: Order) -> None:
        async with self._pool.acquire() as conn:
            async with conn.transaction():
                await conn.execute(
                    """
                    INSERT INTO orders (id, customer_id, status, total_amount, total_currency)
                    VALUES ($1, $2, $3, $4, $5)
                    ON CONFLICT (id) DO UPDATE SET
                        status = EXCLUDED.status,
                        total_amount = EXCLUDED.total_amount
                    """,
                    str(order.id),
                    str(order.customer_id),
                    order.status.value,
                    order.total.amount,
                    order.total.currency,
                )
                # Remove existing items and re-insert (simple upsert strategy)
                await conn.execute("DELETE FROM order_items WHERE order_id = $1", str(order.id))
                await conn.executemany(
                    """
                    INSERT INTO order_items (order_id, product_id, quantity, unit_price_amount, unit_price_currency)
                    VALUES ($1, $2, $3, $4, $5)
                    """,
                    [
                        (
                            str(order.id),
                            item.product_id,
                            item.quantity,
                            item.unit_price.amount,
                            item.unit_price.currency,
                        )
                        for item in order.items
                    ],
                )

    async def find_by_id(self, order_id: OrderId) -> Order | None:
        async with self._pool.acquire() as conn:
            row = await conn.fetchrow(
                "SELECT id, customer_id, status, total_amount, total_currency FROM orders WHERE id = $1",
                str(order_id),
            )
            if row is None:
                return None
            item_rows = await conn.fetch(
                "SELECT product_id, quantity, unit_price_amount, unit_price_currency FROM order_items WHERE order_id = $1",
                str(order_id),
            )
            return self._to_domain(row, item_rows)

    async def find_by_customer_id(self, customer_id: UserId) -> list[Order]:
        async with self._pool.acquire() as conn:
            rows = await conn.fetch(
                "SELECT id, customer_id, status, total_amount, total_currency FROM orders WHERE customer_id = $1",
                str(customer_id),
            )
            orders = []
            for row in rows:
                item_rows = await conn.fetch(
                    "SELECT product_id, quantity, unit_price_amount, unit_price_currency FROM order_items WHERE order_id = $1",
                    row["id"],
                )
                orders.append(self._to_domain(row, item_rows))
            return orders

    def _to_domain(self, row: asyncpg.Record, item_rows: list[asyncpg.Record]) -> Order:
        """Map DB rows to an Order domain object via reconstitute() — no invariant re-validation."""
        items = [
            OrderItem(
                product_id=r["product_id"],
                quantity=r["quantity"],
                unit_price=Money(
                    amount=Decimal(str(r["unit_price_amount"])),
                    currency=r["unit_price_currency"],
                ),
            )
            for r in item_rows
        ]
        return Order.reconstitute(
            order_id=OrderId.from_string(row["id"]),
            customer_id=UserId.from_string(row["customer_id"]),
            items=items,
            total=Money(
                amount=Decimal(str(row["total_amount"])),
                currency=row["total_currency"],
            ),
            status=OrderStatus(row["status"]),
        )
```

---

## Outbound Adapter — Stripe (`src/adapters/outbound/stripe_payment_gateway.py`)

```python
from __future__ import annotations
from decimal import Decimal
import stripe
from src.ports.outbound.payment_gateway import PaymentGateway, PaymentRequest, PaymentResult


class StripePaymentGateway:
    """Implements PaymentGateway using the Stripe Python SDK."""

    def __init__(self, api_key: str) -> None:
        stripe.api_key = api_key

    async def charge(self, request: PaymentRequest) -> PaymentResult:
        try:
            # Stripe amounts are in smallest currency unit (cents)
            amount_cents = int(request.amount * 100)
            intent = stripe.PaymentIntent.create(
                amount=amount_cents,
                currency=request.currency.lower(),
                payment_method=request.payment_method_token,
                confirm=True,
                metadata={"order_id": request.order_id},
            )
            return PaymentResult(
                success=intent.status == "succeeded",
                transaction_id=intent.id,
                failure_reason=None if intent.status == "succeeded" else intent.last_payment_error,
            )
        except stripe.CardError as exc:
            return PaymentResult(
                success=False,
                transaction_id=None,
                failure_reason=exc.user_message,
            )
        except stripe.StripeError as exc:
            # Infrastructure error — re-raise for the use case to handle
            raise RuntimeError(f"Stripe error: {exc}") from exc

    async def refund(self, transaction_id: str, amount: Decimal) -> PaymentResult:
        try:
            refund = stripe.Refund.create(
                payment_intent=transaction_id,
                amount=int(amount * 100),
            )
            return PaymentResult(
                success=refund.status == "succeeded",
                transaction_id=refund.id,
                failure_reason=None,
            )
        except stripe.StripeError as exc:
            raise RuntimeError(f"Stripe refund error: {exc}") from exc
```

---

## Key rules illustrated here

- Pydantic `BaseModel` schemas live in the inbound adapter — never in domain or application
- Route handlers map `DomainError` → HTTP 422, `OrderNotFoundError` → HTTP 404
- `PostgresOrderRepository._to_domain()` calls `Order.reconstitute()` — never `Order.create()`
- Each asyncpg adapter receives an injected `asyncpg.Pool` — never creates its own connection
- `StripePaymentGateway` maps card errors to `PaymentResult(success=False)` and maps infrastructure errors to `RuntimeError` (not domain errors)
- No business logic anywhere in this file — only translation
