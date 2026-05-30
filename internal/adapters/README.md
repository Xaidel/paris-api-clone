# Adapters Layer

Adapters are the **bridge** between the application core and the outside world.
They translate — from transport protocols into application commands, and from
application results into transport responses.

---

## Two directions

| Directory | Direction | Receives from | Calls |
|---|---|---|---|
| `inbound/` | Outside → App | HTTP, gRPC, WebSocket, CLI, queues | Inbound ports |
| `outbound/` | App → Outside | Application use cases | Postgres, Redis, Stripe, Kafka, etc. |

---

## The rule: translate, don't decide

An adapter's only job is **translation**. It never:
- Enforces business rules
- Makes decisions based on domain state
- Calls another adapter directly

All decisions and rules live in the domain and application layers.

---

## Organization

Adapters are organized by direction first (`inbound/`, `outbound/`). When a
direction contains many adapters of different technologies, add a subdirectory
per technology:

```
adapters/
├── inbound/
│   ├── http/
│   │   ├── HttpOrderAdapter
│   │   └── HttpUserAdapter
│   └── grpc/
│       └── GrpcOrderAdapter
└── outbound/
    ├── postgres/
    │   ├── PostgresOrderRepository
    │   └── PostgresUserRepository
    └── redis/
        └── RedisSessionStore
```

Keep it flat until nesting is needed.

---

## Further reading

- [`AGENTS.md`](AGENTS.md) — machine-readable layer contract
- [`inbound/EXAMPLE.md`](inbound/EXAMPLE.md)
- [`outbound/EXAMPLE.md`](outbound/EXAMPLE.md)
