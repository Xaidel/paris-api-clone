# Application Services — Pseudocode Example

A `NotificationService` that is shared across multiple use cases.

---

## The problem it solves

Both `PlaceOrderUseCase` and `CancelOrderUseCase` need to notify customers.
Duplicating the notification logic in each use case creates drift. An application
service centralizes it.

---

## The service

```
ApplicationService NotificationService {

  // Inject outbound ports only — never adapters or infrastructure
  constructor(
    notificationGateway: NotificationGateway,   // outbound port
    userRepository:      UserRepository          // outbound port
  )

  // Send an order confirmation to the customer
  notifyOrderPlaced(orderId: String, customerId: String, total: {amount: Decimal, currency: String}):
    user = userRepository.findById(UserId(customerId))
    if user is null:
      raise NotFoundError("User not found: " + customerId)

    notificationGateway.send(Notification(
      recipient: user.email.toString(),
      subject:   "Order confirmed",
      body:      "Your order " + orderId + " for " + total.currency + " " + total.amount + " has been placed."
    ))

  // Notify about a cancellation
  notifyOrderCancelled(orderId: String, customerId: String, reason: String):
    user = userRepository.findById(UserId(customerId))
    if user is null:
      raise NotFoundError("User not found: " + customerId)

    notificationGateway.send(Notification(
      recipient: user.email.toString(),
      subject:   "Order cancelled",
      body:      "Your order " + orderId + " was cancelled. Reason: " + reason
    ))
}
```

---

## How use cases consume it

```
UseCase PlaceOrderUseCase {
  constructor(
    orderRepository:      OrderRepository,
    eventPublisher:       EventPublisher,
    notificationService:  NotificationService   // injected application service
  )

  execute(command: PlaceOrderCommand) → PlaceOrderResult:
    // ... place order logic ...
    orderRepository.save(order)

    // Delegate notification to the shared service
    notificationService.notifyOrderPlaced(
      orderId:    order.id.toString(),
      customerId: command.customerId,
      total:      {amount: order.total.amount, currency: order.total.currency}
    )

    return PlaceOrderResult(...)
}

UseCase CancelOrderUseCase {
  constructor(
    orderRepository:     OrderRepository,
    eventPublisher:      EventPublisher,
    notificationService: NotificationService
  )

  execute(command: CancelOrderCommand) → CancelOrderResult:
    // ... cancel order logic ...
    orderRepository.save(order)

    notificationService.notifyOrderCancelled(
      orderId:    command.orderId,
      customerId: order.customerId.toString(),
      reason:     command.reason
    )

    return CancelOrderResult(...)
}
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Shared logic across use cases | `NotificationService` used by both `PlaceOrder` and `CancelOrder` |
| Port injection | `notificationGateway`, `userRepository` injected — never instantiated |
| No transport knowledge | Service sends a plain `Notification` — not an HTTP request or email SDK call |
| No business rules | Service orchestrates, it does not decide domain behavior |
| Dependency direction | Use cases depend on `NotificationService`; service depends only on ports |
