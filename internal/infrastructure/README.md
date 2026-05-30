# Infrastructure Layer

Infrastructure is the **composition root** — the place where all layers are
assembled into a running application. It is the only layer that knows about
every other layer.

---

## What lives here

| Subdirectory | Purpose |
|---|---|
| `config/` | Load and validate environment configuration into typed structs |
| `di/` | Dependency injection — instantiate and wire all components |
| `observability/` | Initialize logging, tracing, and metrics before startup |

---

## The golden rule

**Infrastructure wires. It never decides.**

Any conditional logic based on domain state, any business rule, any data
transformation — all of that belongs in the domain, application, or adapter layers.
Infrastructure only assembles.

---

## Why infrastructure is the composition root

Every other layer receives its dependencies through its constructor. They are
unaware of the concrete classes that implement those dependencies. Only the
infrastructure layer knows the full picture — it is the only place where:

- `PostgresOrderRepository` is instantiated (not `OrderRepository` the interface)
- `StripePaymentGateway` is instantiated (not `PaymentGateway` the interface)
- Use cases are instantiated and given their port dependencies
- Inbound adapters are registered with their transport server

This is what makes every other layer independently testable and swappable.

---

## The dependency direction

```
src/infrastructure/  →  src/adapters/
src/infrastructure/  →  src/application/
src/infrastructure/  →  src/ports/
src/infrastructure/  →  src/domain/
```

All arrows point FROM infrastructure. No other layer imports infrastructure.

---

## Further reading

- [`AGENTS.md`](AGENTS.md) — machine-readable layer contract
- [`config/EXAMPLE.md`](config/EXAMPLE.md)
- [`di/EXAMPLE.md`](di/EXAMPLE.md)
- [`observability/EXAMPLE.md`](observability/EXAMPLE.md)
