package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateTransactionCommand requests the creation of a single transaction.
type CreateTransactionCommand struct {
	ClassificationTask  string
	Product             string
	ProcessedYear       int
	ProcessedMonth      int
	DMCIB               string
	DMC                 string
	PartnerBank         string
	ReferenceNumber     string
	TransactionValue    string
	TransactionCount    int
	GoodsDescription    string
	GoodsClassification string
	ApplicantCountry    string
	BeneficiaryCountry  string
	SourceCountry       string
	DestinationCountry  string
	TenorDescription    string
	ESCategory          string
	PAAlignment         string
	ActorUserID         string
	ActorGroupID        string
}

// CreateTransactionPort creates a single transaction.
type CreateTransactionPort interface {
	Execute(ctx context.Context, command CreateTransactionCommand) (outboundports.TransactionResult, error)
}
