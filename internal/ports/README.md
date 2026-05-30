# Ports Layer

Ports are the **formal boundary** of the application core. They are interfaces
(contracts) — no implementation, no logic.

---

## Two directions, two kinds of ports

| Directory | Port type | Direction | Who calls it | Who implements it |
|---|---|---|---|---|
| `inbound/` | Inbound port | Outside → App | Inbound adapters (HTTP, gRPC, WS, CLI) | Application use cases |
| `outbound/` | Outbound port | App → Outside | Application use cases | Outbound adapters (DB, API, cache) |

---

## Why ports exist

Without ports, the application layer would directly reference concrete implementations
(a Postgres repository, an HTTP client). This creates a hard dependency on
infrastructure, making the application layer impossible to test in isolation and
impossible to swap implementations.

Ports invert this dependency:

```
// Without ports (BAD — application imports infrastructure)
use-case imports PostgresOrderRepository

// With ports (GOOD — application imports a contract; adapter implements it)
use-case imports OrderRepository (interface)
PostgresOrderRepository implements OrderRepository
```

The use case is now testable with an in-memory `OrderRepository`. The database
implementation is swappable without touching any application code.

---

## What does NOT belong here

- Implementation code of any kind
- Transport-specific types (HTTP Request, gRPC Context)
- Database-specific types (SQL Row, ORM Model)
- Framework types

---

## Further reading

- [`AGENTS.md`](AGENTS.md) — machine-readable layer contract
- [`inbound/EXAMPLE.md`](inbound/EXAMPLE.md)
- [`outbound/EXAMPLE.md`](outbound/EXAMPLE.md)
