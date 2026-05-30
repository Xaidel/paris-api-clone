package ports

import "context"

// GetTransactionNavigationQuery requests the previous and next transaction IDs
// around one transaction within an optional filtered transaction scope.
type GetTransactionNavigationQuery struct {
	ID                  string
	Classification      string
	Step                string
	UploadID            string
	CreatedAtFrom       string
	CreatedAtTo         string
	ApplicantCountry    string
	BeneficiaryCountry  string
	SourceCountry       string
	DestinationCountry  string
	TransactionCountMin string
	TransactionCountMax string
	Status              string
	ActorUserID         string
	ActorGroupID        string
}

// GetTransactionNavigationResult returns the neighboring transaction IDs.
type GetTransactionNavigationResult struct {
	TransactionID string
	PreviousID    *string
	NextID        *string
}

// GetTransactionNavigationPort gets the neighboring transaction IDs.
type GetTransactionNavigationPort interface {
	Execute(ctx context.Context, query GetTransactionNavigationQuery) (GetTransactionNavigationResult, error)
}
