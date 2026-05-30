# Entities

An **entity** is a domain object with a **unique identity** that persists across
state changes. Two entities are equal if and only if they share the same identity —
regardless of their other attributes.

---

## Rules

1. Every entity has an identity (typically a value object like `OrderId`)
2. The entity enforces its own invariants — invalid state must be unrepresentable
3. Behavior lives on the entity, not in external services
4. The entity emits domain events when significant things happen
5. Entities never perform I/O

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
