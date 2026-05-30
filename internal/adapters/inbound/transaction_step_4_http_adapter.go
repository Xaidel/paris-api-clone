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

// HttpTransactionStep4Adapter exposes aggregated transaction step 4 review over HTTP.
type HttpTransactionStep4Adapter struct {
	createTransactionStep4 inboundports.CreateTransactionStep4Port
}

type createTransactionStep4Request struct {
	SectorID          string `json:"sector_id"`
	AdditionalContext string `json:"additional_context"`
	IsHighEmitting    *bool  `json:"is_high_emitting"`
}

type transactionStep4Response struct {
	Question          string    `json:"question"`
	Answer            string    `json:"answer"`
	TransactionID     string    `json:"transaction_id"`
	SectorID          string    `json:"sector_id"`
	AdditionalContext string    `json:"additional_context"`
	IsHighEmitting    bool      `json:"is_high_emitting"`
	Classification    string    `json:"classification"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// NewHttpTransactionStep4Adapter builds an HttpTransactionStep4Adapter.
func NewHttpTransactionStep4Adapter(createTransactionStep4 inboundports.CreateTransactionStep4Port) *HttpTransactionStep4Adapter {
	return &HttpTransactionStep4Adapter{createTransactionStep4: createTransactionStep4}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpTransactionStep4Adapter) RegisterRoutes(r *gin.Engine) {
	step4 := r.Group("/api/v1/transactions/:id/step-4")
	step4.POST("", a.CreateTransactionStep4)
}

// CreateTransactionStep4 handles POST /api/v1/transactions/:id/step-4.
func (a *HttpTransactionStep4Adapter) CreateTransactionStep4(c *gin.Context) {
	requestBody, err := decodeJSONBody[createTransactionStep4Request](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	command, err := newCreateTransactionStep4Command(c.Param("id"), requestBody, c.GetHeader(actorUserIDHeader), c.GetHeader(actorGroupIDHeader))
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.createTransactionStep4.Execute(c.Request.Context(), command)
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: toTransactionStep4Response(result)})
}

func newCreateTransactionStep4Command(transactionIDValue string, requestBody createTransactionStep4Request, actorUserID, actorGroupID string) (inboundports.CreateTransactionStep4Command, error) {
	transactionID, err := valueobjects.TransactionIDFromString(transactionIDValue)
	if err != nil {
		return inboundports.CreateTransactionStep4Command{}, err
	}

	var validationErrors []domain.FieldValidationError

	if strings.TrimSpace(requestBody.SectorID) == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("sector_id", "required", "sector_id is required"))
	}

	if strings.TrimSpace(requestBody.AdditionalContext) == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("additional_context", "required", "additional_context is required"))
	}

	if requestBody.IsHighEmitting == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("is_high_emitting", "required", "is_high_emitting is required"))
	}

	if validationErr := domain.NewValidationError(validationErrors); validationErr != nil {
		return inboundports.CreateTransactionStep4Command{}, validationErr
	}

	sectorID, err := valueobjects.SectorIDFromString(requestBody.SectorID)
	if err != nil {
		return inboundports.CreateTransactionStep4Command{}, err
	}

	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext(requestBody.AdditionalContext)
	if err != nil {
		return inboundports.CreateTransactionStep4Command{}, err
	}

	return inboundports.CreateTransactionStep4Command{
		TransactionID:     transactionID,
		SectorID:          sectorID,
		AdditionalContext: additionalContext,
		IsHighEmitting:    requestBody.IsHighEmitting,
		ActorUserID:       actorUserID,
		ActorGroupID:      actorGroupID,
	}, nil
}

func toTransactionStep4Response(result ports.TransactionStep4Result) transactionStep4Response {
	return transactionStep4Response{
		Question:          step4Question,
		Answer:            toStep4Answer(result.IsHighEmitting),
		TransactionID:     result.TransactionID.String(),
		SectorID:          result.SectorID.String(),
		AdditionalContext: result.AdditionalContext.String(),
		IsHighEmitting:    result.IsHighEmitting,
		Classification:    result.Classification.String(),
		Status:            result.Status.String(),
		CreatedAt:         result.CreatedAt,
		UpdatedAt:         result.UpdatedAt,
	}
}

func toStep4Answer(isHighEmitting bool) string {
	if isHighEmitting {
		return stepAnswerNotAligned
	}

	return stepAnswerAligned
}

func (a *HttpTransactionStep4Adapter) handleError(c *gin.Context, err error) {
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
