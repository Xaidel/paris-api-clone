# Observability — Pseudocode Example

Structured logging, distributed tracing, and metrics initialization.

---

## 1. Structured logger

```
// Initialize a structured logger. Returns a Logger that all components use.
Function initLogger(logLevel: String, environment: String) → Logger:

  level = parseLogLevel(logLevel)   // "debug" → DEBUG, "info" → INFO, etc.

  logger = StructuredLogger(
    level:  level,
    format: if environment == "development" then PRETTY else JSON,
    fields: {
      service:     "my-service",
      environment: environment,
      version:     APP_VERSION      // injected at build time
    }
  )

  // Usage example (not executed here — illustrates how it's used elsewhere):
  //
  // logger.info("Order placed", {order_id: "abc", customer_id: "xyz"})
  // → {"level":"info","msg":"Order placed","order_id":"abc","customer_id":"xyz",
  //    "service":"my-service","trace_id":"...","span_id":"...","ts":"..."}

  return logger
```

---

## 2. Distributed tracing (OpenTelemetry)

```
// Initialize OpenTelemetry tracing. Returns a Tracer.
Function initTracer(environment: String) → Tracer:

  exporter = if environment == "production":
    OTLPExporter(endpoint: requireEnv("OTEL_EXPORTER_OTLP_ENDPOINT"))
  else:
    ConsoleExporter()   // print spans to stdout in dev/test

  provider = TracerProvider(
    resource: Resource({
      service.name:    "my-service",
      service.version: APP_VERSION,
      deployment.environment: environment
    }),
    sampler:  ParentBasedSampler(root: AlwaysOnSampler()),
    exporter: exporter
  )

  // Register as global tracer
  GlobalTracer.set(provider)

  return provider.getTracer("my-service")

  // Usage example (not executed here):
  //
  // span = tracer.startSpan("PlaceOrderUseCase.execute")
  // span.setAttribute("order.id", orderId)
  // try:
  //   ...
  //   span.setStatus(OK)
  // catch e:
  //   span.recordException(e)
  //   span.setStatus(ERROR)
  // finally:
  //   span.end()
```

---

## 3. Metrics (Prometheus)

```
// Initialize Prometheus metrics. Returns a MetricsRegistry.
Function initMetrics(environment: String) → MetricsRegistry:

  registry = PrometheusRegistry()

  // Pre-register standard metrics (add application-specific ones as needed)
  registry.register(Counter(
    name:   "http_requests_total",
    help:   "Total HTTP requests received",
    labels: ["method", "path", "status_code"]
  ))

  registry.register(Histogram(
    name:    "http_request_duration_seconds",
    help:    "HTTP request duration in seconds",
    labels:  ["method", "path"],
    buckets: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5]
  ))

  registry.register(Counter(
    name:   "domain_events_published_total",
    help:   "Total domain events published",
    labels: ["event_type"]
  ))

  registry.register(Counter(
    name:   "use_case_errors_total",
    help:   "Total use case errors by type",
    labels: ["use_case", "error_type"]
  ))

  // Expose /metrics endpoint (registered in HTTP server in di/)
  metricsHandler = PrometheusHttpHandler(registry)

  return registry
```

---

## 4. Health check handler

```
// Returns a health check handler that validates live dependencies.
// Registered at GET /health or GET /healthz in the HTTP server.
Function healthCheckHandler(db: DatabaseConnection, kafka: KafkaProducer) → HttpHandler:

  return Function(request: HttpRequest) → HttpResponse:
    checks = {}

    // Check database connectivity
    try:
      db.ping()
      checks["database"] = "ok"
    catch Exception as e:
      checks["database"] = "error: " + e.message

    // Check Kafka connectivity
    try:
      kafka.ping()
      checks["kafka"] = "ok"
    catch Exception as e:
      checks["kafka"] = "error: " + e.message

    allOk = all(v == "ok" for v in checks.values())

    return HttpResponse(
      status: if allOk then 200 else 503,
      body:   {status: if allOk then "ok" else "degraded", checks: checks}
    )
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Structured logging | JSON format in production; pretty format in development |
| Observability-first | All three initialized before `wire()` runs |
| Standard fields | `service`, `environment`, `version` on every log line |
| OpenTelemetry | Vendor-agnostic tracing; swap exporter without changing code |
| Pre-registered metrics | Counter and histogram defined up front — not scattered in adapters |
| Health checks | Live dependency probes (not just "is the process alive") |
| No business logic | Config and wiring only — no domain decisions |
