package usecases

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateTransactionUseCase creates a new single transaction record.
type CreateTransactionUseCase struct {
	repository         outboundports.TransactionRepository
	processingQueue    outboundports.TransactionProcessingQueue
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	validator          *domainservices.TransactionFileValidator
	newID              func() (valueobjects.TransactionID, error)
	now                func() time.Time
}

// NewCreateTransactionUseCase builds a CreateTransactionUseCase.
func NewCreateTransactionUseCase(repository outboundports.TransactionRepository, processingQueue outboundports.TransactionProcessingQueue, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory, validator *domainservices.TransactionFileValidator) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		repository:         repository,
		processingQueue:    processingQueue,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		validator:          validator,
		newID:              valueobjects.NewTransactionID,
		now:                time.Now,
	}
}

const defaultTransactionClassificationTask = outboundports.TransactionClassifyReactTaskName

// Execute creates and persists a transaction record.
func (uc *CreateTransactionUseCase) Execute(ctx context.Context, command inboundports.CreateTransactionCommand) (outboundports.TransactionResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.TransactionResult{}, err
	}

	if err := uc.validate(command); err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("validating transaction record: %w", err)
	}

	createdBy, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("parsing actor user id: %w", err)
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("generating transaction id: %w", err)
	}

	createdAt := uc.now()
	transaction, err := entities.NewTransaction(
		id,
		command.Product,
		command.ProcessedYear,
		command.ProcessedMonth,
		command.DMCIB,
		command.DMC,
		command.PartnerBank,
		command.ReferenceNumber,
		command.TransactionValue,
		command.TransactionCount,
		command.GoodsDescription,
		command.GoodsClassification,
		command.ApplicantCountry,
		command.BeneficiaryCountry,
		command.SourceCountry,
		command.DestinationCountry,
		command.TenorDescription,
		command.ESCategory,
		command.PAAlignment,
		createdAt,
	)
	if err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("creating transaction record: %w", err)
	}
	transaction.SetCreatedBy(createdBy)

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, transaction, command.ActorUserID); err != nil {
			return fmt.Errorf("creating transaction record: %w", err)
		}

		if err := uc.processingQueue.Enqueue(txCtx, transactionClassificationTask(command.ClassificationTask), transaction.ID()); err != nil {
			return fmt.Errorf("queueing transaction for processing: %w", err)
		}

		if err := transaction.MarkProcessing(uc.now()); err != nil {
			return fmt.Errorf("marking transaction processing: %w", err)
		}

		if err := uc.repository.Update(txCtx, transaction); err != nil {
			return fmt.Errorf("updating transaction status: %w", err)
		}

		if err := transaction.RecordCreated(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording transaction creation event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, transaction.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing transaction events: %w", err)
		}

		return nil
	}); err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("creating transaction transaction: %w", err)
	}

	return newTransactionResult(transaction, nil, nil, nil), nil
}

func transactionClassificationTask(taskName string) string {
	// Legacy embedding and keyword classification is temporarily disabled.
	// Keep accepting the old task name at the API boundary, but remap it to the
	// ReAct worker until the deprecated queue and worker are removed.
	if taskName == outboundports.TransactionClassifyReactTaskName || taskName == outboundports.TransactionClassifyTaskName {
		return outboundports.TransactionClassifyReactTaskName
	}

	return defaultTransactionClassificationTask
}

func (uc *CreateTransactionUseCase) validate(command inboundports.CreateTransactionCommand) error {
	if uc.validator == nil {
		return nil
	}

	columns := uc.validator.Schema().Columns()
	headers := make([]string, 0, len(columns))
	row := make([]string, 0, len(columns))
	// Rebuild a single synthetic CSV row so the use case and bulk-upload flow use
	// the exact same schema validator and error vocabulary.
	for _, column := range columns {
		headers = append(headers, column.Name())
		row = append(row, transactionCommandValueForColumn(command, column.Name()))
	}

	report := uc.validator.Validate(headers, [][]string{row})
	if report.Valid() {
		return nil
	}

	fieldErrors := make([]domain.FieldValidationError, 0, len(report.Errors()))
	// Translate schema column names back into API field names so adapters can
	// return stable validation payloads without knowing spreadsheet terminology.
	for _, validationError := range report.Errors() {
		fieldErrors = append(fieldErrors, domain.NewFieldValidationError(transactionFieldNameForColumn(validationError.ColumnName()), validationError.Code(), validationError.Message()))
	}

	return domain.NewValidationError(fieldErrors)
}

func transactionCommandValueForColumn(command inboundports.CreateTransactionCommand, columnName string) string {
	// Keep the schema-column mapping centralized so create commands stay aligned
	// with file uploads that use the same validator.
	switch columnName {
	case "Product":
		return command.Product
	case "Year":
		return strconv.Itoa(command.ProcessedYear)
	case "Month":
		return strconv.Itoa(command.ProcessedMonth)
	case "DMC:IB":
		return command.DMCIB
	case "DMC":
		return command.DMC
	case "Partner Bank":
		return command.PartnerBank
	case "Reference Number":
		return command.ReferenceNumber
	case "Value of Transactions":
		return command.TransactionValue
	case "No. of Transactions":
		return strconv.Itoa(command.TransactionCount)
	case "Goods Description":
		return command.GoodsDescription
	case "Goods Classification (Sector)":
		return command.GoodsClassification
	case "Applicant (CG/RPA) or Sub-Borrower (RCF) Country":
		return command.ApplicantCountry
	case "Beneficiary Country":
		return command.BeneficiaryCountry
	case "Source":
		return command.SourceCountry
	case "Destination":
		return command.DestinationCountry
	case "Tenor > 1 year":
		return command.TenorDescription
	case "E&S Category":
		return command.ESCategory
	case "PA Alignment":
		return command.PAAlignment
	default:
		return ""
	}
}

func transactionFieldNameForColumn(columnName string) string {
	// Accept legacy column labels here so schema wording can evolve without
	// breaking the public validation contract.
	switch columnName {
	case "Product":
		return "product"
	case "Year":
		return "processed_year"
	case "Month":
		return "processed_month"
	case "DMC:IB", "DMC : IB":
		return "dmc_ib"
	case "DMC":
		return "dmc"
	case "Partner Bank":
		return "partner_bank"
	case "Reference Number", "Reference No.", "Reference No", "Ref":
		return "reference_number"
	case "Value of Transactions", "Value of Transaction":
		return "transaction_value"
	case "No. of Transactions":
		return "transaction_count"
	case "Goods Description":
		return "goods_description"
	case "Goods Classification (Sector)", "Goods Classification":
		return "goods_classification"
	case "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Applicant (CG/RPA) or Sub-borrower (RCF) Country":
		return "applicant_country"
	case "Beneficiary Country":
		return "beneficiary_country"
	case "Source":
		return "source_country"
	case "Destination":
		return "destination_country"
	case "Tenor > 1 year":
		return "tenor_description"
	case "E&S Category":
		return "es_category"
	case "PA Alignment":
		return "pa_alignment"
	default:
		return columnName
	}
}
