# Value Objects — Pseudocode Examples

Three value objects demonstrating different patterns: a simple wrapper, a
compound value, and an identifier.

---

## 1. Simple wrapper — Email

Wraps a primitive (`String`) and enforces a domain constraint (valid email format).

```
ValueObject Email {
  value: String

  construct(raw: String):
    trimmed = raw.trim()
    if trimmed is empty           → raise DomainError("Email cannot be empty")
    if trimmed does not match
       basic email pattern        → raise DomainError("Invalid email format: " + raw)
    self.value = trimmed.lowercase()

  equals(other: Email) → Boolean:
    return self.value == other.value

  toString() → String:
    return self.value
}
```

**Usage**:
```
// Valid
email = Email("User@Example.com")  // stored as "user@example.com"

// Invalid — raises DomainError at construction time, not at use time
email = Email("not-an-email")      // raises DomainError("Invalid email format")
email = Email("")                   // raises DomainError("Email cannot be empty")
```

---

## 2. Compound value — Money

Combines two related primitives that must always travel together.

```
ValueObject Money {
  amount:   Decimal
  currency: String   // ISO 4217 code, e.g. "USD", "EUR", "GBP"

  construct(amount: Decimal, currency: String):
    if amount < 0           → raise DomainError("Amount cannot be negative")
    if currency.length != 3 → raise DomainError("Currency must be a 3-letter ISO 4217 code")
    self.amount   = amount.roundTo(2 decimal places)
    self.currency = currency.uppercase()

  add(other: Money) → Money:
    if self.currency != other.currency:
      raise DomainError("Cannot add " + self.currency + " and " + other.currency)
    return Money(self.amount + other.amount, self.currency)

  subtract(other: Money) → Money:
    if self.currency != other.currency:
      raise DomainError("Cannot subtract different currencies")
    result = self.amount - other.amount
    if result < 0 → raise DomainError("Result would be negative")
    return Money(result, self.currency)

  isGreaterThan(other: Money) → Boolean:
    if self.currency != other.currency:
      raise DomainError("Cannot compare different currencies")
    return self.amount > other.amount

  equals(other: Money) → Boolean:
    return self.amount == other.amount AND self.currency == other.currency

  toString() → String:
    return self.currency + " " + self.amount.format(2 decimals)

  // Money.zero(currency) — factory for a neutral element
  static zero(currency: String) → Money:
    return Money(0, currency)
}
```

**Usage**:
```
price    = Money(9.99, "USD")
tax      = Money(0.80, "USD")
total    = price.add(tax)        // Money(10.79, "USD")
shipping = Money(5.00, "EUR")

total.add(shipping)              // raises DomainError — currency mismatch
```

---

## 3. Entity identifier — OrderId

Wraps a UUID to give identifiers a meaningful type, preventing accidental mixing
of identifiers from different entity types.

```
ValueObject OrderId {
  value: String   // UUID v4

  construct(raw: String):
    if raw does not match UUID v4 pattern:
      raise DomainError("Invalid OrderId: " + raw)
    self.value = raw.lowercase()

  equals(other: OrderId) → Boolean:
    return self.value == other.value

  toString() → String:
    return self.value

  // OrderId.generate() — factory that produces a new unique ID
  static generate() → OrderId:
    return OrderId(generateUUID())
}
```

**Why a typed ID?**

Without typed IDs, this is silently valid:

```
ship(orderId: String, userId: String, warehouseId: String)

// This compiles but is semantically wrong:
ship(userId, warehouseId, orderId)
```

With typed IDs, the type system catches the mistake:

```
ship(orderId: OrderId, userId: UserId, warehouseId: WarehouseId)

// This now fails at compile time or construction time:
ship(UserId("..."), WarehouseId("..."), OrderId("..."))
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Validation at construction | All three: invalid values raise `DomainError` immediately |
| Immutability | Operations (`add`, `subtract`) return new instances, never mutate |
| Value equality | `equals()` compares fields, not references |
| No identity field | None of these have an `id` — they ARE values |
| Primitive obsession fix | `Email` vs `String`, `Money` vs `Float`, `OrderId` vs `String` |
| Cross-currency guard | `Money.add()` and `Money.subtract()` enforce currency homogeneity |
| Static factories | `Money.zero()`, `OrderId.generate()` — clean construction API |
