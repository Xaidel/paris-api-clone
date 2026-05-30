# ADR-0001: Use Hexagonal Architecture

**Date**: 2026-03-03
**Status**: Accepted
**Deciders**: Project founders

---

## Context

Backend services tend to accumulate coupling over time. Business logic leaks into
HTTP handlers. Database queries appear in domain objects. Changing a storage
technology requires touching business logic. Adding a new transport (gRPC alongside
REST) requires duplicating logic.

In the era of AI-assisted development, this problem is amplified: AI agents generate
code quickly but without inherent knowledge of where things belong. Without explicit
structural constraints, agents place logic wherever seems most convenient, accelerating
the accumulation of coupling.

We needed an architecture that:
1. Enforces a strict separation between business logic and infrastructure
2. Makes any transport (HTTP, gRPC, WebSocket, CLI, queue consumer) addable without
   touching business logic
3. Makes any storage technology swappable without touching business logic
4. Is testable at the domain and application layers without starting servers or databases
5. Provides explicit, checkable rules that can be communicated to AI agents

---

## Decision

Adopt **Hexagonal Architecture** (Ports & Adapters) as the structural pattern for
all backend services using this template.

The architecture is organized into five layers with strict dependency rules:
- `domain/` — pure business logic, zero dependencies
- `application/` — orchestration, depends only on domain and ports
- `ports/` — interface contracts, no implementations
- `adapters/` — concrete implementations of ports
- `infrastructure/` — composition root, wires all layers

The Dependency Rule is enforced: dependencies point inward only. The domain depends
on nothing. Infrastructure depends on everything. No layer imports from a layer
further out than itself.

---

## Alternatives considered

### Option A: Layered (N-tier) Architecture

A traditional Presentation → Service → Repository stack.

- **Pros**: Familiar, widely understood, low setup cost
- **Cons**: Business logic tends to leak into the service layer. The database
  schema drives the domain model (ORM models become the domain). Adding a new
  transport requires threading through all layers. Testing requires mocking the
  full stack.

### Option B: CQRS + Event Sourcing

Separate read and write models, with state derived from an event log.

- **Pros**: Excellent audit trail, scalable read models, natural event-driven design
- **Cons**: Higher complexity, unfamiliar to many engineers, overkill for most CRUD-heavy
  services, significant infrastructure requirements (event store)

### Option C (chosen): Hexagonal Architecture

- **Pros**: Business logic is genuinely isolated and testable. Transport and storage
  are swappable. The architecture scales from simple to complex without structural
  changes. Rules are explicit and enforceable. Compatible with CQRS patterns if needed later.
- **Cons**: More initial structure than layered architecture. Requires discipline
  to maintain port-first discipline. Unfamiliar to engineers who have only worked
  in MVC frameworks.

---

## Consequences

### Positive
- Business logic can be tested without starting a server or a database
- Adding a new transport (e.g. gRPC alongside existing REST) requires only a new
  inbound adapter — zero changes to domain or application
- Swapping storage technology requires only a new outbound adapter — zero changes
  to domain or application
- The architecture is self-documenting: a file's layer tells you its role and constraints
- AI agents can be given explicit rules (AGENTS.md) that they can verify against

### Negative
- More boilerplate upfront: ports must be defined before use cases, use cases before
  adapters. This is intentional friction that prevents shortcuts.
- Engineers familiar only with MVC or service-layer patterns need to adjust their
  mental model
- Small, simple services may feel over-engineered — but the template can be adopted
  partially for simple cases

### Neutral / risks
- Port proliferation: if ports are defined too granularly, there are too many
  small interfaces. Mitigation: one port per meaningful external concern, not one
  per method.
- Risk that teams treat this as boilerplate to work around, rather than a constraint
  to respect. Mitigation: AGENTS.md and layer-specific contracts make the rules explicit.
