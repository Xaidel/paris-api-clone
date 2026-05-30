# Use Cases — Pseudocode Examples

Two use cases: a command (`PlaceOrderUseCase`) and a query (`GetOrderUseCase`).

---

## Command input / output types

```
// Command: plain data in — primitives only, no domain objects
Command PlaceOrderCommand {
  customerId: String
  items: List<{productId: String, quantity: Integer, priceAmount: Decimal, priceCurrency: String}>
}

// Result: plain data out — primitives or simple DTOs
Result PlaceOrderResult {
  orderId:    String
  status:     String
  total:      {amount: Decimal, currency: String}
  occurredAt: Timestamp
}
```

---

## 1. Command Use Case — PlaceOrderUseCase

```
UseCase PlaceOrderUseCase {

  // Inject outbound ports — never concrete adapters or infrastructure
  constructor(
    orderRepository:  OrderRepository,    // outbound port
    eventPublisher:   EventPublisher,     // outbound port
    pricingService:   PricingService      // domain service (pure, no I/O)
  )

  execute(command: PlaceOrderCommand) → PlaceOrderResult:

    // 1. Validate and construct domain types from raw input
    customerId = UserId(command.customerId)

    items = []
    for each raw in command.items:
      price = Money(raw.priceAmount, raw.priceCurrency)
      items.append(OrderItem(raw.productId, raw.quantity, price))

    // 2. Apply domain service (pure business logic — no I/O)
    total = pricingService.calculateTotal(items, discounts=[])

    // 3. Create entity — entity enforces invariants, accumulates events
    orderId = OrderId.generate()
    order   = Order.construct(orderId, customerId, items, total)

    // 4. Persist via outbound port (never call adapter directly)
    orderRepository.save(order)

    // 5. Publish domain events accumulated by the entity
    for event in order.events:
      eventPublisher.publish(event)

    // 6. Return plain result — no domain objects leaked to caller
    return PlaceOrderResult(
      orderId:    order.id.toString(),
      status:     order.status.toString(),
      total:      {amount: order.total.amount, currency: order.total.currency},
      occurredAt: now()
    )
}
```

---

## 2. Query Use Case — GetOrderUseCase

```
// Query input
Query OrderQuery {
  orderId:    String
  requesterId: String   // for authorization checks
}

// Query result (a read DTO — no domain object exposed)
Result OrderResult {
  orderId:    String
  customerId: String
  status:     String
  items: List<{
    productId:   String
    productName: String
    quantity:    Integer
    price:       {amount: Decimal, currency: String}
  }>
  total:      {amount: Decimal, currency: String}
  placedAt:   Timestamp
}

// Use case
UseCase GetOrderUseCase {

  constructor(
    orderRepository: OrderRepository    // outbound port
  )

  execute(query: OrderQuery) → OrderResult:

    // 1. Load from outbound port
    orderId = OrderId(query.orderId)
    order   = orderRepository.findById(orderId)

    if order is null:
      raise NotFoundError("Order not found: " + query.orderId)

    // 2. (Optional) authorization check — still no transport logic
    if order.customerId.toString() != query.requesterId:
      raise ForbiddenError("Access denied")

    // 3. Map domain entity to read DTO — no domain object exposed
    return OrderResult(
      orderId:    order.id.toString(),
      customerId: order.customerId.toString(),
      status:     order.status.toString(),
      items:      order.items.map(item → {
                    productId:   item.productId,
                    productName: item.productName,
                    quantity:    item.quantity,
                    price:       {amount: item.price.amount, currency: item.price.currency}
                  }),
      total:      {amount: order.total.amount, currency: order.total.currency},
      placedAt:   order.placedAt
    )
}
```

---

## What calls a use case

A use case is called by an **inbound adapter** — never directly by another use case.

```
// In HttpOrderAdapter (inbound adapter — NOT in application/):
HttpOrderAdapter {
  POST /orders → handler(httpRequest):
    command = PlaceOrderCommand(
      customerId: httpRequest.body.customer_id,
      items:      httpRequest.body.items
    )
    result = placeOrderUseCase.execute(command)
    return HttpResponse(201, {order_id: result.orderId})
}
```

The adapter translates transport → command, calls the use case, then translates
result → transport response. The use case knows nothing about HTTP.

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Primitives in, primitives out | `PlaceOrderCommand` and `PlaceOrderResult` |
| No domain objects in/out | Domain objects (`Order`, `OrderId`) live only inside the use case |
| Port injection | `orderRepository`, `eventPublisher` injected — never instantiated |
| Domain logic in domain | `Order.construct()` and `pricingService.calculateTotal()` do the work |
| Event dispatch after persist | `for event in order.events: eventPublisher.publish(event)` |
| No transport knowledge | No HTTP, no gRPC, no status codes — adapter handles that |
| Application error types | `NotFoundError`, `ForbiddenError` — not HTTP 404/403 |
