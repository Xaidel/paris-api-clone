# TypeScript — Domain Layer Examples

Real TypeScript code illustrating `src/domain/` patterns.
All examples use the Order / Payment domain.

---

## Errors (`src/domain/errors.ts`)

```typescript
export class DomainError extends Error {
  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
    // Maintains proper prototype chain in transpiled ES5
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

export class InvalidOrderItemsError extends DomainError {
  readonly kind = "invalid-order-items" as const;
  constructor() {
    super("Order must have at least one item");
  }
}

export class InvalidOrderStateError extends DomainError {
  readonly kind = "invalid-order-state" as const;
  constructor(current: string, attempted: string) {
    super(`Cannot transition from ${current} to ${attempted}`);
  }
}

export class OrderNotFoundError extends DomainError {
  readonly kind = "order-not-found" as const;
  constructor(orderId: string) {
    super(`Order ${orderId} not found`);
  }
}

export class CurrencyMismatchError extends DomainError {
  readonly kind = "currency-mismatch" as const;
  constructor(left: string, right: string) {
    super(`Currency mismatch: ${left} vs ${right}`);
  }
}

export class InvalidMoneyError extends DomainError {
  readonly kind = "invalid-money" as const;
}
```

---

## Value Objects

### `src/domain/value-objects/order-id.ts`

```typescript
import { randomUUID } from "crypto";
import { DomainError } from "../errors";

export class OrderId {
  readonly value: string;

  private constructor(value: string) {
    this.value = value;
  }

  static generate(): OrderId {
    return new OrderId(randomUUID());
  }

  static fromString(raw: string): OrderId {
    if (!raw || raw.trim().length === 0) {
      throw new DomainError("OrderId cannot be empty");
    }
    // Basic UUID format validation
    const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
    if (!uuidRegex.test(raw)) {
      throw new DomainError(`Invalid OrderId format: ${raw}`);
    }
    return new OrderId(raw);
  }

  equals(other: OrderId): boolean {
    return this.value === other.value;
  }

  toString(): string {
    return this.value;
  }
}
```

### `src/domain/value-objects/money.ts`

```typescript
import Decimal from "decimal.js";
import { CurrencyMismatchError, InvalidMoneyError } from "../errors";

export class Money {
  readonly amount: Decimal;
  readonly currency: string;

  private constructor(amount: Decimal, currency: string) {
    this.amount = amount;
    this.currency = currency;
  }

  static create(amount: number | string | Decimal, currency: string): Money {
    const d = new Decimal(amount);
    if (d.isNegative()) throw new InvalidMoneyError("Amount cannot be negative");
    if (!currency || currency.trim().length === 0)
      throw new InvalidMoneyError("Currency is required");
    return new Money(d, currency.toUpperCase());
  }

  static zero(currency: string): Money {
    return new Money(new Decimal(0), currency.toUpperCase());
  }

  add(other: Money): Money {
    if (this.currency !== other.currency)
      throw new CurrencyMismatchError(this.currency, other.currency);
    return new Money(this.amount.plus(other.amount), this.currency);
  }

  multiply(factor: number): Money {
    return new Money(this.amount.times(factor), this.currency);
  }

  equals(other: Money): boolean {
    return this.amount.equals(other.amount) && this.currency === other.currency;
  }

  toString(): string {
    return `${this.amount.toFixed(2)} ${this.currency}`;
  }
}
```

### `src/domain/value-objects/user-id.ts`

```typescript
import { randomUUID } from "crypto";
import { DomainError } from "../errors";

export class UserId {
  readonly value: string;

  private constructor(value: string) {
    this.value = value;
  }

  static generate(): UserId {
    return new UserId(randomUUID());
  }

  static fromString(raw: string): UserId {
    if (!raw || raw.trim().length === 0) {
      throw new DomainError("UserId cannot be empty");
    }
    return new UserId(raw);
  }

  equals(other: UserId): boolean {
    return this.value === other.value;
  }

  toString(): string {
    return this.value;
  }
}
```

---

## Domain Events (`src/domain/events/`)

### `src/domain/events/domain-event.ts`

```typescript
import { randomUUID } from "crypto";

export abstract class DomainEvent {
  readonly eventId: string;
  readonly occurredAt: Date;
  abstract readonly eventType: string;

  constructor() {
    this.eventId = randomUUID();
    this.occurredAt = new Date();
  }
}
```

### `src/domain/events/order-placed.ts`

```typescript
import { DomainEvent } from "./domain-event";
import type { OrderId } from "../value-objects/order-id";
import type { UserId } from "../value-objects/user-id";
import type { Money } from "../value-objects/money";

export class OrderPlaced extends DomainEvent {
  readonly eventType = "order.placed" as const;

  constructor(
    readonly orderId: OrderId,
    readonly customerId: UserId,
    readonly total: Money,
  ) {
    super();
  }
}
```

---

## Entity (`src/domain/entities/order.ts`)

```typescript
import { InvalidOrderItemsError, InvalidOrderStateError } from "../errors";
import type { DomainEvent } from "../events/domain-event";
import { OrderPlaced } from "../events/order-placed";
import { OrderConfirmed } from "../events/order-confirmed";
import { OrderId } from "../value-objects/order-id";
import { UserId } from "../value-objects/user-id";
import { Money } from "../value-objects/money";

export type OrderStatus = "PENDING" | "CONFIRMED" | "CANCELLED";

export interface OrderItem {
  readonly productId: string;
  readonly quantity: number;
  readonly unitPrice: Money;
  readonly subtotal: Money;
}

function makeOrderItem(
  productId: string,
  quantity: number,
  unitPrice: Money,
): OrderItem {
  return {
    productId,
    quantity,
    unitPrice,
    get subtotal() {
      return unitPrice.multiply(quantity);
    },
  };
}

export class Order {
  readonly id: OrderId;
  readonly customerId: UserId;
  private _items: OrderItem[];
  private _total: Money;
  private _status: OrderStatus;
  private _events: DomainEvent[];

  private constructor(
    id: OrderId,
    customerId: UserId,
    items: OrderItem[],
    total: Money,
    status: OrderStatus,
    events: DomainEvent[],
  ) {
    this.id = id;
    this.customerId = customerId;
    this._items = items;
    this._total = total;
    this._status = status;
    this._events = events;
  }

  // --- Factories ---

  static create(customerId: UserId, items: OrderItem[]): Order {
    if (items.length === 0) throw new InvalidOrderItemsError();
    const total = items.reduce(
      (acc, item) => acc.add(item.subtotal),
      Money.zero("USD"),
    );
    const id = OrderId.generate();
    const order = new Order(id, customerId, [...items], total, "PENDING", []);
    order._events.push(new OrderPlaced(id, customerId, total));
    return order;
  }

  static reconstitute(
    id: OrderId,
    customerId: UserId,
    items: OrderItem[],
    total: Money,
    status: OrderStatus,
  ): Order {
    return new Order(id, customerId, [...items], total, status, []);
  }

  // --- Properties ---

  get items(): readonly OrderItem[] { return this._items; }
  get total(): Money               { return this._total; }
  get status(): OrderStatus        { return this._status; }

  // --- State transitions ---

  confirm(): void {
    if (this._status !== "PENDING")
      throw new InvalidOrderStateError(this._status, "CONFIRMED");
    this._status = "CONFIRMED";
    this._events.push(new OrderConfirmed(this.id));
  }

  cancel(reason: string): void {
    if (this._status === "CANCELLED")
      throw new InvalidOrderStateError(this._status, "CANCELLED");
    this._status = "CANCELLED";
  }

  // --- Event collection ---

  drainEvents(): DomainEvent[] {
    const events = [...this._events];
    this._events = [];
    return events;
  }

  equals(other: Order): boolean {
    return this.id.equals(other.id);
  }
}

// Re-export factory helper so adapters can construct OrderItems
export { makeOrderItem };
```

---

## Key rules illustrated here

- `private constructor` on all value objects and entities — instantiation via static factories only
- `readonly` fields on all value objects — no mutation after construction
- `decimal.js` for money — never `number`
- `DomainError` subclasses have `readonly kind` discriminants for type narrowing
- `Order.create()` enforces invariants and emits events; `Order.reconstitute()` does neither
- `drainEvents()` returns and clears — the use case calls this after `save()`
- No framework imports anywhere in this directory
