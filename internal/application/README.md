# Application Layer

The application layer **orchestrates** the domain. It is the glue between
the pure business logic (domain) and the outside world (adapters).

---

## What lives here

| Subdirectory | Contents |
|---|---|
| `use-cases/` | Named, single-purpose operations. One file per use case. |
| `services/` | Cross-cutting application concerns spanning multiple use cases. |

---

## What this layer does

1. Accepts a command or query (a plain data structure — no transport types)
2. Loads domain objects via outbound port interfaces (e.g. `OrderRepository`)
3. Calls domain logic (entity methods, domain services)
4. Persists changes via outbound port interfaces
5. Publishes domain events via an outbound port (`EventPublisher`)
6. Returns a result (plain data — no domain objects, no transport types)

---

## What this layer does NOT do

- Enforce business rules — the domain does that
- Know what triggered it (HTTP? gRPC? a queue message?)
- Write SQL, call HTTP APIs, or touch the file system directly
- Import from `adapters/` or `infrastructure/`

---

## The dependency direction

```
src/adapters/inbound/  →  src/application/use-cases/  →  src/ports/outbound/
                                    ↓
                              src/domain/
```

The application layer is called by inbound adapters and calls outbound ports.
It never knows about concrete adapter implementations.

---

## Further reading

- [`AGENTS.md`](AGENTS.md) — machine-readable layer contract
- [`use-cases/EXAMPLE.md`](use-cases/EXAMPLE.md)
- [`services/EXAMPLE.md`](services/EXAMPLE.md)
