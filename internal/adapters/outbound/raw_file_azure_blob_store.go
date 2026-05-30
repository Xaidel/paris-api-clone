package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

const azureBlobRawFileStoreProvider = "azure_blob"

type azureBlobClient interface {
	CreateContainer(ctx context.Context, containerName string) error
	UploadBuffer(ctx context.Context, containerName, blobName string, buffer []byte) error
	DownloadStream(ctx context.Context, containerName, blobName string) (*blob.DownloadStreamResponse, error)
	DeleteBlob(ctx context.Context, containerName, blobName string) error
}

// AzureBlobRawFileStore stores raw uploaded files in Azure Blob Storage.
type AzureBlobRawFileStore struct {
	client        azureBlobClient
	containerName string
}

var _ ports.RawFileStore = (*AzureBlobRawFileStore)(nil)

// NewAzureBlobRawFileStore builds an AzureBlobRawFileStore.
func NewAzureBlobRawFileStore(connectionString, containerName string) (*AzureBlobRawFileStore, error) {
	client, err := newAzureBlobClient(connectionString)
	if err != nil {
		return nil, fmt.Errorf("creating azure blob client: %w", err)
	}

	return newAzureBlobRawFileStore(client, containerName)
}

func newAzureBlobRawFileStore(client azureBlobClient, containerName string) (*AzureBlobRawFileStore, error) {
	if client == nil {
		return nil, fmt.Errorf("azure blob client is required")
	}

	normalizedContainerName := strings.TrimSpace(containerName)
	if normalizedContainerName == "" {
		return nil, fmt.Errorf("azure blob container name is required")
	}

	return &AzureBlobRawFileStore{client: client, containerName: normalizedContainerName}, nil
}

// Store uploads the raw file to Azure Blob Storage.
func (s *AzureBlobRawFileStore) Store(ctx context.Context, command ports.StoreRawFileCommand) (ports.StoreRawFileResult, error) {
	if err := s.client.CreateContainer(ctx, s.containerName); err != nil {
		return ports.StoreRawFileResult{}, fmt.Errorf("creating azure blob container: %w", err)
	}

	key := strings.Trim(command.UploadID+"/"+sanitizeFileName(command.FileName), "/")
	if err := s.client.UploadBuffer(ctx, s.containerName, key, command.FileBytes); err != nil {
		return ports.StoreRawFileResult{}, fmt.Errorf("uploading file to azure blob storage: %w", err)
	}

	return ports.StoreRawFileResult{Provider: azureBlobRawFileStoreProvider, Key: key}, nil
}

// Read downloads a raw file from Azure Blob Storage.
func (s *AzureBlobRawFileStore) Read(ctx context.Context, command ports.ReadRawFileCommand) (ports.ReadRawFileResult, error) {
	response, err := s.client.DownloadStream(ctx, s.containerName, command.Key)
	if err != nil {
		var responseErr *azcore.ResponseError
		if errors.As(err, &responseErr) && (bloberror.HasCode(responseErr, bloberror.BlobNotFound) || bloberror.HasCode(responseErr, bloberror.ContainerNotFound)) {
			return ports.ReadRawFileResult{}, os.ErrNotExist
		}

		return ports.ReadRawFileResult{}, fmt.Errorf("downloading file from azure blob storage: %w", err)
	}

	if response == nil || response.Body == nil {
		return ports.ReadRawFileResult{}, fmt.Errorf("downloading file from azure blob storage: empty response body")
	}
	defer response.Body.Close()

	fileBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return ports.ReadRawFileResult{}, fmt.Errorf("reading azure blob download stream: %w", err)
	}

	contentType := ""
	if response.ContentType != nil && strings.TrimSpace(*response.ContentType) != "" {
		contentType = strings.TrimSpace(*response.ContentType)
	}

	return ports.ReadRawFileResult{FileBytes: fileBytes, ContentType: contentType}, nil
}

// Delete removes a raw file from Azure Blob Storage.
func (s *AzureBlobRawFileStore) Delete(ctx context.Context, command ports.DeleteRawFileCommand) error {
	if err := s.client.DeleteBlob(ctx, s.containerName, command.Key); err != nil {
		return fmt.Errorf("deleting file from azure blob storage: %w", err)
	}

	return nil
}

type azureBlobSDKClient struct {
	client *azblob.Client
}

// CreateContainer ensures the target container exists before uploads proceed.
func (c *azureBlobSDKClient) CreateContainer(ctx context.Context, containerName string) error {
	if _, err := c.client.CreateContainer(ctx, containerName, nil); err != nil && !bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return err
	}

	return nil
}

// UploadBuffer writes the supplied file contents to the configured blob path.
func (c *azureBlobSDKClient) UploadBuffer(ctx context.Context, containerName, blobName string, buffer []byte) error {
	if _, err := c.client.UploadBuffer(ctx, containerName, blobName, buffer, nil); err != nil {
		return err
	}

	return nil
}

// DownloadStream opens a download stream for the configured blob path.
func (c *azureBlobSDKClient) DownloadStream(ctx context.Context, containerName, blobName string) (*blob.DownloadStreamResponse, error) {
	response, err := c.client.DownloadStream(ctx, containerName, blobName, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// DeleteBlob removes the blob and ignores not-found states for idempotent cleanup.
func (c *azureBlobSDKClient) DeleteBlob(ctx context.Context, containerName, blobName string) error {
	if _, err := c.client.DeleteBlob(ctx, containerName, blobName, nil); err != nil && !bloberror.HasCode(err, bloberror.BlobNotFound) && !bloberror.HasCode(err, bloberror.ContainerNotFound) {
		return err
	}

	return nil
}

func newAzureBlobClient(connectionString string) (azureBlobClient, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	return &azureBlobSDKClient{client: client}, nil
}

var _ azureBlobClient = (*azureBlobSDKClient)(nil)
