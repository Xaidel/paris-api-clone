# Development Workflow

This document describes how to implement features in this codebase — for both
human engineers and AI agents. It covers the port-first development loop,
how to review code (human or agent-generated), and how to prompt an AI agent
to build features correctly.

---

## Part 1: Port-First Development Loop

**Always define the port before the implementation.**

This is the most important rule in hexagonal architecture. Defining the port first
forces you to think about the contract before the implementation — what the
application needs, not how it will be provided.

### The loop

```
1. DEFINE → Define the outbound port interface (src/ports/outbound/)
2. USE    → Write the use case against the port (src/application/use-cases/)
3. TEST   → Write a unit test using an in-memory port implementation
4. IMPLEMENT → Implement the adapter (src/adapters/outbound/)
5. TEST AGAIN → Write an integration test for the adapter
6. WIRE   → Wire everything in the composition root (src/infrastructure/di/)
```

### Step-by-step example: adding order persistence

**Step 1: Define the port**

Before writing any use case or Postgres code, define what the application needs:

```
// src/ports/outbound/order-repository
Interface OrderRepository {
  save(order: Order) → void
  findById(id: OrderId) → Order | null
  findByCustomerId(customerId: UserId) → List<Order>
}
```

**Step 2: Write the use case**

Write `PlaceOrderUseCase` using only the port interface. The use case is complete
and testable before any database code exists.

**Step 3: Unit test**

Write a unit test using `InMemoryOrderRepository`. This validates the use case
logic without a database.

**Step 4: Implement the adapter**

Write `PostgresOrderRepository implements OrderRepository`. Now that the contract
is locked, the implementation can be built without affecting the use case.

**Step 5: Integration test**

Write tests for `PostgresOrderRepository` against a real Postgres instance
(Dockerized). Validate that it correctly implements the port contract.

**Step 6: Wire**

In `infrastructure/di/`, replace `InMemoryOrderRepository` with
`PostgresOrderRepository` in the `wire()` function.

---

## Part 2: Self-Audit Checklist

Run this checklist on any code change before submitting — whether written by
a human or an AI agent.

### Layer placement
- [ ] Every new file is in the correct layer directory
- [ ] The file's location matches the code it contains (use the decision flowchart in `AGENTS.md`)

### Dependency rule
- [ ] No layer imports from a layer it is forbidden to import from (check `docs/dependency-rules.md`)
- [ ] No adapter imports `infrastructure/`
- [ ] No domain or application code imports `adapters/` or `infrastructure/`
- [ ] No domain code imports `application/` or `ports/`

### Naming
- [ ] New concepts follow the naming convention table in `docs/naming-conventions.md`
- [ ] Use cases are named `{Verb}{Noun}UseCase`
- [ ] Adapters are named `{Technology}{Noun}{PortSuffix}`
- [ ] Ports are named per their type (`{Verb}{Noun}Port`, `{Noun}Repository`, etc.)

### Business logic isolation
- [ ] No business rules in adapters
- [ ] No business rules in infrastructure
- [ ] No SQL in domain or application layers
- [ ] No transport types (HTTP, gRPC) in domain, application, or port layers

### Port discipline
- [ ] Every new outbound dependency has a corresponding port interface
- [ ] Every new inbound entry point has a corresponding inbound port
- [ ] Ports contain only interface definitions — no implementation

### Wiring
- [ ] New components are wired in `infrastructure/di/`
- [ ] Dependencies are injected through constructors — not instantiated inside classes

---

## Part 3: AI Agent Workflow

### How to instruct an agent to implement a feature

When asking an AI agent to implement a feature in this codebase, provide this
context in your prompt:

```
Read AGENTS.md before writing any code.
Read the AGENTS.md in the specific layer(s) you will work in.
Follow the port-first workflow: define the port first, then the use case, then the adapter.
Run the self-audit checklist in AGENTS.md Section 7 before submitting.
```

### Recommended prompt structure

```
CONTEXT:
- This codebase uses Hexagonal Architecture. Read AGENTS.md.
- Read src/{layer}/AGENTS.md for the layer you are working in.
- Read src/{layer}/{sublayer}/EXAMPLE.md for the pattern.

TASK:
- Implement [feature name]
- Start by defining the outbound port in src/ports/outbound/
- Then write the use case in src/application/use-cases/
- Then implement the adapter in src/adapters/outbound/
- Wire in src/infrastructure/di/

CONSTRAINTS:
- Follow naming conventions in docs/naming-conventions.md
- Do not add business logic to adapters or infrastructure
- Do not import adapters/ from application/ or domain/
- Run the self-audit checklist before finishing
```

### Splitting work across agents

For large features, split work by layer. Each agent loads only the AGENTS.md
for its layer:

| Agent task | Context to load |
|---|---|
| Define ports | `AGENTS.md` + `src/ports/AGENTS.md` + `src/ports/*/EXAMPLE.md` |
| Implement use case | `AGENTS.md` + `src/application/AGENTS.md` + `src/application/use-cases/EXAMPLE.md` |
| Implement adapter | `AGENTS.md` + `src/adapters/AGENTS.md` + `src/adapters/outbound/EXAMPLE.md` |
| Wire infrastructure | `AGENTS.md` + `src/infrastructure/AGENTS.md` + `src/infrastructure/di/EXAMPLE.md` |

### Reviewing agent-generated code

Before accepting agent output, run through the checklist in Part 2. Pay particular
attention to:

1. **Dependency violations** — agents often take shortcuts by importing concrete
   classes instead of ports, or by importing infrastructure from adapters.

2. **Business logic placement** — agents tend to put validation logic in HTTP
   handlers rather than in domain entities. Check that invariants are in the domain.

3. **Naming drift** — agents may use different names than the conventions specify.
   Check naming against `docs/naming-conventions.md`.

4. **Missing port definition** — agents may implement an adapter without defining
   the port interface first. The port must exist before the adapter.

5. **Incomplete wiring** — agents often forget to update `infrastructure/di/`.
   New components are not active until they are wired.

### Common agent failure modes

| Failure mode | Signal | Fix |
|---|---|---|
| Business logic in adapter | HTTP handler contains domain conditionals | Move invariants to domain entity |
| SQL in use case | `SELECT`, `INSERT` in application layer | Define a repository port; move SQL to adapter |
| Concrete import instead of port | Use case imports `PostgresOrderRepository` | Change import to `OrderRepository` port |
| Missing port | Adapter exists but no corresponding interface | Add port interface to `ports/outbound/` |
| Missing wiring | New component not in `infrastructure/di/` | Add to `wire()` function |
| Naming inconsistency | `OrderHandler` instead of `HttpOrderAdapter` | Rename to follow `{Technology}{Noun}Adapter` |

---

## Part 4: ADR-Driven Development

For any significant architectural decision — adding a new transport, changing the
event model, introducing a new external dependency — write an ADR first.

ADR location: `docs/adr/`
ADR template: `docs/adr/template.md`

Writing the ADR before implementing forces you to articulate the problem, the
alternatives considered, and the consequences. This is especially valuable when
working with AI agents, because the ADR becomes context for future agent sessions.
