# Go — Test Examples

Real Go code illustrating test patterns with `testing`, `testify`, and `testcontainers-go`.

---

## In-Memory Port Implementations (`tests/utils/in_memory_order_repository.go`)

```go
package utils

import (
	"context"
	"sync"

	"github.com/example/app/src/domain"
)

// InMemoryOrderRepository is an in-memory implementation of ports.OrderRepository.
// Used in unit tests — satisfies the same interface as the real adapter.
// No mocking library required.
type InMemoryOrderRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Order
}

// NewInMemoryOrderRepository constructs an empty repository.
func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{store: make(map[string]*domain.Order)}
}

func (r *InMemoryOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[order.ID().String()] = order
	return nil
}

func (r *InMemoryOrderRepository) FindByID(_ context.Context, id domain.OrderID) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.store[id.String()], nil // nil if not found — matches port contract
}

func (r *InMemoryOrderRepository) FindByCustomerID(_ context.Context, customerID domain.UserID) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Order
	for _, o := range r.store {
		if o.CustomerID().Equal(customerID) {
			result = append(result, o)
		}
	}
	return result, nil
}

// AllOrders is a test helper — not part of the port contract.
func (r *InMemoryOrderRepository) AllOrders() []*domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Order, 0, len(r.store))
	for _, o := range r.store {
		result = append(result, o)
	}
	return result
}
```

```go
// tests/utils/in_memory_event_publisher.go
package utils

import (
	"context"
	"sync"

	"github.com/example/app/src/domain"
)

// InMemoryEventPublisher stores published events for test assertions.
type InMemoryEventPublisher struct {
	mu        sync.Mutex
	Published []domain.DomainEvent
}

func (p *InMemoryEventPublisher) Publish(_ context.Context, event domain.DomainEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Published = append(p.Published, event)
	return nil
}

func (p *InMemoryEventPublisher) PublishAll(_ context.Context, events []domain.DomainEvent) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Published = append(p.Published, events...)
	return nil
}
```

---

## Unit Tests — Domain Entity (`tests/unit/order_test.go`)

```go
package unit

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/app/src/domain"
)

func makeItem(t *testing.T, price string) domain.OrderItem {
	t.Helper()
	amount, _ := decimal.NewFromString(price)
	money, err := domain.NewMoney(amount, "USD")
	require.NoError(t, err)
	return domain.OrderItem{ProductID: "prod-1", Quantity: 2, UnitPrice: money}
}

func makeOrder(t *testing.T) *domain.Order {
	t.Helper()
	customerID := domain.NewUserID()
	order, err := domain.NewOrder(customerID, []domain.OrderItem{makeItem(t, "9.99")})
	require.NoError(t, err)
	return order
}

func TestOrderCreation_EmitsOrderPlacedEvent(t *testing.T) {
	order := makeOrder(t)
	events := order.DrainEvents()
	require.Len(t, events, 1)
	placed, ok := events[0].(domain.OrderPlaced)
	require.True(t, ok, "expected OrderPlaced event")
	assert.True(t, placed.OrderID.Equal(order.ID()))
}

func TestOrderCreation_StatusIsPending(t *testing.T) {
	order := makeOrder(t)
	assert.Equal(t, domain.OrderStatusPending, order.Status())
}

func TestOrderCreation_RejectsEmptyItems(t *testing.T) {
	_, err := domain.NewOrder(domain.NewUserID(), nil)
	assert.ErrorIs(t, err, domain.ErrInvalidOrderItems)
}

func TestOrderCreation_CalculatesTotalCorrectly(t *testing.T) {
	amount1, _ := decimal.NewFromString("10.00")
	price1, _ := domain.NewMoney(amount1, "USD")
	amount2, _ := decimal.NewFromString("5.00")
	price2, _ := domain.NewMoney(amount2, "USD")
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 2, UnitPrice: price1},
		{ProductID: "p2", Quantity: 1, UnitPrice: price2},
	}
	order, err := domain.NewOrder(domain.NewUserID(), items)
	require.NoError(t, err)
	assert.Equal(t, "25.00", order.Total().Amount().StringFixed(2))
}

func TestOrderConfirm_TransitionsToConfirmed(t *testing.T) {
	order := makeOrder(t)
	order.DrainEvents()
	err := order.Confirm()
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusConfirmed, order.Status())
	events := order.DrainEvents()
	require.Len(t, events, 1)
	_, ok := events[0].(domain.OrderConfirmed)
	assert.True(t, ok)
}

func TestOrderConfirm_RejectsAlreadyCancelledOrder(t *testing.T) {
	order := makeOrder(t)
	require.NoError(t, order.Cancel("test"))
	err := order.Confirm()
	assert.ErrorIs(t, err, domain.ErrInvalidOrderState)
}

func TestDrainEvents_ClearsEventList(t *testing.T) {
	order := makeOrder(t)
	first := order.DrainEvents()
	second := order.DrainEvents()
	assert.Len(t, first, 1)
	assert.Len(t, second, 0)
}

func TestReconstitueOrder_EmitsNoEvents(t *testing.T) {
	original := makeOrder(t)
	order := domain.ReconstitueOrder(
		original.ID(),
		original.CustomerID(),
		original.Items(),
		original.Total(),
		original.Status(),
	)
	assert.Len(t, order.DrainEvents(), 0)
}
```

---

## Unit Tests — Use Case (`tests/unit/place_order_use_case_test.go`)

```go
package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/app/src/application/usecases"
	"github.com/example/app/src/domain"
	"github.com/example/app/src/ports"
	"github.com/example/app/tests/utils"
)

func validCommand() ports.PlaceOrderCommand {
	return ports.PlaceOrderCommand{
		CustomerID: "00000000-0000-0000-0000-000000000001",
		Items: []ports.PlaceOrderItemInput{
			{
				ProductID:         "prod-1",
				Quantity:          2,
				UnitPriceAmount:   "9.99",
				UnitPriceCurrency: "USD",
			},
		},
	}
}

func TestPlaceOrderUseCase_PlacesOrderAndReturnsResult(t *testing.T) {
	repo := utils.NewInMemoryOrderRepository()
	pub := &utils.InMemoryEventPublisher{}
	uc := usecases.NewPlaceOrderUseCase(repo, pub)

	result, err := uc.Execute(context.Background(), validCommand())
	require.NoError(t, err)

	assert.NotEmpty(t, result.OrderID)
	assert.Equal(t, "PENDING", result.Status)
	assert.Equal(t, "19.98", result.TotalAmount)
	assert.Len(t, repo.AllOrders(), 1)
}

func TestPlaceOrderUseCase_PublishesOrderPlacedEvent(t *testing.T) {
	repo := utils.NewInMemoryOrderRepository()
	pub := &utils.InMemoryEventPublisher{}
	uc := usecases.NewPlaceOrderUseCase(repo, pub)

	_, err := uc.Execute(context.Background(), validCommand())
	require.NoError(t, err)

	require.Len(t, pub.Published, 1)
	_, ok := pub.Published[0].(domain.OrderPlaced)
	assert.True(t, ok)
}

func TestPlaceOrderUseCase_RejectsEmptyItems(t *testing.T) {
	uc := usecases.NewPlaceOrderUseCase(
		utils.NewInMemoryOrderRepository(),
		&utils.InMemoryEventPublisher{},
	)
	cmd := ports.PlaceOrderCommand{
		CustomerID: "00000000-0000-0000-0000-000000000001",
		Items:      nil,
	}
	_, err := uc.Execute(context.Background(), cmd)
	assert.ErrorIs(t, err, domain.ErrInvalidOrderItems)
}
```

---

## Integration Test — Repository (`tests/integration/postgres_order_repository_test.go`)

```go
package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/example/app/src/adapters"
	"github.com/example/app/src/domain"
)

const migrations = `
CREATE TABLE IF NOT EXISTS orders (
	id             TEXT PRIMARY KEY,
	customer_id    TEXT NOT NULL,
	status         TEXT NOT NULL,
	total_amount   NUMERIC NOT NULL,
	total_currency TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS order_items (
	id                   SERIAL PRIMARY KEY,
	order_id             TEXT NOT NULL REFERENCES orders(id),
	product_id           TEXT NOT NULL,
	quantity             INTEGER NOT NULL,
	unit_price_amount    NUMERIC NOT NULL,
	unit_price_currency  TEXT NOT NULL
);
`

func setupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase("testdb"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections")),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, migrations)
	require.NoError(t, err)

	return pool
}

func makeTestOrder(t *testing.T) *domain.Order {
	t.Helper()
	amount, _ := decimal.NewFromString("10.00")
	price, err := domain.NewMoney(amount, "USD")
	require.NoError(t, err)
	order, err := domain.NewOrder(domain.NewUserID(), []domain.OrderItem{
		{ProductID: "prod-1", Quantity: 2, UnitPrice: price},
	})
	require.NoError(t, err)
	return order
}

func TestPostgresOrderRepository_SaveAndFindByID(t *testing.T) {
	pool := setupPostgres(t)
	repo := adapters.NewPostgresOrderRepository(pool)
	ctx := context.Background()

	order := makeTestOrder(t)
	require.NoError(t, repo.Save(ctx, order))

	found, err := repo.FindByID(ctx, order.ID())
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.True(t, found.ID().Equal(order.ID()))
	assert.Equal(t, domain.OrderStatusPending, found.Status())
	assert.Len(t, found.Items(), 1)
}

func TestPostgresOrderRepository_FindByID_ReturnsNilWhenNotFound(t *testing.T) {
	pool := setupPostgres(t)
	repo := adapters.NewPostgresOrderRepository(pool)

	found, err := repo.FindByID(context.Background(), domain.NewOrderID())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestPostgresOrderRepository_UpdatesStatusOnResave(t *testing.T) {
	pool := setupPostgres(t)
	repo := adapters.NewPostgresOrderRepository(pool)
	ctx := context.Background()

	order := makeTestOrder(t)
	require.NoError(t, repo.Save(ctx, order))
	require.NoError(t, order.Confirm())
	require.NoError(t, repo.Save(ctx, order))

	found, err := repo.FindByID(ctx, order.ID())
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusConfirmed, found.Status())
}
```

---

## Integration Test — HTTP Adapter (`tests/integration/http_order_adapter_test.go`)

```go
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/app/src/adapters"
	"github.com/example/app/src/application/usecases"
	"github.com/example/app/tests/utils"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	repo := utils.NewInMemoryOrderRepository()
	pub := &utils.InMemoryEventPublisher{}
	placeOrderUC := usecases.NewPlaceOrderUseCase(repo, pub)
	getOrderUC := usecases.NewGetOrderUseCase(repo)
	orderAdapter := adapters.NewHttpOrderAdapter(placeOrderUC, getOrderUC)
	r := gin.New()
	orderAdapter.RegisterRoutes(r)
	return r
}

func TestHttpOrderAdapter_PlaceOrder_Returns201(t *testing.T) {
	router := setupTestRouter()
	body, _ := json.Marshal(map[string]any{
		"customer_id": "00000000-0000-0000-0000-000000000001",
		"items": []map[string]any{
			{"product_id": "prod-1", "quantity": 2, "unit_price_amount": "9.99", "unit_price_currency": "USD"},
		},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/orders/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["order_id"])
	assert.Equal(t, "PENDING", resp["status"])
}

func TestHttpOrderAdapter_PlaceOrder_Returns422ForEmptyItems(t *testing.T) {
	router := setupTestRouter()
	body, _ := json.Marshal(map[string]any{
		"customer_id": "00000000-0000-0000-0000-000000000001",
		"items":       []any{},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/orders/", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHttpOrderAdapter_GetOrder_Returns404ForUnknownID(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/orders/00000000-0000-0000-0000-000000000099?requesting_user_id=00000000-0000-0000-0000-000000000001",
		nil,
	)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
```

---

## Key rules illustrated here

- In-memory repositories implement the same `interface` as the real adapter — no test doubles library needed
- Unit tests use `testify/assert` and `testify/require` — `require` on setup steps, `assert` on assertions
- Integration tests use `testcontainers-go` for a real Postgres DB — no mocking of SQL
- HTTP adapter tests use `httptest.NewRecorder()` and a real Gin router — no real HTTP port
- `t.Cleanup()` handles resource teardown — no `defer` in test helpers that might not run
- Each test function is independent — no shared state between test functions
