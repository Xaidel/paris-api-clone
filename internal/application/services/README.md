# Application Services

An **application service** handles cross-cutting concerns that span multiple
use cases, or coordinates workflows that don't fit neatly into a single named action.

---

## Rules

1. Same import rules as use cases — no transport, no DB, no adapters
2. Stateless where possible
3. Injected with outbound ports in constructor
4. Used sparingly — prefer use cases for most logic

---

## When to use an application service

Use an application service when:
- A concern (e.g. notification dispatch) is shared across multiple use cases
- A workflow involves multiple sequential use cases with shared state
- Cross-cutting logic (e.g. audit logging, idempotency checks) needs a home

**If in doubt, use a use case** — application services are the exception, not the rule.

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
