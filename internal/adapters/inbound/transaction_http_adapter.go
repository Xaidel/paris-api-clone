package adapters

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpTransactionAdapter exposes transaction CRUD over HTTP.
type HttpTransactionAdapter struct {
	createTransaction        inboundports.CreateTransactionPort
	getTransaction           inboundports.GetTransactionPort
	getTransactionNavigation inboundports.GetTransactionNavigationPort
	listTransactions         inboundports.ListTransactionsPort
	deleteTransaction        inboundports.DeleteTransactionPort
}

type transactionRequest struct {
	ClassificationTask  string `json:"classification_task"`
	Product             string `json:"product"`
	ProcessedYear       int    `json:"processed_year"`
	ProcessedMonth      int    `json:"processed_month"`
	DMCIB               string `json:"dmc_ib"`
	DMC                 string `json:"dmc"`
	PartnerBank         string `json:"partner_bank"`
	ReferenceNumber     string `json:"reference_number"`
	TransactionValue    string `json:"transaction_value"`
	TransactionCount    int    `json:"transaction_count"`
	GoodsDescription    string `json:"goods_description"`
	GoodsClassification string `json:"goods_classification"`
	ApplicantCountry    string `json:"applicant_country"`
	BeneficiaryCountry  string `json:"beneficiary_country"`
	SourceCountry       string `json:"source_country"`
	DestinationCountry  string `json:"destination_country"`
	TenorDescription    string `json:"tenor_description"`
	ESCategory          string `json:"es_category"`
	PAAlignment         string `json:"pa_alignment"`
}

type fieldErrorResponse struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type transactionResponse struct {
	ID                          string                           `json:"id"`
	UploadID                    string                           `json:"upload_id"`
	BatchID                     string                           `json:"batch_id"`
	ExitStepNumber              int                              `json:"exit_step_number"`
	ExitClassification          string                           `json:"exit_classification"`
	Status                      string                           `json:"status"`
	TransactionData             transactionDataResponse          `json:"transaction_data"`
	AutomatedClassificationData *automatedClassificationResponse `json:"automated_classification_data"`
	Step4ClassificationData     *step4ClassificationDataResponse `json:"step_4_classification_data"`
	Step5ClassificationData     *step5ClassificationDataResponse `json:"step_5_classification_data"`
	CreatedAt                   string                           `json:"created_at"`
	UpdatedAt                   string                           `json:"updated_at"`
}

type transactionDataResponse struct {
	RowNumber           int    `json:"row_number,omitempty"`
	Product             string `json:"product"`
	ProcessedYear       int    `json:"processed_year"`
	ProcessedMonth      int    `json:"processed_month"`
	DMCIB               string `json:"dmc_ib"`
	DMC                 string `json:"dmc"`
	PartnerBank         string `json:"partner_bank"`
	ReferenceNumber     string `json:"reference_number"`
	TransactionValue    string `json:"transaction_value"`
	FailureReason       string `json:"failure_reason"`
	TransactionCount    int    `json:"transaction_count"`
	GoodsDescription    string `json:"goods_description"`
	GoodsClassification string `json:"goods_classification"`
	ApplicantCountry    string `json:"applicant_country"`
	BeneficiaryCountry  string `json:"beneficiary_country"`
	SourceCountry       string `json:"source_country"`
	DestinationCountry  string `json:"destination_country"`
	TenorDescription    string `json:"tenor_description"`
	ESCategory          string `json:"es_category"`
	PAAlignment         string `json:"pa_alignment"`
	CreatedBy           string `json:"created_by"`
}

type listTransactionsResponse struct {
	Transactions []transactionResponse `json:"transactions"`
}

type transactionNavigationResponse struct {
	TransactionID string  `json:"transaction_id"`
	PreviousID    *string `json:"previous_id"`
	NextID        *string `json:"next_id"`
}

type automatedClassificationResponse struct {
	Step1  automatedStepResponse  `json:"step1"`
	Step2  *automatedStepResponse `json:"step2"`
	Step3  *automatedStepResponse `json:"step3"`
	Result string                 `json:"result"`
	Reason string                 `json:"reason"`
}

type automatedStepResponse struct {
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	Confidence *int   `json:"confidence,omitempty"`
}

type step4ClassificationDataResponse struct {
	IdentifiedSector      string `json:"identified_sector"`
	AdditionalInformation string `json:"additional_information"`
	Result                string `json:"result"`
}

type step5ClassificationDataResponse struct {
	ScreeningQuestion1 *screeningQuestionResponse `json:"screening_question_1"`
	ScreeningQuestion2 *screeningQuestionResponse `json:"screening_question_2"`
	ReviewerNotes      *string                    `json:"reviewer_notes,omitempty"`
	IsFinal            bool                       `json:"is_final"`
	Result             string                     `json:"result"`
}

type screeningQuestionResponse struct {
	Question      string `json:"question"`
	Answer        string `json:"answer"`
	Justification string `json:"justification"`
}

const (
	reactPipelineResultVersion = "react_v1"
	step1Question              = "Is it part of the Paris Agreement Alignment Exclusion List?"
	step2Question              = "Is it part of the ADB's Universally Aligned or U1 List?"
	step3Question              = "Does the transaction have a tenor of less than one year and be categorized as low risk as per ADB's E&S Risk Categorization?"
	step4Question              = "Does it fall in the high-emitting sectors list?"
	stepAnswerAligned          = "aligned"
	stepAnswerNotAligned       = "not-aligned"
	stepAnswerNextStep         = "next-step"
	automatedStepAnswerYes     = "yes"
	automatedStepAnswerNo      = "no"
)

// NewHttpTransactionAdapter builds an HttpTransactionAdapter.
func NewHttpTransactionAdapter(createTransaction inboundports.CreateTransactionPort, getTransaction inboundports.GetTransactionPort, getTransactionNavigation inboundports.GetTransactionNavigationPort, listTransactions inboundports.ListTransactionsPort, deleteTransaction inboundports.DeleteTransactionPort) *HttpTransactionAdapter {
	return &HttpTransactionAdapter{createTransaction: createTransaction, getTransaction: getTransaction, getTransactionNavigation: getTransactionNavigation, listTransactions: listTransactions, deleteTransaction: deleteTransaction}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpTransactionAdapter) RegisterRoutes(r *gin.Engine) {
	transactions := r.Group("/api/v1/transactions")
	transactions.POST("", a.CreateTransaction)
	transactions.GET("", a.ListTransactions)
	transactions.GET("/:id", a.GetTransaction)
	transactions.GET("/:id/navigation", a.GetTransactionNavigation)
	transactions.DELETE("/:id", a.DeleteTransaction)
}

// CreateTransaction handles POST /api/v1/transactions.
func (a *HttpTransactionAdapter) CreateTransaction(c *gin.Context) {
	requestBody, err := decodeJSONBody[transactionRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createTransaction.Execute(c.Request.Context(), inboundports.CreateTransactionCommand{
		ClassificationTask:  requestBody.ClassificationTask,
		Product:             requestBody.Product,
		ProcessedYear:       requestBody.ProcessedYear,
		ProcessedMonth:      requestBody.ProcessedMonth,
		DMCIB:               requestBody.DMCIB,
		DMC:                 requestBody.DMC,
		PartnerBank:         requestBody.PartnerBank,
		ReferenceNumber:     requestBody.ReferenceNumber,
		TransactionValue:    requestBody.TransactionValue,
		TransactionCount:    requestBody.TransactionCount,
		GoodsDescription:    requestBody.GoodsDescription,
		GoodsClassification: requestBody.GoodsClassification,
		ApplicantCountry:    requestBody.ApplicantCountry,
		BeneficiaryCountry:  requestBody.BeneficiaryCountry,
		SourceCountry:       requestBody.SourceCountry,
		DestinationCountry:  requestBody.DestinationCountry,
		TenorDescription:    requestBody.TenorDescription,
		ESCategory:          requestBody.ESCategory,
		PAAlignment:         requestBody.PAAlignment,
		ActorUserID:         c.GetHeader(actorUserIDHeader),
		ActorGroupID:        c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: toTransactionResponse(result)})
}

// GetTransaction handles GET /api/v1/transactions/:id.
func (a *HttpTransactionAdapter) GetTransaction(c *gin.Context) {
	result, err := a.getTransaction.Execute(c.Request.Context(), inboundports.GetTransactionQuery{ID: c.Param("id"), ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toTransactionResponse(result)})
}

// GetTransactionNavigation handles GET /api/v1/transactions/:id/navigation.
func (a *HttpTransactionAdapter) GetTransactionNavigation(c *gin.Context) {
	result, err := a.getTransactionNavigation.Execute(c.Request.Context(), inboundports.GetTransactionNavigationQuery{
		ID:                  c.Param("id"),
		Classification:      c.Query("classification"),
		Step:                c.Query("step"),
		UploadID:            c.Query("upload_id"),
		CreatedAtFrom:       c.Query("created_at_from"),
		CreatedAtTo:         c.Query("created_at_to"),
		ApplicantCountry:    c.Query("applicant_country"),
		BeneficiaryCountry:  c.Query("beneficiary_country"),
		SourceCountry:       c.Query("source_country"),
		DestinationCountry:  c.Query("destination_country"),
		TransactionCountMin: c.Query("transaction_count_min"),
		TransactionCountMax: c.Query("transaction_count_max"),
		Status:              c.Query("status"),
		ActorUserID:         c.GetHeader(actorUserIDHeader),
		ActorGroupID:        c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toTransactionNavigationResponse(result)})
}

// ListTransactions handles GET /api/v1/transactions.
func (a *HttpTransactionAdapter) ListTransactions(c *gin.Context) {
	result, err := a.listTransactions.Execute(c.Request.Context(), inboundports.ListTransactionsQuery{
		UploadID:            c.Query("upload_id"),
		CreatedAtFrom:       c.Query("created_at_from"),
		CreatedAtTo:         c.Query("created_at_to"),
		ApplicantCountry:    c.Query("applicant_country"),
		BeneficiaryCountry:  c.Query("beneficiary_country"),
		SourceCountry:       c.Query("source_country"),
		DestinationCountry:  c.Query("destination_country"),
		TransactionCountMin: c.Query("transaction_count_min"),
		TransactionCountMax: c.Query("transaction_count_max"),
		Classification:      c.Query("classification"),
		Status:              c.Query("status"),
		SortBy:              c.Query("sort_by"),
		SortOrder:           c.Query("sort_order"),
		ActorUserID:         c.GetHeader(actorUserIDHeader),
		ActorGroupID:        c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toListTransactionsResponse(result)})
}

// DeleteTransaction handles DELETE /api/v1/transactions/:id.
func (a *HttpTransactionAdapter) DeleteTransaction(c *gin.Context) {
	_, err := a.deleteTransaction.Execute(c.Request.Context(), inboundports.DeleteTransactionCommand{ID: c.Param("id"), ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *HttpTransactionAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var validationErr *domain.ValidationError
	var domainErr *domain.DomainError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &validationErr):
		fieldErrors := make([]fieldErrorResponse, 0, len(validationErr.Fields()))
		for _, fieldErr := range validationErr.Fields() {
			fieldErrors = append(fieldErrors, fieldErrorResponse{Field: fieldErr.Field(), Code: fieldErr.Code(), Message: fieldErr.Message()})
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "validation_error", "message": validationErr.Error(), "fields": fieldErrors}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: "internal server error"}})
	}
}

func toListTransactionsResponse(result inboundports.ListTransactionsResult) listTransactionsResponse {
	var transactions []transactionResponse
	if len(result.Transactions) > 0 {
		transactions = make([]transactionResponse, 0, len(result.Transactions))
		for _, transaction := range result.Transactions {
			transactions = append(transactions, toTransactionResponse(transaction))
		}
	}

	return listTransactionsResponse{Transactions: transactions}
}

func toTransactionNavigationResponse(result inboundports.GetTransactionNavigationResult) transactionNavigationResponse {
	return transactionNavigationResponse{
		TransactionID: result.TransactionID,
		PreviousID:    result.PreviousID,
		NextID:        result.NextID,
	}
}

func toTransactionResponse(result ports.TransactionResult) transactionResponse {
	return transactionResponse{
		ID:                          result.ID,
		UploadID:                    result.UploadID,
		BatchID:                     batchIDForTransaction(result.PipelineResult),
		ExitStepNumber:              exitStepNumberForTransaction(result),
		ExitClassification:          exitClassificationForTransaction(result),
		Status:                      result.Status,
		TransactionData:             toTransactionDataResponse(result),
		AutomatedClassificationData: automatedClassificationDataForTransaction(result.PipelineResult, result.TenorDescription),
		Step4ClassificationData:     toStep4ClassificationDataResponse(result.Step4Classification),
		Step5ClassificationData:     toStep5ClassificationDataResponse(result.Step5Classification),
		CreatedAt:                   result.CreatedAt,
		UpdatedAt:                   result.UpdatedAt,
	}
}

func toTransactionDataResponse(result ports.TransactionResult) transactionDataResponse {
	return transactionDataResponse{
		RowNumber:           result.RowNumber,
		Product:             result.Product,
		ProcessedYear:       result.ProcessedYear,
		ProcessedMonth:      result.ProcessedMonth,
		DMCIB:               result.DMCIB,
		DMC:                 result.DMC,
		PartnerBank:         result.PartnerBank,
		ReferenceNumber:     result.ReferenceNumber,
		TransactionValue:    result.TransactionValue,
		FailureReason:       result.FailureReason,
		TransactionCount:    result.TransactionCount,
		GoodsDescription:    result.GoodsDescription,
		GoodsClassification: result.GoodsClassification,
		ApplicantCountry:    result.ApplicantCountry,
		BeneficiaryCountry:  result.BeneficiaryCountry,
		SourceCountry:       result.SourceCountry,
		DestinationCountry:  result.DestinationCountry,
		TenorDescription:    result.TenorDescription,
		ESCategory:          result.ESCategory,
		PAAlignment:         result.PAAlignment,
		CreatedBy:           result.CreatedBy,
	}
}

func batchIDForTransaction(result *ports.PipelineResultDetails) string {
	if result == nil {
		return ""
	}

	return result.BatchID
}

func exitClassificationForTransaction(result ports.TransactionResult) string {
	classification := strings.TrimSpace(result.Classification)
	if result.PipelineResult == nil || !strings.EqualFold(strings.TrimSpace(result.Status), "ai-reviewed") {
		return classification
	}

	automated := automatedClassificationDataForTransaction(result.PipelineResult, result.TenorDescription)
	if automated == nil {
		return classification
	}

	return classificationToAPIValue(automated.Result, classification)
}

func exitStepNumberForTransaction(result ports.TransactionResult) int {
	classification := normalizeClassificationValue(result.Classification)
	if classification != stepAnswerAligned && classification != stepAnswerNotAligned {
		return 0
	}

	if result.Step5Classification != nil {
		return 5
	}

	if result.Step4Classification != nil {
		return 4
	}

	if result.PipelineResult == nil {
		return 0
	}

	exitStepNumber := resolveExitStepNumber(result.PipelineResult)
	return exitStepNumber
}

func automatedClassificationDataForTransaction(result *ports.PipelineResultDetails, tenorDescription string) *automatedClassificationResponse {
	if result == nil {
		return nil
	}

	exitStepNumber := resolveExitStepNumber(result)
	return toAutomatedClassificationResponse(result, tenorDescription, exitStepNumber)
}

func toStep4ClassificationDataResponse(result *ports.TransactionStep4Details) *step4ClassificationDataResponse {
	if result == nil {
		return nil
	}

	return &step4ClassificationDataResponse{
		IdentifiedSector:      result.IdentifiedSector,
		AdditionalInformation: result.AdditionalInformation,
		Result:                result.Result,
	}
}

func toStep5ClassificationDataResponse(result *ports.TransactionStep5Details) *step5ClassificationDataResponse {
	if result == nil {
		return nil
	}

	return &step5ClassificationDataResponse{
		ScreeningQuestion1: &screeningQuestionResponse{
			Question:      step5ScreeningQuestion1Text,
			Answer:        toStep5Answer(result.ScreeningQuestion1.Answer),
			Justification: result.ScreeningQuestion1.Justification,
		},
		ScreeningQuestion2: &screeningQuestionResponse{
			Question:      step5ScreeningQuestion2Text,
			Answer:        toStep5Answer(result.ScreeningQuestion2.Answer),
			Justification: result.ScreeningQuestion2.Justification,
		},
		ReviewerNotes: result.ReviewerNotes,
		IsFinal:       result.IsFinal,
		Result:        result.Result,
	}
}

func toAutomatedClassificationResponse(result *ports.PipelineResultDetails, tenorDescription string, exitStepNumber int) *automatedClassificationResponse {
	if result == nil {
		return nil
	}

	step1 := toAutomatedStep1Response(result)
	var step2 *automatedStepResponse
	if exitStepNumber >= 2 {
		step2 = toAutomatedStep2Response(result)
	}

	var step3 *automatedStepResponse
	if exitStepNumber >= 3 {
		step3 = toAutomatedStep3Response(result, tenorDescription)
	}

	resolvedResult := resolveAutomatedClassificationResult(step1, step2, step3, result)

	return &automatedClassificationResponse{
		Step1:  step1,
		Step2:  step2,
		Step3:  step3,
		Result: resolvedResult,
		Reason: resolveAutomatedClassificationReason(result, exitStepNumber, resolvedResult),
	}
}

func toAutomatedStep1Response(result *ports.PipelineResultDetails) automatedStepResponse {
	if result != nil && result.Version == reactPipelineResultVersion {
		return automatedStepResponse{
			Question:   step1Question,
			Answer:     toAutomatedStepAnswer(result.NotAlignedListMatch),
			Confidence: automatedConfidencePointer(result.NotAlignedListMatchConfidence),
		}
	}

	return automatedStepResponse{
		Question: step1Question,
		Answer:   toAutomatedStepAnswer(isAlignedStepResult(result.Step1Result)),
	}
}

func toAutomatedStep2Response(result *ports.PipelineResultDetails) *automatedStepResponse {
	if result == nil {
		return nil
	}

	if result.Version == reactPipelineResultVersion {
		return &automatedStepResponse{
			Question:   step2Question,
			Answer:     toAutomatedStepAnswer(result.AlignedListMatch),
			Confidence: automatedConfidencePointer(result.AlignedListMatchConfidence),
		}
	}

	if result.Step2Result == nil {
		return nil
	}

	return &automatedStepResponse{
		Question: step2Question,
		Answer:   toAutomatedStepAnswer(isAlignedStepResult(*result.Step2Result)),
	}
}

func toAutomatedStep3Response(result *ports.PipelineResultDetails, tenorDescription string) *automatedStepResponse {
	if result == nil {
		return nil
	}

	if result.Version == reactPipelineResultVersion {
		return &automatedStepResponse{
			Question: step3Question,
			Answer:   toAutomatedStepAnswer(inferStep3LowRisk(tenorDescription)),
		}
	}

	if result.Step3Result == nil {
		return nil
	}

	return &automatedStepResponse{
		Question: step3Question,
		Answer:   toAutomatedStepAnswer(isPositiveBooleanStepResult(*result.Step3Result)),
	}
}

func resolveExitStepNumber(result *ports.PipelineResultDetails) int {
	if result == nil {
		return 0
	}

	if result.ExitStep >= 1 && result.ExitStep <= 5 {
		return result.ExitStep
	}

	if result.Step3Result != nil {
		return 3
	}

	if result.Step2Result != nil {
		return 2
	}

	classification := normalizeClassificationValue(result.FinalClassification)

	switch classification {
	case stepAnswerNotAligned:
		if result.Version == reactPipelineResultVersion && result.NotAlignedListMatch {
			return 1
		}
		return 1
	case stepAnswerAligned:
		if result.Version == reactPipelineResultVersion && result.AlignedListMatch {
			return 2
		}
		return 3
	case stepAnswerNextStep:
		return 3
	default:
		return 1
	}
}

func resolveAutomatedClassificationResult(step1 automatedStepResponse, step2, step3 *automatedStepResponse, result *ports.PipelineResultDetails) string {
	// Step 1 yes with later-step data should be impossible, but if malformed
	// pipeline output reaches this adapter we keep step 1 precedence.
	if step1.Answer == automatedStepAnswerYes {
		return stepAnswerNotAligned
	}

	if step2 != nil && step2.Answer == automatedStepAnswerYes {
		return stepAnswerAligned
	}

	if step3 != nil {
		if step3.Answer == automatedStepAnswerYes {
			return stepAnswerAligned
		}
		return stepAnswerNextStep
	}

	if result != nil {
		classification := normalizeClassificationValue(result.FinalClassification)
		if classification != "" {
			return classification
		}
	}

	return stepAnswerNextStep
}

func resolveAutomatedClassificationReason(result *ports.PipelineResultDetails, exitStepNumber int, automatedResult string) string {
	if result == nil {
		return ""
	}

	rawReason := firstAutomatedReason(result, exitStepNumber)
	if exitStepNumber != 3 {
		return rawReason
	}

	cleanReason := sanitizeAutomatedReason(rawReason, automatedResult)

	return joinReasonParts(cleanReason, step3OutcomeReasonSuffix(automatedResult))
}

func firstAutomatedReason(result *ports.PipelineResultDetails, exitStepNumber int) string {
	if result == nil {
		return ""
	}

	if reason := strings.TrimSpace(result.Reason); reason != "" {
		return reason
	}

	if exitStepNumber >= 3 && result.Step3Result != nil {
		if reason := strings.TrimSpace(result.Step3Result.Reason); reason != "" {
			return reason
		}
	}

	if exitStepNumber >= 2 && result.Step2Result != nil {
		if reason := strings.TrimSpace(result.Step2Result.Reason); reason != "" {
			return reason
		}
	}

	if reason := strings.TrimSpace(result.Step1Result.Reason); reason != "" {
		return reason
	}

	return ""
}

func sanitizeAutomatedReason(reason string, automatedResult string) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return ""
	}

	removeVerificationClauses := normalizeClassificationValue(automatedResult) == stepAnswerAligned
	parts := strings.Split(trimmed, ";")
	kept := make([]string, 0, len(parts))
	removedAny := false
	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		if normalized == "" {
			removedAny = true
			continue
		}

		if isKnownDebugReasonTrailer(normalized) {
			removedAny = true
			continue
		}

		if removeVerificationClauses && isAlignedVerificationReasonPart(normalized) {
			removedAny = true
			continue
		}

		kept = append(kept, normalized)
	}

	if !removedAny {
		return trimmed
	}

	if len(kept) == 0 {
		return ""
	}

	return strings.Join(kept, "; ")
}

func isAlignedVerificationReasonPart(reasonPart string) bool {
	normalized := strings.ToLower(strings.TrimSpace(reasonPart))
	normalized = strings.TrimRight(normalized, ".")

	switch normalized {
	case "further verification needed", "further detail required":
		return true
	default:
		return false
	}
}

func isKnownDebugReasonTrailer(reasonPart string) bool {
	switch strings.ToLower(strings.TrimSpace(reasonPart)) {
	case "legacy step 3 fallback boolean result: true", "legacy step 3 fallback boolean result: false":
		return true
	default:
		return false
	}
}

func step3OutcomeReasonSuffix(automatedResult string) string {
	switch normalizeClassificationValue(automatedResult) {
	case stepAnswerAligned:
		return "This transaction is considered a low-risk transaction."
	case stepAnswerNextStep:
		return "This transaction is considered a high-risk transaction and should proceed to step 4 for further review."
	default:
		return ""
	}
}

func joinReasonParts(reason string, suffix string) string {
	trimmedReason := strings.TrimSpace(reason)
	trimmedSuffix := strings.TrimSpace(suffix)

	if trimmedReason == "" {
		return trimmedSuffix
	}
	if trimmedSuffix == "" {
		return trimmedReason
	}
	if strings.HasSuffix(trimmedReason, trimmedSuffix) {
		return trimmedReason
	}

	return strings.TrimRight(trimmedReason, ".") + ". " + trimmedSuffix
}

func inferStep3LowRisk(tenorDescription string) bool {
	normalizedTenor := strings.ToUpper(strings.TrimSpace(tenorDescription))
	if normalizedTenor == "N" {
		return true
	}

	return !strings.HasPrefix(normalizedTenor, "Y")
}

func isAlignedStepResult(result ports.StepResultDetails) bool {
	return strings.EqualFold(strings.TrimSpace(result.StepAlignment), stepAnswerAligned)
}

func isPositiveBooleanStepResult(result ports.StepResultDetails) bool {
	if result.BooleanResult != nil {
		return *result.BooleanResult
	}

	return isAlignedStepResult(result)
}

func toAutomatedStepAnswer(value bool) string {
	if value {
		return automatedStepAnswerYes
	}

	return automatedStepAnswerNo
}

func normalizeClassificationValue(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "aligned":
		return stepAnswerAligned
	case "not_aligned", "not-aligned":
		return stepAnswerNotAligned
	case "next_step", "next-step":
		return stepAnswerNextStep
	default:
		return ""
	}
}

func classificationToAPIValue(value string, fallback string) string {
	switch normalizeClassificationValue(value) {
	case stepAnswerAligned:
		return "aligned"
	case stepAnswerNotAligned:
		return "not_aligned"
	case stepAnswerNextStep:
		return "next_step"
	default:
		return fallback
	}
}

func automatedConfidencePointer(value int) *int {
	return &value
}
