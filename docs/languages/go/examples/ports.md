# Go — Ports Layer Examples

Real Go code illustrating `src/ports/` patterns using Go interfaces.

---

## Inbound Ports (`src/ports/inbound/`)

### `src/ports/inbound/place_order_port.go`

```go
package ports

import "context"

// PlaceOrderItemInput is a primitive DTO for a single order line.
// No domain types leak into the inbound port — all fields are primitives.
type PlaceOrderItemInput struct {
	ProductID         string
	Quantity          int
	UnitPriceAmount   string // string preserves decimal precision
	UnitPriceCurrency string
}

// PlaceOrderCommand is the command sent to PlaceOrderPort.Execute.
type PlaceOrderCommand struct {
	CustomerID string
	Items      []PlaceOrderItemInput
}

// PlaceOrderResult carries the outcome of a successful PlaceOrder operation.
type PlaceOrderResult struct {
	OrderID       string
	Status        string
	TotalAmount   string
	TotalCurrency string
}

// PlaceOrderPort is the inbound port for placing a new order.
// Use cases implement this interface; server handlers call it.
type PlaceOrderPort interface {
	Execute(ctx context.Context, cmd PlaceOrderCommand) (*PlaceOrderResult, error)
}
```

### `src/ports/inbound/get_order_port.go`

```go
package ports

import "context"

// GetOrderQuery identifies which order to retrieve and who is requesting.
type GetOrderQuery struct {
	OrderID         string
	RequestingUserID string
}

// OrderItemView is a read-model DTO for a single order line.
type OrderItemView struct {
	ProductID         string
	Quantity          int
	UnitPriceAmount   string
	UnitPriceCurrency string
	SubtotalAmount    string
}

// GetOrderResult is the read-model returned by GetOrderPort.Execute.
type GetOrderResult struct {
	OrderID       string
	CustomerID    string
	Status        string
	TotalAmount   string
	TotalCurrency string
	Items         []OrderItemView
}

// GetOrderPort is the inbound port for retrieving an existing order.
type GetOrderPort interface {
	Execute(ctx context.Context, query GetOrderQuery) (*GetOrderResult, error)
}
```

---

## Outbound Ports (`src/ports/outbound/`)

### `src/ports/outbound/order_repository.go`

```go
package ports

import (
	"context"

	"github.com/example/app/src/domain"
)

// OrderRepository is the outbound port for persisting and querying orders.
// The application layer depends on this interface; outbound adapters implement it.
type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
	FindByCustomerID(ctx context.Context, customerID domain.UserID) ([]*domain.Order, error)
}
```

### `src/ports/outbound/payment_gateway.go`

```go
package ports

import "context"

// PaymentRequest carries the data needed to charge a customer.
type PaymentRequest struct {
	OrderID            string
	AmountAmount       string
	AmountCurrency     string
	PaymentMethodToken string
}

// PaymentResult carries the outcome of a charge or refund attempt.
type PaymentResult struct {
	Success       bool
	TransactionID string // empty string when not successful
	FailureReason string // empty string on success
}

// PaymentGateway is the outbound port for processing payments.
type PaymentGateway interface {
	Charge(ctx context.Context, req PaymentRequest) (*PaymentResult, error)
	Refund(ctx context.Context, transactionID string, amount string) (*PaymentResult, error)
}
```

### `src/ports/outbound/event_publisher.go`

```go
package ports

import (
	"context"

	"github.com/example/app/src/domain"
)

// EventPublisher is the outbound port for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, event domain.DomainEvent) error
	PublishAll(ctx context.Context, events []domain.DomainEvent) error
}
```

### `src/ports/outbound/notification_gateway.go`

```go
package ports

import "context"

// NotificationMessage is the data sent to a notification recipient.
type NotificationMessage struct {
	RecipientID string
	Subject     string
	Body        string
}

// NotificationGateway is the outbound port for sending notifications.
type NotificationGateway interface {
	Send(ctx context.Context, msg NotificationMessage) error
}
```

---

## Key rules illustrated here

- All ports are Go `interface` types — one file per port, one interface per file
- `context.Context` is the first parameter on every port method — idiomatic Go
- Inbound port commands/results use primitives (`string`, `int`) — no domain types
- Outbound port methods use domain types (`*domain.Order`, `domain.OrderID`) — called from the application layer
- No struct fields, no embedded types, no default implementations — interfaces only
- Package name is `ports` — flat, not `ports/inbound` or `ports/outbound` (sub-packages are acceptable but increase import complexity; keep flat unless the project grows large)
