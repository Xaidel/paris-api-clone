# Domain Services

A **domain service** encapsulates business logic that does not naturally belong
to a single entity or value object. It is stateless and operates only on domain
types passed to it as arguments.

---

## Rules

1. **Stateless** — no instance state; all inputs come in as parameters
2. **Pure domain** — operates only on domain entities and value objects
3. **No I/O** — never reads from or writes to external systems
4. **No ports** — does not call repositories, gateways, or any outbound port
5. **Named for what it does** — `PricingService`, `TaxCalculator`, `InventoryAllocator`

---

## When to use a domain service vs an entity method

Use an entity method when:
- The logic belongs to the lifecycle of that entity
- Only one entity is involved

Use a domain service when:
- The logic involves multiple entities of different types
- The logic doesn't clearly "belong to" any one entity
- The logic is a pure calculation (tax rate, pricing, allocation)

---

## When NOT to use a domain service

Do not use a domain service to:
- Call a repository (that's an application use case)
- Call an external API (that's an outbound adapter)
- Transform HTTP requests (that's an inbound adapter)

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
