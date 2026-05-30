# Tests

This directory contains all automated tests, organized by **scope** — not by
the layer they test. Each scope has different rules for what infrastructure is
available and what can be substituted.

---

## Three test scopes

| Scope | Directory | What it tests | Infrastructure used |
|---|---|---|---|
| Unit | `unit/` | Domain + application logic | None — pure in-memory |
| Integration | `integration/` | Adapter contracts | Real or containerized (DB, cache, queue) |
| End-to-end | `e2e/` | Full user scenarios | Full stack via inbound adapter |

---

## The testing pyramid

```
          ┌─────────┐
          │   e2e   │  ← few, slow, high confidence
          ├─────────┤
          │integrat.│  ← moderate, adapter contracts
          ├─────────┤
          │  unit   │  ← many, fast, domain/application
          └─────────┘
```

Unit tests form the base — many, fast, no I/O. Integration tests validate
adapter contracts. E2E tests validate complete user flows.

---

## Test doubles strategy

| Layer tested | What to substitute |
|---|---|
| Domain entities | Nothing — test them directly (no deps) |
| Use cases | Inject in-memory implementations of outbound ports |
| Inbound adapters | Start a test server; call via HTTP/gRPC client in tests |
| Outbound adapters | Test against a real DB/cache/queue in a container |

---

## Further reading

- [`unit/README.md`](unit/README.md)
- [`integration/README.md`](integration/README.md)
- [`e2e/README.md`](e2e/README.md)
