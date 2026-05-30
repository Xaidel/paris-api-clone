package usecases

import (
	"context"
	"testing"

	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestValidateTransactionFileUseCaseExecute verifies the validate transaction file use case execute behavior and the expected outcome asserted below.
func TestValidateTransactionFileUseCaseExecute(t *testing.T) {
	t.Parallel()

	useCase := NewValidateTransactionFileUseCase(domainservices.NewTransactionFileValidator(valueobjects.TransactionFileSchemaV1()))

	result, err := useCase.Execute(context.Background(), inboundports.ValidateTransactionFileCommand{
		Headers: []string{
			"Product",
			"Year",
			"Month",
			"DMC:IB",
			"DMC",
			"Partner Bank",
			"Reference Number",
			"Value of Transactions",
			"No. of Transactions",
			"Goods Description",
			"Goods Classification (Sector)",
			"Applicant (CG/RPA) or Sub-Borrower (RCF) Country",
			"Beneficiary Country",
			"Source",
			"Destination",
			"Tenor > 1 year",
			"E&S Category",
			"PA Alignment",
		},
		Rows: [][]string{{
			"CG",
			"2026",
			"4",
			"IB",
			"DMC",
			"Partner Bank",
			"REF-1",
			"698,436.80",
			"not-a-count",
			"Industrial machinery",
			"Capital Goods",
			"Philippines",
			"Japan",
			"Thailand",
			"Philippines",
			"N",
			"",
			"PA Aligned",
		}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.SchemaVersion != "transaction-file-v1" {
		t.Fatalf("result.SchemaVersion = %q, want %q", result.SchemaVersion, "transaction-file-v1")
	}

	if len(result.Columns) != 18 {
		t.Fatalf("len(result.Columns) = %d, want %d", len(result.Columns), 18)
	}

	if result.Columns[0].Name != "Product" {
		t.Fatalf("result.Columns[0].Name = %q, want %q", result.Columns[0].Name, "Product")
	}

	if result.Columns[8].Name != "No. of Transactions" {
		t.Fatalf("result.Columns[8].Name = %q, want %q", result.Columns[8].Name, "No. of Transactions")
	}

	if result.Valid {
		t.Fatal("result.Valid = true, want false")
	}

	if len(result.Errors) != 1 {
		t.Fatalf("len(result.Errors) = %d, want %d", len(result.Errors), 1)
	}

	if result.Errors[0].RowNumber != 2 {
		t.Fatalf("result.Errors[0].RowNumber = %d, want %d", result.Errors[0].RowNumber, 2)
	}

	if result.Errors[0].ColumnName != "No. of Transactions" {
		t.Fatalf("result.Errors[0].ColumnName = %q, want %q", result.Errors[0].ColumnName, "No. of Transactions")
	}
}
