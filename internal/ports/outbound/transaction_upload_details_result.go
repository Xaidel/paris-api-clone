package ports

// TransactionUploadDetailsResult exposes one upload with its accepted transactions.
type TransactionUploadDetailsResult struct {
	TransactionUploadResult
	Transactions []TransactionResult
}
