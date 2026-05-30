package db

import (
	"context"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
)

// TestNewPoolInvalidURL verifies the new pool invalid url behavior and the expected outcome asserted below.
func TestNewPoolInvalidURL(t *testing.T) {
	t.Parallel()

	_, err := NewPool(context.Background(), config.DatabaseConfig{
		URL:         "://invalid",
		PingTimeout: time.Second,
	})
	if err == nil {
		t.Fatal("expected NewPool() to fail for invalid URL")
	}
}
