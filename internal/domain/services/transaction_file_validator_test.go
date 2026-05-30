package services

import (
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestTransactionFileValidatorValidate verifies the transaction file validator validate behavior and the expected outcome asserted below.
func TestTransactionFileValidatorValidate(t *testing.T) {
	t.Parallel()

	headers := []string{
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
	}

	validRow := []string{
		"CG",
		"2026",
		"4",
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698,436.80",
		"1",
		"Industrial machinery",
		"Capital Goods",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		"N",
		"",
		"PA Aligned",
	}

	tests := []struct {
		name         string
		headers      []string
		rows         [][]string
		assertReport func(t *testing.T, report valueobjects.TransactionFileValidationReport)
	}{
		{
			name:    "accepts valid table",
			headers: headers,
			rows:    [][]string{validRow},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatal("report.Valid() = false, want true")
				}
			},
		},
		{
			name: "accepts source-of-truth alias header",
			headers: []string{
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
				"Goods Classification",
				"Applicant (CG/RPA) or Sub-borrower (RCF) Country",
				"Beneficiary Country",
				"Source",
				"Destination",
				"Tenor > 1 year",
				"E&S Category",
				"PA Alignment",
			},
			rows: [][]string{validRow},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatal("report.Valid() = false, want true")
				}
			},
		},
		{
			name: "accepts missing optional es category column",
			headers: []string{
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
				"PA Alignment",
			},
			rows: [][]string{{
				"CG",
				"2026",
				"4",
				"IB",
				"DMC",
				"Partner Bank",
				"REF-1",
				"698,436.80",
				"1",
				"Industrial machinery",
				"Capital Goods",
				"Philippines",
				"Japan",
				"Thailand",
				"Philippines",
				"N",
				"PA Aligned",
			}},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatalf("report.Valid() = false, want true; errors = %v", report.Errors())
				}
			},
		},
		{
			name: "accepts missing optional pa alignment column",
			headers: []string{
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
			},
			rows: [][]string{validRow[:17]},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatalf("report.Valid() = false, want true; errors = %v", report.Errors())
				}
			},
		},
		{
			name:    "reports type mismatches with row and column references",
			headers: headers,
			rows: [][]string{{
				"CG",
				"2026",
				"4",
				"IB",
				"DMC",
				"Partner Bank",
				"REF-1",
				"698,436.80",
				"bad-count",
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
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				errors := report.Errors()
				if len(errors) != 1 {
					t.Fatalf("len(report.Errors()) = %d, want %d", len(errors), 1)
				}

				if errors[0].ColumnName() != "No. of Transactions" {
					t.Fatalf("errors[0].ColumnName() = %q, want %q", errors[0].ColumnName(), "No. of Transactions")
				}
			},
		},
		{
			name:    "ignores trailing blank rows from spreadsheets",
			headers: headers,
			rows: [][]string{
				validRow,
				nil,
				{"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""},
			},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatalf("report.Valid() = false, want true with blank spreadsheet rows ignored; errors = %v", report.Errors())
				}
			},
		},
		{
			name: "ignores artifact rows when reference number is empty",
			headers: []string{
				"Product",
				"Year",
				"Month",
				"DMC : IB",
				"DMC",
				"Partner Bank",
				"Reference No.",
				"Value of Transaction",
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
			rows: [][]string{
				{"URPA", "2025", "2", "VIE : Vietnam Eximbank", "VIE", "Standard Chartered Bank", "TF958C7511", "698,436.80", "1", "WOOD PELLET", "Raw/Non-Energy Com", "SOUTH KOREA", "VIET NAM", "VIET NAM", "SOUTH KOREA", "N", "", "PA Aligned"},
				{"", "", "", "", "", "", "", "", "", "", "Goods Classification (Sector)", "", "", "", "", "", "", ""},
			},
			assertReport: func(t *testing.T, report valueobjects.TransactionFileValidationReport) {
				t.Helper()

				if !report.Valid() {
					t.Fatalf("report.Valid() = false, want true when reference number is empty on artifact rows; errors = %v", report.Errors())
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewTransactionFileValidator(valueobjects.TransactionFileSchemaV1())
			report := validator.Validate(tt.headers, tt.rows)
			tt.assertReport(t, report)
		})
	}
}
