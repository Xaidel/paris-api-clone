# Go — Domain Layer Examples

Real Go code illustrating `src/domain/` patterns.
All examples use the Order / Payment domain.

---

## Errors (`src/domain/errors.go`)

```go
package domain

import "fmt"

// DomainError is the base struct for all typed domain errors.
// Use errors.As() in adapters to distinguish them from infrastructure errors.
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Sentinel errors for conditions that carry no extra data.
var (
	ErrInvalidOrderItems = &DomainError{Code: "INVALID_ORDER_ITEMS", Message: "order must have at least one item"}
	ErrInvalidOrderState = &DomainError{Code: "INVALID_ORDER_STATE", Message: "state transition not permitted"}
)

// OrderNotFoundError carries the ID that was not found.
type OrderNotFoundError struct {
	OrderID string
}

func (e *OrderNotFoundError) Error() string {
	return fmt.Sprintf("order %s not found", e.OrderID)
}

// CurrencyMismatchError carries the two mismatched currencies.
type CurrencyMismatchError struct {
	Left  string
	Right string
}

func (e *CurrencyMismatchError) Error() string {
	return fmt.Sprintf("currency mismatch: %s vs %s", e.Left, e.Right)
}
```

---

## Value Objects

### `src/domain/order_id.go`

```go
package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// OrderID is a typed wrapper around a UUID.
// Unexported field prevents direct construction — use NewOrderID or OrderIDFromString.
type OrderID struct {
	value uuid.UUID
}

// NewOrderID generates a new random OrderID.
func NewOrderID() OrderID {
	return OrderID{value: uuid.New()}
}

// OrderIDFromString parses a UUID string into an OrderID.
func OrderIDFromString(s string) (OrderID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return OrderID{}, fmt.Errorf("invalid order id %q: %w", s, err)
	}
	return OrderID{value: id}, nil
}

func (id OrderID) String() string          { return id.value.String() }
func (id OrderID) Equal(other OrderID) bool { return id.value == other.value }
```

### `src/domain/user_id.go`

```go
package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

func NewUserID() UserID {
	return UserID{value: uuid.New()}
}

func UserIDFromString(s string) (UserID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("invalid user id %q: %w", s, err)
	}
	return UserID{value: id}, nil
}

func (id UserID) String() string          { return id.value.String() }
func (id UserID) Equal(other UserID) bool { return id.value == other.value }
```

### `src/domain/money.go`

```go
package domain

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Money is a value type. All methods use value receivers — it is never mutated.
type Money struct {
	amount   decimal.Decimal
	currency string
}

// NewMoney creates Money after validating amount and currency.
func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
	if amount.IsNegative() {
		return Money{}, &DomainError{Code: "INVALID_MONEY", Message: "amount cannot be negative"}
	}
	if currency == "" {
		return Money{}, &DomainError{Code: "INVALID_MONEY", Message: "currency is required"}
	}
	return Money{amount: amount, currency: currency}, nil
}

// ZeroMoney returns a zero-value Money for the given currency.
func ZeroMoney(currency string) Money {
	return Money{amount: decimal.Zero, currency: currency}
}

func (m Money) Amount() decimal.Decimal { return m.amount }
func (m Money) Currency() string        { return m.currency }

// Add returns a new Money that is the sum of m and other.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, &CurrencyMismatchError{Left: m.currency, Right: other.currency}
	}
	return Money{amount: m.amount.Add(other.amount), currency: m.currency}, nil
}

// Multiply returns a new Money scaled by the integer factor.
func (m Money) Multiply(factor int) Money {
	return Money{amount: m.amount.Mul(decimal.NewFromInt(int64(factor))), currency: m.currency}
}

// Equal returns true if amount and currency are both equal.
func (m Money) Equal(other Money) bool {
	return m.amount.Equal(other.amount) && m.currency == other.currency
}

func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.amount.StringFixed(2), m.currency)
}
```

---

## Domain Events (`src/domain/events.go`)

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the base interface for all domain events.
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
}

// baseEvent holds common fields for all events.
type baseEvent struct {
	eventID    string
	occurredAt time.Time
}

func newBaseEvent() baseEvent {
	return baseEvent{
		eventID:    uuid.New().String(),
		occurredAt: time.Now().UTC(),
	}
}

func (e baseEvent) EventID() string       { return e.eventID }
func (e baseEvent) OccurredAt() time.Time { return e.occurredAt }

// OrderPlaced is emitted when a new order is successfully created.
type OrderPlaced struct {
	baseEvent
	OrderID    OrderID
	CustomerID UserID
	Total      Money
}

func (e OrderPlaced) EventType() string { return "order.placed" }

// OrderConfirmed is emitted when an order transitions to CONFIRMED.
type OrderConfirmed struct {
	baseEvent
	OrderID OrderID
}

func (e OrderConfirmed) EventType() string { return "order.confirmed" }
```

---

## Entity (`src/domain/order.go`)

```go
package domain

import "fmt"

// OrderStatus represents the lifecycle state of an order.
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// OrderItem represents a single line in an order.
type OrderItem struct {
	ProductID string
	Quantity  int
	UnitPrice Money
}

// Subtotal returns the line total for this item.
func (i OrderItem) Subtotal() Money {
	return i.UnitPrice.Multiply(i.Quantity)
}

// Order is the aggregate root.
// All fields are unexported — read via getter methods.
type Order struct {
	id         OrderID
	customerID UserID
	items      []OrderItem
	total      Money
	status     OrderStatus
	events     []DomainEvent
}

// --- Constructors ---

// NewOrder creates a valid Order, enforces all invariants, and emits OrderPlaced.
func NewOrder(customerID UserID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrInvalidOrderItems
	}

	total := ZeroMoney("USD")
	for _, item := range items {
		var err error
		total, err = total.Add(item.Subtotal())
		if err != nil {
			return nil, fmt.Errorf("calculating total: %w", err)
		}
	}

	id := NewOrderID()
	o := &Order{
		id:         id,
		customerID: customerID,
		items:      append([]OrderItem{}, items...),
		total:      total,
		status:     OrderStatusPending,
		events:     []DomainEvent{},
	}
	o.events = append(o.events, OrderPlaced{
		baseEvent:  newBaseEvent(),
		OrderID:    id,
		CustomerID: customerID,
		Total:      total,
	})
	return o, nil
}

// ReconstitueOrder hydrates an Order from storage.
// It skips invariant validation and emits no domain events.
func ReconstitueOrder(id OrderID, customerID UserID, items []OrderItem, total Money, status OrderStatus) *Order {
	return &Order{
		id:         id,
		customerID: customerID,
		items:      append([]OrderItem{}, items...),
		total:      total,
		status:     status,
		events:     []DomainEvent{},
	}
}

// --- Getters ---

func (o *Order) ID() OrderID         { return o.id }
func (o *Order) CustomerID() UserID  { return o.customerID }
func (o *Order) Items() []OrderItem  { return append([]OrderItem{}, o.items...) }
func (o *Order) Total() Money        { return o.total }
func (o *Order) Status() OrderStatus { return o.status }

// --- State transitions ---

// Confirm transitions a PENDING order to CONFIRMED.
func (o *Order) Confirm() error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("%w: cannot confirm order in status %s", ErrInvalidOrderState, o.status)
	}
	o.status = OrderStatusConfirmed
	o.events = append(o.events, OrderConfirmed{
		baseEvent: newBaseEvent(),
		OrderID:   o.id,
	})
	return nil
}

// Cancel transitions an order to CANCELLED unless it is already cancelled.
func (o *Order) Cancel(reason string) error {
	if o.status == OrderStatusCancelled {
		return fmt.Errorf("%w: order is already cancelled", ErrInvalidOrderState)
	}
	o.status = OrderStatusCancelled
	return nil
}

// --- Event collection ---

// DrainEvents returns all pending domain events and clears the internal list.
func (o *Order) DrainEvents() []DomainEvent {
	events := append([]DomainEvent{}, o.events...)
	o.events = o.events[:0]
	return events
}

// Equal compares orders by ID.
func (o *Order) Equal(other *Order) bool {
	return o.id.Equal(other.id)
}
```

---

## Key rules illustrated here

- Unexported fields on value objects and entities — all access via exported getters
- Value receivers on `Money` — it is copied, never mutated through a method call
- `NewOrder()` returns `(*Order, error)` — never panics
- `ReconstitueOrder()` skips invariant checks and emits no events; called only from outbound adapters
- `DrainEvents()` returns a copy and clears — the use case calls this after `Save()`
- No `import` of any HTTP, DB, or framework package anywhere in this file
