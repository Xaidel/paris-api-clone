package di

import (
	"context"
	"strings"
	"testing"

	outboundadapters "github.com/gyud-adb/paris-api/internal/adapters/outbound"
	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
)

// TestBuildRawFileStoreLocal verifies the build raw file store local behavior and the expected outcome asserted below.
func TestBuildRawFileStoreLocal(t *testing.T) {
	t.Parallel()

	store, err := buildRawFileStore(context.Background(), config.StorageConfig{Provider: "local", LocalTransactionPath: t.TempDir()})
	if err != nil {
		t.Fatalf("buildRawFileStore() error = %v", err)
	}

	if _, ok := store.(*outboundadapters.LocalRawFileStore); !ok {
		t.Fatalf("store type = %T, want %T", store, &outboundadapters.LocalRawFileStore{})
	}
}

// TestBuildRawFileStoreAzureBlobRequiresValidClient verifies the build raw file store azure blob requires valid client behavior and the expected outcome asserted below.
func TestBuildRawFileStoreAzureBlobRequiresValidClient(t *testing.T) {
	t.Parallel()

	_, err := buildRawFileStore(context.Background(), config.StorageConfig{Provider: "azure_blob", AzureBlobConnection: "bad-connection", AzureBlobContainer: "container"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "creating azure blob client") {
		t.Fatalf("err.Error() = %q, want substring %q", err.Error(), "creating azure blob client")
	}
}

// TestBuildRawFileStoreRejectsUnsupportedProvider verifies the build raw file store rejects unsupported provider behavior and the expected outcome asserted below.
func TestBuildRawFileStoreRejectsUnsupportedProvider(t *testing.T) {
	t.Parallel()

	_, err := buildRawFileStore(context.Background(), config.StorageConfig{Provider: "s3"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "unsupported raw file storage provider") {
		t.Fatalf("err.Error() = %q, want substring %q", err.Error(), "unsupported raw file storage provider")
	}
}
