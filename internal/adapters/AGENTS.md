# AGENTS.md — Adapters Layer Contract

You are working in `src/adapters/`. Read this before writing any code here.

---

## What this layer is

Adapters are **concrete implementations** of ports. They translate between the
application's domain language and the outside world's technical protocols.

- **Inbound adapters** (`adapters/inbound/`) — receive input from outside and call the application
- **Outbound adapters** (`adapters/outbound/`) — implement what the application needs from outside

An adapter has exactly **one job**: translate. It contains no business logic.

---

## Allowed imports

| Allowed | Why |
|---|---|
| `src/ports/inbound/` | Inbound adapters call inbound port interfaces |
| `src/ports/outbound/` | Outbound adapters implement outbound port interfaces |
| `src/application/` | Inbound adapters may reference use case input/output types |
| `src/domain/` | Read-only: value types and error types only |
| Transport libraries | HTTP framework, gRPC stubs, WebSocket library, CLI parser |
| Database drivers / ORMs | Postgres driver, Redis client, ORM — in outbound adapters only |
| External SDK libraries | Stripe SDK, AWS SDK, Twilio SDK — in outbound adapters only |

---

## Forbidden imports — hard rules

| Forbidden | Why |
|---|---|
| `src/infrastructure/` | Infrastructure wires adapters; adapters must not depend on infrastructure |
| Business logic from `src/domain/` beyond value types | Adapters translate; domain enforces rules |
| Another adapter (inbound or outbound) | Adapters communicate through ports, never directly |

---

## What belongs here

### Inbound adapters (`inbound/`)

Receive input from an external caller and invoke the application layer through an inbound port.

Responsibilities:
- Parse and validate the transport-level input (HTTP body, gRPC message, WS frame, CLI args)
- Map transport input → port command/query type
- Call the inbound port (`PlaceOrderPort.execute(command)`)
- Map the result → transport-level response
- Map application errors → transport-level error codes

### Outbound adapters (`outbound/`)

Implement an outbound port by calling an external system.

Responsibilities:
- Implement the outbound port interface
- Map domain types → external system format (SQL, API request, queue message)
- Call the external system (DB, API, cache, queue)
- Map the response → domain types
- Handle external system errors and translate to domain/application errors

---

## One adapter, one port rule

Each adapter class/struct should implement **exactly one port**. If a Postgres
adapter implements both `OrderRepository` and `UserRepository`, split it into
`PostgresOrderRepository` and `PostgresUserRepository`. They can share a
database connection object injected via the constructor.

---

## Naming conventions

| Concept | Pattern | Examples |
|---|---|---|
| Inbound adapter | `{Technology}{Noun}Adapter` | `HttpOrderAdapter`, `GrpcOrderAdapter`, `WsEventAdapter`, `CliImportAdapter` |
| Outbound repository | `{Technology}{Noun}Repository` | `PostgresOrderRepository`, `MongoUserRepository` |
| Outbound gateway | `{Technology}{Noun}Gateway` | `StripePaymentGateway`, `TwilioSMSGateway` |
| Outbound store | `{Technology}{Noun}Store` | `RedisSessionStore`, `MemcachedCacheStore` |
| Outbound publisher | `{Technology}EventPublisher` | `KafkaEventPublisher`, `RabbitMQEventPublisher` |

---

## Error mapping rule

Adapters are responsible for translating errors across the boundary:

**Inbound direction** (external error → application error):
```
transport error (malformed JSON, missing field) → application validation error
```

**Outbound direction** (application error → transport response):
```
application NotFoundError  → HTTP 404 / gRPC NOT_FOUND / WS error frame
application ForbiddenError → HTTP 403 / gRPC PERMISSION_DENIED
domain DomainError         → HTTP 422 / gRPC INVALID_ARGUMENT
```

The inbound adapter handles the mapping. The use case never knows the transport.

---

## Self-audit checklist for this layer

Before submitting code in `src/adapters/`:

- [ ] No import from `src/infrastructure/`
- [ ] No business logic — only translation and protocol handling
- [ ] No adapter calls another adapter directly
- [ ] Inbound adapters call only inbound ports (not use cases directly by class name)
- [ ] Outbound adapters implement exactly one outbound port interface
- [ ] Naming follows `{Technology}{Noun}{PortSuffix}` pattern
- [ ] Error mapping is complete — all application errors have a transport-level mapping
- [ ] No domain rules re-implemented here — domain object enforces them
- [ ] Database connection or HTTP client is injected, not instantiated inside the adapter

---

## See also

- [`inbound/EXAMPLE.md`](inbound/EXAMPLE.md)
- [`outbound/EXAMPLE.md`](outbound/EXAMPLE.md)
- Root [`AGENTS.md`](../../AGENTS.md) for the full architecture contract
