# Unit Tests

Unit tests cover **domain** and **application** layer logic — the code that
contains business rules and orchestration.

---

## What to test here

| Target | How |
|---|---|
| Entity invariants | Construct entities with invalid inputs; assert domain errors raised |
| Entity behavior | Call entity methods; assert state changes and events emitted |
| Value object validation | Construct with invalid values; assert errors |
| Value object operations | Test arithmetic, equality, formatting |
| Domain services | Call with domain objects; assert correct output |
| Use cases | Inject in-memory port implementations; assert orchestration correctness |

---

## Rules

1. **No I/O** — no database, no network, no file system, no external services
2. **No framework** — no HTTP server, no gRPC server
3. **In-memory test doubles** — substitute outbound ports with in-memory implementations
4. **Fast** — every test should complete in milliseconds
5. **Isolated** — each test is independent; no shared mutable state between tests

---

## Test double pattern

Use **in-memory implementations** of outbound ports, not mocks/stubs that
just record calls.

```
// In-memory OrderRepository for tests — implements the same port the real adapter does
InMemoryOrderRepository implements OrderRepository {
  store: Map<String, Order> = {}

  save(order: Order) → void:
    store[order.id.toString()] = order

  findById(id: OrderId) → Order | null:
    return store[id.toString()] or null

  findByCustomerId(customerId: UserId) → List<Order>:
    return store.values().filter(o → o.customerId.equals(customerId))
}
```

In-memory implementations:
- Are reusable across tests
- Actually exercise the interface contract
- Are fast (no I/O)
- Can be initialized with fixture data

---

## Example test structure (pseudocode)

```
TestSuite PlaceOrderUseCaseTests {

  setup():
    orderRepo     = InMemoryOrderRepository()
    eventPublisher = InMemoryEventPublisher()
    pricingService = PricingService()   // real domain service — no I/O
    useCase = PlaceOrderUseCase(orderRepo, eventPublisher, pricingService)

  test "places an order with valid items":
    command = PlaceOrderCommand(
      customerId: "user-123",
      items: [{productId: "prod-1", quantity: 2, priceAmount: 9.99, priceCurrency: "USD"}]
    )
    result = useCase.execute(command)

    assert result.orderId is not null
    assert result.status == "PENDING"
    assert result.total.currency == "USD"

    savedOrder = orderRepo.findById(OrderId(result.orderId))
    assert savedOrder is not null
    assert savedOrder.items.length == 1

    publishedEvents = eventPublisher.published
    assert publishedEvents.length == 1
    assert publishedEvents[0].eventType == "order.placed"

  test "raises error when items list is empty":
    command = PlaceOrderCommand(customerId: "user-123", items: [])
    assert raises DomainError: useCase.execute(command)

  test "raises error when customerId is invalid":
    command = PlaceOrderCommand(customerId: "", items: [...])
    assert raises DomainError: useCase.execute(command)
}

TestSuite OrderEntityTests {

  test "emits OrderPlaced event on construction":
    order = Order.construct(OrderId.generate(), UserId("u1"), [validItem()], Money(10, "USD"))
    assert order.events.length == 1
    assert order.events[0] is OrderPlaced

  test "raises when confirming a non-pending order":
    order = Order.construct(...)
    order.cancel("test")
    assert raises DomainError: order.confirm()

  test "cannot be constructed with empty items":
    assert raises DomainError: Order.construct(OrderId.generate(), UserId("u1"), [], Money(0, "USD"))
}
```
