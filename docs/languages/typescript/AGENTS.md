# AGENTS.md — TypeScript Language Supplement

Load this after `AGENTS.md` (root) and `src/{layer}/AGENTS.md`.
This file is authoritative for TypeScript idioms and tooling. It never overrides
architecture rules — if a conflict exists, the architecture rule wins.

**Stack**: TanStack Start · postgres.js · bun test · TSyringe (complex DI only)

---

## Barrel Files

Each layer directory exports a public `index.ts` barrel file. Other layers import from the barrel, never from internal module files directly. This enforces encapsulation at the module boundary.

```typescript
// src/domain/index.ts — domain public API
export { Order } from "./entities/order";
export { OrderId } from "./value-objects/order-id";
export { Money } from "./value-objects/money";
export { UserId } from "./value-objects/user-id";
export { OrderStatus } from "./entities/order";
export { DomainError, OrderNotFoundError, InvalidOrderItemsError } from "./errors";
```

```typescript
// Correct — import from the barrel
import { Order, OrderId, DomainError } from "@domain";
import { OrderRepository } from "@ports/outbound";

// WRONG — importing from an internal module file
import { Order } from "@domain/entities/order";
import { OrderRepository } from "@ports/outbound/order-repository";
```

**Rules:**
- Every layer directory (`src/domain/`, `src/ports/inbound/`, `src/ports/outbound/`, `src/adapters/inbound/`, `src/adapters/outbound/`, `src/application/use-cases/`, `src/infrastructure/`) has an `index.ts`
- `index.ts` re-exports only what other layers need — keep internal types unexported
- Never `export * from "./..."` — explicit named exports only; wildcard re-exports defeat encapsulation

---

## Path Aliases

Configure `tsconfig.json` path aliases to avoid `../../` relative import chains across layers.

**Required aliases** — add to `compilerOptions` in `tsconfig.json`:

```json
{
  "compilerOptions": {
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "target": "ES2022",
    "baseUrl": ".",
    "paths": {
      "@domain": ["src/domain/index.ts"],
      "@domain/*": ["src/domain/*"],
      "@ports/*": ["src/ports/*"],
      "@application/*": ["src/application/*"],
      "@adapters/*": ["src/adapters/*"],
      "@infrastructure/*": ["src/infrastructure/*"]
    }
  }
}
```

The bundler (Vite / Bun) must also resolve these aliases. For Vite:

```typescript
// vite.config.ts
import { resolve } from "path";

export default {
  resolve: {
    alias: {
      "@domain": resolve(__dirname, "src/domain/index.ts"),
      "@ports": resolve(__dirname, "src/ports"),
      "@application": resolve(__dirname, "src/application"),
      "@adapters": resolve(__dirname, "src/adapters"),
      "@infrastructure": resolve(__dirname, "src/infrastructure"),
    },
  },
};
```

**Rules:**
- All cross-layer imports MUST use path aliases, never relative `../../` paths
- Path aliases are configured in `tsconfig.json` AND the bundler config — both are required
- Domain may not import from any other layer even via alias — aliases do not relax the dependency rule

---

## 1. Port / Interface Mechanism

Use TypeScript `interface` for all port definitions. Never use `abstract class`.

```typescript
// src/ports/outbound/order-repository.ts
export interface OrderRepository {
  save(order: Order): Promise<void>;
  findById(orderId: OrderId): Promise<Order | null>;
  findByCustomerId(customerId: UserId): Promise<Order[]>;
}
```

**Rules:**
- ALL ports (inbound and outbound) MUST be `interface`, never `abstract class`
- Port interfaces live in `src/ports/` only — never inline them in adapters or use cases
- Concrete adapter classes implement the interface with `implements OrderRepository` — this is checked by the compiler
- Inbound ports may also be expressed as interfaces that use cases implement directly
- Never expose framework types (e.g. `Request`, `Response`) in port interfaces

---

## 2. Value Object Immutability

Use a `class` with `private constructor`, `readonly` fields, and a static `create()` factory.

```typescript
// src/domain/value-objects/money.ts
import Decimal from "decimal.js";

export class Money {
  readonly amount: Decimal;
  readonly currency: string;

  private constructor(amount: Decimal, currency: string) {
    this.amount = amount;
    this.currency = currency;
  }

  static create(amount: number | string, currency: string): Money {
    const d = new Decimal(amount);
    if (d.isNegative()) throw new InvalidMoneyError("Amount cannot be negative");
    if (!currency) throw new InvalidMoneyError("Currency is required");
    return new Money(d, currency);
  }

  static zero(currency: string): Money {
    return new Money(new Decimal(0), currency);
  }

  add(other: Money): Money {
    if (this.currency !== other.currency)
      throw new CurrencyMismatchError(this.currency, other.currency);
    return new Money(this.amount.plus(other.amount), this.currency);
  }

  equals(other: Money): boolean {
    return this.amount.equals(other.amount) && this.currency === other.currency;
  }
}
```

**Rules:**
- `private constructor` prevents direct instantiation — use static factory methods
- All fields are `readonly` — no mutation after construction
- Use `decimal.js` for monetary values — never `number` for money
- Provide `equals()` — do not rely on reference equality (`===`)
- Never use plain `type` aliases for value objects that need validation logic; use `class`

---

## 3. Entity Identity and Equality

Entities have a typed ID value object. Equality is by ID only.

```typescript
// src/domain/value-objects/order-id.ts
import { randomUUID } from "crypto";

export class OrderId {
  readonly value: string;

  private constructor(value: string) {
    this.value = value;
  }

  static generate(): OrderId {
    return new OrderId(randomUUID());
  }

  static fromString(raw: string): OrderId {
    if (!raw) throw new DomainError("OrderId cannot be empty");
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

Entity equality — implement an `equals()` method comparing by ID:

```typescript
export class Order {
  constructor(readonly id: OrderId, ...) {}

  equals(other: Order): boolean {
    return this.id.equals(other.id);
  }
}
```

TypeScript does not support operator overloading; always use `.equals()`, never `===` for entities.

---

## 4. Domain Error Types

Use a class hierarchy with a `readonly kind` discriminant for exhaustive `switch` narrowing.

```typescript
// src/domain/errors.ts
export class DomainError extends Error {
  constructor(message: string) {
    super(message);
    this.name = this.constructor.name;
  }
}

export class InvalidOrderItemsError extends DomainError {
  readonly kind = "invalid-order-items" as const;
  constructor() { super("Order must have at least one item"); }
}

export class OrderNotFoundError extends DomainError {
  readonly kind = "order-not-found" as const;
  constructor(orderId: string) { super(`Order ${orderId} not found`); }
}

export class InvalidOrderStateError extends DomainError {
  readonly kind = "invalid-order-state" as const;
  constructor(current: string, attempted: string) {
    super(`Cannot transition from ${current} to ${attempted}`);
  }
}
```

The `kind` discriminant enables type-safe error mapping in adapters:

```typescript
} catch (err) {
  if (err instanceof OrderNotFoundError) return notFound(err.message);
  if (err instanceof InvalidOrderItemsError) return badRequest(err.message);
  throw err; // re-throw unexpected errors
}
```

---

## 5. Null / Absence Handling

- Use `T | null` at all domain and port boundaries — never `undefined`
- Enable `strict: true` and `exactOptionalPropertyTypes: true` in tsconfig
- Return `null` from repository `findById()` when not found — do not throw
- The use case raises `OrderNotFoundError` after receiving `null`

```typescript
// Port: returns null, not undefined, not throws
findById(orderId: OrderId): Promise<Order | null>;

// Use case: raises after null check
const order = await this.orderRepo.findById(command.orderId);
if (order === null) throw new OrderNotFoundError(command.orderId.toString());
```

---

## 6. Error Propagation

TypeScript uses `throw` / `try-catch`. Do not use `Result` types in domain or application.

```
domain          → throws DomainError subclasses
application     → throws DomainError (re-thrown) or ApplicationError
inbound adapter → catch DomainError → HTTP 422 / 404 / etc.
                  catch ApplicationError → HTTP 400 / 403 / etc.
                  catch unknown → HTTP 500
```

If the team adopts `neverthrow`, it is permitted in adapters only — never in domain or application.

---

## 7. Reconstitute Pattern

```typescript
export class Order {
  private constructor(
    readonly id: OrderId,
    readonly customerId: UserId,
    private _items: OrderItem[],
    private _total: Money,
    private _status: OrderStatus,
    private _events: DomainEvent[],
  ) {}

  // Creates a new order — enforces invariants, emits OrderPlaced
  static create(customerId: UserId, items: OrderItem[]): Order {
    if (items.length === 0) throw new InvalidOrderItemsError();
    const total = items.reduce((acc, i) => acc.add(i.subtotal), Money.zero("USD"));
    const order = new Order(OrderId.generate(), customerId, items, total, "PENDING", []);
    order._events.push(new OrderPlaced(order.id, customerId, total));
    return order;
  }

  // Hydrates from storage — skips invariant re-validation, emits no events
  static reconstitute(
    id: OrderId,
    customerId: UserId,
    items: OrderItem[],
    total: Money,
    status: OrderStatus,
  ): Order {
    return new Order(id, customerId, items, total, status, []);
  }
}
```

---

## 8. Async Conventions

TanStack Start server functions and postgres.js are async. Use `async/await` throughout.

- All outbound port methods return `Promise<T>` — they perform I/O
- Inbound port methods may be sync or async depending on the use case
- Use case `execute()` methods are `async` when they call outbound ports
- TanStack Start server functions (`createServerFn`) are always `async`
- Never mix `.then().catch()` chains with `async/await` in the same function

---

## 9. Dependency Injection

### Simple case — manual constructor wiring

```typescript
// src/infrastructure/di/wire.ts
export async function wire(config: AppConfig) {
  const sql = postgres(config.database.url);

  const orderRepo = new PostgresOrderRepository(sql);
  const paymentGateway = new StripePaymentGateway(config.stripe.secretKey);
  const eventPublisher = new KafkaEventPublisher(config.kafka.brokers);

  const placeOrderUseCase = new PlaceOrderUseCase(orderRepo, paymentGateway, eventPublisher);
  const getOrderUseCase = new GetOrderUseCase(orderRepo);

  return { placeOrderUseCase, getOrderUseCase };
}
```

### Complex case — TSyringe

Use `tsyringe` when manual wiring exceeds ~100 lines or lifecycle management is needed.

```typescript
import { container, injectable, inject } from "tsyringe";

@injectable()
class PostgresOrderRepository implements OrderRepository {
  constructor(@inject("SqlClient") private sql: Sql) {}
  // ...
}

container.register("SqlClient", { useValue: postgres(config.database.url) });
container.register("OrderRepository", { useClass: PostgresOrderRepository });

const repo = container.resolve<OrderRepository>("OrderRepository");
```

Do NOT use `@inject` decorators in domain or application layer classes.

---

## 10. File and Package Naming

| Concept | File name | Export name |
|---|---|---|
| Entity | `order.ts` | `Order` |
| Value object | `money.ts`, `order-id.ts` | `Money`, `OrderId` |
| Domain event | `order-placed.ts` | `OrderPlaced` |
| Domain error | `errors.ts` (all domain errors) | `DomainError`, `OrderNotFoundError` |
| Use case | `place-order-use-case.ts` | `PlaceOrderUseCase` |
| Inbound port | `place-order-port.ts` | `PlaceOrderPort` |
| Outbound port | `order-repository.ts` | `OrderRepository` |
| Inbound adapter | `http-order-adapter.ts` | route handler functions |
| Outbound adapter | `postgres-order-repository.ts` | `PostgresOrderRepository` |
| Config | `config.ts` | `AppConfig`, `loadConfig` |

**Rules:**
- `kebab-case` for all file names
- `PascalCase` for all class, interface, and type names
- `camelCase` for all functions and variables
- `SCREAMING_SNAKE_CASE` for constants
- One primary export per file (supporting types may coexist)
- Use ESM — `import`/`export` only, never `require()`

---

## 11. TanStack Start Integration

TanStack Start server functions (`createServerFn`) are **inbound adapters** and live in `src/adapters/inbound/`. TanStack route files (in `routes/`) are thin shells that import and re-expose those adapters — they are not adapters themselves.

```typescript
// src/adapters/inbound/order-server-fns.ts  ← ADAPTER (owns the logic)
import { createServerFn } from "@tanstack/start";
import { z } from "zod";

const placeOrderSchema = z.object({
  customerId: z.string().min(1),
  items: z.array(orderItemSchema).min(1),
});

export const placeOrderFn = createServerFn({ method: "POST" })
  .validator(placeOrderSchema)
  .handler(async ({ data }) => {
    const result = await placeOrderPort.execute(toCommand(data));
    return toResponse(result);
  });
```

```typescript
// routes/orders/index.ts  ← ROUTE FILE (thin shell, no logic)
import { placeOrderFn } from "@adapters/inbound/order-server-fns";
export { placeOrderFn };
```

**Rules:**
- Server functions call port interfaces — never import use case classes directly
- Use `zod` for input validation in server functions (schema lives in the adapter)
- Map `DomainError` subclasses to appropriate HTTP error responses using TanStack's error utilities
- Zod schemas are adapter-layer concerns — do not leak them into domain or application

---

## 12. Forbidden Patterns

| Pattern | Why forbidden |
|---|---|
| `abstract class` for ports | Use `interface` — abstract class implies inheritance, interface is structural |
| `number` for monetary values | Use `decimal.js` — floating point loses precision |
| `undefined` at domain/port boundaries | Use `T \| null` — undefined is ambiguous |
| `any` anywhere | Breaks type safety entirely — use `unknown` and narrow |
| `!` non-null assertion in domain/application | Fix the type instead of asserting |
| `neverthrow` in domain/application | Permitted in adapters only |
| `require()` | Use ESM `import` — this codebase is ESM-only |
| `import ... from "../../adapters/..."` in application | Application never imports from adapters |
| Relative `../../` cross-layer imports | Use path aliases (`@domain`, `@ports/*`, etc.) |
| `export * from "./..."` in barrel files | Use explicit named exports — wildcard exports defeat encapsulation |
| Inline SQL in use cases | SQL belongs in outbound adapters only |

---

## 13. Tooling

| Tool | Purpose | Minimum config |
|---|---|---|
| `tsc` | Type checking | `strict: true`, `exactOptionalPropertyTypes: true` |
| `eslint` + `@typescript-eslint` | Linting | `@typescript-eslint/strict` ruleset |
| `prettier` | Formatting | default config, run via `eslint-plugin-prettier` |
| `bun test` | Test runner | built-in, no config file needed |
| `@testcontainers/postgresql` | Real DB in integration tests | — |

Minimum `tsconfig.json`:
```json
{
  "compilerOptions": {
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "target": "ES2022",
    "baseUrl": ".",
    "paths": {
      "@domain": ["src/domain/index.ts"],
      "@domain/*": ["src/domain/*"],
      "@ports/*": ["src/ports/*"],
      "@application/*": ["src/application/*"],
      "@adapters/*": ["src/adapters/*"],
      "@infrastructure/*": ["src/infrastructure/*"]
    }
  }
}
```

> **Note**: `baseUrl` and `paths` are required — without them all cross-layer imports via `@domain`, `@ports/*`, etc. will fail at compile time. The bundler alias config (see the Path Aliases section above) must match these entries.

---

## 14. Self-Audit Checklist

Before submitting any TypeScript code:

- [ ] All ports are `interface`, not `abstract class`
- [ ] All value object fields are `readonly`; constructed via `static create()`
- [ ] `decimal.js` used for monetary amounts, never `number`
- [ ] `T | null` used at all domain/port boundaries, never `undefined`
- [ ] Domain errors extend `DomainError` with a `readonly kind` discriminant
- [ ] Use case `execute()` returns `Promise<Result>` when it calls outbound ports
- [ ] `Order.reconstitute()` exists and emits no domain events
- [ ] Only inbound adapters contain `try/catch` blocks that map to HTTP errors
- [ ] Server functions are in `src/adapters/inbound/`; route files are thin shells that import them
- [ ] Zod schemas live in the adapter layer, not in domain or application
- [ ] Every layer directory has an `index.ts` barrel file with explicit named exports
- [ ] All cross-layer imports use path aliases (`@domain`, `@ports/*`, etc.), not relative `../../` paths
- [ ] Path aliases are configured in both `tsconfig.json` and the bundler config
- [ ] `tsc --noEmit` passes with no errors
- [ ] `eslint` passes with no errors

---

## See also

- [`examples/domain.md`](examples/domain.md)
- [`examples/application.md`](examples/application.md)
- [`examples/ports.md`](examples/ports.md)
- [`examples/adapters.md`](examples/adapters.md)
- [`examples/infrastructure.md`](examples/infrastructure.md)
- [`examples/tests.md`](examples/tests.md)
- Root [`AGENTS.md`](../../../AGENTS.md)
