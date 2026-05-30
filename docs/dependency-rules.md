# Dependency Rules

This document states the dependency rules precisely, in a format that both
humans and AI agents can use to verify correctness.

---

## The Dependency Rule

> Source code dependencies must always point **inward** — toward the domain.
> No layer may import from a layer that is further out than itself.

```
infrastructure → adapters → ports → application → domain
(outermost)                                        (innermost)
```

---

## Layer dependency table

Use this table to verify any import. Find the layer of the file you are writing in
(left column), then check whether the layer you want to import from is in its
"Allowed" column.

| Layer | Path | May import from | May NOT import from |
|---|---|---|---|
| `domain` | `src/domain/` | Nothing (zero external deps) | `application`, `ports`, `adapters`, `infrastructure` |
| `application` | `src/application/` | `domain`, `ports` | `adapters`, `infrastructure` |
| `ports` | `src/ports/` | `domain` | `application`, `adapters`, `infrastructure` |
| `adapters` | `src/adapters/` | `ports`, `application`, `domain` (value types only) | `infrastructure` |
| `infrastructure` | `src/infrastructure/` | Everything | — |

---

## Decision tree for any import

```
You are in file X and want to import from Y.

1. What layer is file X in?
2. What layer is file Y in?
3. Is Y in the "May import from" column for X's layer?
   └─ YES → import is valid
   └─ NO  → the import violates the dependency rule

   Fix options:
   a. Move the code in Y to a layer that X is allowed to import from
   b. Extract an interface (port) that X can depend on instead
   c. Reconsider whether the code in X belongs in a different layer
```

---

## Common violations and how to fix them

### Domain imports application

```
// VIOLATION: domain entity imports a use case
import PlaceOrderUseCase from "src/application/use-cases/place-order"

// FIX: the domain never needs the application layer.
// If the domain needs to trigger something, it emits a domain event.
// The application layer listens for events and acts.
```

### Application imports adapter

```
// VIOLATION: use case imports a concrete repository
import PostgresOrderRepository from "src/adapters/outbound/postgres-order-repository"

// FIX: use case imports the port (interface) only.
// The concrete class is injected at runtime by infrastructure/di/.
import OrderRepository from "src/ports/outbound/order-repository"
```

### Adapter imports infrastructure

```
// VIOLATION: adapter accesses the DI container to resolve dependencies
import container from "src/infrastructure/di/container"
db = container.resolve("database")

// FIX: inject the dependency through the constructor.
// Infrastructure/di/ passes the database to the adapter when constructing it.
constructor(db: DatabaseConnection)
```

### Domain performs I/O

```
// VIOLATION: entity loads related data from a repository
Entity Order {
  confirm():
    payment = paymentRepository.findByOrderId(self.id)  // I/O in domain!
    ...
}

// FIX: load in the use case, pass to the domain object.
UseCase ConfirmOrderUseCase {
  execute(command):
    payment = paymentRepository.findByOrderId(command.orderId)
    order   = orderRepository.findById(command.orderId)
    order.confirmWithPayment(payment)   // domain receives data, doesn't fetch it
}
```

### Port exposes transport type

```
// VIOLATION: inbound port uses an HTTP type
Interface PlaceOrderPort {
  execute(request: HttpRequest) → HttpResponse  // transport type in port!
}

// FIX: port uses domain/primitive types only.
Interface PlaceOrderPort {
  execute(command: PlaceOrderCommand) → PlaceOrderResult
}
```

---

## Automated enforcement

The dependency rules should be enforced by a linter or architecture test:

| Language | Tool |
|---|---|
| Python | `import-linter` (`importlinter`) |
| Go | `go-arch-lint` or `depguard` |
| Rust | Cargo workspace member visibility (`pub(crate)` / module structure) |
| TypeScript | `eslint-plugin-import`, `ts-arch`, or `dependency-cruiser` |
| Java / Kotlin | `ArchUnit` |

Configure the tool to fail the build if any forbidden import is detected.
This makes the architecture self-enforcing.
