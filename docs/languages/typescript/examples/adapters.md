# TypeScript — Adapters Layer Examples

Real TypeScript code illustrating `src/adapters/` patterns with TanStack Start server functions and postgres.js.

---

## Inbound Adapter — TanStack Start (`src/adapters/inbound/order-server-fns.ts`)

```typescript
import { createServerFn } from "@tanstack/start";
import { z } from "zod";
import { UserId } from "../../domain/value-objects/user-id";
import { Money } from "../../domain/value-objects/money";
import { makeOrderItem } from "../../domain/entities/order";
import {
  OrderNotFoundError,
  DomainError,
} from "../../domain/errors";
import type { PlaceOrderPort } from "../../ports/inbound/place-order-port";
import type { GetOrderPort } from "../../ports/inbound/get-order-port";

// --- Zod schemas (adapter layer only — never leak into domain or application) ---

const orderItemInputSchema = z.object({
  productId: z.string().min(1),
  quantity: z.number().int().positive(),
  unitPriceAmount: z.string().regex(/^\d+(\.\d{1,2})?$/, "Must be a valid decimal"),
  unitPriceCurrency: z.string().length(3),
});

const placeOrderSchema = z.object({
  customerId: z.string().uuid(),
  items: z.array(orderItemInputSchema).min(1, "Order must have at least one item"),
});

const getOrderSchema = z.object({
  orderId: z.string().uuid(),
  requestingUserId: z.string().uuid(),
});

// --- Server function factories ---
// Ports are module-level singletons injected at startup from infrastructure/di/wire.ts

let _placeOrderPort: PlaceOrderPort;
let _getOrderPort: GetOrderPort;

export function initOrderServerFns(
  placeOrderPort: PlaceOrderPort,
  getOrderPort: GetOrderPort,
): void {
  _placeOrderPort = placeOrderPort;
  _getOrderPort = getOrderPort;
}

// --- Server functions ---

export const placeOrderFn = createServerFn({ method: "POST" })
  .validator(placeOrderSchema)
  .handler(async ({ data }) => {
    try {
      const result = await _placeOrderPort.execute({
        customerId: data.customerId,
        items: data.items.map((i) => ({
          productId: i.productId,
          quantity: i.quantity,
          unitPriceAmount: i.unitPriceAmount,
          unitPriceCurrency: i.unitPriceCurrency,
        })),
      });
      return { success: true as const, data: result };
    } catch (err) {
      if (err instanceof OrderNotFoundError) {
        throw new Error(`404:${err.message}`);
      }
      if (err instanceof DomainError) {
        throw new Error(`422:${err.message}`);
      }
      throw err;
    }
  });

export const getOrderFn = createServerFn({ method: "GET" })
  .validator(getOrderSchema)
  .handler(async ({ data }) => {
    try {
      const result = await _getOrderPort.execute({
        orderId: data.orderId,
        requestingUserId: data.requestingUserId,
      });
      return { success: true as const, data: result };
    } catch (err) {
      if (err instanceof OrderNotFoundError) {
        throw new Error(`404:${err.message}`);
      }
      if (err instanceof DomainError) {
        throw new Error(`422:${err.message}`);
      }
      throw err;
    }
  });
```

---

## Outbound Adapter — postgres.js (`src/adapters/outbound/postgres-order-repository.ts`)

```typescript
import type { Sql, Row } from "postgres";
import Decimal from "decimal.js";
import { Order, makeOrderItem } from "../../domain/entities/order";
import type { OrderStatus } from "../../domain/entities/order";
import { OrderId } from "../../domain/value-objects/order-id";
import { UserId } from "../../domain/value-objects/user-id";
import { Money } from "../../domain/value-objects/money";
import type { OrderRepository } from "../../ports/outbound/order-repository";

export class PostgresOrderRepository implements OrderRepository {
  constructor(private readonly sql: Sql) {}

  async save(order: Order): Promise<void> {
    // Use a transaction to keep orders + items consistent
    await this.sql.begin(async (tx) => {
      await tx`
        INSERT INTO orders (id, customer_id, status, total_amount, total_currency)
        VALUES (
          ${order.id.toString()},
          ${order.customerId.toString()},
          ${order.status},
          ${order.total.amount.toString()},
          ${order.total.currency}
        )
        ON CONFLICT (id) DO UPDATE SET
          status        = EXCLUDED.status,
          total_amount  = EXCLUDED.total_amount,
          total_currency = EXCLUDED.total_currency
      `;

      // Delete + re-insert items (simple upsert strategy)
      await tx`DELETE FROM order_items WHERE order_id = ${order.id.toString()}`;

      if (order.items.length > 0) {
        const rows = order.items.map((item) => ({
          order_id: order.id.toString(),
          product_id: item.productId,
          quantity: item.quantity,
          unit_price_amount: item.unitPrice.amount.toString(),
          unit_price_currency: item.unitPrice.currency,
        }));
        await tx`INSERT INTO order_items ${tx(rows)}`;
      }
    });
  }

  async findById(orderId: OrderId): Promise<Order | null> {
    const [orderRow] = await this.sql<
      Array<{
        id: string;
        customer_id: string;
        status: string;
        total_amount: string;
        total_currency: string;
      }>
    >`
      SELECT id, customer_id, status, total_amount, total_currency
      FROM orders
      WHERE id = ${orderId.toString()}
    `;

    if (!orderRow) return null;

    const itemRows = await this.sql<
      Array<{
        product_id: string;
        quantity: number;
        unit_price_amount: string;
        unit_price_currency: string;
      }>
    >`
      SELECT product_id, quantity, unit_price_amount, unit_price_currency
      FROM order_items
      WHERE order_id = ${orderId.toString()}
    `;

    return this.toDomain(orderRow, itemRows);
  }

  async findByCustomerId(customerId: UserId): Promise<Order[]> {
    const orderRows = await this.sql<
      Array<{
        id: string;
        customer_id: string;
        status: string;
        total_amount: string;
        total_currency: string;
      }>
    >`
      SELECT id, customer_id, status, total_amount, total_currency
      FROM orders
      WHERE customer_id = ${customerId.toString()}
    `;

    const orders: Order[] = [];
    for (const row of orderRows) {
      const itemRows = await this.sql<
        Array<{
          product_id: string;
          quantity: number;
          unit_price_amount: string;
          unit_price_currency: string;
        }>
      >`
        SELECT product_id, quantity, unit_price_amount, unit_price_currency
        FROM order_items
        WHERE order_id = ${row.id}
      `;
      orders.push(this.toDomain(row, itemRows));
    }
    return orders;
  }

  private toDomain(
    row: {
      id: string;
      customer_id: string;
      status: string;
      total_amount: string;
      total_currency: string;
    },
    itemRows: Array<{
      product_id: string;
      quantity: number;
      unit_price_amount: string;
      unit_price_currency: string;
    }>,
  ): Order {
    const items = itemRows.map((r) =>
      makeOrderItem(
        r.product_id,
        r.quantity,
        Money.create(new Decimal(r.unit_price_amount), r.unit_price_currency),
      ),
    );

    return Order.reconstitute(
      OrderId.fromString(row.id),
      UserId.fromString(row.customer_id),
      items,
      Money.create(new Decimal(row.total_amount), row.total_currency),
      row.status as OrderStatus,
    );
  }
}
```

---

## Outbound Adapter — Stripe (`src/adapters/outbound/stripe-payment-gateway.ts`)

```typescript
import Stripe from "stripe";
import Decimal from "decimal.js";
import type { PaymentGateway, PaymentRequest, PaymentResult } from "../../ports/outbound/payment-gateway";

export class StripePaymentGateway implements PaymentGateway {
  private readonly stripe: Stripe;

  constructor(secretKey: string) {
    this.stripe = new Stripe(secretKey, { apiVersion: "2024-11-20.acacia" });
  }

  async charge(request: PaymentRequest): Promise<PaymentResult> {
    try {
      // Stripe amounts are in the smallest currency unit (e.g. cents for USD)
      const amountCents = new Decimal(request.amountAmount)
        .times(100)
        .toInteger()
        .toNumber();

      const intent = await this.stripe.paymentIntents.create({
        amount: amountCents,
        currency: request.amountCurrency.toLowerCase(),
        payment_method: request.paymentMethodToken,
        confirm: true,
        metadata: { order_id: request.orderId },
      });

      return {
        success: intent.status === "succeeded",
        transactionId: intent.id,
        failureReason:
          intent.status !== "succeeded"
            ? (intent.last_payment_error?.message ?? "unknown failure")
            : null,
      };
    } catch (err) {
      if (err instanceof Stripe.errors.StripeCardError) {
        return {
          success: false,
          transactionId: null,
          failureReason: err.message,
        };
      }
      // Infrastructure failure — re-throw as unknown for the adapter's caller to handle
      throw new Error(`Stripe charge failed: ${String(err)}`);
    }
  }

  async refund(transactionId: string, amount: string): Promise<PaymentResult> {
    try {
      const amountCents = new Decimal(amount).times(100).toInteger().toNumber();
      const refund = await this.stripe.refunds.create({
        payment_intent: transactionId,
        amount: amountCents,
      });
      return {
        success: refund.status === "succeeded",
        transactionId: refund.id,
        failureReason: null,
      };
    } catch (err) {
      throw new Error(`Stripe refund failed: ${String(err)}`);
    }
  }
}
```

---

## Key rules illustrated here

- Zod schemas live entirely in the inbound adapter — never in domain or application
- `PostgresOrderRepository.toDomain()` calls `Order.reconstitute()` — never `Order.create()`
- Each adapter receives an injected dependency (`Sql`, `Stripe`) — never creates its own connection
- `StripePaymentGateway` maps card errors to `PaymentResult(success: false)` and wraps infrastructure errors as generic `Error` — never as `DomainError`
- Server functions call port interfaces — never use case classes directly by name
- No business logic anywhere in this file — only translation between transport/DB and domain types
