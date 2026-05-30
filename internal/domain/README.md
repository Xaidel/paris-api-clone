# Domain Layer

The domain layer is the **core of the application**. It is entirely independent of
how the application is delivered (HTTP, gRPC, CLI) or how data is stored (Postgres,
Redis, files). It expresses the business in pure code.

---

## What lives here

| Subdirectory | Contents |
|---|---|
| `entities/` | Objects with identity and lifecycle. They enforce business invariants. |
| `value-objects/` | Immutable, self-validating values. Equality is by value, not reference. |
| `events/` | Immutable records of things that happened. Named in past tense. |
| `services/` | Stateless business logic that spans multiple entities. |

---

## The rule: zero dependencies

The domain layer imports **nothing** from other layers and **no external frameworks**.
It uses only:

- The language's standard library
- Other types within `src/domain/`

This is what makes the domain layer testable in isolation and portable across
delivery mechanisms.

---

## What does NOT belong here

- HTTP request/response types
- Database queries or ORM models
- Framework decorators, annotations, or lifecycle hooks
- Logging calls (use domain events to signal significant occurrences)
- Any operation that requires I/O

---

## Relationship to other layers

```
src/domain/  ◀──  src/application/  (application orchestrates domain)
src/domain/  ◀──  src/ports/        (ports reference domain types)
src/domain/  ◀──  src/adapters/     (adapters use domain value types)
```

The arrows point **toward** domain. Domain never points outward.

---

## Further reading

- [`AGENTS.md`](AGENTS.md) — machine-readable layer contract
- [`entities/EXAMPLE.md`](entities/EXAMPLE.md)
- [`value-objects/EXAMPLE.md`](value-objects/EXAMPLE.md)
- [`events/EXAMPLE.md`](events/EXAMPLE.md)
- [`services/EXAMPLE.md`](services/EXAMPLE.md)
