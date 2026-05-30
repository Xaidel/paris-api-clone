package di

import (
	"context"
	"net/http"
	"testing"
	"time"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
)

// This test locks the Phase 5 config reshape at the same constructor seam used
// by Bootstrap so ReAct keeps reading its live OpenAI transport settings from
// ClassificationConfig.
func TestBuildReActChatModelUsesClassificationConfig(t *testing.T) {
	originalFactory := newOpenAIReActChatModel
	t.Cleanup(func() {
		newOpenAIReActChatModel = originalFactory
	})

	cfg := config.ClassificationConfig{
		OpenAIAPIKey:        "test-key",
		OpenAIBaseURL:       "https://example.test",
		OpenAIAPIVersion:    "2025-01-01-preview",
		OpenAIUseAzure:      true,
		ReactModel:          "gpt-4o-mini",
		ReactRequestTimeout: 30 * time.Second,
	}

	var captured *openaimodel.ChatModelConfig
	newOpenAIReActChatModel = func(_ context.Context, chatConfig *openaimodel.ChatModelConfig) (*openaimodel.ChatModel, error) {
		captured = chatConfig
		return nil, nil
	}

	if _, err := buildReActChatModel(context.Background(), cfg); err != nil {
		t.Fatalf("buildReActChatModel() error = %v", err)
	}

	if captured == nil {
		t.Fatal("captured config = nil")
	}

	if captured.APIKey != cfg.OpenAIAPIKey {
		t.Fatalf("APIKey = %q, want %q", captured.APIKey, cfg.OpenAIAPIKey)
	}

	if captured.BaseURL != cfg.OpenAIBaseURL {
		t.Fatalf("BaseURL = %q, want %q", captured.BaseURL, cfg.OpenAIBaseURL)
	}

	if captured.APIVersion != cfg.OpenAIAPIVersion {
		t.Fatalf("APIVersion = %q, want %q", captured.APIVersion, cfg.OpenAIAPIVersion)
	}

	if captured.ByAzure != cfg.OpenAIUseAzure {
		t.Fatalf("ByAzure = %t, want %t", captured.ByAzure, cfg.OpenAIUseAzure)
	}

	if captured.Model != cfg.ReactModel {
		t.Fatalf("Model = %q, want %q", captured.Model, cfg.ReactModel)
	}

	if captured.Timeout != cfg.ReactRequestTimeout {
		t.Fatalf("Timeout = %v, want %v", captured.Timeout, cfg.ReactRequestTimeout)
	}
}

// This test checks that HTTP server construction carries configuration through
// to the concrete http.Server without additional router requirements.
func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	server := NewHTTPServer(":9000", config.HTTPConfig{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}, http.NewServeMux())

	if server.Addr != ":9000" {
		t.Fatalf("Addr = %q", server.Addr)
	}
}

// This test keeps shutdown nil-safe so cleanup paths can call it defensively
// during partially failed bootstrap sequences.
func TestApplicationShutdownNil(t *testing.T) {
	t.Parallel()

	var application *Application
	if err := application.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

// This test fixes the address-formatting contract used by bootstrap when it
// turns the configured port into a net/http listener address.
func TestServerAddress(t *testing.T) {
	t.Parallel()

	if got := serverAddress("9000"); got != ":9000" {
		t.Fatalf("serverAddress() = %q", got)
	}
}
