# AGENTS.md — Infrastructure Layer Contract

You are working in `src/infrastructure/`. Read this before writing any code here.

---

## What this layer is

Infrastructure is the **composition root** — the place where all layers are wired
together. It is the only layer allowed to import from every other layer.

Infrastructure contains:
- **Dependency injection** — wiring adapters to ports to use cases
- **Configuration** — loading environment variables into typed config structs
- **Observability bootstrap** — initializing logging, tracing, and metrics
- **Application entrypoint** — starting the server, listening on ports

Infrastructure contains **zero business logic**. If you find yourself writing a
conditional based on domain state in this layer, it belongs somewhere else.

---

## Allowed imports

| Allowed | Why |
|---|---|
| `src/domain/` | Needed to pass domain types during wiring |
| `src/application/` | Instantiates use cases |
| `src/ports/` | References port interfaces when wiring |
| `src/adapters/` | Instantiates concrete adapter implementations |
| Any external library | DB drivers, HTTP servers, gRPC servers, logging libraries, config loaders |

---

## Forbidden behavior (not imports, but logic)

| Forbidden | Why |
|---|---|
| Business logic of any kind | This layer only wires; it never decides |
| Domain invariant enforcement | Domain enforces its own rules |
| Transport-specific request handling | That belongs in inbound adapters |
| SQL queries | That belongs in outbound adapters |

---

## The composition root pattern

Infrastructure is the **only place** where concrete classes are instantiated and
wired together. All other layers receive their dependencies through constructor
injection (or equivalent).

```
// Infrastructure wires everything — no other layer does this
db         = PostgresConnection(config.database)
orderRepo  = PostgresOrderRepository(db)          // outbound adapter
eventPub   = KafkaEventPublisher(kafka, config)   // outbound adapter
placeOrder = PlaceOrderUseCase(orderRepo, eventPub) // application use case
httpAdapter = HttpOrderAdapter(placeOrder)          // inbound adapter
server.register(POST, "/orders", httpAdapter.handlePlaceOrder)
```

---

## What belongs here

### `config/`
- Load configuration from environment variables, config files, or secrets managers
- Produce typed config structs (`DatabaseConfig`, `ServerConfig`, `AppConfig`)
- Validate required configuration at startup — fail fast if missing

### `di/`
- Instantiate all adapters, use cases, application services, domain services
- Wire dependencies through constructor injection
- Register inbound adapters with their transport server (HTTP router, gRPC server, WS handler)

### `observability/`
- Initialize structured logging (with correlation IDs, trace IDs)
- Initialize distributed tracing (OpenTelemetry or equivalent)
- Initialize metrics collection (Prometheus, StatsD, or equivalent)
- Register health check endpoints

---

## Naming conventions

| Concept | Pattern | Examples |
|---|---|---|
| Config struct | `{Scope}Config` | `DatabaseConfig`, `ServerConfig`, `AppConfig`, `KafkaConfig` |
| Composition root function | `wire()`, `bootstrap()`, `setup()` | `wire()`, `bootstrap()` |
| Entrypoint | `main()`, `run()`, `start()` | Language-idiomatic |

---

## Startup failure rule

**Fail fast at startup if configuration is invalid or required dependencies are unavailable.**

Do not start the server if:
- A required environment variable is missing
- A database connection cannot be established
- A required external service is unreachable (on startup health checks)

Use readiness vs liveness probes for runtime health — not startup logic.

---

## Self-audit checklist for this layer

Before submitting code in `src/infrastructure/`:

- [ ] No business logic — only wiring, configuration, and bootstrap
- [ ] No domain invariant checks
- [ ] No SQL queries
- [ ] No transport request/response handling
- [ ] Config is typed — no raw `os.getenv("FOO")` scattered through adapters
- [ ] All dependencies injected via constructor — no global singletons
- [ ] Application fails fast on missing or invalid configuration
- [ ] Observability (logging, tracing) initialized before any adapters start

---

## See also

- [`config/EXAMPLE.md`](config/EXAMPLE.md)
- [`di/EXAMPLE.md`](di/EXAMPLE.md)
- [`observability/EXAMPLE.md`](observability/EXAMPLE.md)
- Root [`AGENTS.md`](../../AGENTS.md) for the full architecture contract
