# TypeScript — Application Layer Examples

Real TypeScript code illustrating `src/application/` patterns.

---

## Use Case: Place Order (`src/application/use-cases/place-order-use-case.ts`)

```typescript
import { Order, makeOrderItem } from "../../domain/entities/order";
import { UserId } from "../../domain/value-objects/user-id";
import { Money } from "../../domain/value-objects/money";
import type { PlaceOrderPort, PlaceOrderCommand, PlaceOrderResult } from "../../ports/inbound/place-order-port";
import type { OrderRepository } from "../../ports/outbound/order-repository";
import type { EventPublisher } from "../../ports/outbound/event-publisher";

export class PlaceOrderUseCase implements PlaceOrderPort {
  constructor(
    private readonly orderRepo: OrderRepository,
    private readonly eventPublisher: EventPublisher,
  ) {}

  async execute(command: PlaceOrderCommand): Promise<PlaceOrderResult> {
    // 1. Map primitives → domain types
    const customerId = UserId.fromString(command.customerId);
    const items = command.items.map((i) =>
      makeOrderItem(
        i.productId,
        i.quantity,
        Money.create(i.unitPriceAmount, i.unitPriceCurrency),
      ),
    );

    // 2. Create domain entity (invariants enforced inside Order.create())
    const order = Order.create(customerId, items);

    // 3. Persist (port call — no knowledge of Postgres)
    await this.orderRepo.save(order);

    // 4. Publish events (port call — no knowledge of Kafka)
    await this.eventPublisher.publishAll(order.drainEvents());

    // 5. Return primitive result DTO
    return {
      orderId: order.id.toString(),
      status: order.status,
      totalAmount: order.total.amount.toFixed(2),
      totalCurrency: order.total.currency,
    };
  }
}
```

---

## Use Case: Get Order (`src/application/use-cases/get-order-use-case.ts`)

```typescript
import { OrderNotFoundError } from "../../domain/errors";
import { OrderId } from "../../domain/value-objects/order-id";
import type { GetOrderPort, GetOrderQuery, GetOrderResult } from "../../ports/inbound/get-order-port";
import type { OrderRepository } from "../../ports/outbound/order-repository";

export class GetOrderUseCase implements GetOrderPort {
  constructor(private readonly orderRepo: OrderRepository) {}

  async execute(query: GetOrderQuery): Promise<GetOrderResult> {
    // 1. Parse identifier
    const orderId = OrderId.fromString(query.orderId);

    // 2. Load from repository
    const order = await this.orderRepo.findById(orderId);
    if (order === null) throw new OrderNotFoundError(query.orderId);

    // 3. Map domain entity → result DTO
    return {
      orderId: order.id.toString(),
      customerId: order.customerId.toString(),
      status: order.status,
      totalAmount: order.total.amount.toFixed(2),
      totalCurrency: order.total.currency,
      items: order.items.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
        unitPriceAmount: item.unitPrice.amount.toFixed(2),
        unitPriceCurrency: item.unitPrice.currency,
        subtotalAmount: item.subtotal.amount.toFixed(2),
      })),
    };
  }
}
```

---

## Application Service (`src/application/services/notification-service.ts`)

```typescript
import { OrderNotFoundError } from "../../domain/errors";
import type { OrderId } from "../../domain/value-objects/order-id";
import type { OrderRepository } from "../../ports/outbound/order-repository";
import type { NotificationGateway } from "../../ports/outbound/notification-gateway";

export class NotificationService {
  constructor(
    private readonly notificationGateway: NotificationGateway,
    private readonly orderRepo: OrderRepository,
  ) {}

  async notifyOrderPlaced(orderId: OrderId): Promise<void> {
    const order = await this.orderRepo.findById(orderId);
    if (order === null) throw new OrderNotFoundError(orderId.toString());

    await this.notificationGateway.send({
      recipientId: order.customerId.toString(),
      subject: "Your order has been placed",
      body: `Order ${orderId} for ${order.total} is confirmed.`,
    });
  }

  async notifyOrderCancelled(orderId: OrderId, reason: string): Promise<void> {
    const order = await this.orderRepo.findById(orderId);
    if (order === null) throw new OrderNotFoundError(orderId.toString());

    await this.notificationGateway.send({
      recipientId: order.customerId.toString(),
      subject: "Your order has been cancelled",
      body: `Order ${orderId} was cancelled: ${reason}`,
    });
  }
}
```

---

## Key rules illustrated here

- Use cases receive port interfaces in the constructor — never concrete adapter classes
- `PlaceOrderUseCase` declares `implements PlaceOrderPort` — checked by the TypeScript compiler
- `execute()` takes a command interface and returns a result interface (primitives only)
- Domain entity construction happens inside the use case — not in the adapter
- Events are drained from the entity and published after the entity is persisted
- Use cases do NOT catch exceptions — `DomainError` propagates to the adapter
- Application services depend only on port interfaces — never on adapters or infrastructure
