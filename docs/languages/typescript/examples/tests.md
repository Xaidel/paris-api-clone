# TypeScript — Test Examples

Real TypeScript code illustrating test patterns with `bun test` and `@testcontainers/postgresql`.

---

## In-Memory Port Implementations (`tests/utils/in-memory-order-repository.ts`)

```typescript
import type { OrderRepository } from "../../src/ports/outbound/order-repository";
import type { Order } from "../../src/domain/entities/order";
import type { OrderId } from "../../src/domain/value-objects/order-id";
import type { UserId } from "../../src/domain/value-objects/user-id";

/**
 * In-memory implementation of OrderRepository.
 * Used in unit tests — satisfies the same interface the real adapter does.
 * No mocking library required.
 */
export class InMemoryOrderRepository implements OrderRepository {
  private readonly store = new Map<string, Order>();

  async save(order: Order): Promise<void> {
    this.store.set(order.id.toString(), order);
  }

  async findById(orderId: OrderId): Promise<Order | null> {
    return this.store.get(orderId.toString()) ?? null;
  }

  async findByCustomerId(customerId: UserId): Promise<Order[]> {
    return [...this.store.values()].filter((o) =>
      o.customerId.equals(customerId),
    );
  }

  // Test helper — not part of the port contract
  get allOrders(): Order[] {
    return [...this.store.values()];
  }
}
```

```typescript
// tests/utils/in-memory-event-publisher.ts
import type { EventPublisher } from "../../src/ports/outbound/event-publisher";
import type { DomainEvent } from "../../src/domain/events/domain-event";

/**
 * In-memory implementation of EventPublisher.
 * Stores published events so tests can assert on them.
 */
export class InMemoryEventPublisher implements EventPublisher {
  readonly published: DomainEvent[] = [];

  async publish(event: DomainEvent): Promise<void> {
    this.published.push(event);
  }

  async publishAll(events: DomainEvent[]): Promise<void> {
    this.published.push(...events);
  }
}
```

---

## Unit Tests — Domain Entity (`tests/unit/order-entity.test.ts`)

```typescript
import { describe, it, expect } from "bun:test";
import Decimal from "decimal.js";
import { Order, makeOrderItem } from "../../src/domain/entities/order";
import { Money } from "../../src/domain/value-objects/money";
import { UserId } from "../../src/domain/value-objects/user-id";
import { OrderId } from "../../src/domain/value-objects/order-id";
import {
  InvalidOrderItemsError,
  InvalidOrderStateError,
} from "../../src/domain/errors";
import { OrderPlaced } from "../../src/domain/events/order-placed";
import { OrderConfirmed } from "../../src/domain/events/order-confirmed";

function makeItem(price = "9.99") {
  return makeOrderItem("prod-1", 2, Money.create(new Decimal(price), "USD"));
}

function makeOrder(items = [makeItem()]) {
  return Order.create(UserId.generate(), items);
}

describe("Order creation", () => {
  it("emits an OrderPlaced event on creation", () => {
    const order = makeOrder();
    const events = order.drainEvents();
    expect(events).toHaveLength(1);
    expect(events[0]).toBeInstanceOf(OrderPlaced);
    expect((events[0] as OrderPlaced).orderId.equals(order.id)).toBe(true);
  });

  it("status is PENDING after creation", () => {
    const order = makeOrder();
    expect(order.status).toBe("PENDING");
  });

  it("throws InvalidOrderItemsError when items is empty", () => {
    expect(() => Order.create(UserId.generate(), [])).toThrow(
      InvalidOrderItemsError,
    );
  });

  it("calculates total correctly", () => {
    const items = [
      makeOrderItem("prod-1", 2, Money.create(new Decimal("10.00"), "USD")),
      makeOrderItem("prod-2", 1, Money.create(new Decimal("5.00"), "USD")),
    ];
    const order = makeOrder(items);
    expect(order.total.amount.toString()).toBe("25");
  });
});

describe("Order state transitions", () => {
  it("confirms a PENDING order", () => {
    const order = makeOrder();
    order.drainEvents(); // clear OrderPlaced
    order.confirm();
    const events = order.drainEvents();
    expect(order.status).toBe("CONFIRMED");
    expect(events).toHaveLength(1);
    expect(events[0]).toBeInstanceOf(OrderConfirmed);
  });

  it("throws when confirming a CANCELLED order", () => {
    const order = makeOrder();
    order.cancel("test");
    expect(() => order.confirm()).toThrow(InvalidOrderStateError);
  });

  it("draining events clears the list", () => {
    const order = makeOrder();
    const first = order.drainEvents();
    const second = order.drainEvents();
    expect(first).toHaveLength(1);
    expect(second).toHaveLength(0);
  });
});

describe("Order.reconstitute", () => {
  it("does not emit events when reconstituted", () => {
    const original = makeOrder();
    const order = Order.reconstitute(
      original.id,
      original.customerId,
      [...original.items],
      original.total,
      original.status,
    );
    expect(order.drainEvents()).toHaveLength(0);
  });

  it("preserves all fields when reconstituted", () => {
    const original = makeOrder();
    const order = Order.reconstitute(
      original.id,
      original.customerId,
      [...original.items],
      original.total,
      "CONFIRMED",
    );
    expect(order.id.equals(original.id)).toBe(true);
    expect(order.status).toBe("CONFIRMED");
  });
});
```

---

## Unit Tests — Use Case (`tests/unit/place-order-use-case.test.ts`)

```typescript
import { describe, it, expect, beforeEach } from "bun:test";
import Decimal from "decimal.js";
import { PlaceOrderUseCase } from "../../src/application/use-cases/place-order-use-case";
import { InMemoryOrderRepository } from "../utils/in-memory-order-repository";
import { InMemoryEventPublisher } from "../utils/in-memory-event-publisher";
import { InvalidOrderItemsError } from "../../src/domain/errors";
import { OrderPlaced } from "../../src/domain/events/order-placed";
import type { PlaceOrderCommand } from "../../src/ports/inbound/place-order-port";

function validCommand(): PlaceOrderCommand {
  return {
    customerId: "00000000-0000-0000-0000-000000000001",
    items: [
      {
        productId: "prod-1",
        quantity: 2,
        unitPriceAmount: "9.99",
        unitPriceCurrency: "USD",
      },
    ],
  };
}

describe("PlaceOrderUseCase", () => {
  let orderRepo: InMemoryOrderRepository;
  let eventPublisher: InMemoryEventPublisher;
  let useCase: PlaceOrderUseCase;

  beforeEach(() => {
    orderRepo = new InMemoryOrderRepository();
    eventPublisher = new InMemoryEventPublisher();
    useCase = new PlaceOrderUseCase(orderRepo, eventPublisher);
  });

  it("places an order and returns a result", async () => {
    const result = await useCase.execute(validCommand());

    expect(result.orderId).toBeTruthy();
    expect(result.status).toBe("PENDING");
    expect(result.totalAmount).toBe("19.98");
    expect(orderRepo.allOrders).toHaveLength(1);
  });

  it("publishes an OrderPlaced event", async () => {
    await useCase.execute(validCommand());

    expect(eventPublisher.published).toHaveLength(1);
    expect(eventPublisher.published[0]).toBeInstanceOf(OrderPlaced);
  });

  it("throws InvalidOrderItemsError when items is empty", async () => {
    const command: PlaceOrderCommand = {
      customerId: "00000000-0000-0000-0000-000000000001",
      items: [],
    };
    await expect(useCase.execute(command)).rejects.toBeInstanceOf(
      InvalidOrderItemsError,
    );
  });
});
```

---

## Integration Test — Repository (`tests/integration/postgres-order-repository.test.ts`)

```typescript
import { describe, it, expect, beforeAll, afterAll, beforeEach } from "bun:test";
import { PostgreSqlContainer, type StartedPostgreSqlContainer } from "@testcontainers/postgresql";
import postgres from "postgres";
import Decimal from "decimal.js";
import { PostgresOrderRepository } from "../../src/adapters/outbound/postgres-order-repository";
import { Order, makeOrderItem } from "../../src/domain/entities/order";
import { Money } from "../../src/domain/value-objects/money";
import { UserId } from "../../src/domain/value-objects/user-id";
import { OrderId } from "../../src/domain/value-objects/order-id";

const MIGRATIONS = `
  CREATE TABLE IF NOT EXISTS orders (
    id            TEXT PRIMARY KEY,
    customer_id   TEXT NOT NULL,
    status        TEXT NOT NULL,
    total_amount  NUMERIC NOT NULL,
    total_currency TEXT NOT NULL
  );
  CREATE TABLE IF NOT EXISTS order_items (
    id                   SERIAL PRIMARY KEY,
    order_id             TEXT NOT NULL REFERENCES orders(id),
    product_id           TEXT NOT NULL,
    quantity             INTEGER NOT NULL,
    unit_price_amount    NUMERIC NOT NULL,
    unit_price_currency  TEXT NOT NULL
  );
`;

function makeOrder() {
  return Order.create(UserId.generate(), [
    makeOrderItem("prod-1", 2, Money.create(new Decimal("10.00"), "USD")),
  ]);
}

describe("PostgresOrderRepository", () => {
  let container: StartedPostgreSqlContainer;
  let sql: ReturnType<typeof postgres>;
  let repo: PostgresOrderRepository;

  beforeAll(async () => {
    container = await new PostgreSqlContainer("postgres:16").start();
    sql = postgres(container.getConnectionUri());
    await sql.unsafe(MIGRATIONS);
    repo = new PostgresOrderRepository(sql);
  });

  afterAll(async () => {
    await sql.end();
    await container.stop();
  });

  beforeEach(async () => {
    // Clean between tests — order matters due to FK constraint
    await sql`DELETE FROM order_items`;
    await sql`DELETE FROM orders`;
  });

  it("saves and retrieves an order by ID", async () => {
    const order = makeOrder();
    await repo.save(order);

    const found = await repo.findById(order.id);
    expect(found).not.toBeNull();
    expect(found!.id.equals(order.id)).toBe(true);
    expect(found!.status).toBe("PENDING");
    expect(found!.items).toHaveLength(1);
  });

  it("returns null for an unknown ID", async () => {
    const found = await repo.findById(OrderId.generate());
    expect(found).toBeNull();
  });

  it("updates status when order is re-saved", async () => {
    const order = makeOrder();
    await repo.save(order);
    order.confirm();
    await repo.save(order);

    const found = await repo.findById(order.id);
    expect(found!.status).toBe("CONFIRMED");
  });

  it("finds all orders for a customer", async () => {
    const customer = UserId.generate();
    const other = UserId.generate();
    const order1 = Order.create(customer, [makeOrderItem("p1", 1, Money.create(new Decimal("5"), "USD"))]);
    const order2 = Order.create(customer, [makeOrderItem("p2", 1, Money.create(new Decimal("5"), "USD"))]);
    const order3 = Order.create(other, [makeOrderItem("p3", 1, Money.create(new Decimal("5"), "USD"))]);
    for (const o of [order1, order2, order3]) await repo.save(o);

    const results = await repo.findByCustomerId(customer);
    expect(results).toHaveLength(2);
  });
});
```

---

## Integration Test — Inbound Adapter (`tests/integration/order-server-fns.test.ts`)

```typescript
import { describe, it, expect, beforeEach } from "bun:test";
import { PlaceOrderUseCase } from "../../src/application/use-cases/place-order-use-case";
import { GetOrderUseCase } from "../../src/application/use-cases/get-order-use-case";
import { InMemoryOrderRepository } from "../utils/in-memory-order-repository";
import { InMemoryEventPublisher } from "../utils/in-memory-event-publisher";
import { initOrderServerFns, placeOrderFn, getOrderFn } from "../../src/adapters/inbound/order-server-fns";

// NOTE: Server functions are tested here by calling their .handler directly.
// This validates the full adapter → use case → domain path without a real HTTP server.

describe("Order server functions", () => {
  let orderRepo: InMemoryOrderRepository;

  beforeEach(() => {
    orderRepo = new InMemoryOrderRepository();
    const eventPublisher = new InMemoryEventPublisher();
    initOrderServerFns(
      new PlaceOrderUseCase(orderRepo, eventPublisher),
      new GetOrderUseCase(orderRepo),
    );
  });

  it("placeOrderFn returns the new order", async () => {
    const result = await placeOrderFn({
      data: {
        customerId: "00000000-0000-0000-0000-000000000001",
        items: [
          {
            productId: "prod-1",
            quantity: 2,
            unitPriceAmount: "9.99",
            unitPriceCurrency: "USD",
          },
        ],
      },
    });
    expect(result.data.status).toBe("PENDING");
    expect(result.data.orderId).toBeTruthy();
  });

  it("getOrderFn throws a 404 error for an unknown order", async () => {
    await expect(
      getOrderFn({
        data: {
          orderId: "00000000-0000-0000-0000-000000000099",
          requestingUserId: "00000000-0000-0000-0000-000000000001",
        },
      }),
    ).rejects.toThrow("404:");
  });

  it("placeOrderFn throws a 422 error for an empty items list", async () => {
    await expect(
      placeOrderFn({
        data: {
          customerId: "00000000-0000-0000-0000-000000000001",
          items: [],
        },
      }),
    ).rejects.toThrow("422:");
  });
});
```

---

## `bunfig.toml` (test configuration)

```toml
[test]
# Timeout per test in milliseconds (increase for integration tests with containers)
timeout = 60000
# Run tests in parallel within files; sequential across files by default
```

---

## Key rules illustrated here

- In-memory repositories implement the same `interface` as the real adapter — structural typing, no inheritance
- Unit tests use `describe`/`it` from `bun:test` — no external test framework
- Integration tests use `@testcontainers/postgresql` for a real DB — no mocking of SQL
- Inbound adapter integration tests call server function handlers directly — no real HTTP server needed
- `beforeEach` cleans tables between tests — prevents state bleed
- `beforeAll`/`afterAll` manage the container lifecycle — started once per describe block
