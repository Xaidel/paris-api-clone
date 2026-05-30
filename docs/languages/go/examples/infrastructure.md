# Go — Infrastructure Layer Examples

Real Go code illustrating `src/infrastructure/` patterns with Gin and pgx.

---

## Config (`src/infrastructure/config/config.go`)

```go
package config

import (
	"fmt"
	"os"
	"strconv"
)

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}

func optionalEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// DatabaseConfig holds connection parameters for the primary database.
type DatabaseConfig struct {
	DSN            string
	MaxConnections int32
}

// StripeConfig holds credentials for the Stripe payment gateway.
type StripeConfig struct {
	APIKey        string
	WebhookSecret string
}

// KafkaConfig holds connection parameters for event publishing.
type KafkaConfig struct {
	BootstrapServers string
	OrdersTopic      string
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port  int
	Debug bool
}

// AppConfig is the root configuration struct assembled from environment variables.
type AppConfig struct {
	Database DatabaseConfig
	Stripe   StripeConfig
	Kafka    KafkaConfig
	Server   ServerConfig
}

// LoadFromEnv reads all required and optional env vars and returns a populated AppConfig.
// Returns an error if any required variable is missing — fail fast at startup.
func LoadFromEnv() (*AppConfig, error) {
	dbDSN, err := requireEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}
	stripeAPIKey, err := requireEnv("STRIPE_API_KEY")
	if err != nil {
		return nil, err
	}
	stripeWebhook, err := requireEnv("STRIPE_WEBHOOK_SECRET")
	if err != nil {
		return nil, err
	}
	kafkaBrokers, err := requireEnv("KAFKA_BOOTSTRAP_SERVERS")
	if err != nil {
		return nil, err
	}

	maxConn, _ := strconv.ParseInt(optionalEnv("DB_MAX_CONNECTIONS", "10"), 10, 32)
	port, _ := strconv.Atoi(optionalEnv("PORT", "8080"))

	return &AppConfig{
		Database: DatabaseConfig{
			DSN:            dbDSN,
			MaxConnections: int32(maxConn),
		},
		Stripe: StripeConfig{
			APIKey:        stripeAPIKey,
			WebhookSecret: stripeWebhook,
		},
		Kafka: KafkaConfig{
			BootstrapServers: kafkaBrokers,
			OrdersTopic:      optionalEnv("KAFKA_ORDERS_TOPIC", "orders.events"),
		},
		Server: ServerConfig{
			Port:  port,
			Debug: optionalEnv("DEBUG", "false") == "true",
		},
	}, nil
}
```

---

## Manual DI Wiring (`src/infrastructure/di/wire.go`)

```go
package di

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/app/src/adapters"
	"github.com/example/app/src/application/usecases"
	"github.com/example/app/src/infrastructure/config"
)

// Application bundles the HTTP server and resources that require cleanup.
type Application struct {
	Server *http.Server
	pool   *pgxpool.Pool
}

// Shutdown gracefully closes infrastructure resources.
func (a *Application) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down http server: %w", err)
	}
	a.pool.Close()
	return nil
}

// Wire assembles the full application from config.
// This is the composition root — the only place that imports from every layer.
func Wire(cfg *config.AppConfig) (*Application, error) {
	// 1. External connections
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("creating db pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// 2. Outbound adapters
	orderRepo := adapters.NewPostgresOrderRepository(pool)
	paymentGateway := adapters.NewStripePaymentGateway(cfg.Stripe.APIKey)
	eventPublisher := adapters.NewKafkaEventPublisher(cfg.Kafka.BootstrapServers)

	// 3. Use cases
	placeOrderUC := usecases.NewPlaceOrderUseCase(orderRepo, eventPublisher)
	getOrderUC := usecases.NewGetOrderUseCase(orderRepo)

	// 4. Inbound adapters
	orderAdapter := adapters.NewHttpOrderAdapter(placeOrderUC, getOrderUC)

	// 5. Router
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	orderAdapter.RegisterRoutes(router)

	// 6. HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	return &Application{Server: server, pool: pool}, nil
}
```

### `src/infrastructure/di/main.go`

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/app/src/infrastructure/config"
	"github.com/example/app/src/infrastructure/di"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	app, err := di.Wire(cfg)
	if err != nil {
		log.Fatalf("wiring application: %v", err)
	}

	// Start server in a goroutine so we can listen for shutdown signals
	go func() {
		log.Printf("listening on %s", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	// Block until SIGTERM or SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("shutdown complete")
}
```

---

## Observability (`src/infrastructure/observability/logging.go`)

```go
package observability

import (
	"log/slog"
	"os"

	"github.com/example/app/src/infrastructure/config"
)

// ConfigureLogging sets up structured logging based on the server config.
// Call once at startup before wiring adapters.
func ConfigureLogging(cfg *config.ServerConfig) {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}
```

---

## Key rules illustrated here

- `LoadFromEnv()` reads ALL env vars — no `os.Getenv()` calls anywhere outside `config/`
- Missing required vars return an error at startup — `main()` calls `log.Fatalf`, failing fast
- `Wire()` is the sole composition root — the only file in the codebase that imports from every layer
- `Application.Shutdown()` closes the DB pool — no leaked connections
- No `init()` functions for wiring — everything is explicit in `Wire()`
- Graceful shutdown listens for OS signals before terminating
