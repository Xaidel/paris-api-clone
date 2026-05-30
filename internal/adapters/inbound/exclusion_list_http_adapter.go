package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpExclusionListAdapter exposes exclusion list CRUD over HTTP.
type HttpExclusionListAdapter struct {
	createEntry inboundports.CreateExclusionListPort
	getEntry    inboundports.GetExclusionListPort
	listEntries inboundports.ListExclusionListPort
	updateEntry inboundports.UpdateExclusionListPort
	deleteEntry inboundports.DeleteExclusionListPort
}

type exclusionListRequest struct {
	ActivityType string `json:"activity_type"`
}

// NewHttpExclusionListAdapter builds an HttpExclusionListAdapter.
func NewHttpExclusionListAdapter(
	createEntry inboundports.CreateExclusionListPort,
	getEntry inboundports.GetExclusionListPort,
	listEntries inboundports.ListExclusionListPort,
	updateEntry inboundports.UpdateExclusionListPort,
	deleteEntry inboundports.DeleteExclusionListPort,
) *HttpExclusionListAdapter {
	return &HttpExclusionListAdapter{
		createEntry: createEntry,
		getEntry:    getEntry,
		listEntries: listEntries,
		updateEntry: updateEntry,
		deleteEntry: deleteEntry,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpExclusionListAdapter) RegisterRoutes(r *gin.Engine) {
	exclusionList := r.Group("/api/v1/u2-exclusion-list")
	exclusionList.POST("", a.CreateEntry)
	exclusionList.GET("", a.ListEntries)
	exclusionList.GET("/:id", a.GetEntry)
	exclusionList.PUT("/:id", a.UpdateEntry)
	exclusionList.DELETE("/:id", a.DeleteEntry)
}

// CreateEntry handles POST /api/v1/u2-exclusion-list.
func (a *HttpExclusionListAdapter) CreateEntry(c *gin.Context) {
	requestBody, err := decodeJSONBody[exclusionListRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createEntry.Execute(c.Request.Context(), inboundports.CreateExclusionListCommand{
		ActivityType: requestBody.ActivityType,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: result})
}

// GetEntry handles GET /api/v1/u2-exclusion-list/:id.
func (a *HttpExclusionListAdapter) GetEntry(c *gin.Context) {
	result, err := a.getEntry.Execute(c.Request.Context(), inboundports.GetExclusionListQuery{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// ListEntries handles GET /api/v1/u2-exclusion-list.
func (a *HttpExclusionListAdapter) ListEntries(c *gin.Context) {
	result, err := a.listEntries.Execute(c.Request.Context(), inboundports.ListExclusionListQuery{
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// UpdateEntry handles PUT /api/v1/u2-exclusion-list/:id.
func (a *HttpExclusionListAdapter) UpdateEntry(c *gin.Context) {
	requestBody, err := decodeJSONBody[exclusionListRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.updateEntry.Execute(c.Request.Context(), inboundports.UpdateExclusionListCommand{
		ID:           c.Param("id"),
		ActivityType: requestBody.ActivityType,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteEntry handles DELETE /api/v1/u2-exclusion-list/:id.
func (a *HttpExclusionListAdapter) DeleteEntry(c *gin.Context) {
	_, err := a.deleteEntry.Execute(c.Request.Context(), inboundports.DeleteExclusionListCommand{
		ID:           c.Param("id"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *HttpExclusionListAdapter) handleError(c *gin.Context, err error) {
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
