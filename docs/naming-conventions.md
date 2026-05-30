# Naming Conventions

Consistent naming is essential for maintainability — especially when multiple
AI agents and human engineers contribute to the same codebase across many sessions.
These conventions are **required**, not optional.

---

## Why naming matters for agent-assisted development

When naming is inconsistent, agents:

- Create duplicate abstractions with different names for the same concept
- Place code in the wrong layer because the file name doesn't signal its role
- Fail to find existing types and recreate them

When naming is consistent, agents:

- Can find the correct file to edit without searching
- Can infer what a file contains from its name alone
- Can verify their output against a checklist

---

## Convention table

| Concept | Pattern | Anti-pattern | Examples |
| --- | --- | --- | --- |
| Entity | `{Noun}` | `{Noun}Model`, `{Noun}Object` | `Order`, `User`, `Invoice`, `Shipment` |
| Value Object | `{Noun}` (descriptive) | `{Noun}VO`, `{Noun}Value` | `Money`, `Email`, `OrderId`, `Address` |
| Domain Event | `{Noun}{PastVerb}` | `{Noun}Event`, `On{Noun}{Verb}` | `OrderPlaced`, `PaymentFailed`, `UserDeactivated` |
| Domain Service | `{Noun}Service` or `{Noun}{Verb}er` | `{Noun}Manager`, `{Noun}Helper` | `PricingService`, `TaxCalculator`, `InventoryAllocator` |
| Use Case (command) | `{Verb}{Noun}UseCase` | `{Noun}Handler`, `{Noun}Controller` | `PlaceOrderUseCase`, `CancelOrderUseCase` |
| Use Case (query) | `Get{Noun}UseCase`, `List{Noun}UseCase` | `{Noun}Fetcher`, `{Noun}Reader` | `GetOrderUseCase`, `ListOrdersUseCase` |
| Command input | `{Verb}{Noun}Command` | `{Noun}Request`, `{Verb}{Noun}Input` | `PlaceOrderCommand`, `CancelOrderCommand` |
| Query input | `{Noun}Query` | `{Noun}Filter`, `{Noun}Params` | `OrderQuery`, `UserSearchQuery` |
| Result / DTO | `{Noun}Result` or `{Noun}Dto` | `{Noun}Response`, `{Noun}Output` | `PlaceOrderResult`, `OrderResult` |
| Inbound Port | `{Verb}{Noun}Port` | `I{Verb}{Noun}`, `{Verb}{Noun}Interface` | `PlaceOrderPort`, `GetOrderPort`, `StreamEventsPort` |
| Repository (outbound port) | `{Noun}Repository` | `{Noun}Repo`, `{Noun}Dao`, `I{Noun}Store` | `OrderRepository`, `UserRepository` |
| Gateway (outbound port) | `{Noun}Gateway` | `{Noun}Client`, `{Noun}Api` | `PaymentGateway`, `EmailGateway`, `SMSGateway` |
| Store / Cache (outbound port) | `{Noun}Store` | `{Noun}Cache`, `{Noun}Bucket` | `SessionStore`, `CacheStore` |
| Event publisher (outbound port) | `EventPublisher` | `EventBus`, `EventDispatcher` | `EventPublisher` |
| Inbound adapter | `{Technology}{Noun}Adapter` | `{Noun}Controller`, `{Noun}Handler` | `HttpOrderAdapter`, `GrpcOrderAdapter`, `WsEventAdapter` |
| Outbound repository adapter | `{Technology}{Noun}Repository` | `{Noun}Repository{Technology}` | `PostgresOrderRepository`, `MongoUserRepository` |
| Outbound gateway adapter | `{Technology}{Noun}Gateway` | `{Noun}GatewayImpl` | `StripePaymentGateway`, `TwilioSMSGateway` |
| Outbound store adapter | `{Technology}{Noun}Store` | `{Noun}StoreImpl` | `RedisSessionStore`, `MemcachedCacheStore` |
| Outbound publisher adapter | `{Technology}EventPublisher` | `{Noun}Publisher{Technology}` | `KafkaEventPublisher`, `RabbitMQEventPublisher` |
| Application service | `{Noun}Service` | `{Noun}Manager`, `{Noun}Facade` | `NotificationService`, `AuditService` |
| Config struct | `{Scope}Config` | `{Scope}Settings`, `{Scope}Conf` | `DatabaseConfig`, `ServerConfig`, `AppConfig` |
| Domain error | `{Noun}Error` or `DomainError` | `{Noun}Exception`, `{Noun}Fault` | `DomainError`, `InvalidOrderError` |
| Application error | `{Condition}Error` | `{Condition}Exception`, `AppError` | `NotFoundError`, `ConflictError`, `ForbiddenError` |

---

## File naming

Use the language's idiomatic file naming convention, but keep the conceptual
name consistent with the table above.

| Language | Convention | Example |
| --- | --- | --- |
| Python | `snake_case.py` | `place_order_use_case.py` |
| Go | `snake_case.go` | `order_place_use_case.go` |
| Rust | `snake_case.rs` | `place_order_use_case.rs` |
| TypeScript | `camelCase.ts` or `kebab-case.ts` | `place-order-use-case.ts` |

One concept per file. The file name matches the primary type it defines.

For Go, "match" means the file should carry the same primary concept as the
exported type, not that the word order must be identical. Keep Go file names
topic-first for editor grouping, while keeping exported type names aligned with
the conceptual naming table.

### Go cross-layer file stems

In Go, keep exported type names verb-first when that matches the concept table,
but prefer topic-first file names so related files group together in the editor.

Use this pattern for cross-layer file names:

- inbound ports: `{noun}_{verb}_port.go`
- repository ports: `{noun}_repository.go`
- gateway ports: `{noun}_gateway.go`
- store ports: `{noun}_store.go`
- application use cases: `{noun}_{verb}_use_case.go`
- application services: `{noun}_service.go` or `{noun}_{verb}er.go`
- inbound adapters: `{noun}_{technology}_adapter.go`
- outbound repository adapters: `{noun}_{technology}_repository.go`
- outbound gateway adapters: `{noun}_{technology}_gateway.go`
- outbound store adapters: `{noun}_{technology}_store.go`
- tests: mirror the production file stem and append `_test.go`

Examples:

- `bug_report_create_port.go`
- `bug_report_create_use_case.go`
- `bug_report_http_adapter.go`
- `bug_report_postgres_repository.go`
- `bug_report_create_use_case_test.go`

This keeps file browsing topic-first while preserving the existing conceptual
type names such as `CreateBugReportPort`, `CreateBugReportUseCase`,
`HttpBugReportAdapter`, and `PostgresBugReportRepository`.

---

## Naming in tests

Test files mirror the source file they test, with a test suffix:

| Language | Test file convention | Example |
| --- | --- | --- |
| Python | `test_{source}.py` | `test_place_order_use_case.py` |
| Go | `{source}_test.go` | `order_place_use_case_test.go` |
| Rust | Inline `#[cfg(test)]` or `tests/{source}.rs` | `tests/place_order_use_case.rs` |
| TypeScript | `{source}.test.ts` | `place-order-use-case.test.ts` |

---

## Anti-patterns to avoid

| Anti-pattern | Why it's wrong |
| --- | --- |
| `OrderManager` | "Manager" is vague — is it a domain service, a use case, or an adapter? Use `PricingService`, `PlaceOrderUseCase`, or `PostgresOrderRepository` |
| `OrderHelper` | "Helper" signals unclear responsibility — refactor to a named abstraction |
| `OrderController` | Controller is an MVC term — use `HttpOrderAdapter` to signal layer and transport |
| `IOrderRepository` | The `I` prefix is a Java/C# habit — the port name itself signals it's an interface |
| `OrderRepositoryImpl` | The `Impl` suffix is redundant — use the technology name: `PostgresOrderRepository` |
| `OrderDTO` | "DTO" is too generic — use `OrderResult`, `OrderQuery`, or `PlaceOrderCommand` |
| `OrderCreateUseCase` | Noun-first file names do not imply noun-first type names; keep file names topic-first but keep conceptual type names verb-first, such as `order_create_use_case.go` with `CreateOrderUseCase` |
| `OrderService` (for everything) | "Service" catches too much — be specific about whether it's a domain service, application service, or gateway |
