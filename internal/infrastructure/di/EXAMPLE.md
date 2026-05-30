# Dependency Injection — Pseudocode Example

The composition root wires all layers together. This is the only file in the
entire codebase that knows about every layer's concrete classes simultaneously.

---

## The wiring function

```
// Wire all components and return a configured, ready-to-start Application.
// Called once at startup from main().
Function wire(config: AppConfig, logger: Logger, tracer: Tracer) → Application:

  // ── 1. Infrastructure: external connections ──────────────────────────────

  db = PostgresConnection(
    connectionString: config.database.connectionString(),
    maxConns:         config.database.maxConns
  )

  kafkaProducer = KafkaProducer(
    brokers:       config.kafka.brokers,
    transactional: true
  )

  kafkaConsumer = KafkaConsumer(
    brokers:       config.kafka.brokers,
    consumerGroup: config.kafka.consumerGroup
  )

  // ── 2. Outbound adapters (implement outbound ports) ───────────────────────

  orderRepository  = PostgresOrderRepository(db)           // implements OrderRepository
  userRepository   = PostgresUserRepository(db)            // implements UserRepository
  paymentGateway   = StripePaymentGateway(                 // implements PaymentGateway
    stripeClient: StripeClient(config.stripe.apiKey),
    config:       config.stripe
  )
  sessionStore     = RedisSessionStore(                     // implements SessionStore
    redisClient: RedisClient(config.redis.url)
  )
  eventPublisher   = KafkaEventPublisher(                   // implements EventPublisher
    producer: kafkaProducer,
    config:   config.kafka
  )
  notificationGw   = SendGridEmailGateway(                  // implements NotificationGateway
    client: SendGridClient(config.sendgrid.apiKey)
  )

  // ── 3. Domain services (pure, no I/O — no injection needed) ──────────────

  pricingService      = PricingService()
  inventoryAllocator  = InventoryAllocator()

  // ── 4. Application services ───────────────────────────────────────────────

  notificationService = NotificationService(
    notificationGateway: notificationGw,
    userRepository:      userRepository
  )

  // ── 5. Use cases (application layer) ─────────────────────────────────────

  placeOrderUseCase  = PlaceOrderUseCase(       // implements PlaceOrderPort
    orderRepository:     orderRepository,
    paymentGateway:      paymentGateway,
    eventPublisher:      eventPublisher,
    pricingService:      pricingService,
    notificationService: notificationService
  )

  getOrderUseCase    = GetOrderUseCase(         // implements GetOrderPort
    orderRepository: orderRepository
  )

  cancelOrderUseCase = CancelOrderUseCase(      // implements CancelOrderPort
    orderRepository:     orderRepository,
    eventPublisher:      eventPublisher,
    notificationService: notificationService
  )

  // ── 6. Inbound adapters ───────────────────────────────────────────────────

  // HTTP adapter: inject ports (use cases), not concrete classes
  httpOrderAdapter = HttpOrderAdapter(
    placeOrderPort: placeOrderUseCase,   // PlaceOrderUseCase implements PlaceOrderPort
    getOrderPort:   getOrderUseCase,
    cancelOrderPort: cancelOrderUseCase
  )

  // gRPC adapter (same use cases, different transport)
  grpcOrderAdapter = GrpcOrderAdapter(
    placeOrderPort: placeOrderUseCase,
    getOrderPort:   getOrderUseCase
  )

  // Kafka consumer (message queue → use case)
  kafkaOrderConsumer = KafkaOrderConsumer(
    placeOrderPort: placeOrderUseCase,
    consumer:       kafkaConsumer,
    topic:          config.kafka.orderEventsTopic
  )

  // ── 7. Servers / transports ───────────────────────────────────────────────

  httpServer = HttpServer(
    address: config.server.address(),
    routes: [
      Route(POST, "/orders",      httpOrderAdapter.handlePlaceOrder),
      Route(GET,  "/orders/{id}", httpOrderAdapter.handleGetOrder),
      Route(DELETE, "/orders/{id}", httpOrderAdapter.handleCancelOrder),
      Route(GET,  "/health",      healthCheckHandler(db, kafkaProducer))
    ]
  )

  grpcServer = GrpcServer(
    address: config.grpc.address(),
    services: [
      grpcOrderAdapter.registerWith(grpcServer)
    ]
  )

  // ── 8. Return runnable application ───────────────────────────────────────

  return Application(
    servers:   [httpServer, grpcServer],
    consumers: [kafkaOrderConsumer],
    onShutdown: Function():
      db.close()
      kafkaProducer.close()
      kafkaConsumer.close()
  )
```

---

## The main entrypoint

```
// main.go / main.py / main.rs / index.ts — language-idiomatic entrypoint

Function main():

  // 1. Load and validate config — fail fast
  config = loadConfig()

  // 2. Initialize observability before anything else
  logger = initLogger(config.logLevel, config.environment)
  tracer = initTracer(config.environment)
  metrics = initMetrics(config.environment)

  // 3. Wire all components
  app = wire(config, logger, tracer)

  // 4. Start all servers and consumers
  app.start()

  // 5. Wait for shutdown signal (SIGTERM, SIGINT)
  waitForSignal()

  // 6. Graceful shutdown
  app.shutdown(timeoutSec: config.server.shutdownTimeout)

  logger.info("Application shutdown complete")
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Single composition root | `wire()` is the only place all layers are assembled |
| Constructor injection | Every component receives dependencies, never instantiates them |
| No global singletons | `db`, `kafkaProducer` are passed as arguments, not globals |
| Port, not class | `httpOrderAdapter` receives `PlaceOrderPort`, not `PlaceOrderUseCase` |
| Graceful shutdown | `onShutdown` closes connections before process exits |
| Fail fast | `loadConfig()` raises on missing env vars before `wire()` runs |
| Observability first | Logger and tracer initialized before any adapter or use case |
| Multiple transports, same use cases | HTTP and gRPC adapters both use the same `placeOrderUseCase` |
