# Domain Events — Pseudocode Examples

Three events from an order management context, illustrating the base event
pattern and specific event payloads.

---

## Base event structure

All domain events share a common structure. Define a base type and extend it.

```
// Base type for all domain events
DomainEvent {
  eventId:       String     // unique ID for this event occurrence (UUID)
  occurredAt:    Timestamp  // when it happened — set at construction, immutable
  eventType:     String     // discriminator, e.g. "order.placed", "payment.failed"

  construct():
    self.eventId    = generateUUID()
    self.occurredAt = now()
    // eventType is set by each concrete event
}
```

---

## Concrete events

### OrderPlaced

Emitted when a new order is successfully created.

```
Event OrderPlaced extends DomainEvent {
  orderId:   OrderId
  customerId: UserId
  items:     List<OrderItem>
  total:     Money

  construct(orderId: OrderId, customerId: UserId, items: List<OrderItem>, total: Money):
    super()                              // sets eventId, occurredAt
    self.eventType  = "order.placed"
    self.orderId    = orderId
    self.customerId = customerId
    self.items      = items              // snapshot of items at this moment
    self.total      = total
}
```

### OrderConfirmed

Emitted when an order transitions from PENDING to CONFIRMED.

```
Event OrderConfirmed extends DomainEvent {
  orderId:     OrderId
  confirmedAt: Timestamp

  construct(orderId: OrderId):
    super()
    self.eventType   = "order.confirmed"
    self.orderId     = orderId
    self.confirmedAt = self.occurredAt
}
```

### PaymentFailed

Emitted when a payment attempt for an order was rejected.

```
Event PaymentFailed extends DomainEvent {
  orderId:      OrderId
  attemptId:    String
  failureCode:  String    // e.g. "insufficient_funds", "card_expired"
  failureMessage: String  // human-readable, for logging and support

  construct(orderId: OrderId, attemptId: String, failureCode: String, failureMessage: String):
    super()
    self.eventType      = "payment.failed"
    self.orderId        = orderId
    self.attemptId      = attemptId
    self.failureCode    = failureCode
    self.failureMessage = failureMessage
}
```

---

## How events flow through the system

```
// 1. Entity accumulates events (domain layer)
Entity Order {
  events: List<DomainEvent> = []

  place(customerId, items):
    // ... enforce invariants ...
    self.events.append(OrderPlaced(self.id, customerId, items, self.total))
}

// 2. Use case reads and dispatches (application layer — NOT domain)
UseCase PlaceOrderUseCase {
  execute(command: PlaceOrderCommand) → OrderId:
    order = Order.place(command.customerId, command.items)
    orderRepository.save(order)

    for event in order.events:
      eventPublisher.publish(event)   // through an outbound port

    return order.id
}

// 3. Subscribers react (adapters / other bounded contexts)
// EventSubscriber listens for "order.placed" and triggers downstream actions:
// - send confirmation email
// - reserve inventory
// - notify warehouse
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Past tense naming | `OrderPlaced`, `OrderConfirmed`, `PaymentFailed` |
| Immutability | All fields set in constructor; no setters |
| Base type | `DomainEvent` provides `eventId`, `occurredAt`, `eventType` |
| Minimal payload | Only the data needed to describe what happened |
| Snapshot, not reference | `items` is copied at event time, not referenced |
| Entity accumulates | `order.events.append(...)` — entity does not dispatch |
| Application dispatches | Use case calls `eventPublisher.publish(event)` after saving |
