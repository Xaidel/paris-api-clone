package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	transactionUploadStatusUploadedValue = "uploaded"
	transactionUploadStatusFailedValue   = "failed"
)

// TransactionUploadStatus describes the upload processing state of a transaction.
type TransactionUploadStatus struct {
	value string
}

// UploadedTransactionUploadStatus returns the uploaded value.
func UploadedTransactionUploadStatus() TransactionUploadStatus {
	return TransactionUploadStatus{value: transactionUploadStatusUploadedValue}
}

// FailedTransactionUploadStatus returns the failed value.
func FailedTransactionUploadStatus() TransactionUploadStatus {
	return TransactionUploadStatus{value: transactionUploadStatusFailedValue}
}

// TransactionUploadStatusFromString parses and normalizes a transaction upload status value.
func TransactionUploadStatusFromString(value string) (TransactionUploadStatus, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case transactionUploadStatusUploadedValue:
		return UploadedTransactionUploadStatus(), nil
	case transactionUploadStatusFailedValue:
		return FailedTransactionUploadStatus(), nil
	default:
		return TransactionUploadStatus{}, domain.ErrInvalidTransactionUploadStatus
	}
}

// String returns the canonical transaction upload status value.
func (s TransactionUploadStatus) String() string {
	return s.value
}

// Equal reports whether two transaction upload statuses are equal.
func (s TransactionUploadStatus) Equal(other TransactionUploadStatus) bool {
	return s.value == other.value
}
