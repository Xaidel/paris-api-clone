# AGENTS.md — Application Layer Contract

You are working in `src/application/`. Read this before writing any code here.

---

## What this layer is

The application layer **orchestrates** the domain. It contains:

- **Use Cases** — named, single-purpose operations representing user or system actions
- **Application Services** — cross-cutting coordination that spans multiple use cases

This layer is the boundary between "what the business does" (domain) and "how
it is triggered or where data comes from" (adapters). It coordinates domain objects
and ports but contains **no business rules itself**.

---

## Allowed imports

| Allowed | Why |
|---|---|
| `src/domain/` | Entities, value objects, events, domain services |
| `src/ports/inbound/` | This layer implements inbound port interfaces |
| `src/ports/outbound/` | This layer calls outbound port interfaces (never implementations) |
| Standard library | Exceptions, collections, async primitives |

---

## Forbidden imports — hard rules

| Forbidden | Why |
|---|---|
| HTTP, gRPC, WebSocket, CLI libraries | Application layer must not know the delivery mechanism |
| Database drivers, ORMs, SQL | Application layer must not know the storage mechanism |
| `src/adapters/` | Adapters depend on application, not the other way around |
| `src/infrastructure/` | Infrastructure depends on application, not the other way around |
| Any framework (Flask, Gin, Express, etc.) | Application layer is framework-free |

---

## What belongs here

### Use Cases (`use-cases/`)

Each use case is a **single, named action** that a user or system can perform.

- Accepts a command or query object as input
- Calls domain objects (entities, value objects, domain services)
- Calls outbound ports (repository, gateway, event publisher) — never adapters
- Returns a result or raises an application error
- Has **no business logic** — it delegates to the domain
- One use case = one file = one class/function

**Command use cases**: mutate state (`PlaceOrderUseCase`, `CancelOrderUseCase`)
**Query use cases**: read state (`GetOrderUseCase`, `ListOrdersUseCase`)

### Application Services (`services/`)

Application services coordinate concerns that span multiple use cases:

- Cross-cutting logic like notification dispatch after multiple actions
- Workflow orchestration across bounded contexts
- Only use when logic genuinely doesn't fit a single use case

---

## Naming conventions

| Concept | Pattern | Examples |
|---|---|---|
| Command use case | `{Verb}{Noun}UseCase` | `PlaceOrderUseCase`, `CancelOrderUseCase` |
| Query use case | `Get{Noun}UseCase`, `List{Noun}UseCase` | `GetOrderUseCase`, `ListOrdersUseCase` |
| Command input | `{Verb}{Noun}Command` | `PlaceOrderCommand`, `CancelOrderCommand` |
| Query input | `{Noun}Query` | `OrderQuery`, `UserSearchQuery` |
| Result | `{Noun}Result` | `PlaceOrderResult`, `OrderResult` |
| Application service | `{Noun}Service` | `NotificationService` |

---

## Use case structure rule

Each use case must follow this structure:

```
// Input
Command {Verb}{Noun}Command {
  // validated primitive fields — no domain objects
  // (domain objects are constructed inside the use case)
}

// Output
Result {Verb}{Noun}Result {
  // only primitives or simple DTOs — no domain objects leaked outward
}

// Use case
UseCase {Verb}{Noun}UseCase {
  // inject outbound ports in constructor, never adapters or infrastructure

  execute(command: {Verb}{Noun}Command) → {Verb}{Noun}Result:
    // 1. Load domain objects via outbound ports
    // 2. Call domain logic (entities, domain services)
    // 3. Persist via outbound ports
    // 4. Publish domain events via outbound ports
    // 5. Return result
}
```

---

## Application errors vs domain errors

- **Domain errors** — raised by domain objects when invariants are violated
  (e.g. `DomainError("Order must have at least one item")`)
- **Application errors** — raised by use cases for orchestration failures
  (e.g. `NotFoundError("Order not found")`, `ConflictError("Order already confirmed")`)
- **Never** raise HTTP errors (404, 409, 500) from this layer — adapters map errors to transport codes

---

## Self-audit checklist for this layer

Before submitting code in `src/application/`:

- [ ] No imports from `adapters/` or `infrastructure/`
- [ ] No transport library imports (HTTP, gRPC, WS, CLI parsers)
- [ ] No database drivers, ORMs, or SQL
- [ ] No business logic — only orchestration; domain objects enforce rules
- [ ] Use cases are named `{Verb}{Noun}UseCase`
- [ ] Each use case has exactly one `execute()` method
- [ ] Command/query inputs contain primitives, not domain objects
- [ ] Results contain primitives or DTOs, not domain objects
- [ ] Outbound ports are injected in constructor, not instantiated inside use cases
- [ ] Domain events are published through an outbound port (`EventPublisher`), not dispatched directly

---

## See also

- [`use-cases/EXAMPLE.md`](use-cases/EXAMPLE.md)
- [`services/EXAMPLE.md`](services/EXAMPLE.md)
- Root [`AGENTS.md`](../../AGENTS.md) for the full architecture contract
