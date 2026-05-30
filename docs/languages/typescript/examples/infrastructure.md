# TypeScript — Infrastructure Layer Examples

Real TypeScript code illustrating `src/infrastructure/` patterns with TanStack Start and postgres.js.

---

## Config (`src/infrastructure/config/config.ts`)

```typescript
function requireEnv(key: string): string {
  const value = process.env[key];
  if (!value) throw new Error(`Required environment variable ${key} is not set`);
  return value;
}

function optionalEnv(key: string, defaultValue: string): string {
  return process.env[key] ?? defaultValue;
}

export interface DatabaseConfig {
  readonly url: string;
  readonly maxConnections: number;
}

export interface StripeConfig {
  readonly secretKey: string;
  readonly webhookSecret: string;
}

export interface KafkaConfig {
  readonly brokers: string[];
  readonly ordersTopic: string;
}

export interface ServerConfig {
  readonly port: number;
  readonly debug: boolean;
}

export interface AppConfig {
  readonly database: DatabaseConfig;
  readonly stripe: StripeConfig;
  readonly kafka: KafkaConfig;
  readonly server: ServerConfig;
}

export function loadConfig(): AppConfig {
  return {
    database: {
      url: requireEnv("DATABASE_URL"),
      maxConnections: parseInt(optionalEnv("DB_MAX_CONNECTIONS", "10"), 10),
    },
    stripe: {
      secretKey: requireEnv("STRIPE_SECRET_KEY"),
      webhookSecret: requireEnv("STRIPE_WEBHOOK_SECRET"),
    },
    kafka: {
      brokers: requireEnv("KAFKA_BROKERS").split(","),
      ordersTopic: optionalEnv("KAFKA_ORDERS_TOPIC", "orders.events"),
    },
    server: {
      port: parseInt(optionalEnv("PORT", "3000"), 10),
      debug: optionalEnv("DEBUG", "false") === "true",
    },
  };
}
```

---

## Simple DI — Manual Wiring (`src/infrastructure/di/wire.ts`)

Use this for most projects. No library required.

```typescript
import postgres from "postgres";
import { loadConfig, type AppConfig } from "../config/config";
import { PostgresOrderRepository } from "../../adapters/outbound/postgres-order-repository";
import { StripePaymentGateway } from "../../adapters/outbound/stripe-payment-gateway";
import { KafkaEventPublisher } from "../../adapters/outbound/kafka-event-publisher";
import { PlaceOrderUseCase } from "../../application/use-cases/place-order-use-case";
import { GetOrderUseCase } from "../../application/use-cases/get-order-use-case";
import { initOrderServerFns } from "../../adapters/inbound/order-server-fns";
import type { PlaceOrderPort } from "../../ports/inbound/place-order-port";
import type { GetOrderPort } from "../../ports/inbound/get-order-port";

export interface WiredApp {
  readonly sql: ReturnType<typeof postgres>;
  readonly placeOrderPort: PlaceOrderPort;
  readonly getOrderPort: GetOrderPort;
  shutdown(): Promise<void>;
}

export async function wire(config?: AppConfig): Promise<WiredApp> {
  const cfg = config ?? loadConfig();

  // 1. External connections
  const sql = postgres(cfg.database.url, {
    max: cfg.database.maxConnections,
    onnotice: () => {}, // suppress notices in production
  });

  // 2. Outbound adapters
  const orderRepo = new PostgresOrderRepository(sql);
  const paymentGateway = new StripePaymentGateway(cfg.stripe.secretKey);
  const eventPublisher = new KafkaEventPublisher(cfg.kafka.brokers);

  // 3. Use cases (implement inbound ports)
  const placeOrderUseCase = new PlaceOrderUseCase(orderRepo, eventPublisher);
  const getOrderUseCase = new GetOrderUseCase(orderRepo);

  // 4. Register use cases with inbound adapters (server functions)
  initOrderServerFns(placeOrderUseCase, getOrderUseCase);

  return {
    sql,
    placeOrderPort: placeOrderUseCase,
    getOrderPort: getOrderUseCase,
    async shutdown() {
      await sql.end();
      await eventPublisher.disconnect();
    },
  };
}
```

### `src/infrastructure/di/main.ts`

```typescript
import { wire } from "./wire";
import { loadConfig } from "../config/config";

async function main(): Promise<void> {
  const config = loadConfig();
  const app = await wire(config);

  // TanStack Start handles the HTTP server — this file only wires dependencies.
  // For non-TanStack contexts (CLIs, workers), start your server here.

  process.on("SIGTERM", async () => {
    console.log("SIGTERM received, shutting down...");
    await app.shutdown();
    process.exit(0);
  });

  process.on("SIGINT", async () => {
    console.log("SIGINT received, shutting down...");
    await app.shutdown();
    process.exit(0);
  });
}

main().catch((err) => {
  console.error("Fatal startup error:", err);
  process.exit(1);
});
```

---

## Complex DI — TSyringe (`src/infrastructure/di/container.ts`)

Use this when the manual wiring file exceeds ~100 lines or lifecycle management is needed.

```typescript
import "reflect-metadata";
import { container, injectable, inject } from "tsyringe";
import postgres from "postgres";
import { loadConfig } from "../config/config";
import { PostgresOrderRepository } from "../../adapters/outbound/postgres-order-repository";
import { StripePaymentGateway } from "../../adapters/outbound/stripe-payment-gateway";
import { KafkaEventPublisher } from "../../adapters/outbound/kafka-event-publisher";
import { PlaceOrderUseCase } from "../../application/use-cases/place-order-use-case";
import { GetOrderUseCase } from "../../application/use-cases/get-order-use-case";
import type { OrderRepository } from "../../ports/outbound/order-repository";
import type { EventPublisher } from "../../ports/outbound/event-publisher";
import type { PaymentGateway } from "../../ports/outbound/payment-gateway";

// Token symbols — avoids stringly-typed registration
export const TOKENS = {
  SqlClient: Symbol("SqlClient"),
  OrderRepository: Symbol("OrderRepository"),
  PaymentGateway: Symbol("PaymentGateway"),
  EventPublisher: Symbol("EventPublisher"),
} as const;

export async function buildContainer() {
  const config = loadConfig();

  // Register infrastructure resources
  const sql = postgres(config.database.url, { max: config.database.maxConnections });
  container.registerInstance(TOKENS.SqlClient, sql);

  // Register adapters against port tokens
  container.register<OrderRepository>(TOKENS.OrderRepository, {
    useFactory: (c) => new PostgresOrderRepository(c.resolve(TOKENS.SqlClient)),
  });
  container.register<PaymentGateway>(TOKENS.PaymentGateway, {
    useFactory: () => new StripePaymentGateway(config.stripe.secretKey),
  });
  container.register<EventPublisher>(TOKENS.EventPublisher, {
    useFactory: () => new KafkaEventPublisher(config.kafka.brokers),
  });

  // Register use cases
  container.register(PlaceOrderUseCase, {
    useFactory: (c) =>
      new PlaceOrderUseCase(
        c.resolve<OrderRepository>(TOKENS.OrderRepository),
        c.resolve<EventPublisher>(TOKENS.EventPublisher),
      ),
  });
  container.register(GetOrderUseCase, {
    useFactory: (c) =>
      new GetOrderUseCase(c.resolve<OrderRepository>(TOKENS.OrderRepository)),
  });

  return container;
}
```

> **Note:** Do NOT use `@inject` decorators in domain or application layer classes.
> Injection via `@inject` is only permitted on adapter and infrastructure classes.

---

## Observability (`src/infrastructure/observability/logger.ts`)

```typescript
import type { AppConfig } from "../config/config";

export interface Logger {
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, context?: Record<string, unknown>): void;
  debug(message: string, context?: Record<string, unknown>): void;
}

function formatEntry(
  level: string,
  message: string,
  context?: Record<string, unknown>,
): string {
  const entry = {
    time: new Date().toISOString(),
    level,
    message,
    ...context,
  };
  return JSON.stringify(entry);
}

export function createLogger(config: AppConfig): Logger {
  const minLevel = config.server.debug ? "debug" : "info";
  const levels = ["debug", "info", "warn", "error"];
  const minIndex = levels.indexOf(minLevel);

  function log(level: string, message: string, context?: Record<string, unknown>): void {
    if (levels.indexOf(level) >= minIndex) {
      process.stdout.write(formatEntry(level, message, context) + "\n");
    }
  }

  return {
    info: (m, c) => log("info", m, c),
    warn: (m, c) => log("warn", m, c),
    error: (m, c) => log("error", m, c),
    debug: (m, c) => log("debug", m, c),
  };
}
```

---

## Key rules illustrated here

- `loadConfig()` reads ALL env vars — no `process.env` access anywhere else in the codebase
- Missing required vars throw at startup — fail fast, never silently
- Manual `wire()` covers most projects; TSyringe `buildContainer()` is the complex-case alternative
- `shutdown()` closes the DB connection — no leaked resources
- Domain and application layer classes are never decorated with `@inject` — only infrastructure performs wiring
- The logger interface is defined in infrastructure — domain and application never import it
