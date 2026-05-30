package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpTransactionFeedbackAdapter exposes transaction feedback over HTTP.
type HttpTransactionFeedbackAdapter struct {
	upsertFeedback inboundports.UpsertTransactionFeedbackPort
	deleteFeedback inboundports.DeleteTransactionFeedbackPort
	getFeedback    inboundports.GetTransactionFeedbackPort
}

// NewHttpTransactionFeedbackAdapter builds an HttpTransactionFeedbackAdapter.
func NewHttpTransactionFeedbackAdapter(
	upsertFeedback inboundports.UpsertTransactionFeedbackPort,
	deleteFeedback inboundports.DeleteTransactionFeedbackPort,
	getFeedback inboundports.GetTransactionFeedbackPort,
) *HttpTransactionFeedbackAdapter {
	return &HttpTransactionFeedbackAdapter{
		upsertFeedback: upsertFeedback,
		deleteFeedback: deleteFeedback,
		getFeedback:    getFeedback,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpTransactionFeedbackAdapter) RegisterRoutes(r *gin.Engine) {
	feedback := r.Group("/api/v1/transactions/:id/feedback")
	feedback.PUT("", a.UpsertFeedback)
	feedback.DELETE("", a.DeleteFeedback)
	feedback.GET("", a.GetFeedback)
}

// UpsertFeedback handles PUT /api/v1/transactions/:id/feedback?kind=thumbs_up.
func (a *HttpTransactionFeedbackAdapter) UpsertFeedback(c *gin.Context) {
	transactionID, err := valueobjects.TransactionIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	kind, err := valueobjects.FeedbackKindFromString(c.Query("kind"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.upsertFeedback.Execute(c.Request.Context(), inboundports.UpsertTransactionFeedbackCommand{
		TransactionID: transactionID,
		Kind:          kind,
		ActorUserID:   c.GetHeader(actorUserIDHeader),
		ActorGroupID:  c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteFeedback handles DELETE /api/v1/transactions/:id/feedback.
func (a *HttpTransactionFeedbackAdapter) DeleteFeedback(c *gin.Context) {
	transactionID, err := valueobjects.TransactionIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	err = a.deleteFeedback.Execute(c.Request.Context(), inboundports.DeleteTransactionFeedbackCommand{
		TransactionID: transactionID,
		ActorUserID:   c.GetHeader(actorUserIDHeader),
		ActorGroupID:  c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetFeedback handles GET /api/v1/transactions/:id/feedback.
func (a *HttpTransactionFeedbackAdapter) GetFeedback(c *gin.Context) {
	transactionID, err := valueobjects.TransactionIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.getFeedback.Execute(c.Request.Context(), inboundports.GetTransactionFeedbackQuery{
		TransactionID: transactionID,
		ActorUserID:   c.GetHeader(actorUserIDHeader),
		ActorGroupID:  c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

func (a *HttpTransactionFeedbackAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var domainErr *domain.DomainError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		logInternalRequestError(c, err)
		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: "internal server error"}})
	}
}
