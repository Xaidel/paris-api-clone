# AGENTS.md — Ports Layer Contract

You are working in `src/ports/`. Read this before writing any code here.

---

## What this layer is

Ports are **contracts** — interface definitions that describe what the application
can do (inbound) and what it needs (outbound). They contain **no implementation**.

- **Inbound ports** (`ports/inbound/`) — what the application offers to the outside world
- **Outbound ports** (`ports/outbound/`) — what the application needs from the outside world

Ports are the formal boundary of the application core. They decouple the application
from both its delivery mechanism and its infrastructure.

---

## Allowed imports

| Allowed | Why |
|---|---|
| `src/domain/` | Ports reference domain types (entities, value objects, events) |
| Standard library | Interface primitives, collections, error types |

---

## Forbidden imports — hard rules

| Forbidden | Why |
|---|---|
| `src/adapters/` | Adapters implement ports; ports must not depend on adapters |
| `src/infrastructure/` | Infrastructure depends on ports, not the other way around |
| `src/application/` | Application implements/uses ports; ports must not depend on application |
| HTTP, gRPC, WS, or any transport library | Ports are transport-agnostic |
| Database drivers, ORMs, query builders | Ports are storage-agnostic |
| Any framework | Ports are framework-free |

---

## What belongs here

### Inbound ports (`inbound/`)

An inbound port is the **interface that the application exposes** to the outside world.
Inbound adapters (HTTP handlers, gRPC handlers, queue consumers) call inbound ports.
Application use cases **implement** inbound ports.

```
// Example: the contract the application offers for placing orders
InboundPort PlaceOrderPort {
  execute(command: PlaceOrderCommand) → PlaceOrderResult
}
```

### Outbound ports (`outbound/`)

An outbound port is the **interface that the application needs** from the outside world.
Outbound adapters (database repositories, API gateways, caches) **implement** outbound ports.
Application use cases call outbound ports.

```
// Example: the contract the application needs for order persistence
OutboundPort OrderRepository {
  save(order: Order) → void
  findById(id: OrderId) → Order | null
  findByCustomerId(customerId: UserId) → List<Order>
}
```

---

## Naming conventions

| Concept | Pattern | Examples |
|---|---|---|
| Inbound port | `{Verb}{Noun}Port` | `PlaceOrderPort`, `AuthenticateUserPort`, `StreamEventsPort` |
| Repository port | `{Noun}Repository` | `OrderRepository`, `UserRepository`, `ProductRepository` |
| Gateway port | `{Noun}Gateway` | `PaymentGateway`, `EmailGateway`, `SMSGateway` |
| Store / Cache port | `{Noun}Store` | `SessionStore`, `CacheStore` |
| Event publisher port | `EventPublisher` | `EventPublisher` |
| Notification port | `{Noun}Gateway` or `{Noun}Notifier` | `NotificationGateway` |

---

## The port-first rule

**Define the port before writing the use case or the adapter.**

Why: defining the port forces you to think about the contract — what the application
needs — before thinking about how to implement it. This prevents implementation
details from leaking into the application layer.

Order of work:
1. Define the outbound port (`ports/outbound/`)
2. Write the use case against the port (`application/use-cases/`)
3. Implement the adapter (`adapters/outbound/`)
4. Wire in the DI container (`infrastructure/di/`)

---

## Self-audit checklist for this layer

Before submitting code in `src/ports/`:

- [ ] Port files contain only interface/contract definitions — zero implementation code
- [ ] No imports from `adapters/` or `infrastructure/`
- [ ] No transport-specific types (no HTTP Request/Response, no gRPC types)
- [ ] No database-specific types (no SQL Row, no ORM Model)
- [ ] Inbound ports are named `{Verb}{Noun}Port`
- [ ] Repository ports are named `{Noun}Repository`
- [ ] Gateway ports are named `{Noun}Gateway`
- [ ] Port method signatures use only domain types or primitives

---

## See also

- [`inbound/EXAMPLE.md`](inbound/EXAMPLE.md)
- [`outbound/EXAMPLE.md`](outbound/EXAMPLE.md)
- Root [`AGENTS.md`](../../AGENTS.md) for the full architecture contract
