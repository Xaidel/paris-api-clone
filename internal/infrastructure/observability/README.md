# Observability

The `observability/` directory bootstraps **structured logging**, **distributed
tracing**, and **metrics** before the application starts serving requests.

---

## Rules

1. Initialize observability before any adapter or use case runs
2. Structured logging only — no `print()` or unstructured log strings
3. Correlation IDs and trace IDs attached to every log line
4. No business logic in observability setup
5. Health check endpoints live here (or in `di/`) — not in application layer

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
