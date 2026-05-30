package ports

const (
	TransactionUploadSkippedRowReasonMalformed           = "malformed"
	TransactionUploadSkippedRowReasonNotValidTransaction = "not_valid_transaction"
)

// TransactionUploadSkippedRow exposes one spreadsheet row that was not imported.
type TransactionUploadSkippedRow struct {
	RowNumber int
	Reason    string
}

// TransactionUploadResult exposes one accepted upload record.
type TransactionUploadResult struct {
	ID              string
	GroupID         string
	FileName        string
	FileFormat      string
	ContentMD5      string
	StorageProvider string
	StorageKey      string
	SchemaVersion   string
	Status          string
	RowCount        int
	UploadedAt      string
}
