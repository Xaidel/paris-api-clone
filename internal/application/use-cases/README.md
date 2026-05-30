# Use Cases

A **use case** is a single, named operation that a user or system can perform.
It represents one unit of application behavior.

---

## Rules

1. One use case per file, per class/function
2. One `execute()` method — the only public entry point
3. Input is a plain command or query object (primitives only — no domain objects)
4. Output is a plain result object (primitives or DTOs — no domain objects leaked out)
5. No business logic — delegate to domain entities and domain services
6. Outbound ports are injected, never instantiated inside the use case
7. Always publish domain events collected from entities after persisting

---

## Command vs Query

| Type | Mutates state | Returns |
|---|---|---|
| Command | Yes | Success confirmation or ID |
| Query | No | Data (DTO or list of DTOs) |

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
