package entities

import (
	"encoding/hex"
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionUpload is the immutable record of an accepted transaction file upload.
type TransactionUpload struct {
	aggregateRoot
	id              valueobjects.UploadID
	groupID         valueobjects.GroupID
	fileName        string
	fileFormat      string
	contentMD5      string
	storageProvider string
	storageKey      string
	schemaVersion   string
	status          valueobjects.TransactionUploadStatus
	rowCount        int
	uploadedAt      time.Time
}

// RecordPreviewed records the upload preview event.
func (u *TransactionUpload) RecordPreviewed(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.PreviewTransactionUploadEventType, map[string]any{
		"action":    "preview",
		"resource":  "transaction_upload",
		"upload_id": u.ID().String(),
		"file_name": u.FileName(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordDownloaded records the upload download event.
func (u *TransactionUpload) RecordDownloaded(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DownloadTransactionUploadEventType, map[string]any{
		"action":    "download",
		"resource":  "transaction_upload",
		"upload_id": u.ID().String(),
		"file_name": u.FileName(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordCreated records the upload creation event.
func (u *TransactionUpload) RecordCreated(actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(u.uploadedAt, actorUserID, actorGroupID, events.CreateTransactionUploadEventType, map[string]any{
		"action":           "create",
		"resource":         "transaction_upload",
		"upload_id":        u.ID().String(),
		"file_name":        u.FileName(),
		"file_format":      u.FileFormat(),
		"schema_version":   u.SchemaVersion(),
		"row_count":        u.RowCount(),
		"storage_provider": u.StorageProvider(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordRead records the upload read event.
func (u *TransactionUpload) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetTransactionUploadEventType, map[string]any{
		"action":    "read",
		"resource":  "transaction_upload",
		"upload_id": u.ID().String(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the upload deletion event.
func (u *TransactionUpload) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteTransactionUploadEventType, map[string]any{
		"action":    "delete",
		"resource":  "transaction_upload",
		"upload_id": u.ID().String(),
		"file_name": u.FileName(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// NewTransactionUpload creates a valid transaction upload record.
func NewTransactionUpload(
	id valueobjects.UploadID,
	groupID valueobjects.GroupID,
	fileName string,
	fileFormat string,
	contentMD5 string,
	storageProvider string,
	storageKey string,
	schemaVersion string,
	status valueobjects.TransactionUploadStatus,
	rowCount int,
	uploadedAt time.Time,
) (*TransactionUpload, error) {
	if _, err := valueobjects.UploadIDFromString(id.String()); err != nil {
		return nil, err
	}

	if _, err := valueobjects.GroupIDFromString(groupID.String()); err != nil {
		return nil, err
	}

	normalizedFileName := strings.TrimSpace(fileName)
	if normalizedFileName == "" {
		return nil, domain.ErrInvalidUploadFile
	}

	normalizedFormat := strings.TrimSpace(strings.ToLower(fileFormat))
	if normalizedFormat == "" {
		return nil, domain.ErrInvalidFileFormat
	}

	normalizedHash := strings.TrimSpace(strings.ToLower(contentMD5))
	if len(normalizedHash) != 32 {
		return nil, domain.ErrInvalidFileHash
	}

	if _, err := hex.DecodeString(normalizedHash); err != nil {
		return nil, domain.ErrInvalidFileHash
	}

	normalizedStorageProvider := strings.TrimSpace(storageProvider)
	if normalizedStorageProvider == "" {
		return nil, domain.ErrInvalidStorage
	}

	normalizedStorageKey := strings.TrimSpace(storageKey)
	if normalizedStorageKey == "" {
		return nil, domain.ErrInvalidStorage
	}

	normalizedSchemaVersion := strings.TrimSpace(schemaVersion)
	if normalizedSchemaVersion == "" {
		return nil, domain.ErrInvalidSchema
	}

	normalizedStatus, err := valueobjects.TransactionUploadStatusFromString(status.String())
	if err != nil {
		return nil, err
	}

	if rowCount < 0 {
		return nil, domain.ErrInvalidRowCount
	}

	if rowCount == 0 && !normalizedStatus.Equal(valueobjects.FailedTransactionUploadStatus()) {
		return nil, domain.ErrInvalidRowCount
	}

	if rowCount > 0 && !normalizedStatus.Equal(valueobjects.UploadedTransactionUploadStatus()) {
		return nil, domain.ErrInvalidRowCount
	}

	if uploadedAt.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &TransactionUpload{
		id:              id,
		groupID:         groupID,
		fileName:        normalizedFileName,
		fileFormat:      normalizedFormat,
		contentMD5:      normalizedHash,
		storageProvider: normalizedStorageProvider,
		storageKey:      normalizedStorageKey,
		schemaVersion:   normalizedSchemaVersion,
		status:          normalizedStatus,
		rowCount:        rowCount,
		uploadedAt:      uploadedAt,
	}, nil
}

// ReconstituteTransactionUpload rebuilds a transaction upload from storage.
func ReconstituteTransactionUpload(
	id valueobjects.UploadID,
	groupID valueobjects.GroupID,
	fileName string,
	fileFormat string,
	contentMD5 string,
	storageProvider string,
	storageKey string,
	schemaVersion string,
	status valueobjects.TransactionUploadStatus,
	rowCount int,
	uploadedAt time.Time,
) (*TransactionUpload, error) {
	if _, err := valueobjects.GroupIDFromString(groupID.String()); err != nil {
		return nil, err
	}

	return &TransactionUpload{
		id:              id,
		groupID:         groupID,
		fileName:        fileName,
		fileFormat:      fileFormat,
		contentMD5:      contentMD5,
		storageProvider: storageProvider,
		storageKey:      storageKey,
		schemaVersion:   schemaVersion,
		status:          status,
		rowCount:        rowCount,
		uploadedAt:      uploadedAt,
	}, nil
}

// GroupID returns the owning group identifier.
func (u *TransactionUpload) GroupID() valueobjects.GroupID {
	return u.groupID
}

// ID returns the upload identifier.
func (u *TransactionUpload) ID() valueobjects.UploadID {
	return u.id
}

// FileName returns the original file name.
func (u *TransactionUpload) FileName() string {
	return u.fileName
}

// FileFormat returns the parsed file format.
func (u *TransactionUpload) FileFormat() string {
	return u.fileFormat
}

// ContentMD5 returns the MD5 content hash.
func (u *TransactionUpload) ContentMD5() string {
	return u.contentMD5
}

// StorageProvider returns the raw file storage provider.
func (u *TransactionUpload) StorageProvider() string {
	return u.storageProvider
}

// StorageKey returns the raw file storage key.
func (u *TransactionUpload) StorageKey() string {
	return u.storageKey
}

// SchemaVersion returns the schema version used during validation.
func (u *TransactionUpload) SchemaVersion() string {
	return u.schemaVersion
}

// Status returns the upload processing status.
func (u *TransactionUpload) Status() string {
	return u.status.String()
}

// RowCount returns the count of accepted data rows.
func (u *TransactionUpload) RowCount() int {
	return u.rowCount
}

// UploadedAt returns the upload acceptance timestamp.
func (u *TransactionUpload) UploadedAt() time.Time {
	return u.uploadedAt
}
