# Go — Application Layer Examples

Real Go code illustrating `src/application/` patterns with use cases and application services.

---

## Use Case: Place Order (`src/application/usecases/place_order.go`)

```go
package usecases

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/example/app/src/domain"
	"github.com/example/app/src/ports"
)

// PlaceOrderUseCase orchestrates order placement.
// It implements ports.PlaceOrderPort — verified by the compiler at the DI wiring site.
type PlaceOrderUseCase struct {
	orderRepo      ports.OrderRepository
	eventPublisher ports.EventPublisher
}

// NewPlaceOrderUseCase constructs the use case with its required ports.
func NewPlaceOrderUseCase(
	orderRepo ports.OrderRepository,
	eventPublisher ports.EventPublisher,
) *PlaceOrderUseCase {
	return &PlaceOrderUseCase{
		orderRepo:      orderRepo,
		eventPublisher: eventPublisher,
	}
}

// Execute carries out the PlaceOrder operation.
func (uc *PlaceOrderUseCase) Execute(ctx context.Context, cmd ports.PlaceOrderCommand) (*ports.PlaceOrderResult, error) {
	// 1. Parse and validate the customer ID
	customerID, err := domain.UserIDFromString(cmd.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("parsing customer id: %w", err)
	}

	// 2. Map item DTOs → domain OrderItem values
	items := make([]domain.OrderItem, 0, len(cmd.Items))
	for _, i := range cmd.Items {
		amount, err := decimal.NewFromString(i.UnitPriceAmount)
		if err != nil {
			return nil, fmt.Errorf("parsing unit price for product %s: %w", i.ProductID, err)
		}
		unitPrice, err := domain.NewMoney(amount, i.UnitPriceCurrency)
		if err != nil {
			return nil, fmt.Errorf("creating money for product %s: %w", i.ProductID, err)
		}
		items = append(items, domain.OrderItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			UnitPrice: unitPrice,
		})
	}

	// 3. Create the domain aggregate (invariants enforced inside NewOrder)
	order, err := domain.NewOrder(customerID, items)
	if err != nil {
		return nil, err // domain error — let adapter map to HTTP status
	}

	// 4. Persist (port call — no knowledge of Postgres)
	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("saving order: %w", err)
	}

	// 5. Publish events (port call — no knowledge of Kafka)
	events := order.DrainEvents()
	if err := uc.eventPublisher.PublishAll(ctx, events); err != nil {
		return nil, fmt.Errorf("publishing events: %w", err)
	}

	// 6. Return primitive result DTO
	return &ports.PlaceOrderResult{
		OrderID:       order.ID().String(),
		Status:        string(order.Status()),
		TotalAmount:   order.Total().Amount().StringFixed(2),
		TotalCurrency: order.Total().Currency(),
	}, nil
}
```

---

## Use Case: Get Order (`src/application/usecases/get_order.go`)

```go
package usecases

import (
	"context"
	"fmt"

	"github.com/example/app/src/domain"
	"github.com/example/app/src/ports"
)

// GetOrderUseCase retrieves a single order by ID.
type GetOrderUseCase struct {
	orderRepo ports.OrderRepository
}

// NewGetOrderUseCase constructs the use case with its required repository.
func NewGetOrderUseCase(orderRepo ports.OrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{orderRepo: orderRepo}
}

// Execute carries out the GetOrder operation.
func (uc *GetOrderUseCase) Execute(ctx context.Context, query ports.GetOrderQuery) (*ports.GetOrderResult, error) {
	// 1. Parse the order ID
	orderID, err := domain.OrderIDFromString(query.OrderID)
	if err != nil {
		return nil, fmt.Errorf("parsing order id: %w", err)
	}

	// 2. Load from repository
	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("finding order: %w", err)
	}
	if order == nil {
		return nil, &domain.OrderNotFoundError{OrderID: query.OrderID}
	}

	// 3. Map domain entity → result DTO
	items := make([]ports.OrderItemView, 0, len(order.Items()))
	for _, item := range order.Items() {
		items = append(items, ports.OrderItemView{
			ProductID:         item.ProductID,
			Quantity:          item.Quantity,
			UnitPriceAmount:   item.UnitPrice.Amount().StringFixed(2),
			UnitPriceCurrency: item.UnitPrice.Currency(),
			SubtotalAmount:    item.Subtotal().Amount().StringFixed(2),
		})
	}

	return &ports.GetOrderResult{
		OrderID:       order.ID().String(),
		CustomerID:    order.CustomerID().String(),
		Status:        string(order.Status()),
		TotalAmount:   order.Total().Amount().StringFixed(2),
		TotalCurrency: order.Total().Currency(),
		Items:         items,
	}, nil
}
```

---

## Application Service (`src/application/services/notification_service.go`)

```go
package services

import (
	"context"
	"fmt"

	"github.com/example/app/src/domain"
	"github.com/example/app/src/ports"
)

// NotificationService orchestrates cross-cutting notification logic.
// It is not a use case because it is called from multiple use cases,
// not directly from an inbound adapter.
type NotificationService struct {
	orderRepo            ports.OrderRepository
	notificationGateway  ports.NotificationGateway
}

// NewNotificationService constructs the service with its required ports.
func NewNotificationService(
	orderRepo ports.OrderRepository,
	notificationGateway ports.NotificationGateway,
) *NotificationService {
	return &NotificationService{
		orderRepo:           orderRepo,
		notificationGateway: notificationGateway,
	}
}

// NotifyOrderPlaced sends a confirmation notification for a placed order.
func (s *NotificationService) NotifyOrderPlaced(ctx context.Context, orderID domain.OrderID) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("finding order for notification: %w", err)
	}
	if order == nil {
		return &domain.OrderNotFoundError{OrderID: orderID.String()}
	}

	return s.notificationGateway.Send(ctx, ports.NotificationMessage{
		RecipientID: order.CustomerID().String(),
		Subject:     "Your order has been placed",
		Body:        fmt.Sprintf("Order %s for %s is confirmed.", orderID, order.Total()),
	})
}
```

---

## Key rules illustrated here

- Use cases receive port interfaces in their constructor — never concrete adapter types
- `PlaceOrderUseCase` satisfies `ports.PlaceOrderPort` — Go's implicit interface satisfaction is verified at the assignment site in DI wiring
- `Execute()` takes command/query structs and returns result structs with primitives only
- Domain entity construction happens inside the use case — not in the inbound adapter
- Events are drained from the entity and published after the entity is persisted
- Use cases do NOT call `errors.As()` — they let domain errors propagate to the adapter
- Application services depend only on port interfaces — never on adapters or infrastructure
- No goroutines or channels in use cases — Go concurrency belongs in infrastructure
