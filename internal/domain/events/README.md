# Domain Events

A **domain event** is an immutable record of something significant that happened
in the domain. Events are facts — they cannot be changed after they occur.

---

## Rules

1. **Past tense naming** — `OrderPlaced`, not `PlaceOrder` or `OrderPlace`
2. **Immutable** — no setters, no mutation after construction
3. **Carry minimal data** — only what's needed to describe what happened
4. **No behavior** — events are data carriers, not behavior carriers
5. **Emitted by entities** — entities accumulate events; the application layer dispatches them
6. **No I/O** — events are plain data structures

---

## When to emit an event

Emit a domain event when:

- An entity changes state in a way that other parts of the system might care about
- A significant business fact has occurred (not just a CRUD operation)

Examples:
- `OrderPlaced` — a new order entered the system
- `PaymentFailed` — a payment attempt was rejected
- `UserDeactivated` — a user account was disabled
- `InventoryDepleted` — stock of a product reached zero

---

## When NOT to emit an event

- For internal, non-significant state transitions
- For every field update (not every setter needs an event)
- When no other part of the system would ever need to react

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
