package entities

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// Transaction is a persisted transaction record.
type Transaction struct {
	aggregateRoot
	id                  valueobjects.TransactionID
	uploadID            *valueobjects.UploadID
	rowNumber           *int
	product             string
	processedYear       int
	processedMonth      int
	dmcIB               string
	dmc                 string
	partnerBank         string
	referenceNumber     string
	transactionValue    string
	classification      valueobjects.TransactionClassification
	status              valueobjects.TransactionStatus
	pipelineResult      *valueobjects.PipelineResult
	failureReason       string
	transactionCount    int
	goodsDescription    string
	goodsClassification string
	applicantCountry    string
	beneficiaryCountry  string
	sourceCountry       string
	destinationCountry  string
	tenorDescription    string
	esCategory          string
	paAlignment         string
	createdBy           valueobjects.UserID
	createdAt           time.Time
	updatedAt           time.Time
}

// RecordCreated records the transaction creation event.
func (t *Transaction) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateTransactionEventType, map[string]any{
		"action":               "create",
		"resource":             "transaction",
		"target_id":            t.ID().String(),
		"product":              t.Product(),
		"processed_year":       t.ProcessedYear(),
		"processed_month":      t.ProcessedMonth(),
		"dmc_ib":               t.DMCIB(),
		"dmc":                  t.DMC(),
		"partner_bank":         t.PartnerBank(),
		"reference_number":     t.ReferenceNumber(),
		"transaction_value":    t.TransactionValue(),
		"classification":       t.Classification(),
		"status":               t.Status(),
		"transaction_count":    t.TransactionCount(),
		"goods_description":    t.GoodsDescription(),
		"goods_classification": t.GoodsClassification(),
		"applicant_country":    t.ApplicantCountry(),
		"beneficiary_country":  t.BeneficiaryCountry(),
		"source_country":       t.SourceCountry(),
		"destination_country":  t.DestinationCountry(),
		"tenor_description":    t.TenorDescription(),
		"es_category":          t.ESCategory(),
		"pa_alignment":         t.PAAlignment(),
	})
	if err != nil {
		return err
	}

	t.recordDomainEvent(event)
	return nil
}

// RecordRead records the transaction read event.
func (t *Transaction) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetTransactionEventType, map[string]any{
		"action":    "read",
		"resource":  "transaction",
		"target_id": t.ID().String(),
	})
	if err != nil {
		return err
	}

	t.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the transaction deletion event.
func (t *Transaction) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteTransactionEventType, map[string]any{
		"action":    "delete",
		"resource":  "transaction",
		"target_id": t.ID().String(),
	})
	if err != nil {
		return err
	}

	t.recordDomainEvent(event)
	return nil
}

// NewTransaction creates a valid standalone transaction record.
func NewTransaction(
	id valueobjects.TransactionID,
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
	createdAt time.Time,
) (*Transaction, error) {
	classification := valueobjects.UnclassifiedTransactionClassification()
	status := valueobjects.ProcessingTransactionStatus()

	return newTransactionWithReviewState(
		id,
		nil,
		nil,
		product,
		processedYear,
		processedMonth,
		dmcIB,
		dmc,
		partnerBank,
		referenceNumber,
		transactionValue,
		classification,
		status,
		nil,
		transactionCount,
		goodsDescription,
		goodsClassification,
		applicantCountry,
		beneficiaryCountry,
		sourceCountry,
		destinationCountry,
		tenorDescription,
		esCategory,
		paAlignment,
		createdAt,
	)
}

func newTransactionWithReviewState(
	id valueobjects.TransactionID,
	uploadID *valueobjects.UploadID,
	rowNumber *int,
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
	classification valueobjects.TransactionClassification,
	status valueobjects.TransactionStatus,
	pipelineResult *valueobjects.PipelineResult,
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
	createdAt time.Time,
) (*Transaction, error) {
	// Validate identifier, timestamps, and workflow state before normalizing the
	// user-provided metadata. This keeps both create and upload constructors on
	// the same invariant path.
	if _, err := valueobjects.TransactionIDFromString(id.String()); err != nil {
		return nil, err
	}

	if createdAt.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	if _, err := valueobjects.TransactionClassificationFromString(classification.String()); err != nil {
		return nil, err
	}

	if _, err := valueobjects.TransactionStatusFromString(status.String()); err != nil {
		return nil, err
	}

	metadata, err := normalizeTransactionMetadata(
		product,
		processedYear,
		processedMonth,
		dmcIB,
		dmc,
		partnerBank,
		referenceNumber,
		transactionValue,
	)
	if err != nil {
		return nil, err
	}

	fields, err := normalizeTransactionFields(
		transactionCount,
		goodsDescription,
		goodsClassification,
		applicantCountry,
		beneficiaryCountry,
		sourceCountry,
		destinationCountry,
		tenorDescription,
		esCategory,
		paAlignment,
	)
	if err != nil {
		return nil, err
	}

	return &Transaction{
		id:                  id,
		uploadID:            uploadID,
		rowNumber:           rowNumber,
		product:             metadata.product,
		processedYear:       metadata.processedYear,
		processedMonth:      metadata.processedMonth,
		dmcIB:               metadata.dmcIB,
		dmc:                 metadata.dmc,
		partnerBank:         metadata.partnerBank,
		referenceNumber:     metadata.referenceNumber,
		transactionValue:    metadata.transactionValue,
		classification:      classification,
		status:              status,
		pipelineResult:      clonePipelineResult(pipelineResult),
		failureReason:       "",
		transactionCount:    fields.transactionCount,
		goodsDescription:    fields.goodsDescription,
		goodsClassification: fields.goodsClassification,
		applicantCountry:    fields.applicantCountry,
		beneficiaryCountry:  fields.beneficiaryCountry,
		sourceCountry:       fields.sourceCountry,
		destinationCountry:  fields.destinationCountry,
		tenorDescription:    fields.tenorDescription,
		esCategory:          fields.esCategory,
		paAlignment:         fields.paAlignment,
		createdAt:           createdAt,
		updatedAt:           createdAt,
	}, nil
}

// NewUploadedTransaction creates a valid uploaded transaction row.
func NewUploadedTransaction(
	id valueobjects.TransactionID,
	uploadID valueobjects.UploadID,
	rowNumber int,
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
	createdAt time.Time,
) (*Transaction, error) {
	classification := valueobjects.UnclassifiedTransactionClassification()
	status := valueobjects.ProcessingTransactionStatus()

	if _, err := valueobjects.UploadIDFromString(uploadID.String()); err != nil {
		return nil, err
	}

	if rowNumber < 2 {
		return nil, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("row_number", "invalid_value", "row_number must be greater than or equal to 2"),
		})
	}

	return newTransactionWithReviewState(
		id,
		&uploadID,
		&rowNumber,
		product,
		processedYear,
		processedMonth,
		dmcIB,
		dmc,
		partnerBank,
		referenceNumber,
		transactionValue,
		classification,
		status,
		nil,
		transactionCount,
		goodsDescription,
		goodsClassification,
		applicantCountry,
		beneficiaryCountry,
		sourceCountry,
		destinationCountry,
		tenorDescription,
		esCategory,
		paAlignment,
		createdAt,
	)
}

// ReconstituteTransaction rebuilds a transaction from storage.
func ReconstituteTransaction(
	id valueobjects.TransactionID,
	uploadID *valueobjects.UploadID,
	rowNumber *int,
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
	classification valueobjects.TransactionClassification,
	status valueobjects.TransactionStatus,
	pipelineResult *valueobjects.PipelineResult,
	failureReason string,
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
	createdBy valueobjects.UserID,
	createdAt time.Time,
	updatedAt time.Time,
) *Transaction {
	return &Transaction{
		id:                  id,
		uploadID:            uploadID,
		rowNumber:           rowNumber,
		product:             product,
		processedYear:       processedYear,
		processedMonth:      processedMonth,
		dmcIB:               dmcIB,
		dmc:                 dmc,
		partnerBank:         partnerBank,
		referenceNumber:     referenceNumber,
		transactionValue:    transactionValue,
		classification:      classification,
		status:              status,
		pipelineResult:      clonePipelineResult(pipelineResult),
		failureReason:       strings.TrimSpace(failureReason),
		transactionCount:    transactionCount,
		goodsDescription:    goodsDescription,
		goodsClassification: goodsClassification,
		applicantCountry:    applicantCountry,
		beneficiaryCountry:  beneficiaryCountry,
		sourceCountry:       sourceCountry,
		destinationCountry:  destinationCountry,
		tenorDescription:    tenorDescription,
		esCategory:          esCategory,
		paAlignment:         paAlignment,
		createdBy:           createdBy,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

// Product returns the uploaded product classification.
func (t *Transaction) Product() string {
	return t.product
}

// ProcessedYear returns the uploaded transaction year.
func (t *Transaction) ProcessedYear() int {
	return t.processedYear
}

// ProcessedMonth returns the uploaded transaction month.
func (t *Transaction) ProcessedMonth() int {
	return t.processedMonth
}

// DMCIB returns the uploaded DMC:IB value.
func (t *Transaction) DMCIB() string {
	return t.dmcIB
}

// DMC returns the uploaded DMC value.
func (t *Transaction) DMC() string {
	return t.dmc
}

// PartnerBank returns the uploaded partner bank value.
func (t *Transaction) PartnerBank() string {
	return t.partnerBank
}

// ReferenceNumber returns the uploaded reference number.
func (t *Transaction) ReferenceNumber() string {
	return t.referenceNumber
}

// TransactionValue returns the uploaded transaction value as provided.
func (t *Transaction) TransactionValue() string {
	return t.transactionValue
}

// Classification returns the transaction review classification.
func (t *Transaction) Classification() string {
	return t.classification.String()
}

// ClassificationValue returns the transaction review classification value object.
func (t *Transaction) ClassificationValue() valueobjects.TransactionClassification {
	return t.classification
}

// Status returns the transaction review lifecycle status.
func (t *Transaction) Status() string {
	return t.status.String()
}

// StatusValue returns the transaction review lifecycle status value object.
func (t *Transaction) StatusValue() valueobjects.TransactionStatus {
	return t.status
}

// PipelineResult returns the classification pipeline result when present.
func (t *Transaction) PipelineResult() *valueobjects.PipelineResult {
	return clonePipelineResult(t.pipelineResult)
}

// FailureReason returns the persisted pipeline failure reason when present.
func (t *Transaction) FailureReason() string {
	return t.failureReason
}

// ID returns the transaction identifier.
func (t *Transaction) ID() valueobjects.TransactionID {
	return t.id
}

// UploadID returns the parent upload identifier when present.
func (t *Transaction) UploadID() *valueobjects.UploadID {
	return t.uploadID
}

// RowNumber returns the spreadsheet row number when present.
func (t *Transaction) RowNumber() *int {
	return t.rowNumber
}

// TransactionCount returns the number of transactions.
func (t *Transaction) TransactionCount() int {
	return t.transactionCount
}

// GoodsDescription returns the goods description.
func (t *Transaction) GoodsDescription() string {
	return t.goodsDescription
}

// GoodsClassification returns the goods classification.
func (t *Transaction) GoodsClassification() string {
	return t.goodsClassification
}

// ApplicantCountry returns the applicant or sub-borrower country.
func (t *Transaction) ApplicantCountry() string {
	return t.applicantCountry
}

// BeneficiaryCountry returns the beneficiary country.
func (t *Transaction) BeneficiaryCountry() string {
	return t.beneficiaryCountry
}

// SourceCountry returns the source country value.
func (t *Transaction) SourceCountry() string {
	return t.sourceCountry
}

// DestinationCountry returns the destination country value.
func (t *Transaction) DestinationCountry() string {
	return t.destinationCountry
}

// TenorDescription returns the tenor description value.
func (t *Transaction) TenorDescription() string {
	return t.tenorDescription
}

// ESCategory returns the E&S category value.
func (t *Transaction) ESCategory() string {
	return t.esCategory
}

// PAAlignment returns the PA alignment value.
func (t *Transaction) PAAlignment() string {
	return t.paAlignment
}

// CreatedAt returns the persistence timestamp.
func (t *Transaction) CreatedAt() time.Time {
	return t.createdAt
}

// CreatedBy returns the creator user identifier.
func (t *Transaction) CreatedBy() string {
	return t.createdBy.String()
}

// SetCreatedBy sets the creator user identifier.
func (t *Transaction) SetCreatedBy(createdBy valueobjects.UserID) {
	t.createdBy = createdBy
}

// UpdatedAt returns the latest persistence timestamp.
func (t *Transaction) UpdatedAt() time.Time {
	return t.updatedAt
}

// Update updates the mutable transaction fields.
func (t *Transaction) Update(
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
	classification string,
	status string,
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
	updatedAt time.Time,
) error {
	if updatedAt.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	nextClassification, err := valueobjects.TransactionClassificationFromString(classification)
	if err != nil {
		return err
	}

	nextStatus, err := valueobjects.TransactionStatusFromString(status)
	if err != nil {
		return err
	}

	metadata, err := normalizeTransactionMetadata(
		product,
		processedYear,
		processedMonth,
		dmcIB,
		dmc,
		partnerBank,
		referenceNumber,
		transactionValue,
	)
	if err != nil {
		return err
	}

	fields, err := normalizeTransactionFields(
		transactionCount,
		goodsDescription,
		goodsClassification,
		applicantCountry,
		beneficiaryCountry,
		sourceCountry,
		destinationCountry,
		tenorDescription,
		esCategory,
		paAlignment,
	)
	if err != nil {
		return err
	}

	t.product = metadata.product
	t.processedYear = metadata.processedYear
	t.processedMonth = metadata.processedMonth
	t.dmcIB = metadata.dmcIB
	t.dmc = metadata.dmc
	t.partnerBank = metadata.partnerBank
	t.referenceNumber = metadata.referenceNumber
	t.transactionValue = metadata.transactionValue
	t.classification = nextClassification
	t.status = nextStatus
	t.transactionCount = fields.transactionCount
	t.goodsDescription = fields.goodsDescription
	t.goodsClassification = fields.goodsClassification
	t.applicantCountry = fields.applicantCountry
	t.beneficiaryCountry = fields.beneficiaryCountry
	t.sourceCountry = fields.sourceCountry
	t.destinationCountry = fields.destinationCountry
	t.tenorDescription = fields.tenorDescription
	t.esCategory = fields.esCategory
	t.paAlignment = fields.paAlignment
	t.updatedAt = updatedAt

	return nil
}

// MarkProcessing moves the transaction to the processing status.
func (t *Transaction) MarkProcessing(updatedAt time.Time) error {
	if updatedAt.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	t.status = valueobjects.ProcessingTransactionStatus()
	t.updatedAt = updatedAt
	return nil
}

// MarkClassified updates the transaction with a completed pipeline result.
func (t *Transaction) MarkClassified(classification valueobjects.TransactionClassification, pipelineResult valueobjects.PipelineResult, updatedAt time.Time) error {
	if !classification.Equal(valueobjects.AlignedTransactionClassification()) &&
		!classification.Equal(valueobjects.NotAlignedTransactionClassification()) &&
		!classification.Equal(valueobjects.NextStepTransactionClassification()) {
		return domain.ErrInvalidTransactionClassification
	}

	return t.markReviewed(classification, valueobjects.AIReviewedTransactionStatus(), pipelineResult, updatedAt)
}

// MarkClassifiedFromPreviousTransaction updates the transaction with a reused classification result.
func (t *Transaction) MarkClassifiedFromPreviousTransaction(classification valueobjects.TransactionClassification, pipelineResult valueobjects.PipelineResult, updatedAt time.Time) error {
	if !classification.Equal(valueobjects.AlignedTransactionClassification()) && !classification.Equal(valueobjects.NotAlignedTransactionClassification()) {
		return domain.ErrInvalidTransactionClassification
	}

	return t.markReviewed(classification, valueobjects.FromPreviousTransactionsTransactionStatus(), pipelineResult, updatedAt)
}

// MarkProfessionallyReviewed updates the transaction with a human-reviewed classification.
func (t *Transaction) MarkProfessionallyReviewed(classification valueobjects.TransactionClassification, updatedAt time.Time) error {
	if updatedAt.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	if !classification.Equal(valueobjects.AlignedTransactionClassification()) &&
		!classification.Equal(valueobjects.NotAlignedTransactionClassification()) &&
		!classification.Equal(valueobjects.NextStepTransactionClassification()) {
		return domain.ErrInvalidTransactionClassification
	}

	t.classification = classification
	t.status = valueobjects.ProfessionallyReviewedTransactionStatus()
	t.failureReason = ""
	t.updatedAt = updatedAt
	return nil
}

// HasStep3NextStepResult reports whether step 3 produced a next_step outcome.
func (t *Transaction) HasStep3NextStepResult() bool {
	if t == nil || t.pipelineResult == nil {
		return false
	}

	if reactResult := t.pipelineResult.React(); reactResult != nil {
		return reactResult.ExitStep() == 3 && reactResult.OverallClassification().Equal(valueobjects.NextStepTransactionClassification())
	}

	if t.pipelineResult.ExitStep() != 3 || !t.pipelineResult.FinalClassification().Equal(valueobjects.NextStepTransactionClassification()) {
		return false
	}

	step3Result := t.pipelineResult.Step3Result()
	if step3Result == nil || !step3Result.StepAlignment().Equal(valueobjects.UnalignedAlignment()) {
		return false
	}

	booleanResult := step3Result.BooleanResult()
	return booleanResult != nil && !*booleanResult
}

func (t *Transaction) markReviewed(
	classification valueobjects.TransactionClassification,
	status valueobjects.TransactionStatus,
	pipelineResult valueobjects.PipelineResult,
	updatedAt time.Time,
) error {
	// Centralize reviewed-state mutation so AI-reviewed and reused-classification
	// flows cannot drift on pipeline-result persistence semantics.
	if updatedAt.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	t.classification = classification
	t.status = status
	t.pipelineResult = clonePipelineResult(&pipelineResult)
	t.failureReason = ""
	t.updatedAt = updatedAt
	return nil
}

// MarkFailed moves the transaction to the failed status.
func (t *Transaction) MarkFailed(failureReason string, updatedAt time.Time) error {
	if updatedAt.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	t.status = valueobjects.FailedTransactionStatus()
	t.pipelineResult = nil
	t.failureReason = strings.TrimSpace(failureReason)
	t.updatedAt = updatedAt
	return nil
}

// Equal reports whether two transactions share the same identity.
func (t *Transaction) Equal(other *Transaction) bool {
	if t == nil || other == nil {
		return false
	}

	return t.id.Equal(other.id)
}

type normalizedTransactionFields struct {
	transactionCount    int
	goodsDescription    string
	goodsClassification string
	applicantCountry    string
	beneficiaryCountry  string
	sourceCountry       string
	destinationCountry  string
	tenorDescription    string
	esCategory          string
	paAlignment         string
}

type normalizedTransactionMetadata struct {
	product          string
	processedYear    int
	processedMonth   int
	dmcIB            string
	dmc              string
	partnerBank      string
	referenceNumber  string
	transactionValue string
}

func normalizeTransactionMetadata(
	product string,
	processedYear int,
	processedMonth int,
	dmcIB string,
	dmc string,
	partnerBank string,
	referenceNumber string,
	transactionValue string,
) (normalizedTransactionMetadata, error) {
	// Normalize whitespace once so both constructors and updates validate the same
	// canonical representation.
	metadata := normalizedTransactionMetadata{
		product:          strings.TrimSpace(product),
		processedYear:    processedYear,
		processedMonth:   processedMonth,
		dmcIB:            strings.TrimSpace(dmcIB),
		dmc:              strings.TrimSpace(dmc),
		partnerBank:      strings.TrimSpace(partnerBank),
		referenceNumber:  strings.TrimSpace(referenceNumber),
		transactionValue: strings.TrimSpace(transactionValue),
	}

	var validationErrors []domain.FieldValidationError
	// Required string fields are checked as a group to keep validation behavior
	// aligned across create, upload, and update paths.
	for _, requirement := range []struct {
		field string
		value string
	}{
		{field: "product", value: metadata.product},
		{field: "dmc_ib", value: metadata.dmcIB},
		{field: "dmc", value: metadata.dmc},
		{field: "partner_bank", value: metadata.partnerBank},
		{field: "reference_number", value: metadata.referenceNumber},
		{field: "transaction_value", value: metadata.transactionValue},
	} {
		if requirement.value == "" {
			validationErrors = append(validationErrors, domain.NewFieldValidationError(requirement.field, "required", requirement.field+" is required"))
		}
	}

	if metadata.processedYear <= 0 {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("processed_year", "invalid_value", "processed_year must be greater than 0"))
	}

	if metadata.processedMonth < 1 || metadata.processedMonth > 12 {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("processed_month", "invalid_value", "processed_month must be between 1 and 12"))
	}

	if _, err := parseTransactionValue(metadata.transactionValue); err != nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("transaction_value", "invalid_value", err.Error()))
	}

	if validationErr := domain.NewValidationError(validationErrors); validationErr != nil {
		return normalizedTransactionMetadata{}, validationErr
	}

	return metadata, nil
}

func parseTransactionValue(value string) (float64, error) {
	// Parse with big.Rat to accept exact decimal input without introducing float
	// rounding into validation logic.
	normalizedValue := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if _, ok := new(big.Rat).SetString(normalizedValue); !ok {
		return 0, fmt.Errorf("transaction_value must be a decimal")
	}

	return 0, nil
}

func normalizeTransactionFields(
	transactionCount int,
	goodsDescription string,
	goodsClassification string,
	applicantCountry string,
	beneficiaryCountry string,
	sourceCountry string,
	destinationCountry string,
	tenorDescription string,
	esCategory string,
	paAlignment string,
) (normalizedTransactionFields, error) {
	// Keep supplemental row fields on a separate normalization path from the core
	// transaction metadata so future schema changes stay localized.
	fields := normalizedTransactionFields{
		transactionCount:    transactionCount,
		goodsDescription:    strings.TrimSpace(goodsDescription),
		goodsClassification: strings.TrimSpace(goodsClassification),
		applicantCountry:    strings.TrimSpace(applicantCountry),
		beneficiaryCountry:  strings.TrimSpace(beneficiaryCountry),
		sourceCountry:       strings.TrimSpace(sourceCountry),
		destinationCountry:  strings.TrimSpace(destinationCountry),
		tenorDescription:    strings.TrimSpace(tenorDescription),
		esCategory:          strings.TrimSpace(esCategory),
		paAlignment:         strings.TrimSpace(paAlignment),
	}

	var validationErrors []domain.FieldValidationError
	if fields.transactionCount < 0 {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("transaction_count", "invalid_value", "transaction_count must be greater than or equal to 0"))
	}

	// These fields are required for downstream classification and auditing even
	// when they are not used to identify the transaction itself.
	for _, requirement := range []struct {
		field string
		value string
	}{
		{field: "goods_description", value: fields.goodsDescription},
		{field: "goods_classification", value: fields.goodsClassification},
		{field: "applicant_country", value: fields.applicantCountry},
		{field: "source_country", value: fields.sourceCountry},
		{field: "destination_country", value: fields.destinationCountry},
		{field: "tenor_description", value: fields.tenorDescription},
	} {
		if requirement.value == "" {
			validationErrors = append(validationErrors, domain.NewFieldValidationError(requirement.field, "required", requirement.field+" is required"))
		}
	}

	if validationErr := domain.NewValidationError(validationErrors); validationErr != nil {
		return normalizedTransactionFields{}, validationErr
	}

	return fields, nil
}

func clonePipelineResult(result *valueobjects.PipelineResult) *valueobjects.PipelineResult {
	// Return a copy so callers cannot mutate the entity's stored pipeline result
	// through an escaped pointer.
	if result == nil {
		return nil
	}

	resultCopy := *result
	return &resultCopy
}
