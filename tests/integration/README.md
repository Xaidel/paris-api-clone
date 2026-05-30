# Integration Tests

Integration tests validate that **adapters correctly implement their port contracts**
against real or containerized external systems.

---

## What to test here

| Target | How |
|---|---|
| Outbound adapters | Test against a real DB, cache, or queue (Docker container in CI) |
| Inbound adapters | Start a test server; call via real HTTP/gRPC/WS client |
| Port contract compliance | Assert the adapter correctly implements the full port interface |

---

## Rules

1. **Test the adapter, not the business logic** — business rules are covered by unit tests
2. **Use real infrastructure** — containers (Postgres, Redis, Kafka) via Docker Compose or testcontainers
3. **Test the full port surface** — every method on the port interface must be tested
4. **Isolate test data** — each test runs in a transaction that is rolled back, or uses unique IDs
5. **No use cases** — integration tests call adapters directly through the port interface

---

## Outbound adapter test pattern (pseudocode)

```
TestSuite PostgresOrderRepositoryTests {

  // Spin up Postgres before the suite, tear down after
  setup():
    db = PostgresConnection(testDatabaseUrl())
    db.runMigrations()
    repo = PostgresOrderRepository(db)

  teardown():
    db.close()

  beforeEach():
    db.beginTransaction()   // each test runs in a transaction

  afterEach():
    db.rollback()           // rolled back — no state bleeds between tests

  test "saves and retrieves an order by id":
    order = buildTestOrder()
    repo.save(order)

    found = repo.findById(order.id)
    assert found is not null
    assert found.id.equals(order.id)
    assert found.status == order.status
    assert found.items.length == order.items.length

  test "returns null for unknown id":
    found = repo.findById(OrderId.generate())
    assert found is null

  test "finds orders by customer id":
    order1 = buildTestOrder(customerId: "cust-1")
    order2 = buildTestOrder(customerId: "cust-1")
    order3 = buildTestOrder(customerId: "cust-2")
    repo.save(order1)
    repo.save(order2)
    repo.save(order3)

    results = repo.findByCustomerId(UserId("cust-1"))
    assert results.length == 2

  test "updates status on re-save":
    order = buildTestOrder()
    repo.save(order)
    order.confirm()
    repo.save(order)

    found = repo.findById(order.id)
    assert found.status == CONFIRMED
}
```

---

## Inbound adapter test pattern (pseudocode)

```
TestSuite HttpOrderAdapterTests {

  setup():
    // Wire with real use cases backed by in-memory ports
    // (not full production infra — just enough to test the HTTP layer)
    orderRepo  = InMemoryOrderRepository()
    publisher  = InMemoryEventPublisher()
    useCase    = PlaceOrderUseCase(orderRepo, publisher, PricingService())
    adapter    = HttpOrderAdapter(placeOrderPort: useCase)
    testServer = startTestHttpServer(adapter)

  teardown():
    testServer.stop()

  test "POST /orders returns 201 with valid body":
    response = httpClient.post(testServer.url + "/orders", body: {
      customer_id: "user-123",
      items: [{product_id: "prod-1", quantity: 1, price: {amount: 9.99, currency: "USD"}}]
    })
    assert response.status == 201
    assert response.body.order_id is not null
    assert response.body.status == "PENDING"

  test "POST /orders returns 400 when customer_id is missing":
    response = httpClient.post(testServer.url + "/orders", body: {
      items: [...]
    })
    assert response.status == 400
    assert response.body.error contains "customer_id"

  test "POST /orders returns 422 when items is empty":
    response = httpClient.post(testServer.url + "/orders", body: {
      customer_id: "user-123",
      items: []
    })
    assert response.status == 422
}
```
