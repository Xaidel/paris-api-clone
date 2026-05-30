# AGENTS.md — Domain Layer Contract

You are working in `src/domain/`. Read this before writing any code here.

---

## What this layer is

The domain layer is the **heart of the application**. It contains:

- **Entities** — objects with identity that change over time
- **Value Objects** — immutable descriptive values with no identity
- **Domain Events** — facts about things that happened
- **Domain Services** — business logic that doesn't belong to a single entity

This layer expresses the business in code. It has **zero dependencies** on any
other layer, any framework, or any infrastructure.

---

## Allowed imports

| Allowed | Examples |
|---|---|
| Standard library primitives | strings, numbers, dates, collections, errors |
| Other domain types | entities, value objects, events within `src/domain/` |
| Nothing else | — |

---

## Forbidden imports — hard rules

| Forbidden | Why |
|---|---|
| HTTP, gRPC, WebSocket, or any transport library | Business logic must not know how it is delivered |
| Database drivers, ORMs, query builders | Business logic must not know how it is stored |
| Application frameworks (Flask, Express, Gin, etc.) | Domain is framework-free |
| `src/ports/` | Ports depend on domain, not the other way around |
| `src/adapters/` | Strictly forbidden — violates dependency rule |
| `src/infrastructure/` | Strictly forbidden — violates dependency rule |
| `src/application/` | Application depends on domain, not the other way around |
| Any I/O operation | No file reads, network calls, logging, or side effects |

---

## What belongs here

### Entities (`entities/`)
- Has a unique identity (e.g. `OrderId`, `UserId`)
- Identity persists even when attributes change
- Enforces invariants — invalid state must be impossible to construct
- Emits domain events when significant things happen
- Contains behavior, not just data

### Value Objects (`value-objects/`)
- No identity — equality is determined by value, not by reference
- Immutable — never mutated after construction
- Self-validating — invalid values cannot be constructed
- Examples: `Money`, `Email`, `PhoneNumber`, `Address`, `OrderId`

### Domain Events (`events/`)
- Named in past tense: something that **happened**
- Immutable — events are facts, they cannot be changed
- Carry only the data needed to describe what occurred
- Examples: `OrderPlaced`, `PaymentFailed`, `UserRegistered`

### Domain Services (`services/`)
- Stateless operations that coordinate multiple entities or value objects
- Use only when the logic does not naturally belong to a single entity
- Never perform I/O — accept what they need as parameters
- Examples: `PricingService`, `TaxCalculator`, `InventoryAllocator`

---

## Naming conventions

| Concept | Pattern | Examples |
|---|---|---|
| Entity | `{Noun}` | `Order`, `User`, `Invoice`, `Shipment` |
| Value Object | `{Noun}` | `Money`, `Email`, `OrderId`, `Address` |
| Domain Event | `{Noun}{PastVerb}` | `OrderPlaced`, `PaymentFailed`, `UserDeactivated` |
| Domain Service | `{Noun}Service` or `{Noun}{Verb}er` | `PricingService`, `TaxCalculator` |

---

## Invariant enforcement rule

**An entity must never be in an invalid state.**

- Validate in the constructor/factory — reject invalid input immediately
- Use value objects to encapsulate validation (e.g. `Email` validates format)
- Raise a domain error (not an HTTP error) when an invariant is violated
- Domain errors are plain error types — no HTTP status codes, no JSON shapes

---

## Event emission rule

- Entities collect events internally and expose them for the application layer to dispatch
- Events are **not dispatched from within the domain** — that would require I/O
- Pattern: entity accumulates events in a list; use case reads and publishes them

---

## Self-audit checklist for this layer

Before submitting code in `src/domain/`:

- [ ] No imports from `ports/`, `adapters/`, `application/`, or `infrastructure/`
- [ ] No framework imports
- [ ] No database or transport library imports
- [ ] No I/O operations (no logging, no network, no file access)
- [ ] Entities enforce their own invariants — invalid state is unrepresentable
- [ ] Value objects are immutable and self-validating
- [ ] Domain events are named in past tense
- [ ] Domain services are stateless and accept all inputs as parameters
- [ ] Error types are domain errors, not HTTP errors or framework exceptions

---

## See also

- [`entities/EXAMPLE.md`](entities/EXAMPLE.md)
- [`value-objects/EXAMPLE.md`](value-objects/EXAMPLE.md)
- [`events/EXAMPLE.md`](events/EXAMPLE.md)
- [`services/EXAMPLE.md`](services/EXAMPLE.md)
- Root [`AGENTS.md`](../../AGENTS.md) for the full architecture contract
