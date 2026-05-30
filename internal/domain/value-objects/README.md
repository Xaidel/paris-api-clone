# Value Objects

A **value object** is an immutable domain concept whose identity is determined
entirely by its value. Two value objects with the same value are interchangeable.

---

## Rules

1. **Immutable** — never mutated after construction; operations return new instances
2. **Self-validating** — invalid values cannot be constructed; validate in the constructor
3. **Value equality** — equal if all attributes are equal (no identity field)
4. **No identity** — value objects are not stored independently; they belong to entities
5. **No I/O** — value objects are pure computations

---

## Common uses

- Wrapping primitive obsession: `Email` instead of `String`, `Money` instead of `Float`
- Enforcing format constraints: `PhoneNumber`, `PostalCode`, `Url`
- Combining related values: `Address` (street + city + country), `DateRange` (start + end)
- Entity identifiers: `OrderId`, `UserId`, `ProductId`

---

## See `EXAMPLE.md` for a complete pseudocode walkthrough
