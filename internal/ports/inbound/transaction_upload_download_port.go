package ports

import "context"

// DownloadTransactionUploadQuery requests one upload file for download.
type DownloadTransactionUploadQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DownloadTransactionUploadResult returns the upload record with file content.
type DownloadTransactionUploadResult struct {
	FileName      string
	ContentType   string
	ContentLength int
	FileBytes     []byte
}

// DownloadTransactionUploadPort downloads a single stored upload file.
type DownloadTransactionUploadPort interface {
	Execute(ctx context.Context, query DownloadTransactionUploadQuery) (DownloadTransactionUploadResult, error)
}
