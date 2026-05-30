# AGENTS.md — Hexagonal Architecture Contract

This file is the authoritative contract for AI agents working in this codebase.
Read this before writing, editing, or moving any code. Every layer has its own
`AGENTS.md` with scoped rules — load the one for the layer you are working in.

---

## 1. Architecture Overview

This codebase follows **Hexagonal Architecture** (also called Ports & Adapters).
The core idea: **business logic has zero knowledge of how it is delivered or stored**.

```
                        ┌─────────────────────────────────┐
  [ HTTP ]              │                                 │
  [ gRPC ]  ──inbound──▶│   ports/inbound                 │
  [ WS   ]  adapters    │         │                       │
  [ CLI  ]              │         ▼                       │
                        │   application/use-cases         │
                        │         │                       │
                        │         ▼                       │
                        │      domain/                    │
                        │   (entities, value-objects,     │
                        │    events, domain services)     │
                        │         │                       │
                        │         ▼                       │
                        │   ports/outbound                │
                        │         │                       │              [ Postgres ]
                        └─────────────────────────────────┘              [ Redis   ]
                                  │                                      [ Stripe  ]
                                  └──outbound──▶ adapters/outbound ──▶   [ S3      ]
                                    adapters                             [ Kafka   ]
```

Infrastructure (`src/infrastructure/`) is the **composition root**: it wires all
layers together at startup. It is the only layer that imports from every other layer.

---

## 2. Layer Map — Quick Reference

| Layer | Path | Allowed imports | Forbidden imports |
|---|---|---|---|
| `domain` | `src/domain/` | Nothing (zero dependencies) | All other layers, all frameworks |
| `application` | `src/application/` | `domain`, `ports` | `adapters`, `infrastructure`, frameworks |
| `ports` | `src/ports/` | `domain` | `adapters`, `infrastructure`, frameworks |
| `adapters` | `src/adapters/` | `ports`, `application`, `domain` (read-only value types) | `infrastructure` |
| `infrastructure` | `src/infrastructure/` | Everything | — |

**The Dependency Rule**: source code dependencies point **inward only**.
`infrastructure` → `adapters` → `ports` → `application` → `domain`
The `domain` layer depends on nothing. No exceptions.

---

## 3. Decision Flowchart — Where Does This Code Go?

When you need to create or place a piece of code, follow this flowchart:

```
Is it a business rule, invariant, or core concept?
  └─ YES → Is it tied to a specific entity?
              └─ YES → src/domain/entities/
              └─ NO  → Is it an immutable descriptive value?
                          └─ YES → src/domain/value-objects/
                          └─ NO  → Is it something that happened (past tense)?
                                      └─ YES → src/domain/events/
                                      └─ NO  → src/domain/services/

Is it orchestration (calling domain + ports in sequence)?
  └─ YES → Is it a single, named user action?
              └─ YES → src/application/use-cases/
              └─ NO  → src/application/services/

Is it a contract / interface with no implementation?
  └─ YES → Does it describe what the app OFFERS to callers?
              └─ YES → src/ports/inbound/
              └─ NO  → Does it describe what the app NEEDS from outside?
                          └─ YES → src/ports/outbound/

Is it a concrete implementation of a port?
  └─ YES → Does it receive input FROM outside (HTTP, gRPC, WS, CLI, queue)?
              └─ YES → src/adapters/inbound/
              └─ NO  → Does it call OUT to external systems (DB, queue, API)?
                          └─ YES → src/adapters/outbound/

Is it wiring, configuration, DI, or observability bootstrap?
  └─ YES → src/infrastructure/

Still unsure? → Define a port first. Ports clarify boundaries.
```

---

## 4. Naming Conventions

Consistent naming is critical for agent-to-agent and agent-to-human handoffs.

| Concept | Pattern | Example |
|---|---|---|
| Entity | `{Noun}` | `Order`, `User`, `Invoice` |
| Value Object | `{Noun}` (descriptive) | `Money`, `Email`, `OrderId` |
| Domain Event | `{Noun}{PastVerb}` | `OrderPlaced`, `PaymentFailed` |
| Domain Service | `{Noun}Service` | `PricingService`, `TaxCalculator` |
| Use Case (command) | `{Verb}{Noun}UseCase` | `PlaceOrderUseCase`, `CancelOrderUseCase` |
| Use Case (query) | `Get{Noun}UseCase` | `GetOrderUseCase`, `ListOrdersUseCase` |
| Inbound Port | `{Verb}{Noun}Port` | `PlaceOrderPort`, `AuthenticateUserPort` |
| Outbound Port | `{Noun}Repository` / `{Noun}Gateway` / `{Noun}Store` | `OrderRepository`, `PaymentGateway`, `SessionStore` |
| Inbound Adapter | `{Technology}{Noun}Adapter` | `HttpOrderAdapter`, `GrpcOrderAdapter`, `WsOrderAdapter` |
| Outbound Adapter | `{Technology}{Noun}{PortSuffix}` | `PostgresOrderRepository`, `RedisSessionStore`, `StripePaymentGateway` |
| Application Service | `{Noun}Service` | `NotificationService` |
| Config struct | `{Scope}Config` | `DatabaseConfig`, `ServerConfig`, `AppConfig` |

**File naming**: use the language's idiomatic convention (`snake_case.py`, `camelCase.ts`,
`PascalCase.go`), but keep the conceptual name consistent with the table above.

---

## 5. Forbidden Patterns — Hard Rules

These are **non-negotiable** regardless of language, deadline, or convenience.

### domain/
- NEVER import HTTP, gRPC, WebSocket, or any transport library
- NEVER import a database driver, ORM, or query builder
- NEVER import an application framework (Flask, Express, Gin, Actix, etc.)
- NEVER reference `ports/`, `adapters/`, or `infrastructure/`
- NEVER perform I/O (no file reads, no network calls, no logging side effects)

### application/
- NEVER import a transport library (HTTP, gRPC, WS, CLI parsers)
- NEVER import a database driver or ORM
- NEVER write SQL or query DSL directly
- NEVER reference `adapters/` or `infrastructure/`
- NEVER contain conditional logic based on the caller's protocol

### ports/
- NEVER contain implementation code — interfaces/contracts only
- NEVER import `adapters/` or `infrastructure/`
- NEVER expose transport-specific types (e.g. HTTP Request/Response objects)

### adapters/
- NEVER contain business logic or domain rules
- NEVER import `infrastructure/` (no DI container access)
- NEVER call another adapter directly — go through a port

### infrastructure/
- NEVER contain business logic
- NEVER be imported by any other layer

---

## 6. Port-First Development Rule

**Always define the port before the implementation.**

Order of work:
1. Define the outbound port interface (`src/ports/outbound/`)
2. Write the use case against the port interface (`src/application/use-cases/`)
3. Implement the adapter (`src/adapters/outbound/`)
4. Wire in infrastructure (`src/infrastructure/di/`)

This order is enforced because it keeps business logic decoupled from implementation
details and makes use cases testable without real infrastructure.

---

## 7. Self-Audit Checklist

Before submitting any code change, verify:

- [ ] Each new file is in the correct layer directory
- [ ] No layer imports from a layer it is forbidden to import from
- [ ] New concepts follow the naming convention table in Section 4
- [ ] No business logic exists in adapters or infrastructure
- [ ] No transport or DB code exists in domain or application
- [ ] Every new outbound dependency has a corresponding port interface
- [ ] Every new inbound entry point has a corresponding inbound port
- [ ] New use cases are named `{Verb}{Noun}UseCase`
- [ ] New adapters are named `{Technology}{PortName}`
- [ ] The composition root in `infrastructure/di/` wires the new component

---

## 8. Layer-Specific AGENTS.md Files

For detailed rules scoped to the layer you are working in, read:

| Layer | Agent contract |
|---|---|
| domain | `src/domain/AGENTS.md` |
| application | `src/application/AGENTS.md` |
| ports | `src/ports/AGENTS.md` |
| adapters | `src/adapters/AGENTS.md` |
| infrastructure | `src/infrastructure/AGENTS.md` |

---

## 9. Language-Specific Supplements

This file and the layer AGENTS.md files are **language-neutral**. When working in
a specific programming language, also load the language supplement:

```
docs/languages/{language}/AGENTS.md
```

**Agent loading order** (each file narrows and specialises the previous):

1. `AGENTS.md` ← this file — architecture rules, layer map, naming conventions
2. `src/{layer}/AGENTS.md` ← layer rules for the layer you are working in
3. `docs/languages/{language}/AGENTS.md` ← idioms, tooling, forbidden patterns for your language

**Precedence rule**: language supplements are authoritative for idioms and tooling
choices. They never override architecture rules. If a language supplement conflicts
with an architecture rule in this file or a layer AGENTS.md, the architecture rule wins.

| Language | Supplement | Key frameworks |
|---|---|---|
| Python | `docs/languages/python/AGENTS.md` | FastAPI, asyncpg |
| TypeScript | `docs/languages/typescript/AGENTS.md` | TanStack Start, postgres.js |
| Go | `docs/languages/go/AGENTS.md` | Gin, pgx |
| Rust | `docs/languages/rust/AGENTS.md` | Axum, sqlx |

Real-code examples for each language and layer live in `docs/languages/{language}/examples/`.

**Language-specific overrides to be aware of:**

- **Go** replaces `src/` with `internal/` (toolchain-enforced private packages) and flattens `ports/inbound/` + `ports/outbound/` into a single `internal/ports/` package. See `docs/languages/go/AGENTS.md` → *Directory Structure*.
- **Async model and error propagation** mechanics vary by language (exception hierarchies in Python/TypeScript; `(T, error)` returns in Go; `Result<T, E>` + `?` in Rust). The architecture contract governs *what* flows across boundaries — the language supplement is authoritative for *how* the language expresses success, failure, and async.

---

## 10. When in Doubt

1. **Read the layer's `AGENTS.md`** for the specific rules
2. **Read the layer's `EXAMPLE.md`** files to see the pseudocode pattern
3. **Define a port** — if the boundary is unclear, a port makes it explicit
4. **Do not shortcut** — the constraint is the point; it is what makes the codebase
   maintainable across multiple agents and humans over time
