package valueobjects

import "slices"

const (
	transactionFileDataTypeText    = "text"
	transactionFileDataTypeInteger = "integer"
	transactionFileDataTypeDecimal = "decimal"
)

// TransactionFileColumn describes a single uploaded transaction file column.
type TransactionFileColumn struct {
	name          string
	description   string
	dataType      string
	required      bool
	aliases       []string
	allowedValues []string
}

// Name returns the column name.
func (c TransactionFileColumn) Name() string {
	return c.name
}

// Description returns the business description for the column.
func (c TransactionFileColumn) Description() string {
	return c.description
}

// DataType returns the expected primitive type for the column.
func (c TransactionFileColumn) DataType() string {
	return c.dataType
}

// Required reports whether the column must be present and populated.
func (c TransactionFileColumn) Required() bool {
	return c.required
}

// AllowedValues returns the accepted values for enumerated columns.
func (c TransactionFileColumn) AllowedValues() []string {
	return slices.Clone(c.allowedValues)
}

// Aliases returns alternate accepted header names for the column.
func (c TransactionFileColumn) Aliases() []string {
	return slices.Clone(c.aliases)
}

func (c TransactionFileColumn) clone() TransactionFileColumn {
	return TransactionFileColumn{
		name:          c.name,
		description:   c.description,
		dataType:      c.dataType,
		required:      c.required,
		aliases:       slices.Clone(c.aliases),
		allowedValues: slices.Clone(c.allowedValues),
	}
}

// TransactionFileSchema describes the hard-coded upload schema.
type TransactionFileSchema struct {
	version string
	columns []TransactionFileColumn
}

// Version returns the schema version identifier.
func (s TransactionFileSchema) Version() string {
	return s.version
}

// Columns returns the ordered schema columns.
func (s TransactionFileSchema) Columns() []TransactionFileColumn {
	columns := make([]TransactionFileColumn, 0, len(s.columns))
	for _, column := range s.columns {
		columns = append(columns, column.clone())
	}

	return columns
}

// TransactionFileSchemaV1 returns the hard-coded schema for uploaded transaction files.
func TransactionFileSchemaV1() TransactionFileSchema {
	return TransactionFileSchema{
		version: "transaction-file-v1",
		columns: []TransactionFileColumn{
			{
				name:        "Product",
				description: "Product classification as supplied in the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Year",
				description: "Year of the transaction as supplied in the uploaded file",
				dataType:    transactionFileDataTypeInteger,
				required:    true,
			},
			{
				name:        "Month",
				description: "Month of the transaction as supplied in the uploaded file",
				dataType:    transactionFileDataTypeInteger,
				required:    true,
			},
			{
				name:        "DMC:IB",
				description: "Indicator",
				dataType:    transactionFileDataTypeText,
				required:    true,
				aliases:     []string{"DMC : IB"},
			},
			{
				name:        "DMC",
				description: "DMC classification as supplied in the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Partner Bank",
				description: "Name of the partner bank as supplied in the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Reference Number",
				description: "Reference number for the transaction as supplied in the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
				aliases:     []string{"Reference No.", "Reference No", "Ref"},
			},
			{
				name:        "Value of Transactions",
				description: "Value of transactions in the uploaded row",
				dataType:    transactionFileDataTypeDecimal,
				required:    true,
				aliases:     []string{"Value of Transaction"},
			},
			{
				name:        "No. of Transactions",
				description: "Count of transactions in the uploaded row",
				dataType:    transactionFileDataTypeInteger,
				required:    false,
			},
			{
				name:        "Goods Description",
				description: "Goods description as provided by client banks",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Goods Classification (Sector)",
				description: "Goods classification sector based on internal classification",
				dataType:    transactionFileDataTypeText,
				required:    true,
				aliases:     []string{"Goods Classification"},
			},
			{
				name:        "Applicant (CG/RPA) or Sub-Borrower (RCF) Country",
				description: "Country of the applicant or sub-borrower",
				dataType:    transactionFileDataTypeText,
				required:    true,
				aliases:     []string{"Applicant (CG/RPA) or Sub-borrower (RCF) Country"},
			},
			{
				name:        "Beneficiary Country",
				description: "Country of the beneficiary",
				dataType:    transactionFileDataTypeText,
				required:    false,
			},
			{
				name:        "Source",
				description: "Source country or origin reference from the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Destination",
				description: "Destination country reference from the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "Tenor > 1 year",
				description: "Tenor classification as supplied in the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    true,
			},
			{
				name:        "E&S Category",
				description: "Environmental and social category when provided",
				dataType:    transactionFileDataTypeText,
				required:    false,
			},
			{
				name:        "PA Alignment",
				description: "Paris alignment classification from the uploaded file",
				dataType:    transactionFileDataTypeText,
				required:    false,
			},
		},
	}
}
