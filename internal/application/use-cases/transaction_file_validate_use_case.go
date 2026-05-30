package usecases

import (
	"context"
	"slices"

	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ValidateTransactionFileUseCase validates a parsed transaction file against the hard-coded schema.
type ValidateTransactionFileUseCase struct {
	validator *domainservices.TransactionFileValidator
}

// NewValidateTransactionFileUseCase builds a ValidateTransactionFileUseCase.
func NewValidateTransactionFileUseCase(validator *domainservices.TransactionFileValidator) *ValidateTransactionFileUseCase {
	return &ValidateTransactionFileUseCase{validator: validator}
}

// Execute validates a parsed table and returns structured validation output.
func (uc *ValidateTransactionFileUseCase) Execute(ctx context.Context, command inboundports.ValidateTransactionFileCommand) (inboundports.ValidateTransactionFileResult, error) {
	_ = ctx

	schema := uc.validator.Schema()
	report := uc.validator.Validate(command.Headers, command.Rows)

	return inboundports.ValidateTransactionFileResult{
		SchemaVersion: report.SchemaVersion(),
		Columns:       toTransactionFileSchemaColumns(schema.Columns()),
		Valid:         report.Valid(),
		Errors:        toTransactionFileValidationErrors(report.Errors()),
	}, nil
}

func toTransactionFileSchemaColumns(columns []valueobjects.TransactionFileColumn) []inboundports.TransactionFileSchemaColumn {
	result := make([]inboundports.TransactionFileSchemaColumn, 0, len(columns))
	for _, column := range columns {
		result = append(result, inboundports.TransactionFileSchemaColumn{
			Name:          column.Name(),
			Description:   column.Description(),
			DataType:      column.DataType(),
			Required:      column.Required(),
			AllowedValues: slices.Clone(column.AllowedValues()),
		})
	}

	return result
}

func toTransactionFileValidationErrors(validationErrors []valueobjects.TransactionFileValidationError) []outboundports.TransactionFileValidationError {
	result := make([]outboundports.TransactionFileValidationError, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		result = append(result, outboundports.TransactionFileValidationError{
			Code:        validationError.Code(),
			Message:     validationError.Message(),
			RowNumber:   validationError.RowNumber(),
			ColumnName:  validationError.ColumnName(),
			ColumnIndex: validationError.ColumnIndex(),
			Value:       validationError.Value(),
		})
	}

	return result
}
