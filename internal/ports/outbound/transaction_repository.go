package ports

import (
	"context"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

const (
	// TransactionSortBy* constants define the supported repository sort keys so
	// adapters and use cases share one query contract.
	TransactionSortByCreatedAt          = "created_at"
	TransactionSortByApplicantCountry   = "applicant_country"
	TransactionSortByBeneficiaryCountry = "beneficiary_country"
	TransactionSortBySourceCountry      = "source_country"
	TransactionSortByDestinationCountry = "destination_country"
	TransactionSortByTransactionCount   = "transaction_count"
	TransactionSortByClassification     = "classification"
	TransactionSortByStatus             = "status"
	TransactionSortOrderAscending       = "asc"
	TransactionSortOrderDescending      = "desc"
)

// TransactionFilter describes supported transaction list query filters.
type TransactionFilter struct {
	// UploadID scopes the query to one upload when present.
	UploadID *valueobjects.UploadID
	// CreatedAtFrom and CreatedAtTo bound the creation timestamp range.
	CreatedAtFrom *time.Time
	CreatedAtTo   *time.Time
	// Country fields filter the persisted transaction metadata exactly as stored.
	ApplicantCountry   *string
	BeneficiaryCountry *string
	SourceCountry      *string
	DestinationCountry *string
	// TransactionCountMin and TransactionCountMax bound the numeric count range.
	TransactionCountMin *int
	TransactionCountMax *int
	// Classification and Status filter by review outcome and lifecycle state.
	Classification *string
	Status         *string
	// SortBy and SortOrder choose one of the supported repository sort modes.
	SortBy    string
	SortOrder string
}

// TransactionNavigationLookup scopes one transaction-neighbor lookup.
// Classification duplicates Filter.Classification intentionally so navigation-
// specific consumers can read the canonical classification directly while
// adapters that already translate TransactionFilter still receive the same
// normalized value through Filter.Classification.
type TransactionNavigationLookup struct {
	TransactionID  valueobjects.TransactionID
	Filter         TransactionFilter
	Classification *string
	Step           *int
}

// TransactionNavigationResult returns one transaction and its neighbors.
type TransactionNavigationResult struct {
	TransactionID string
	PreviousID    *string
	NextID        *string
}

// TransactionRepository persists transaction records.
type TransactionRepository interface {
	// Create stores one new transaction and records the creating actor.
	Create(ctx context.Context, transaction *entities.Transaction, createdByUserID string) error
	// CreateMany stores a batch of uploaded transactions under one actor context.
	CreateMany(ctx context.Context, transactions []*entities.Transaction, createdByUserID string) error
	// Update persists the latest aggregate state for an existing transaction.
	Update(ctx context.Context, transaction *entities.Transaction) error
	// FindByID returns one transaction or nil when it does not exist.
	FindByID(ctx context.Context, id valueobjects.TransactionID) (*entities.Transaction, error)
	// FindHistoricalClassificationByExactGoodsDescription looks for a reusable
	// prior review result with the same normalized goods description.
	FindHistoricalClassificationByExactGoodsDescription(ctx context.Context, query HistoricalTransactionClassificationQuery) (*HistoricalTransactionClassificationMatch, error)
	// GetNavigation returns the previous and next transaction IDs in the filtered scope.
	GetNavigation(ctx context.Context, lookup TransactionNavigationLookup) (*TransactionNavigationResult, error)
	// List returns transactions matching the supplied filter contract.
	List(ctx context.Context, filter TransactionFilter) ([]*entities.Transaction, error)
	// ListByUploadIDs fetches all transactions associated with the given uploads.
	ListByUploadIDs(ctx context.Context, uploadIDs []valueobjects.UploadID) ([]*entities.Transaction, error)
	// HasProcessingByUploadID reports whether an upload still has transactions being processed.
	HasProcessingByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (bool, error)
	// DeleteByID removes one transaction by identity.
	DeleteByID(ctx context.Context, id valueobjects.TransactionID) error
	// DeleteByUploadID removes all transactions that belong to one upload.
	DeleteByUploadID(ctx context.Context, uploadID valueobjects.UploadID) error
}

// HistoricalTransactionClassificationQuery describes one historical reuse lookup.
type HistoricalTransactionClassificationQuery struct {
	// Classifier metadata scopes reuse to compatible historical review outputs.
	ClassifierFamily  string
	ClassifierVersion string
	ResultVersion     string
	// GoodsDescription is the normalized description to match exactly.
	GoodsDescription string
}

// HistoricalTransactionClassificationMatch describes one reusable historical review result.
type HistoricalTransactionClassificationMatch struct {
	// TransactionID identifies the prior reviewed transaction.
	TransactionID valueobjects.TransactionID
	// GoodsDescription echoes the matched normalized description.
	GoodsDescription string
	// Classification and Status describe the reusable review outcome.
	Classification valueobjects.TransactionClassification
	Status         valueobjects.TransactionStatus
	// ReviewResult carries the full stored pipeline result for downstream reuse.
	ReviewResult valueobjects.PipelineResult
}
