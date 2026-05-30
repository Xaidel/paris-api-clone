package observability

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
	"go.uber.org/zap"
)

// TestNewLogger verifies the new logger behavior and the expected outcome asserted below.
func TestNewLogger(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "paris-api.jsonl")

	logger, cleanup, err := NewLogger(config.LogConfig{
		Level:       "info",
		FilePath:    logPath,
		ServiceName: "paris-api",
		Environment: "test",
	})
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}

	logger.Info("hello world", zap.String("request_id", "req-123"))
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}

	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("os.ReadFile() error = %v", err)
	}

	text := string(contents)
	checks := []string{"\"message\":\"hello world\"", "\"service\":\"paris-api\"", "\"request_id\":\"req-123\""}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("log file does not contain %q", check)
		}
	}
}

// TestNewLoggerInvalidLevel verifies the new logger invalid level behavior and the expected outcome asserted below.
func TestNewLoggerInvalidLevel(t *testing.T) {
	t.Parallel()

	if _, _, err := NewLogger(config.LogConfig{Level: "loud", FilePath: filepath.Join(t.TempDir(), "app.log")}); err == nil {
		t.Fatal("expected invalid log level error")
	}
}
