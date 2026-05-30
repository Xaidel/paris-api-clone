package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

const localRawFileStoreProvider = "local"

// LocalRawFileStore stores raw uploaded files on the local filesystem.
type LocalRawFileStore struct {
	basePath string
}

var _ ports.RawFileStore = (*LocalRawFileStore)(nil)

// NewLocalRawFileStore builds a LocalRawFileStore.
func NewLocalRawFileStore(basePath string) *LocalRawFileStore {
	return &LocalRawFileStore{basePath: filepath.Clean(basePath)}
}

// Store saves the raw file under the configured directory.
func (s *LocalRawFileStore) Store(_ context.Context, command ports.StoreRawFileCommand) (ports.StoreRawFileResult, error) {
	if err := os.MkdirAll(s.basePath, 0o755); err != nil {
		return ports.StoreRawFileResult{}, fmt.Errorf("creating local raw file directory: %w", err)
	}

	key := filepath.ToSlash(filepath.Join(command.UploadID, sanitizeFileName(command.FileName)))
	targetPath := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return ports.StoreRawFileResult{}, fmt.Errorf("creating local raw file subdirectory: %w", err)
	}

	if err := os.WriteFile(targetPath, command.FileBytes, 0o600); err != nil {
		return ports.StoreRawFileResult{}, fmt.Errorf("writing local raw file: %w", err)
	}

	return ports.StoreRawFileResult{Provider: localRawFileStoreProvider, Key: key}, nil
}

// Read loads a previously stored local raw file.
func (s *LocalRawFileStore) Read(_ context.Context, command ports.ReadRawFileCommand) (ports.ReadRawFileResult, error) {
	basePath, err := filepath.Abs(s.basePath)
	if err != nil {
		return ports.ReadRawFileResult{}, fmt.Errorf("resolving local raw file base path: %w", err)
	}

	keyPath := filepath.Clean(filepath.FromSlash(command.Key))
	if filepath.IsAbs(keyPath) || filepath.VolumeName(keyPath) != "" || strings.HasPrefix(keyPath, string(filepath.Separator)) {
		return ports.ReadRawFileResult{}, fmt.Errorf("reading local raw file: path traversal is not allowed")
	}

	targetPath, err := filepath.Abs(filepath.Join(basePath, keyPath))
	if err != nil {
		return ports.ReadRawFileResult{}, fmt.Errorf("resolving local raw file path: %w", err)
	}

	relativePath, err := filepath.Rel(basePath, targetPath)
	if err != nil {
		return ports.ReadRawFileResult{}, fmt.Errorf("checking local raw file path: %w", err)
	}

	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return ports.ReadRawFileResult{}, fmt.Errorf("reading local raw file: path traversal is not allowed")
	}

	fileBytes, err := os.ReadFile(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ports.ReadRawFileResult{}, os.ErrNotExist
		}

		return ports.ReadRawFileResult{}, fmt.Errorf("reading local raw file: %w", err)
	}

	return ports.ReadRawFileResult{FileBytes: fileBytes}, nil
}

// Delete removes a previously stored local raw file.
func (s *LocalRawFileStore) Delete(_ context.Context, command ports.DeleteRawFileCommand) error {
	targetPath := filepath.Join(s.basePath, filepath.FromSlash(command.Key))
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing local raw file: %w", err)
	}

	return nil
}

func sanitizeFileName(fileName string) string {
	baseName := strings.TrimSpace(filepath.Base(fileName))
	if baseName == "." || baseName == string(filepath.Separator) || baseName == "" {
		return "upload.bin"
	}

	return strings.ReplaceAll(baseName, " ", "_")
}
