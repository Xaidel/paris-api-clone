package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	transactionStatusPendingValue                = "pending"
	transactionStatusProcessingValue             = "processing"
	transactionStatusAIReviewedValue             = "ai-reviewed"
	transactionStatusFromPreviousTransactions    = "from-previous-transactions"
	transactionStatusProfessionallyReviewedValue = "professionally-reviewed"
	transactionStatusFailedValue                 = "failed"
	transactionStatusUploadedValue               = "uploaded"
	transactionStatusClassifiedValue             = "classified"
)

// TransactionStatus describes the review lifecycle state of a transaction.
type TransactionStatus struct {
	value string
}

// PendingTransactionStatus returns the pending value.
func PendingTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusPendingValue}
}

// ProcessingTransactionStatus returns the processing value.
func ProcessingTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusProcessingValue}
}

// AIReviewedTransactionStatus returns the ai-reviewed value.
func AIReviewedTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusAIReviewedValue}
}

// FromPreviousTransactionsTransactionStatus returns the from-previous-transactions value.
func FromPreviousTransactionsTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusFromPreviousTransactions}
}

// ProfessionallyReviewedTransactionStatus returns the professionally-reviewed value.
func ProfessionallyReviewedTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusProfessionallyReviewedValue}
}

// ClassifiedTransactionStatus returns the ai-reviewed value for legacy callers.
func ClassifiedTransactionStatus() TransactionStatus {
	return AIReviewedTransactionStatus()
}

// FailedTransactionStatus returns the failed value.
func FailedTransactionStatus() TransactionStatus {
	return TransactionStatus{value: transactionStatusFailedValue}
}

// TransactionStatusFromString parses and normalizes a transaction status value.
func TransactionStatusFromString(value string) (TransactionStatus, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case transactionStatusPendingValue:
		return PendingTransactionStatus(), nil
	case transactionStatusProcessingValue:
		return ProcessingTransactionStatus(), nil
	case transactionStatusAIReviewedValue:
		return AIReviewedTransactionStatus(), nil
	case transactionStatusFromPreviousTransactions:
		return FromPreviousTransactionsTransactionStatus(), nil
	case transactionStatusProfessionallyReviewedValue:
		return ProfessionallyReviewedTransactionStatus(), nil
	case transactionStatusFailedValue:
		return FailedTransactionStatus(), nil
	case transactionStatusUploadedValue:
		return TransactionStatus{value: transactionStatusUploadedValue}, nil
	case transactionStatusClassifiedValue:
		return AIReviewedTransactionStatus(), nil
	default:
		return TransactionStatus{}, domain.ErrInvalidTransactionStatus
	}
}

// String returns the canonical transaction status value.
func (s TransactionStatus) String() string {
	return s.value
}

// Equal reports whether two transaction statuses are equal.
func (s TransactionStatus) Equal(other TransactionStatus) bool {
	return s.value == other.value
}
