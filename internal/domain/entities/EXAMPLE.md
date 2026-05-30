# Entity — Pseudocode Example

This example illustrates an `Order` entity. The pseudocode is intentionally
language-agnostic — translate the intent, not the syntax, into your language.

---

## Value Objects used by this entity

```
// OrderId: wraps a UUID string. Validates format on construction.
ValueObject OrderId {
  value: String

  construct(raw: String):
    if raw is not a valid UUID format → raise DomainError("Invalid OrderId format")
    self.value = raw

  equals(other: OrderId) → Boolean:
    return self.value == other.value

  toString() → String:
    return self.value
}

// Money: amount + currency. Immutable. Validates on construction.
ValueObject Money {
  amount: Decimal
  currency: String   // ISO 4217, e.g. "USD", "EUR"

  construct(amount: Decimal, currency: String):
    if amount < 0         → raise DomainError("Money amount cannot be negative")
    if currency is empty  → raise DomainError("Currency is required")
    self.amount   = amount
    self.currency = currency

  add(other: Money) → Money:
    if self.currency != other.currency → raise DomainError("Currency mismatch")
    return Money(self.amount + other.amount, self.currency)

  equals(other: Money) → Boolean:
    return self.amount == other.amount AND self.currency == other.currency
}

// OrderStatus: an enum / discriminated union
Enum OrderStatus {
  PENDING
  CONFIRMED
  SHIPPED
  CANCELLED
}
```

---

## The Entity

```
// Order: has identity (OrderId). Enforces invariants. Emits domain events.
Entity Order {
  id:        OrderId         // identity — immutable after construction
  items:     List<OrderItem> // value objects
  status:    OrderStatus
  total:     Money
  events:    List<DomainEvent>  // accumulated, not dispatched from here

  // --- Factory / constructor ---

  construct(id: OrderId, items: List<OrderItem>):
    if items is empty → raise DomainError("An order must have at least one item")

    self.id     = id
    self.items  = items
    self.status = OrderStatus.PENDING
    self.total  = sum of item.price for each item in items
    self.events = []

    // Emit event to signal creation
    self.events.append(OrderPlaced(orderId: id, items: items, total: self.total))

  // --- Behavior ---

  confirm():
    if self.status != PENDING → raise DomainError("Only pending orders can be confirmed")
    self.status = CONFIRMED
    self.events.append(OrderConfirmed(orderId: self.id))

  cancel(reason: String):
    if self.status == SHIPPED    → raise DomainError("Cannot cancel a shipped order")
    if self.status == CANCELLED  → raise DomainError("Order is already cancelled")
    self.status = CANCELLED
    self.events.append(OrderCancelled(orderId: self.id, reason: reason))

  // --- Identity ---

  equals(other: Order) → Boolean:
    return self.id.equals(other.id)  // identity equality, not value equality
}

// OrderItem: a value object embedded in Order
ValueObject OrderItem {
  productId:   String
  productName: String
  quantity:    Integer
  price:       Money

  construct(productId, productName, quantity, price):
    if quantity <= 0 → raise DomainError("Quantity must be positive")
    // assign fields
}
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Identity via value object | `OrderId` wraps a UUID; equality uses it |
| Invariant enforcement | `items is empty` check in constructor; `status` checks in methods |
| Immutable identity | `id` is set once in constructor and never changed |
| Event accumulation | `self.events.append(...)` — events collected, not dispatched |
| No I/O | Entity never reads from or writes to any external system |
| Value object for money | `Money` carries both amount and currency; prevents currency bugs |
| Behavior on entity | `confirm()`, `cancel()` live on `Order`, not on a service |

---

## What the application layer does with events

The application layer (use case) reads `order.events` after calling a behavior,
then dispatches those events through an outbound port:

```
// In PlaceOrderUseCase (application layer — NOT here):
order = Order.construct(id, items)
orderRepository.save(order)
for event in order.events:
    eventPublisher.publish(event)
```

The domain entity never calls `eventPublisher` directly. It only accumulates.
