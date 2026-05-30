package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListTransactionsQuery requests transaction records with an optional upload filter.
type ListTransactionsQuery struct {
	UploadID            string
	CreatedAtFrom       string
	CreatedAtTo         string
	ApplicantCountry    string
	BeneficiaryCountry  string
	SourceCountry       string
	DestinationCountry  string
	TransactionCountMin string
	TransactionCountMax string
	Classification      string
	Status              string
	SortBy              string
	SortOrder           string
	ActorUserID         string
	ActorGroupID        string
}

// ListTransactionsResult returns transaction records.
type ListTransactionsResult struct {
	Transactions []outboundports.TransactionResult
}

// ListTransactionsPort lists transaction records.
type ListTransactionsPort interface {
	Execute(ctx context.Context, query ListTransactionsQuery) (ListTransactionsResult, error)
}
