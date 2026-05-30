package adapters

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type azureBlobClientStub struct {
	createContainerErr error
	uploadErr          error
	downloadResponse   *blob.DownloadStreamResponse
	downloadErr        error
	deleteErr          error
	containerName      string
	blobName           string
	buffer             []byte
}

func (s *azureBlobClientStub) CreateContainer(_ context.Context, containerName string) error {
	s.containerName = containerName
	return s.createContainerErr
}

func (s *azureBlobClientStub) UploadBuffer(_ context.Context, containerName string, blobName string, buffer []byte) error {
	s.containerName = containerName
	s.blobName = blobName
	s.buffer = append([]byte(nil), buffer...)
	return s.uploadErr
}

func (s *azureBlobClientStub) DownloadStream(_ context.Context, containerName string, blobName string) (*blob.DownloadStreamResponse, error) {
	s.containerName = containerName
	s.blobName = blobName
	return s.downloadResponse, s.downloadErr
}

func (s *azureBlobClientStub) DeleteBlob(_ context.Context, containerName string, blobName string) error {
	s.containerName = containerName
	s.blobName = blobName
	return s.deleteErr
}

// TestAzureBlobRawFileStoreStoreAndDelete verifies the azure blob raw file store store and delete behavior and the expected outcome asserted below.
func TestAzureBlobRawFileStoreStoreAndDelete(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	result, err := store.Store(context.Background(), ports.StoreRawFileCommand{UploadID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", FileName: "transactions.csv", FileBytes: []byte("hello")})
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	if result.Provider != "azure_blob" {
		t.Fatalf("result.Provider = %q, want %q", result.Provider, "azure_blob")
	}

	client.downloadResponse = &blob.DownloadStreamResponse{
		DownloadResponse: blob.DownloadResponse{Body: io.NopCloser(strings.NewReader("hello")), ContentType: strPtr("text/csv")},
	}

	readResult, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: result.Key})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(readResult.FileBytes) != "hello" {
		t.Fatalf("string(readResult.FileBytes) = %q, want %q", string(readResult.FileBytes), "hello")
	}

	if readResult.ContentType != "text/csv" {
		t.Fatalf("readResult.ContentType = %q, want %q", readResult.ContentType, "text/csv")
	}

	if err := store.Delete(context.Background(), ports.DeleteRawFileCommand{Key: result.Key}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// TestNewAzureBlobRawFileStoreRequiresClient verifies the new azure blob raw file store requires client behavior and the expected outcome asserted below.
func TestNewAzureBlobRawFileStoreRequiresClient(t *testing.T) {
	t.Parallel()

	if _, err := newAzureBlobRawFileStore(nil, "container"); err == nil {
		t.Fatal("expected error")
	}
}

// TestAzureBlobRawFileStoreDeleteIgnoresBlobNotFound verifies the azure blob raw file store delete ignores blob not-found behavior and the expected outcome asserted below.
func TestAzureBlobRawFileStoreDeleteIgnoresBlobNotFound(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	if err := store.Delete(context.Background(), ports.DeleteRawFileCommand{Key: "missing"}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

// TestAzureBlobRawFileStoreStoreWrapsErrors verifies the azure blob raw file store store wraps errors behavior and the expected outcome asserted below.
func TestAzureBlobRawFileStoreStoreWrapsErrors(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{uploadErr: errors.New("boom")}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	if _, err := store.Store(context.Background(), ports.StoreRawFileCommand{UploadID: "01962b8f-aeb2-7e03-a8ff-1edce1300201", FileName: "transactions.csv", FileBytes: []byte("hello")}); err == nil {
		t.Fatal("expected error")
	}
}

func TestAzureBlobRawFileStoreReadWrapsDownloadErrors(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{downloadErr: errors.New("boom")}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	if _, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "uploads/file.csv"}); err == nil {
		t.Fatal("Read() error = nil, want wrapped download error")
	}
}

func TestAzureBlobRawFileStoreReadRejectsEmptyBody(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{downloadResponse: &blob.DownloadStreamResponse{}}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	if _, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "uploads/file.csv"}); err == nil {
		t.Fatal("Read() error = nil, want empty body error")
	}
}

func TestAzureBlobRawFileStoreReadUsesConfiguredContainerAndKey(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{downloadResponse: &blob.DownloadStreamResponse{DownloadResponse: blob.DownloadResponse{Body: io.NopCloser(strings.NewReader("hello"))}}}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	readKey := "uploads/read-only-check.csv"
	if _, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: readKey}); err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if client.containerName != "transaction-uploads" {
		t.Fatalf("client.containerName = %q, want %q", client.containerName, "transaction-uploads")
	}

	if client.blobName != readKey {
		t.Fatalf("client.blobName = %q, want %q", client.blobName, readKey)
	}
}

func TestAzureBlobRawFileStoreReadNormalizesMissingBlob(t *testing.T) {
	t.Parallel()

	code := string(bloberror.BlobNotFound)
	client := &azureBlobClientStub{downloadErr: &azcore.ResponseError{ErrorCode: code}}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	_, err = store.Read(context.Background(), ports.ReadRawFileCommand{Key: "uploads/missing.csv"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Read() error = %v, want %v", err, os.ErrNotExist)
	}
}

func TestAzureBlobRawFileStoreReadNormalizesMissingContainer(t *testing.T) {
	t.Parallel()

	code := string(bloberror.ContainerNotFound)
	client := &azureBlobClientStub{downloadErr: &azcore.ResponseError{ErrorCode: code}}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	_, err = store.Read(context.Background(), ports.ReadRawFileCommand{Key: "uploads/missing.csv"})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Read() error = %v, want %v", err, os.ErrNotExist)
	}
}

func TestAzureBlobRawFileStoreReadLeavesContentTypeEmptyWhenMissing(t *testing.T) {
	t.Parallel()

	client := &azureBlobClientStub{downloadResponse: &blob.DownloadStreamResponse{DownloadResponse: blob.DownloadResponse{Body: io.NopCloser(strings.NewReader("hello"))}}}
	store, err := newAzureBlobRawFileStore(client, "transaction-uploads")
	if err != nil {
		t.Fatalf("newAzureBlobRawFileStore() error = %v", err)
	}

	result, err := store.Read(context.Background(), ports.ReadRawFileCommand{Key: "uploads/file.csv"})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if result.ContentType != "" {
		t.Fatalf("result.ContentType = %q, want empty string", result.ContentType)
	}
}

func strPtr(value string) *string {
	return &value
}
