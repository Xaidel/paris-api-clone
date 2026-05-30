package adapters

import (
	"context"
	"testing"

	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/xuri/excelize/v2"
)

// TestTransactionFileParserParseCSV verifies the transaction file parser parse CSV behavior and the expected outcome asserted below.
func TestTransactionFileParserParseCSV(t *testing.T) {
	t.Parallel()

	parser := NewTransactionFileParser()
	result, err := parser.Parse(context.Background(), ports.ParseTransactionFileCommand{FileName: "transactions.csv", FileBytes: []byte("Product,Year,Month,DMC:IB,DMC,Partner Bank,Reference Number,Value of Transactions,No. of Transactions,Goods Description,Goods Classification (Sector),Applicant (CG/RPA) or Sub-Borrower (RCF) Country,Beneficiary Country,Source,Destination,Tenor > 1 year,E&S Category,PA Alignment\nCG,2026,4,IB,DMC,Partner Bank,REF-1,698436.80,1,Goods,Classification,Philippines,Japan,Thailand,Philippines,N,,PA Aligned\n")})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Format != "csv" {
		t.Fatalf("result.Format = %q, want %q", result.Format, "csv")
	}

	if len(result.Rows) != 1 {
		t.Fatalf("len(result.Rows) = %d, want %d", len(result.Rows), 1)
	}
}

// TestTransactionFileParserParseCSVMixedTransactionCounts verifies the parser keeps all non-blank rows regardless of transaction count.
func TestTransactionFileParserParseCSVMixedTransactionCounts(t *testing.T) {
	t.Parallel()

	parser := NewTransactionFileParser()
	result, err := parser.Parse(context.Background(), ports.ParseTransactionFileCommand{FileName: "transactions.csv", FileBytes: []byte("Product,Year,Month,DMC:IB,DMC,Partner Bank,Reference Number,Value of Transactions,No. of Transactions,Goods Description,Goods Classification (Sector),Applicant (CG/RPA) or Sub-Borrower (RCF) Country,Beneficiary Country,Source,Destination,Tenor > 1 year,E&S Category,PA Alignment\nCG,2026,4,IB,DMC,Partner Bank,REF-1,698436.80,1,Goods,Classification,Philippines,Japan,Thailand,Philippines,N,,PA Aligned\nCG,2026,4,IB,DMC,Partner Bank,REF-2,698436.80,2,Goods,Classification,Philippines,Japan,Thailand,Philippines,N,,PA Aligned\n")})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("len(result.Rows) = %d, want %d", len(result.Rows), 2)
	}

	if result.Rows[1][6] != "REF-2" {
		t.Fatalf("result.Rows[1][6] = %q, want %q", result.Rows[1][6], "REF-2")
	}
}

// TestTransactionFileParserParseXLSX verifies the transaction file parser parse XLSX behavior and the expected outcome asserted below.
func TestTransactionFileParserParseXLSX(t *testing.T) {
	t.Parallel()

	workbook := excelize.NewFile()
	defer func() { _ = workbook.Close() }()
	if err := workbook.SetSheetRow("Sheet1", "A1", &[]string{"Product", "Year", "Month", "DMC:IB", "DMC", "Partner Bank", "Reference Number", "Value of Transactions", "No. of Transactions", "Goods Description", "Goods Classification (Sector)", "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Beneficiary Country", "Source", "Destination", "Tenor > 1 year", "E&S Category", "PA Alignment"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}
	if err := workbook.SetSheetRow("Sheet1", "A2", &[]string{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}

	buffer, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer() error = %v", err)
	}

	parser := NewTransactionFileParser()
	result, err := parser.Parse(context.Background(), ports.ParseTransactionFileCommand{FileName: "transactions.xlsx", FileBytes: buffer.Bytes()})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if result.Format != "xlsx" {
		t.Fatalf("result.Format = %q, want %q", result.Format, "xlsx")
	}
}

// TestTransactionFileParserParseXLSXMixedTransactionCounts verifies the parser keeps mixed transaction-count rows for XLSX files.
func TestTransactionFileParserParseXLSXMixedTransactionCounts(t *testing.T) {
	t.Parallel()

	workbook := excelize.NewFile()
	defer func() { _ = workbook.Close() }()
	if err := workbook.SetSheetRow("Sheet1", "A1", &[]string{"Product", "Year", "Month", "DMC:IB", "DMC", "Partner Bank", "Reference Number", "Value of Transactions", "No. of Transactions", "Goods Description", "Goods Classification (Sector)", "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Beneficiary Country", "Source", "Destination", "Tenor > 1 year", "E&S Category", "PA Alignment"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}
	if err := workbook.SetSheetRow("Sheet1", "A2", &[]string{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}
	if err := workbook.SetSheetRow("Sheet1", "A3", &[]string{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-2", "698436.80", "2", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}); err != nil {
		t.Fatalf("SetSheetRow() error = %v", err)
	}

	buffer, err := workbook.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer() error = %v", err)
	}

	parser := NewTransactionFileParser()
	result, err := parser.Parse(context.Background(), ports.ParseTransactionFileCommand{FileName: "transactions.xlsx", FileBytes: buffer.Bytes()})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("len(result.Rows) = %d, want %d", len(result.Rows), 2)
	}
}

// TestSplitParsedRowsTrimsTrailingHeaderArtifacts verifies the split parsed rows trims trailing header artifacts behavior and the expected outcome asserted below.
func TestSplitParsedRowsTrimsTrailingHeaderArtifacts(t *testing.T) {
	t.Parallel()

	headers, rows, err := splitParsedRows([][]string{
		{"No. of Transactions", "Goods Description", "Goods Classification (Sector)"},
		{"1", "Goods", "Classification"},
		{"", "", "Goods Classification (Sector)"},
		{"", "", "Goods Classification (Sector)"},
	})
	if err != nil {
		t.Fatalf("splitParsedRows() error = %v", err)
	}

	if len(headers) != 3 {
		t.Fatalf("len(headers) = %d, want %d", len(headers), 3)
	}

	if len(rows) != 1 {
		t.Fatalf("len(rows) = %d, want %d", len(rows), 1)
	}
}
