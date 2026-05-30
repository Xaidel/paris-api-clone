package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestDatabaseConfigOmitsPgvectorToggle(t *testing.T) {
	t.Parallel()

	if _, exists := reflect.TypeOf(DatabaseConfig{}).FieldByName("UsePgvector"); exists {
		t.Fatal("DatabaseConfig unexpectedly exposes UsePgvector")
	}
}

// TestLoadFromEnvironment verifies the load from environment behavior and the expected outcome asserted below.
func TestLoadFromEnvironment(t *testing.T) {
	// This test verifies the happy-path bootstrap contract: Load must parse a full
	// production environment into strongly typed configuration values.
	t.Setenv("ENV", "production")
	t.Setenv("SERVICE_NAME", "paris-api")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "1m")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("DATABASE_URL", "postgres://admin:admin@localhost:5432/paris?sslmode=disable")
	t.Setenv("DATABASE_PING_TIMEOUT", "5s")
	t.Setenv("MIGRATIONS_PATH", "internal/infrastructure/db/migrations")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE_PATH", "tmp/logs/paris-api.jsonl")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENAI_BASE_URL", "https://example.test")
	t.Setenv("OPENAI_API_VERSION", "2025-01-01-preview")
	t.Setenv("OPENAI_USE_AZURE", "true")
	t.Setenv("REACT_CLASSIFICATION_MODEL", "gpt-4o-mini")
	t.Setenv("REACT_CLASSIFICATION_BATCH_SIZE", "10")
	t.Setenv("REACT_CLASSIFICATION_FLUSH_TIMEOUT", "2s")
	t.Setenv("REACT_CLASSIFICATION_CLASSIFIER_FAMILY", "react")
	t.Setenv("REACT_CLASSIFICATION_CLASSIFIER_VERSION", "v1")
	t.Setenv("REACT_CLASSIFICATION_REQUEST_TIMEOUT", "30s")
	t.Setenv("REACT_CLASSIFICATION_MAX_RETRIES", "2")
	t.Setenv("REACT_CLASSIFICATION_RETRY_BACKOFF", "2s")
	t.Setenv("STORAGE_PROVIDER", "local")
	t.Setenv("LOCAL_TRANSACTION_FILE_PATH", ".tx-files")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Env != "production" {
		t.Fatalf("Env = %q", config.Env)
	}

	if config.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("ReadTimeout = %v", config.HTTP.ReadTimeout)
	}

	if config.Storage.Provider != "local" {
		t.Fatalf("Storage.Provider = %q", config.Storage.Provider)
	}

	if config.Classification.OpenAIAPIKey != "test-key" {
		t.Fatalf("Classification.OpenAIAPIKey = %q", config.Classification.OpenAIAPIKey)
	}

	if config.Classification.OpenAIBaseURL != "https://example.test" {
		t.Fatalf("Classification.OpenAIBaseURL = %q", config.Classification.OpenAIBaseURL)
	}

	if config.Classification.OpenAIAPIVersion != "2025-01-01-preview" {
		t.Fatalf("Classification.OpenAIAPIVersion = %q", config.Classification.OpenAIAPIVersion)
	}

	if !config.Classification.OpenAIUseAzure {
		t.Fatal("Classification.OpenAIUseAzure = false, want true")
	}

	if config.Classification.ReactBatchSize != 10 {
		t.Fatalf("Classification.ReactBatchSize = %d", config.Classification.ReactBatchSize)
	}

	if config.Classification.ReactMaxRetries != 2 {
		t.Fatalf("Classification.ReactMaxRetries = %d", config.Classification.ReactMaxRetries)
	}

	if config.Classification.ReactRetryBackoff != 2*time.Second {
		t.Fatalf("Classification.ReactRetryBackoff = %v", config.Classification.ReactRetryBackoff)
	}
}

// TestLoadLeavesOpenAIFieldsEmptyWhenUnset verifies the config loader preserves
// empty OpenAI transport settings when the environment does not provide them.
func TestLoadLeavesOpenAIFieldsEmptyWhenUnset(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "inherited-key")
	t.Setenv("OPENAI_BASE_URL", "https://inherited.example.test")
	t.Setenv("OPENAI_API_VERSION", "2024-01-01-preview")
	t.Setenv("OPENAI_USE_AZURE", "true")
	clearEnvironment(t, []string{"OPENAI_API_KEY", "OPENAI_BASE_URL", "OPENAI_API_VERSION", "OPENAI_USE_AZURE"})

	t.Setenv("ENV", "production")
	t.Setenv("SERVICE_NAME", "paris-api")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "1m")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("DATABASE_URL", "postgres://admin:admin@localhost:5432/paris?sslmode=disable")
	t.Setenv("DATABASE_PING_TIMEOUT", "5s")
	t.Setenv("MIGRATIONS_PATH", "internal/infrastructure/db/migrations")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE_PATH", "tmp/logs/paris-api.jsonl")
	t.Setenv("STORAGE_PROVIDER", "local")
	t.Setenv("LOCAL_TRANSACTION_FILE_PATH", ".tx-files")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Classification.OpenAIAPIKey != "" {
		t.Fatalf("Classification.OpenAIAPIKey = %q, want empty", config.Classification.OpenAIAPIKey)
	}

	if config.Classification.OpenAIBaseURL != "" {
		t.Fatalf("Classification.OpenAIBaseURL = %q, want empty", config.Classification.OpenAIBaseURL)
	}

	if config.Classification.OpenAIAPIVersion != "" {
		t.Fatalf("Classification.OpenAIAPIVersion = %q, want empty", config.Classification.OpenAIAPIVersion)
	}

	if config.Classification.OpenAIUseAzure {
		t.Fatal("Classification.OpenAIUseAzure = true, want false")
	}
}

// TestLoadDevelopmentDotenv verifies the load development dotenv behavior and the expected outcome asserted below.
func TestLoadDevelopmentDotenv(t *testing.T) {
	// This test ensures development mode reads .env from the working directory and
	// resolves relative paths consistently.
	clearEnvironment(t, []string{
		"SERVICE_NAME",
		"HTTP_PORT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
		"HTTP_SHUTDOWN_TIMEOUT",
		"DATABASE_URL",
		"DATABASE_PING_TIMEOUT",
		"MIGRATIONS_PATH",
		"LOG_LEVEL",
		"LOG_FILE_PATH",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_API_VERSION",
		"OPENAI_USE_AZURE",
		"REACT_CLASSIFICATION_MODEL",
		"REACT_CLASSIFICATION_BATCH_SIZE",
		"REACT_CLASSIFICATION_FLUSH_TIMEOUT",
		"REACT_CLASSIFICATION_CLASSIFIER_FAMILY",
		"REACT_CLASSIFICATION_CLASSIFIER_VERSION",
		"REACT_CLASSIFICATION_REQUEST_TIMEOUT",
		"REACT_CLASSIFICATION_MAX_RETRIES",
		"REACT_CLASSIFICATION_RETRY_BACKOFF",
		"STORAGE_PROVIDER",
		"LOCAL_TRANSACTION_FILE_PATH",
	})
	t.Setenv("ENV", "development")

	tempDir := t.TempDir()
	writeEnvFile(t, tempDir, "SERVICE_NAME=paris-api\nHTTP_PORT=9000\nHTTP_READ_TIMEOUT=5s\nHTTP_WRITE_TIMEOUT=10s\nHTTP_IDLE_TIMEOUT=60s\nHTTP_SHUTDOWN_TIMEOUT=10s\nDATABASE_URL=postgres://admin:admin@localhost:5432/paris?sslmode=disable\nDATABASE_PING_TIMEOUT=5s\nMIGRATIONS_PATH=internal/infrastructure/db/migrations\nLOG_LEVEL=info\nLOG_FILE_PATH=tmp/logs/paris-api.jsonl\nOPENAI_API_KEY=test-key\nOPENAI_BASE_URL=https://dotenv.example.test\nOPENAI_API_VERSION=2025-02-01-preview\nOPENAI_USE_AZURE=true\nREACT_CLASSIFICATION_MODEL=gpt-4o-mini\nREACT_CLASSIFICATION_BATCH_SIZE=10\nREACT_CLASSIFICATION_FLUSH_TIMEOUT=2s\nREACT_CLASSIFICATION_CLASSIFIER_FAMILY=react\nREACT_CLASSIFICATION_CLASSIFIER_VERSION=v1\nREACT_CLASSIFICATION_REQUEST_TIMEOUT=30s\nREACT_CLASSIFICATION_MAX_RETRIES=2\nREACT_CLASSIFICATION_RETRY_BACKOFF=2s\nSTORAGE_PROVIDER=local\nLOCAL_TRANSACTION_FILE_PATH=.tx-files\n")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Log.FilePath != filepath.Clean("tmp/logs/paris-api.jsonl") {
		t.Fatalf("Log.FilePath = %q", config.Log.FilePath)
	}

	if config.Classification.OpenAIBaseURL != "https://dotenv.example.test" {
		t.Fatalf("Classification.OpenAIBaseURL = %q", config.Classification.OpenAIBaseURL)
	}

	if config.Classification.OpenAIAPIVersion != "2025-02-01-preview" {
		t.Fatalf("Classification.OpenAIAPIVersion = %q", config.Classification.OpenAIAPIVersion)
	}

	if !config.Classification.OpenAIUseAzure {
		t.Fatal("Classification.OpenAIUseAzure = false, want true")
	}
}

// TestLoadProductionDoesNotReadDotenv verifies the load production does not read dotenv behavior and the expected outcome asserted below.
func TestLoadProductionDoesNotReadDotenv(t *testing.T) {
	// This test prevents production startup from silently depending on .env files
	// when required environment variables are missing.
	clearEnvironment(t, []string{
		"SERVICE_NAME",
		"HTTP_PORT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
		"HTTP_SHUTDOWN_TIMEOUT",
		"DATABASE_URL",
		"DATABASE_PING_TIMEOUT",
		"MIGRATIONS_PATH",
		"LOG_LEVEL",
		"LOG_FILE_PATH",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_API_VERSION",
		"OPENAI_USE_AZURE",
		"REACT_CLASSIFICATION_MODEL",
		"REACT_CLASSIFICATION_BATCH_SIZE",
		"REACT_CLASSIFICATION_FLUSH_TIMEOUT",
		"REACT_CLASSIFICATION_CLASSIFIER_FAMILY",
		"REACT_CLASSIFICATION_CLASSIFIER_VERSION",
		"REACT_CLASSIFICATION_REQUEST_TIMEOUT",
		"REACT_CLASSIFICATION_MAX_RETRIES",
		"REACT_CLASSIFICATION_RETRY_BACKOFF",
		"STORAGE_PROVIDER",
		"LOCAL_TRANSACTION_FILE_PATH",
	})
	t.Setenv("ENV", "production")

	tempDir := t.TempDir()
	writeEnvFile(t, tempDir, "SERVICE_NAME=from-dotenv\n")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	if _, err := Load(); err == nil {
		t.Fatal("expected Load() to fail when required production env vars are missing")
	}
}

// TestLoadDevelopmentReadsDotenvWithoutPresetEnv verifies the load development reads dotenv without preset env behavior and the expected outcome asserted below.
func TestLoadDevelopmentReadsDotenvWithoutPresetEnv(t *testing.T) {
	// This test verifies development mode can source ENV itself from .env when the
	// process starts without any preset runtime variables.
	clearEnvironment(t, []string{
		"ENV",
		"SERVICE_NAME",
		"HTTP_PORT",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
		"HTTP_SHUTDOWN_TIMEOUT",
		"DATABASE_URL",
		"DATABASE_PING_TIMEOUT",
		"MIGRATIONS_PATH",
		"LOG_LEVEL",
		"LOG_FILE_PATH",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
		"OPENAI_API_VERSION",
		"OPENAI_USE_AZURE",
		"REACT_CLASSIFICATION_MODEL",
		"REACT_CLASSIFICATION_BATCH_SIZE",
		"REACT_CLASSIFICATION_FLUSH_TIMEOUT",
		"REACT_CLASSIFICATION_CLASSIFIER_FAMILY",
		"REACT_CLASSIFICATION_CLASSIFIER_VERSION",
		"REACT_CLASSIFICATION_REQUEST_TIMEOUT",
		"REACT_CLASSIFICATION_MAX_RETRIES",
		"REACT_CLASSIFICATION_RETRY_BACKOFF",
		"STORAGE_PROVIDER",
		"LOCAL_TRANSACTION_FILE_PATH",
	})

	tempDir := t.TempDir()
	writeEnvFile(t, tempDir, "ENV=development\nSERVICE_NAME=paris-api\nHTTP_PORT=9000\nHTTP_READ_TIMEOUT=5s\nHTTP_WRITE_TIMEOUT=10s\nHTTP_IDLE_TIMEOUT=60s\nHTTP_SHUTDOWN_TIMEOUT=10s\nDATABASE_URL=postgres://admin:admin@localhost:5432/paris?sslmode=disable\nDATABASE_PING_TIMEOUT=5s\nMIGRATIONS_PATH=internal/infrastructure/db/migrations\nLOG_LEVEL=info\nLOG_FILE_PATH=tmp/logs/paris-api.jsonl\nOPENAI_API_KEY=test-key\nOPENAI_BASE_URL=https://dotenv.example.test\nOPENAI_API_VERSION=2025-02-01-preview\nOPENAI_USE_AZURE=true\nREACT_CLASSIFICATION_MODEL=gpt-4o-mini\nREACT_CLASSIFICATION_BATCH_SIZE=10\nREACT_CLASSIFICATION_FLUSH_TIMEOUT=2s\nREACT_CLASSIFICATION_CLASSIFIER_FAMILY=react\nREACT_CLASSIFICATION_CLASSIFIER_VERSION=v1\nREACT_CLASSIFICATION_REQUEST_TIMEOUT=30s\nREACT_CLASSIFICATION_MAX_RETRIES=2\nREACT_CLASSIFICATION_RETRY_BACKOFF=2s\nSTORAGE_PROVIDER=local\nLOCAL_TRANSACTION_FILE_PATH=.tx-files\n")

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(originalDir) })

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Env != "development" {
		t.Fatalf("Env = %q, want %q", config.Env, "development")
	}

	if config.Classification.OpenAIBaseURL != "https://dotenv.example.test" {
		t.Fatalf("Classification.OpenAIBaseURL = %q", config.Classification.OpenAIBaseURL)
	}

	if config.Classification.OpenAIAPIVersion != "2025-02-01-preview" {
		t.Fatalf("Classification.OpenAIAPIVersion = %q", config.Classification.OpenAIAPIVersion)
	}

	if !config.Classification.OpenAIUseAzure {
		t.Fatal("Classification.OpenAIUseAzure = false, want true")
	}
}

// TestLoadRejectsNegativeReactMaxRetries verifies startup validation still rejects
// invalid retry tuning after the legacy matching fields are removed.
func TestLoadRejectsNegativeReactMaxRetries(t *testing.T) {
	t.Setenv("ENV", "production")
	t.Setenv("SERVICE_NAME", "paris-api")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "1m")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("DATABASE_URL", "postgres://admin:admin@localhost:5432/paris?sslmode=disable")
	t.Setenv("DATABASE_PING_TIMEOUT", "5s")
	t.Setenv("MIGRATIONS_PATH", "internal/infrastructure/db/migrations")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE_PATH", "tmp/logs/paris-api.jsonl")
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("REACT_CLASSIFICATION_MODEL", "gpt-4o-mini")
	t.Setenv("REACT_CLASSIFICATION_BATCH_SIZE", "10")
	t.Setenv("REACT_CLASSIFICATION_FLUSH_TIMEOUT", "2s")
	t.Setenv("REACT_CLASSIFICATION_CLASSIFIER_FAMILY", "react")
	t.Setenv("REACT_CLASSIFICATION_CLASSIFIER_VERSION", "v1")
	t.Setenv("REACT_CLASSIFICATION_REQUEST_TIMEOUT", "30s")
	t.Setenv("REACT_CLASSIFICATION_MAX_RETRIES", "-1")
	t.Setenv("REACT_CLASSIFICATION_RETRY_BACKOFF", "2s")
	t.Setenv("STORAGE_PROVIDER", "local")
	t.Setenv("LOCAL_TRANSACTION_FILE_PATH", ".tx-files")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid retry error")
	}
}

func clearEnvironment(t *testing.T, keys []string) {
	t.Helper()

	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if exists {
			if err := os.Unsetenv(key); err != nil {
				t.Fatalf("os.Unsetenv(%q) error = %v", key, err)
			}

			key := key
			value := value
			t.Cleanup(func() {
				_ = os.Setenv(key, value)
			})
		}
	}
}

func writeEnvFile(t *testing.T, dir, contents string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(contents), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
}
