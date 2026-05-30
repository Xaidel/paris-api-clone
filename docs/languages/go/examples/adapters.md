# Go — Adapters Layer Examples

Real Go code illustrating `src/adapters/` patterns with Gin and pgx.

---

## Inbound Adapter — Gin (`src/adapters/inbound/http_order_adapter.go`)

```go
package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/app/src/domain"
	"github.com/example/app/src/ports"
)

// HttpOrderAdapter handles HTTP requests for the order resource.
// It holds port interfaces — never concrete use case structs.
type HttpOrderAdapter struct {
	placeOrder ports.PlaceOrderPort
	getOrder   ports.GetOrderPort
}

// NewHttpOrderAdapter constructs the adapter with its required ports.
func NewHttpOrderAdapter(
	placeOrder ports.PlaceOrderPort,
	getOrder ports.GetOrderPort,
) *HttpOrderAdapter {
	return &HttpOrderAdapter{
		placeOrder: placeOrder,
		getOrder:   getOrder,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpOrderAdapter) RegisterRoutes(r *gin.Engine) {
	orders := r.Group("/orders")
	orders.POST("/", a.PlaceOrder)
	orders.GET("/:id", a.GetOrder)
}

// --- Request / Response DTOs (adapter layer only) ---

type placeOrderItemRequest struct {
	ProductID         string `json:"product_id" binding:"required"`
	Quantity          int    `json:"quantity" binding:"required,min=1"`
	UnitPriceAmount   string `json:"unit_price_amount" binding:"required"`
	UnitPriceCurrency string `json:"unit_price_currency" binding:"required,len=3"`
}

type placeOrderRequest struct {
	CustomerID string                  `json:"customer_id" binding:"required,uuid"`
	Items      []placeOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type placeOrderResponse struct {
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	TotalAmount   string `json:"total_amount"`
	TotalCurrency string `json:"total_currency"`
}

type orderItemResponse struct {
	ProductID         string `json:"product_id"`
	Quantity          int    `json:"quantity"`
	UnitPriceAmount   string `json:"unit_price_amount"`
	UnitPriceCurrency string `json:"unit_price_currency"`
	SubtotalAmount    string `json:"subtotal_amount"`
}

type getOrderResponse struct {
	OrderID       string              `json:"order_id"`
	CustomerID    string              `json:"customer_id"`
	Status        string              `json:"status"`
	TotalAmount   string              `json:"total_amount"`
	TotalCurrency string              `json:"total_currency"`
	Items         []orderItemResponse `json:"items"`
}

// --- Handlers ---

// PlaceOrder handles POST /orders/
func (a *HttpOrderAdapter) PlaceOrder(c *gin.Context) {
	var req placeOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]ports.PlaceOrderItemInput, 0, len(req.Items))
	for _, i := range req.Items {
		items = append(items, ports.PlaceOrderItemInput{
			ProductID:         i.ProductID,
			Quantity:          i.Quantity,
			UnitPriceAmount:   i.UnitPriceAmount,
			UnitPriceCurrency: i.UnitPriceCurrency,
		})
	}

	result, err := a.placeOrder.Execute(c.Request.Context(), ports.PlaceOrderCommand{
		CustomerID: req.CustomerID,
		Items:      items,
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, placeOrderResponse{
		OrderID:       result.OrderID,
		Status:        result.Status,
		TotalAmount:   result.TotalAmount,
		TotalCurrency: result.TotalCurrency,
	})
}

// GetOrder handles GET /orders/:id
func (a *HttpOrderAdapter) GetOrder(c *gin.Context) {
	requestingUserID := c.Query("requesting_user_id")

	result, err := a.getOrder.Execute(c.Request.Context(), ports.GetOrderQuery{
		OrderID:          c.Param("id"),
		RequestingUserID: requestingUserID,
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	resp := getOrderResponse{
		OrderID:       result.OrderID,
		CustomerID:    result.CustomerID,
		Status:        result.Status,
		TotalAmount:   result.TotalAmount,
		TotalCurrency: result.TotalCurrency,
	}
	for _, item := range result.Items {
		resp.Items = append(resp.Items, orderItemResponse{
			ProductID:         item.ProductID,
			Quantity:          item.Quantity,
			UnitPriceAmount:   item.UnitPriceAmount,
			UnitPriceCurrency: item.UnitPriceCurrency,
			SubtotalAmount:    item.SubtotalAmount,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// handleError maps domain and infrastructure errors to HTTP status codes.
// All error mapping lives here — never scattered across handlers.
func (a *HttpOrderAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *domain.OrderNotFoundError
	var domainErr *domain.DomainError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
```

---

## Outbound Adapter — pgx (`src/adapters/outbound/postgres_order_repository.go`)

```go
package adapters

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/example/app/src/domain"
)

// PostgresOrderRepository implements ports.OrderRepository using pgx.
type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresOrderRepository constructs the repository with the shared connection pool.
func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

// Save inserts or updates an order and its items in a single transaction.
func (r *PostgresOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit

	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, customer_id, status, total_amount, total_currency)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			status         = EXCLUDED.status,
			total_amount   = EXCLUDED.total_amount,
			total_currency = EXCLUDED.total_currency
	`,
		order.ID().String(),
		order.CustomerID().String(),
		string(order.Status()),
		order.Total().Amount().String(),
		order.Total().Currency(),
	)
	if err != nil {
		return fmt.Errorf("upserting order: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, order.ID().String())
	if err != nil {
		return fmt.Errorf("deleting old items: %w", err)
	}

	for _, item := range order.Items() {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, unit_price_amount, unit_price_currency)
			VALUES ($1, $2, $3, $4, $5)
		`,
			order.ID().String(),
			item.ProductID,
			item.Quantity,
			item.UnitPrice.Amount().String(),
			item.UnitPrice.Currency(),
		)
		if err != nil {
			return fmt.Errorf("inserting item %s: %w", item.ProductID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

// FindByID returns the Order with the given ID, or nil if not found.
func (r *PostgresOrderRepository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, status, total_amount, total_currency
		FROM orders WHERE id = $1
	`, id.String())

	var (
		rawID, rawCustomerID, rawStatus, rawTotalAmount, rawTotalCurrency string
	)
	if err := row.Scan(&rawID, &rawCustomerID, &rawStatus, &rawTotalAmount, &rawTotalCurrency); err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning order row: %w", err)
	}

	itemRows, err := r.pool.Query(ctx, `
		SELECT product_id, quantity, unit_price_amount, unit_price_currency
		FROM order_items WHERE order_id = $1
	`, rawID)
	if err != nil {
		return nil, fmt.Errorf("querying order items: %w", err)
	}
	defer itemRows.Close()

	var items []domain.OrderItem
	for itemRows.Next() {
		var productID, rawUnitAmount, rawUnitCurrency string
		var quantity int
		if err := itemRows.Scan(&productID, &quantity, &rawUnitAmount, &rawUnitCurrency); err != nil {
			return nil, fmt.Errorf("scanning item row: %w", err)
		}
		amount, _ := decimal.NewFromString(rawUnitAmount)
		unitPrice, err := domain.NewMoney(amount, rawUnitCurrency)
		if err != nil {
			return nil, fmt.Errorf("building item money: %w", err)
		}
		items = append(items, domain.OrderItem{
			ProductID: productID,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
	}
	if err := itemRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating item rows: %w", err)
	}

	return toDomainOrder(rawID, rawCustomerID, rawStatus, rawTotalAmount, rawTotalCurrency, items)
}

// FindByCustomerID returns all orders for a given customer.
func (r *PostgresOrderRepository) FindByCustomerID(ctx context.Context, customerID domain.UserID) ([]*domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, customer_id, status, total_amount, total_currency
		FROM orders WHERE customer_id = $1
	`, customerID.String())
	if err != nil {
		return nil, fmt.Errorf("querying orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var rawID, rawCustomerID, rawStatus, rawAmount, rawCurrency string
		if err := rows.Scan(&rawID, &rawCustomerID, &rawStatus, &rawAmount, &rawCurrency); err != nil {
			return nil, fmt.Errorf("scanning order: %w", err)
		}
		// Item loading omitted for brevity — same pattern as FindByID
		order, err := toDomainOrder(rawID, rawCustomerID, rawStatus, rawAmount, rawCurrency, nil)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

// --- helpers ---

func toDomainOrder(rawID, rawCustomerID, rawStatus, rawAmount, rawCurrency string, items []domain.OrderItem) (*domain.Order, error) {
	id, err := domain.OrderIDFromString(rawID)
	if err != nil {
		return nil, fmt.Errorf("parsing order id: %w", err)
	}
	customerID, err := domain.UserIDFromString(rawCustomerID)
	if err != nil {
		return nil, fmt.Errorf("parsing customer id: %w", err)
	}
	totalAmount, _ := decimal.NewFromString(rawAmount)
	total, err := domain.NewMoney(totalAmount, rawCurrency)
	if err != nil {
		return nil, fmt.Errorf("building total money: %w", err)
	}
	return domain.ReconstitueOrder(id, customerID, items, total, domain.OrderStatus(rawStatus)), nil
}

func isNotFound(err error) bool {
	// pgx returns pgx.ErrNoRows when QueryRow finds nothing
	return err != nil && err.Error() == "no rows in result set"
}
```

---

## Outbound Adapter — Stripe (`src/adapters/outbound/stripe_payment_gateway.go`)

```go
package adapters

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/paymentintent"
	"github.com/stripe/stripe-go/v79/refund"

	"github.com/example/app/src/ports"
)

// StripePaymentGateway implements ports.PaymentGateway using the Stripe API.
type StripePaymentGateway struct{}

// NewStripePaymentGateway sets the Stripe API key and returns the gateway.
// Setting stripe.Key is package-level state — acceptable in infrastructure setup.
func NewStripePaymentGateway(apiKey string) *StripePaymentGateway {
	stripe.Key = apiKey
	return &StripePaymentGateway{}
}

// Charge creates a PaymentIntent and confirms it immediately.
func (g *StripePaymentGateway) Charge(ctx context.Context, req ports.PaymentRequest) (*ports.PaymentResult, error) {
	amount, _ := decimal.NewFromString(req.AmountAmount)
	amountCents := amount.Mul(decimal.NewFromInt(100)).IntPart()

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(amountCents),
		Currency:      stripe.String(req.AmountCurrency),
		PaymentMethod: stripe.String(req.PaymentMethodToken),
		Confirm:       stripe.Bool(true),
	}
	params.AddMetadata("order_id", req.OrderID)

	intent, err := paymentintent.New(params)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Type == stripe.ErrorTypeCard {
			return &ports.PaymentResult{
				Success:       false,
				FailureReason: stripeErr.Msg,
			}, nil
		}
		return nil, fmt.Errorf("stripe charge: %w", err)
	}

	return &ports.PaymentResult{
		Success:       intent.Status == stripe.PaymentIntentStatusSucceeded,
		TransactionID: intent.ID,
	}, nil
}

// Refund creates a refund against an existing PaymentIntent.
func (g *StripePaymentGateway) Refund(ctx context.Context, transactionID string, amount string) (*ports.PaymentResult, error) {
	amountDec, _ := decimal.NewFromString(amount)
	amountCents := amountDec.Mul(decimal.NewFromInt(100)).IntPart()

	r, err := refund.New(&stripe.RefundParams{
		PaymentIntent: stripe.String(transactionID),
		Amount:        stripe.Int64(amountCents),
	})
	if err != nil {
		return nil, fmt.Errorf("stripe refund: %w", err)
	}

	return &ports.PaymentResult{
		Success:       r.Status == stripe.RefundStatusSucceeded,
		TransactionID: r.ID,
	}, nil
}
```

---

## Key rules illustrated here

- `handleError()` is a single shared method on `HttpOrderAdapter` — error mapping is never scattered across handlers
- `c.ShouldBindJSON()` is used (not `c.BindJSON`) — it returns an error instead of aborting
- `PostgresOrderRepository.toDomainOrder()` calls `domain.ReconstitueOrder()` — never `domain.NewOrder()`
- Each adapter receives an injected pool or API key — never creates its own connection
- `StripePaymentGateway` maps card errors to `PaymentResult{Success: false}` and wraps other Stripe errors as infrastructure errors
- No business logic in any handler or adapter — only translation between transport/DB and port types
