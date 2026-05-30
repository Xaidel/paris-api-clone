package adapters

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

const (
	step5ScreeningQuestion1Text = "Screening Question 1"
	step5ScreeningQuestion2Text = "Screening Question 2"
	step5AnswerYes              = "yes"
	step5AnswerNo               = "no"
)

// HttpTransactionStep5Adapter exposes transaction step 5 screening over HTTP.
type HttpTransactionStep5Adapter struct {
	createTransactionStep5 inboundports.CreateTransactionStep5Port
}

type createTransactionStep5Request struct {
	ScreeningQuestion1 createTransactionStep5ScreeningQuestionRequest `json:"screening_question_1"`
	ScreeningQuestion2 createTransactionStep5ScreeningQuestionRequest `json:"screening_question_2"`
	ReviewerNotes      *string                                        `json:"reviewer_notes"`
	IsFinal            *bool                                          `json:"is_final"`
}

type createTransactionStep5ScreeningQuestionRequest struct {
	Answer        *bool  `json:"answer"`
	Justification string `json:"justification"`
}

type transactionStep5Response struct {
	TransactionID      string                                    `json:"transaction_id"`
	ScreeningQuestion1 transactionStep5ScreeningQuestionResponse `json:"screening_question_1"`
	ScreeningQuestion2 transactionStep5ScreeningQuestionResponse `json:"screening_question_2"`
	ReviewerNotes      *string                                   `json:"reviewer_notes,omitempty"`
	IsFinal            bool                                      `json:"is_final"`
	Classification     string                                    `json:"classification"`
	Detail             string                                    `json:"detail"`
	CreatedAt          *time.Time                                `json:"created_at,omitempty"`
	UpdatedAt          *time.Time                                `json:"updated_at,omitempty"`
}

type transactionStep5PreviewResponse struct {
	Classification string `json:"classification"`
	Detail         string `json:"detail"`
}

type transactionStep5ScreeningQuestionResponse struct {
	Question      string `json:"question"`
	Answer        string `json:"answer"`
	Justification string `json:"justification"`
}

// NewHttpTransactionStep5Adapter builds an HttpTransactionStep5Adapter.
func NewHttpTransactionStep5Adapter(createTransactionStep5 inboundports.CreateTransactionStep5Port) *HttpTransactionStep5Adapter {
	return &HttpTransactionStep5Adapter{createTransactionStep5: createTransactionStep5}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpTransactionStep5Adapter) RegisterRoutes(r *gin.Engine) {
	step5 := r.Group("/api/v1/transactions/:id/step-5")
	step5.POST("", a.CreateTransactionStep5)
}

// CreateTransactionStep5 handles POST /api/v1/transactions/:id/step-5.
func (a *HttpTransactionStep5Adapter) CreateTransactionStep5(c *gin.Context) {
	requestBody, err := decodeJSONBody[createTransactionStep5Request](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	command, err := newCreateTransactionStep5Command(c.Param("id"), requestBody, c.GetHeader(actorUserIDHeader), c.GetHeader(actorGroupIDHeader))
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.createTransactionStep5.Execute(c.Request.Context(), command)
	if err != nil {
		a.handleError(c, err)
		return
	}

	statusCode := http.StatusOK
	if result.IsFinal {
		statusCode = http.StatusCreated
		c.JSON(statusCode, apiResponse{Data: toTransactionStep5Response(result)})
		return
	}

	c.JSON(statusCode, apiResponse{Data: toTransactionStep5PreviewResponse(result)})
}

func newCreateTransactionStep5Command(transactionIDValue string, requestBody createTransactionStep5Request, actorUserID, actorGroupID string) (inboundports.CreateTransactionStep5Command, error) {
	transactionID, err := valueobjects.TransactionIDFromString(transactionIDValue)
	if err != nil {
		return inboundports.CreateTransactionStep5Command{}, err
	}

	var validationErrors []domain.FieldValidationError

	if requestBody.ScreeningQuestion1.Answer == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_1.answer", "required", "screening_question_1.answer is required"))
	}

	if strings.TrimSpace(requestBody.ScreeningQuestion1.Justification) == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_1.justification", "required", "screening_question_1.justification is required"))
	}

	if requestBody.ScreeningQuestion2.Answer == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_2.answer", "required", "screening_question_2.answer is required"))
	}

	if strings.TrimSpace(requestBody.ScreeningQuestion2.Justification) == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_2.justification", "required", "screening_question_2.justification is required"))
	}

	if requestBody.IsFinal == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("is_final", "required", "is_final is required"))
	}

	if validationErr := domain.NewValidationError(validationErrors); validationErr != nil {
		return inboundports.CreateTransactionStep5Command{}, validationErr
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification(requestBody.ScreeningQuestion1.Justification)
	if err != nil {
		return inboundports.CreateTransactionStep5Command{}, err
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification(requestBody.ScreeningQuestion2.Justification)
	if err != nil {
		return inboundports.CreateTransactionStep5Command{}, err
	}

	return inboundports.CreateTransactionStep5Command{
		TransactionID:                   transactionID,
		ScreeningQuestion1Answer:        requestBody.ScreeningQuestion1.Answer,
		ScreeningQuestion1Justification: question1Justification,
		ScreeningQuestion2Answer:        requestBody.ScreeningQuestion2.Answer,
		ScreeningQuestion2Justification: question2Justification,
		ReviewerNotes:                   valueobjects.NewTransactionStep5ReviewerNotes(requestBody.ReviewerNotes),
		IsFinal:                         requestBody.IsFinal,
		ActorUserID:                     actorUserID,
		ActorGroupID:                    actorGroupID,
	}, nil
}

func toTransactionStep5Response(result ports.TransactionStep5Result) transactionStep5Response {
	return transactionStep5Response{
		TransactionID: result.TransactionID.String(),
		ScreeningQuestion1: transactionStep5ScreeningQuestionResponse{
			Question:      step5ScreeningQuestion1Text,
			Answer:        toStep5Answer(result.ScreeningQuestion1Answer),
			Justification: result.ScreeningQuestion1Justification.String(),
		},
		ScreeningQuestion2: transactionStep5ScreeningQuestionResponse{
			Question:      step5ScreeningQuestion2Text,
			Answer:        toStep5Answer(result.ScreeningQuestion2Answer),
			Justification: result.ScreeningQuestion2Justification.String(),
		},
		ReviewerNotes:  result.ReviewerNotes.String(),
		IsFinal:        result.IsFinal,
		Classification: result.Classification.String(),
		Detail:         result.Detail,
		CreatedAt:      result.CreatedAt,
		UpdatedAt:      result.UpdatedAt,
	}
}

func toTransactionStep5PreviewResponse(result ports.TransactionStep5Result) transactionStep5PreviewResponse {
	return transactionStep5PreviewResponse{
		Classification: result.Classification.String(),
		Detail:         result.Detail,
	}
}

func toStep5Answer(answer bool) string {
	if answer {
		return step5AnswerYes
	}

	return step5AnswerNo
}

func (a *HttpTransactionStep5Adapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var conflictErr *usecases.ConflictError
	var validationErr *domain.ValidationError
	var domainErr *domain.DomainError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &conflictErr):
		c.JSON(http.StatusConflict, apiResponse{Error: &errorResponse{Code: "conflict", Message: conflictErr.Error()}})
	case errors.As(err, &validationErr):
		fieldErrors := make([]fieldErrorResponse, 0, len(validationErr.Fields()))
		for _, fieldErr := range validationErr.Fields() {
			fieldErrors = append(fieldErrors, fieldErrorResponse{Field: fieldErr.Field(), Code: fieldErr.Code(), Message: fieldErr.Message()})
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "validation_error", "message": validationErr.Error(), "fields": fieldErrors}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		logInternalRequestError(c, err)
		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: "internal server error"}})
	}
}
