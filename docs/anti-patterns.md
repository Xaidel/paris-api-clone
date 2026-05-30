# Anti-Patterns

This document catalogs the most common architectural violations — patterns that
seem convenient but break the boundaries that make hexagonal architecture valuable.

Each entry includes: what the pattern is, why it's wrong, and how to fix it.

---

## Domain layer anti-patterns

### 1. Domain entity with I/O

**The pattern**:
```
Entity Order {
  confirm():
    payment = database.query("SELECT * FROM payments WHERE order_id = ?", self.id)
    if payment.status == "captured":
      self.status = CONFIRMED
}
```

**Why it's wrong**: The domain becomes coupled to the database. You cannot test
`confirm()` without a real database. You cannot change the storage technology
without touching domain logic.

**Fix**: Load data in the use case, pass it to the domain object.
```
UseCase ConfirmOrderUseCase {
  execute(command):
    payment = paymentRepository.findByOrderId(command.orderId)
    order   = orderRepository.findById(command.orderId)
    order.confirm(payment)   // domain receives data, doesn't fetch it
}
```

---

### 2. HTTP status codes in domain errors

**The pattern**:
```
Entity Order {
  cancel():
    if self.status == SHIPPED:
      raise HttpError(409, "Cannot cancel a shipped order")
}
```

**Why it's wrong**: Domain errors must not know about HTTP. If you later add a gRPC
transport, the error carries an HTTP status code — meaningless to gRPC.

**Fix**: Raise a plain domain error. Adapters map it to transport codes.
```
Entity Order {
  cancel():
    if self.status == SHIPPED:
      raise DomainError("Cannot cancel a shipped order")
}

// In HttpOrderAdapter:
catch DomainError as e:
  return HttpResponse(422, {error: e.message})

// In GrpcOrderAdapter:
catch DomainError as e:
  raise GrpcError(INVALID_ARGUMENT, e.message)
```

---

### 3. Anemic domain model

**The pattern**:
```
// Entity with no behavior — just data
Entity Order {
  id:     OrderId
  items:  List<OrderItem>
  status: OrderStatus
  total:  Money
  // No methods — all logic is elsewhere
}

// All logic in an "OrderManager" or "OrderService"
Service OrderManager {
  cancelOrder(order: Order, reason: String):
    if order.status == SHIPPED:
      raise Error("...")
    order.status = CANCELLED   // mutating entity from outside
    order.events.append(OrderCancelled(...))
}
```

**Why it's wrong**: Invariants are scattered. Nothing enforces that `order.status`
is only set through valid transitions. Any code can mutate the entity directly.

**Fix**: Put behavior on the entity.
```
Entity Order {
  cancel(reason: String):
    if self.status == SHIPPED:
      raise DomainError("Cannot cancel a shipped order")
    self.status = CANCELLED
    self.events.append(OrderCancelled(self.id, reason))
}
```

---

## Application layer anti-patterns

### 4. Business logic in a use case

**The pattern**:
```
UseCase PlaceOrderUseCase {
  execute(command):
    // Business rule in use case:
    if command.items.length > 50:
      raise Error("Order cannot have more than 50 items")
    // More business rules...
    if command.items.any(i → i.quantity <= 0):
      raise Error("Quantity must be positive")
}
```

**Why it's wrong**: Business rules live in the domain. When they live in use cases,
they are duplicated across multiple use cases, not enforced at the domain boundary,
and tested in integration rather than unit tests.

**Fix**: Move invariants to the domain entity or value object.
```
Entity Order {
  construct(id, customerId, items, total):
    if items.length > 50:      raise DomainError("Order cannot have more than 50 items")
    if items.isEmpty():        raise DomainError("Order must have at least one item")
}

ValueObject OrderItem {
  construct(productId, quantity, price):
    if quantity <= 0: raise DomainError("Quantity must be positive")
}
```

---

### 5. Use case calls another use case

**The pattern**:
```
UseCase PlaceOrderUseCase {
  execute(command):
    order = ...
    // Calling another use case directly:
    notifyCustomerUseCase.execute(NotifyCustomerCommand(order.customerId))
}
```

**Why it's wrong**: Use cases should be independent units. Chaining them creates
coupling, makes each harder to test, and obscures the flow.

**Fix**: Use domain events and an application service, or move shared logic to an
application service injected into both use cases.
```
UseCase PlaceOrderUseCase {
  execute(command):
    order = ...
    orderRepository.save(order)
    eventPublisher.publishAll(order.events)              // emit event
    notificationService.notifyOrderPlaced(order.id, ...) // shared service
}
```

---

## Ports anti-patterns

### 6. Port exposes transport types

**The pattern**:
```
Interface PlaceOrderPort {
  execute(request: HttpRequest) → HttpResponse
}
```

**Why it's wrong**: The port is now tied to HTTP. A CLI or gRPC caller cannot use it.
The port must be the universal interface for all transports.

**Fix**: Use domain types and primitives only.
```
Interface PlaceOrderPort {
  execute(command: PlaceOrderCommand) → PlaceOrderResult
}
```

---

### 7. Port contains implementation code

**The pattern**:
```
Interface OrderRepository {
  findById(id: OrderId) → Order:
    // default implementation in the interface
    row = sqlConnection.query("SELECT ...", id)
    return mapToOrder(row)
}
```

**Why it's wrong**: Ports are contracts. Any implementation detail here is a violation.

**Fix**: Port is interface only. Implementation belongs in an adapter.

---

## Adapters anti-patterns

### 8. Business logic in an inbound adapter

**The pattern**:
```
InboundAdapter HttpOrderAdapter {
  handlePlaceOrder(request: HttpRequest):
    body = request.parseJSON()

    // Business rule in the HTTP handler:
    if body.items.length == 0:
      return HttpResponse(400, {error: "Order must have items"})

    // More business rules:
    total = body.items.sum(i → i.price * i.quantity)
    if total > 10000:
      return HttpResponse(400, {error: "Order total cannot exceed 10,000"})
```

**Why it's wrong**: When you add a gRPC adapter, these rules must be duplicated.
Business rules belong in the domain, not in adapters.

**Fix**: Adapter translates only. Domain enforces rules.
```
InboundAdapter HttpOrderAdapter {
  handlePlaceOrder(request: HttpRequest):
    body    = request.parseJSON()
    command = PlaceOrderCommand(customerId: body.customer_id, items: body.items)

    try:
      result = placeOrderPort.execute(command)   // domain enforces rules here
      return HttpResponse(201, result)
    catch DomainError as e:
      return HttpResponse(422, {error: e.message})
}
```

---

### 9. Adapter calls another adapter

**The pattern**:
```
OutboundAdapter PostgresOrderRepository {
  save(order):
    db.save(order)
    // Calling another adapter directly:
    kafkaAdapter.publish(OrderSavedEvent(order.id))
}
```

**Why it's wrong**: Adapters should communicate through ports. Directly calling
another adapter creates tight coupling between infrastructure implementations.

**Fix**: Event publishing is handled by the use case through the EventPublisher port.
```
UseCase PlaceOrderUseCase {
  execute(command):
    ...
    orderRepository.save(order)      // via port
    eventPublisher.publish(event)    // via port — not called from inside repository
}
```

---

## Infrastructure anti-patterns

### 10. Business logic in infrastructure

**The pattern**:
```
Function wire(config):
  // Business rule in infrastructure:
  if config.environment == "production":
    orderLimit = 1000
  else:
    orderLimit = 9999

  placeOrderUseCase = PlaceOrderUseCase(orderLimit: orderLimit)
```

**Why it's wrong**: Business rules (order limits, validation) must not live in
infrastructure. They become invisible to domain tests and impossible to find.

**Fix**: Business rules in the domain. Config in infrastructure is structural only.
```
Entity Order {
  construct(id, customerId, items, total):
    if items.length > MAX_ITEMS_PER_ORDER:   // constant in domain
      raise DomainError("...")
}
```

---

### 11. Other layers import infrastructure

**The pattern**:
```
// In an adapter — importing the DI container to resolve a dependency
import container from "src/infrastructure/di/container"

OutboundAdapter PostgresOrderRepository {
  save(order):
    logger = container.resolve("logger")   // anti-pattern: adapter accesses DI
    logger.info("Saving order", {id: order.id})
}
```

**Why it's wrong**: Adapters must not know about infrastructure. This creates a
circular dependency risk and defeats the composition root pattern.

**Fix**: Inject the logger through the constructor.
```
OutboundAdapter PostgresOrderRepository {
  constructor(db: DatabaseConnection, logger: Logger)

  save(order):
    self.logger.info("Saving order", {id: order.id})
    // ...
}
```
