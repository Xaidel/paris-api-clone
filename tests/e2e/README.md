# End-to-End Tests

End-to-end (E2E) tests validate **complete user scenarios** through the full
running application stack.

---

## What to test here

| Scenario | What it covers |
|---|---|
| "A customer places an order" | HTTP in → use case → DB → event published |
| "An order is confirmed after payment" | Payment webhook → use case → DB → notification sent |
| "A user authenticates and retrieves their orders" | Auth → session → query → response |

---

## Rules

1. **Test scenarios, not implementation** — E2E tests don't know about layers
2. **Full stack** — real HTTP/gRPC client, real DB, real queue (all containerized)
3. **Minimal set** — only cover critical user journeys; unit tests cover edge cases
4. **Slow by nature** — acceptable; don't try to make E2E tests fast by cutting corners
5. **Isolated data** — each test uses unique identifiers; clean up after or use separate DB

---

## E2E test pattern (pseudocode)

```
TestSuite OrderJourneyE2ETests {

  // Start the full application before the suite
  setup():
    testInfra = DockerCompose.start("docker-compose.test.yml")  // Postgres, Kafka, Redis
    app       = startApplication(testConfigFrom(testInfra))
    client    = HttpClient(baseUrl: app.address())

  teardown():
    app.stop()
    testInfra.stop()

  test "customer places an order end-to-end":
    // 1. Place order via HTTP
    response = client.post("/orders", body: {
      customer_id: "user-" + generateUUID(),
      items: [{
        product_id:    "prod-001",
        quantity:      2,
        price: {amount: 29.99, currency: "USD"}
      }]
    })
    assert response.status == 201
    orderId = response.body.order_id

    // 2. Retrieve the order
    getResponse = client.get("/orders/" + orderId)
    assert getResponse.status == 200
    assert getResponse.body.order_id == orderId
    assert getResponse.body.status == "PENDING"
    assert getResponse.body.items.length == 1

    // 3. Verify domain event was published to Kafka
    event = kafkaTestConsumer.consumeNext(
      topic:   "orders.events",
      timeout: 5 seconds
    )
    assert event.event_type == "order.placed"
    assert event.order_id == orderId

  test "order cannot be placed with zero items":
    response = client.post("/orders", body: {
      customer_id: "user-" + generateUUID(),
      items: []
    })
    assert response.status == 422
    assert response.body.error is not null
}
```

---

## What E2E tests are NOT

- A replacement for unit or integration tests
- A way to test edge cases (there are too many)
- A fast feedback loop (they are slow by design)

Run E2E tests in CI on every pull request against a full containerized stack.
Run unit and integration tests locally and in CI on every commit.
