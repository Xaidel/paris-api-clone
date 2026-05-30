package adapters

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// TestLocalRawFileStoreStoreAndDelete verifies the local raw file store store and delete behavior and the expected outcome asserted below.
func TestLocalRawFileStoreStoreAndDelete(t *testing.T) {
	t.Parallel()

	basePath := filepath.Join(t.TempDir(), ".tx-files")
	store := NewLocalRawFileStore(basePath)

	result, err := store.Store(context.Background(), ports.StoreRawFileCommand{UploadID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", FileName: "transactions.csv", FileBytes: []byte("hello")})
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	targetPath := filepath.Join(basePath, filepath.FromSlash(result.Key))
	if _, err := os.Stat(targetPath); err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	readResult, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: result.Key})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(readResult.FileBytes) != "hello" {
		t.Fatalf("string(readResult.FileBytes) = %q, want %q", string(readResult.FileBytes), "hello")
	}

	if readResult.ContentType != "" {
		t.Fatalf("readResult.ContentType = %q, want empty string", readResult.ContentType)
	}

	if err := store.Delete(context.Background(), ports.DeleteRawFileCommand{Key: result.Key}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestLocalRawFileStoreReadRejectsTraversal(t *testing.T) {
	t.Parallel()

	store := NewLocalRawFileStore(filepath.Join(t.TempDir(), ".tx-files"))

	absolutePath := filepath.Join(t.TempDir(), "outside.txt")
	rootedPath := string(filepath.Separator) + filepath.Join("Windows", "Temp", "outside.txt")

	tests := []struct {
		name string
		key  string
	}{
		{name: "relative traversal", key: "../secrets.txt"},
		{name: "absolute path", key: absolutePath},
		{name: "drive rooted path", key: rootedPath},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: tc.key})
			if err == nil {
				t.Fatalf("Read(%q) error = nil, want traversal rejection", tc.key)
			}

			if !strings.Contains(err.Error(), "path traversal is not allowed") {
				t.Fatalf("Read(%q) error = %v, want traversal rejection message", tc.key, err)
			}
		})
	}
}

func TestLocalRawFileStoreReadMissingFile(t *testing.T) {
	t.Parallel()

	store := NewLocalRawFileStore(filepath.Join(t.TempDir(), ".tx-files"))

	_, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "missing/file.csv"})
	if err == nil {
		t.Fatal("Read() error = nil, want missing file error")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(err, os.ErrNotExist) = false, want true (err = %v)", err)
	}

	if !errors.Is(err, os.ErrNotExist) || err != os.ErrNotExist {
		t.Fatalf("Read() error = %v, want %v", err, os.ErrNotExist)
	}
}
