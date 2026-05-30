# TypeScript — Ports Layer Examples

Real TypeScript code illustrating `src/ports/` patterns using native `interface`.

---

## Inbound Ports (`src/ports/inbound/`)

### `src/ports/inbound/place-order-port.ts`

```typescript
// Command — primitives in; no domain types leak into the inbound port
export interface PlaceOrderItemInput {
  readonly productId: string;
  readonly quantity: number;
  readonly unitPriceAmount: string; // string for Decimal precision
  readonly unitPriceCurrency: string;
}

export interface PlaceOrderCommand {
  readonly customerId: string;
  readonly items: PlaceOrderItemInput[];
}

export interface PlaceOrderResult {
  readonly orderId: string;
  readonly status: string;
  readonly totalAmount: string;
  readonly totalCurrency: string;
}

export interface PlaceOrderPort {
  execute(command: PlaceOrderCommand): Promise<PlaceOrderResult>;
}
```

### `src/ports/inbound/get-order-port.ts`

```typescript
export interface GetOrderQuery {
  readonly orderId: string;
  readonly requestingUserId: string;
}

export interface OrderItemView {
  readonly productId: string;
  readonly quantity: number;
  readonly unitPriceAmount: string;
  readonly unitPriceCurrency: string;
  readonly subtotalAmount: string;
}

export interface GetOrderResult {
  readonly orderId: string;
  readonly customerId: string;
  readonly status: string;
  readonly totalAmount: string;
  readonly totalCurrency: string;
  readonly items: OrderItemView[];
}

export interface GetOrderPort {
  execute(query: GetOrderQuery): Promise<GetOrderResult>;
}
```

---

## Outbound Ports (`src/ports/outbound/`)

### `src/ports/outbound/order-repository.ts`

```typescript
import type { Order } from "../../domain/entities/order";
import type { OrderId } from "../../domain/value-objects/order-id";
import type { UserId } from "../../domain/value-objects/user-id";

export interface OrderRepository {
  save(order: Order): Promise<void>;
  findById(orderId: OrderId): Promise<Order | null>;
  findByCustomerId(customerId: UserId): Promise<Order[]>;
}
```

### `src/ports/outbound/payment-gateway.ts`

```typescript
export interface PaymentRequest {
  readonly orderId: string;
  readonly amountAmount: string;
  readonly amountCurrency: string;
  readonly paymentMethodToken: string;
}

export interface PaymentResult {
  readonly success: boolean;
  readonly transactionId: string | null;
  readonly failureReason: string | null;
}

export interface PaymentGateway {
  charge(request: PaymentRequest): Promise<PaymentResult>;
  refund(transactionId: string, amount: string): Promise<PaymentResult>;
}
```

### `src/ports/outbound/event-publisher.ts`

```typescript
import type { DomainEvent } from "../../domain/events/domain-event";

export interface EventPublisher {
  publish(event: DomainEvent): Promise<void>;
  publishAll(events: DomainEvent[]): Promise<void>;
}
```

---

## Key rules illustrated here

- ALL ports are `interface`, never `abstract class` or `class`
- Port interfaces live in `src/ports/` only — never inline them in adapters
- Inbound port commands/results use primitives (`string`, `number`) — no domain types
- Outbound port methods use domain types (`Order`, `OrderId`) — called from the application layer
- All port methods return `Promise<T>` — the stack is async (postgres.js / TanStack Start)
- `type` imports (`import type`) where the value is not used at runtime — keeps bundles clean
