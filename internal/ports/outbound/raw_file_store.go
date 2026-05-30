package ports

import "context"

// StoreRawFileCommand requests storage of a raw uploaded file.
type StoreRawFileCommand struct {
	UploadID  string
	FileName  string
	FileBytes []byte
}

// StoreRawFileResult returns the raw file storage reference.
//
// Provider is an adapter-defined identifier such as "local" or "azure_blob".
type StoreRawFileResult struct {
	Provider string
	Key      string
}

// ReadRawFileCommand requests reading a stored raw uploaded file.
type ReadRawFileCommand struct {
	Key string
}

// ReadRawFileResult returns the raw uploaded file bytes.
type ReadRawFileResult struct {
	FileBytes   []byte
	ContentType string
}

// DeleteRawFileCommand requests deletion of a raw uploaded file.
type DeleteRawFileCommand struct {
	Key string
}

// RawFileStore stores and deletes raw uploaded files.
type RawFileStore interface {
	Store(ctx context.Context, command StoreRawFileCommand) (StoreRawFileResult, error)
	Read(ctx context.Context, command ReadRawFileCommand) (ReadRawFileResult, error)
	Delete(ctx context.Context, command DeleteRawFileCommand) error
}
