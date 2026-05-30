# Configuration — Pseudocode Example

Typed configuration structs loaded from environment variables, validated at startup.

---

## Config structs

```
// Typed struct for database connection settings
Config DatabaseConfig {
  host:     String
  port:     Integer
  name:     String
  user:     String
  password: String
  maxConns: Integer   // connection pool size
  sslMode:  String    // "require", "disable", etc.

  connectionString() → String:
    return "postgres://" + self.user + ":" + self.password +
           "@" + self.host + ":" + self.port + "/" + self.name +
           "?sslmode=" + self.sslMode
}

// Typed struct for HTTP server settings
Config ServerConfig {
  host:            String
  port:            Integer
  readTimeoutSec:  Integer
  writeTimeoutSec: Integer
  shutdownTimeout: Integer

  address() → String:
    return self.host + ":" + self.port
}

// Typed struct for Kafka settings
Config KafkaConfig {
  brokers:           List<String>
  orderEventsTopic:  String
  paymentEventsTopic: String
  consumerGroup:     String
}

// Top-level app config — composes all sub-configs
Config AppConfig {
  environment: String          // "development", "staging", "production"
  database:    DatabaseConfig
  server:      ServerConfig
  kafka:       KafkaConfig
  logLevel:    String          // "debug", "info", "warn", "error"
}
```

---

## Config loader

```
// Load, validate, and return AppConfig. Fails fast if anything is missing.
Function loadConfig() → AppConfig:

  // Each field loaded from environment; required fields raise on missing
  dbConfig = DatabaseConfig(
    host:     requireEnv("DB_HOST"),
    port:     requireEnvInt("DB_PORT", default: 5432),
    name:     requireEnv("DB_NAME"),
    user:     requireEnv("DB_USER"),
    password: requireEnv("DB_PASSWORD"),
    maxConns: optionalEnvInt("DB_MAX_CONNS", default: 10),
    sslMode:  optionalEnv("DB_SSL_MODE", default: "require")
  )

  serverConfig = ServerConfig(
    host:            optionalEnv("SERVER_HOST", default: "0.0.0.0"),
    port:            optionalEnvInt("SERVER_PORT", default: 8080),
    readTimeoutSec:  optionalEnvInt("SERVER_READ_TIMEOUT", default: 30),
    writeTimeoutSec: optionalEnvInt("SERVER_WRITE_TIMEOUT", default: 30),
    shutdownTimeout: optionalEnvInt("SERVER_SHUTDOWN_TIMEOUT", default: 10)
  )

  kafkaConfig = KafkaConfig(
    brokers:            requireEnvList("KAFKA_BROKERS"),       // comma-separated
    orderEventsTopic:   requireEnv("KAFKA_ORDERS_TOPIC"),
    paymentEventsTopic: requireEnv("KAFKA_PAYMENTS_TOPIC"),
    consumerGroup:      requireEnv("KAFKA_CONSUMER_GROUP")
  )

  return AppConfig(
    environment: optionalEnv("APP_ENV", default: "development"),
    database:    dbConfig,
    server:      serverConfig,
    kafka:       kafkaConfig,
    logLevel:    optionalEnv("LOG_LEVEL", default: "info")
  )

// Helper: require a string env var, raise if missing
Function requireEnv(key: String) → String:
  value = os.getenv(key)
  if value is null or empty:
    raise ConfigError("Required environment variable not set: " + key)
  return value

// Helper: require an integer env var with optional default
Function requireEnvInt(key: String, default: Integer | null) → Integer:
  raw = os.getenv(key)
  if raw is null or empty:
    if default is not null: return default
    raise ConfigError("Required environment variable not set: " + key)
  parsed = parseInt(raw)
  if parsed is invalid:
    raise ConfigError("Environment variable " + key + " must be an integer, got: " + raw)
  return parsed
```

---

## Key patterns illustrated

| Pattern | Where |
|---|---|
| Typed structs | `DatabaseConfig`, `ServerConfig`, `KafkaConfig`, `AppConfig` |
| Required vs optional | `requireEnv()` fails fast; `optionalEnv()` has a default |
| Fail fast at startup | `ConfigError` raised before any server starts |
| No business logic | Config structs are plain data — no domain rules |
| Centralized loading | All `os.getenv()` calls in one place — not scattered through adapters |
| Sub-config composition | `AppConfig` composes smaller typed configs |
