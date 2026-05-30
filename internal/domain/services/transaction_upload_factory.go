package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionUploadFactory builds upload aggregates from parsed file content.
type TransactionUploadFactory struct{}

// NewTransactionUploadFactory builds a TransactionUploadFactory.
func NewTransactionUploadFactory() *TransactionUploadFactory {
	return &TransactionUploadFactory{}
}

// Build constructs a transaction upload aggregate and its transactions.
func (f *TransactionUploadFactory) Build(
	uploadID valueobjects.UploadID,
	groupID valueobjects.GroupID,
	fileName string,
	fileFormat string,
	contentMD5 string,
	storageProvider string,
	storageKey string,
	schemaVersion string,
	headers []string,
	rows [][]string,
	uploadedAt time.Time,
	newTransactionID func() (valueobjects.TransactionID, error),
) (*entities.TransactionUpload, []*entities.Transaction, error) {
	columnIndexes, err := indexTransactionColumns(headers)
	if err != nil {
		return nil, nil, err
	}

	transactions := make([]*entities.Transaction, 0, len(rows))
	for index, row := range rows {
		if shouldSkipPersistedTransactionRow(row, columnIndexes.referenceIndex) {
			continue
		}

		transactionID, err := newTransactionID()
		if err != nil {
			return nil, nil, fmt.Errorf("generating transaction id for row %d: %w", index+2, err)
		}

		transaction, err := newTransactionFromRow(transactionID, uploadID, index+2, row, columnIndexes, uploadedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("creating transaction entity for row %d: %w", index+2, err)
		}

		transactions = append(transactions, transaction)
	}

	upload, err := entities.NewTransactionUpload(uploadID, groupID, fileName, fileFormat, contentMD5, storageProvider, storageKey, schemaVersion, valueobjects.UploadedTransactionUploadStatus(), len(transactions), uploadedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("creating transaction upload entity: %w", err)
	}

	return upload, transactions, nil
}

type transactionColumnIndexes struct {
	productIndex             valueobjects.ColumnIndex
	processedYearIndex       valueobjects.ColumnIndex
	processedMonthIndex      valueobjects.ColumnIndex
	dmcIBIndex               valueobjects.ColumnIndex
	dmcIndex                 valueobjects.ColumnIndex
	partnerBankIndex         valueobjects.ColumnIndex
	referenceNumberIndex     valueobjects.ColumnIndex
	transactionValueIndex    valueobjects.ColumnIndex
	transactionCountIndex    valueobjects.ColumnIndex
	goodsDescriptionIndex    valueobjects.ColumnIndex
	goodsClassificationIndex valueobjects.ColumnIndex
	applicantCountryIndex    valueobjects.ColumnIndex
	beneficiaryCountryIndex  valueobjects.ColumnIndex
	sourceIndex              valueobjects.ColumnIndex
	destinationIndex         valueobjects.ColumnIndex
	tenorIndex               valueobjects.ColumnIndex
	esCategoryIndex          *valueobjects.ColumnIndex
	paAlignmentIndex         *valueobjects.ColumnIndex
	referenceIndex           *valueobjects.ColumnIndex
}

func newTransactionFromRow(transactionID valueobjects.TransactionID, uploadID valueobjects.UploadID, rowNumber int, row []string, indexes transactionColumnIndexes, createdAt time.Time) (*entities.Transaction, error) {
	processedYear, err := strconv.Atoi(strings.TrimSpace(rowValue(row, indexes.processedYearIndex)))
	if err != nil {
		return nil, err
	}

	processedMonth, err := strconv.Atoi(strings.TrimSpace(rowValue(row, indexes.processedMonthIndex)))
	if err != nil {
		return nil, err
	}

	transactionCount, err := strconv.Atoi(strings.TrimSpace(rowValue(row, indexes.transactionCountIndex)))
	if err != nil {
		return nil, err
	}

	return entities.NewUploadedTransaction(
		transactionID,
		uploadID,
		rowNumber,
		rowValue(row, indexes.productIndex),
		processedYear,
		processedMonth,
		rowValue(row, indexes.dmcIBIndex),
		rowValue(row, indexes.dmcIndex),
		rowValue(row, indexes.partnerBankIndex),
		rowValue(row, indexes.referenceNumberIndex),
		rowValue(row, indexes.transactionValueIndex),
		transactionCount,
		rowValue(row, indexes.goodsDescriptionIndex),
		rowValue(row, indexes.goodsClassificationIndex),
		rowValue(row, indexes.applicantCountryIndex),
		rowValue(row, indexes.beneficiaryCountryIndex),
		rowValue(row, indexes.sourceIndex),
		rowValue(row, indexes.destinationIndex),
		rowValue(row, indexes.tenorIndex),
		optionalRowValue(row, indexes.esCategoryIndex),
		optionalRowValue(row, indexes.paAlignmentIndex),
		createdAt,
	)
}

func indexTransactionColumns(headers []string) (transactionColumnIndexes, error) {
	indexedHeaders := make(map[string]valueobjects.ColumnIndex, len(headers))
	for index, header := range headers {
		columnIndex, err := valueobjects.NewColumnIndex(index)
		if err != nil {
			return transactionColumnIndexes{}, fmt.Errorf("building column index for header %q: %w", header, err)
		}

		indexedHeaders[normalizeTransactionHeader(header)] = columnIndex
	}

	lookup := func(candidates ...string) (valueobjects.ColumnIndex, error) {
		for _, candidate := range candidates {
			if index, ok := indexedHeaders[normalizeTransactionHeader(candidate)]; ok {
				return index, nil
			}
		}

		return valueobjects.ColumnIndex{}, fmt.Errorf("required transaction column is missing: %s", candidates[0])
	}

	optionalLookup := func(candidates ...string) *valueobjects.ColumnIndex {
		for _, candidate := range candidates {
			if index, ok := indexedHeaders[normalizeTransactionHeader(candidate)]; ok {
				columnIndex := index
				return &columnIndex
			}
		}

		return nil
	}

	transactionCountIndex, err := lookup("No. of Transactions")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	productIndex, err := lookup("Product")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	processedYearIndex, err := lookup("Year")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	processedMonthIndex, err := lookup("Month")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	dmcIBIndex, err := lookup("DMC:IB", "DMC : IB")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	dmcIndex, err := lookup("DMC")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	partnerBankIndex, err := lookup("Partner Bank")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	referenceNumberIndex, err := lookup("Reference Number", "Reference No.", "Reference No", "Ref")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	transactionValueIndex, err := lookup("Value of Transactions", "Value of Transaction")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	goodsDescriptionIndex, err := lookup("Goods Description")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	goodsClassificationIndex, err := lookup("Goods Classification (Sector)", "Goods Classification")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	applicantCountryIndex, err := lookup("Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Applicant (CG/RPA) or Sub-borrower (RCF) Country")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	beneficiaryCountryIndex, err := lookup("Beneficiary Country")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	sourceIndex, err := lookup("Source")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	destinationIndex, err := lookup("Destination")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	tenorIndex, err := lookup("Tenor > 1 year")
	if err != nil {
		return transactionColumnIndexes{}, err
	}

	esCategoryIndex := optionalLookup("E&S Category")

	paAlignmentIndex := optionalLookup("PA Alignment")

	referenceIndex := optionalLookup("Reference No.", "Reference No", "Ref")

	return transactionColumnIndexes{
		productIndex:             productIndex,
		processedYearIndex:       processedYearIndex,
		processedMonthIndex:      processedMonthIndex,
		dmcIBIndex:               dmcIBIndex,
		dmcIndex:                 dmcIndex,
		partnerBankIndex:         partnerBankIndex,
		referenceNumberIndex:     referenceNumberIndex,
		transactionValueIndex:    transactionValueIndex,
		transactionCountIndex:    transactionCountIndex,
		goodsDescriptionIndex:    goodsDescriptionIndex,
		goodsClassificationIndex: goodsClassificationIndex,
		applicantCountryIndex:    applicantCountryIndex,
		beneficiaryCountryIndex:  beneficiaryCountryIndex,
		sourceIndex:              sourceIndex,
		destinationIndex:         destinationIndex,
		tenorIndex:               tenorIndex,
		esCategoryIndex:          esCategoryIndex,
		paAlignmentIndex:         paAlignmentIndex,
		referenceIndex:           referenceIndex,
	}, nil
}

func shouldSkipPersistedTransactionRow(row []string, referenceIndex *valueobjects.ColumnIndex) bool {
	if referenceIndex == nil {
		return false
	}

	return strings.TrimSpace(rowValue(row, *referenceIndex)) == ""
}

func normalizeTransactionHeader(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	return strings.Join(strings.Fields(normalized), " ")
}

func rowValue(row []string, index valueobjects.ColumnIndex) string {
	if index.Int() >= len(row) {
		return ""
	}

	return row[index.Int()]
}

func optionalRowValue(row []string, index *valueobjects.ColumnIndex) string {
	if index == nil {
		return ""
	}

	return rowValue(row, *index)
}
