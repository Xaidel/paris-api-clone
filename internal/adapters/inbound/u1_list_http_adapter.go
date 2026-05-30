package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpU1ListAdapter exposes U1 list CRUD over HTTP.
type HttpU1ListAdapter struct {
	createEntry inboundports.CreateU1ListPort
	getEntry    inboundports.GetU1ListPort
	listEntries inboundports.ListU1ListPort
	updateEntry inboundports.UpdateU1ListPort
	deleteEntry inboundports.DeleteU1ListPort
}

type u1ListRequest struct {
	Sector                string `json:"sector"`
	EligibleOperationType string `json:"eligible_operation_type"`
	ConditionGuidance     string `json:"condition_guidance"`
}

// NewHttpU1ListAdapter builds an HttpU1ListAdapter.
func NewHttpU1ListAdapter(
	createEntry inboundports.CreateU1ListPort,
	getEntry inboundports.GetU1ListPort,
	listEntries inboundports.ListU1ListPort,
	updateEntry inboundports.UpdateU1ListPort,
	deleteEntry inboundports.DeleteU1ListPort,
) *HttpU1ListAdapter {
	return &HttpU1ListAdapter{
		createEntry: createEntry,
		getEntry:    getEntry,
		listEntries: listEntries,
		updateEntry: updateEntry,
		deleteEntry: deleteEntry,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpU1ListAdapter) RegisterRoutes(r *gin.Engine) {
	u1List := r.Group("/api/v1/u1-list")
	u1List.POST("", a.CreateEntry)
	u1List.GET("", a.ListEntries)
	u1List.GET("/:id", a.GetEntry)
	u1List.PUT("/:id", a.UpdateEntry)
	u1List.DELETE("/:id", a.DeleteEntry)
}

// CreateEntry handles POST /api/v1/u1-list.
func (a *HttpU1ListAdapter) CreateEntry(c *gin.Context) {
	requestBody, err := decodeJSONBody[u1ListRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createEntry.Execute(c.Request.Context(), inboundports.CreateU1ListCommand{
		Sector:                requestBody.Sector,
		EligibleOperationType: requestBody.EligibleOperationType,
		ConditionGuidance:     requestBody.ConditionGuidance,
		ActorUserID:           c.GetHeader(actorUserIDHeader),
		ActorGroupID:          c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: result})
}

// GetEntry handles GET /api/v1/u1-list/:id.
func (a *HttpU1ListAdapter) GetEntry(c *gin.Context) {
	result, err := a.getEntry.Execute(c.Request.Context(), inboundports.GetU1ListQuery{
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

// ListEntries handles GET /api/v1/u1-list.
func (a *HttpU1ListAdapter) ListEntries(c *gin.Context) {
	result, err := a.listEntries.Execute(c.Request.Context(), inboundports.ListU1ListQuery{
		Sector:       c.Query("sector"),
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// UpdateEntry handles PUT /api/v1/u1-list/:id.
func (a *HttpU1ListAdapter) UpdateEntry(c *gin.Context) {
	requestBody, err := decodeJSONBody[u1ListRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.updateEntry.Execute(c.Request.Context(), inboundports.UpdateU1ListCommand{
		ID:                    c.Param("id"),
		Sector:                requestBody.Sector,
		EligibleOperationType: requestBody.EligibleOperationType,
		ConditionGuidance:     requestBody.ConditionGuidance,
		ActorUserID:           c.GetHeader(actorUserIDHeader),
		ActorGroupID:          c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteEntry handles DELETE /api/v1/u1-list/:id.
func (a *HttpU1ListAdapter) DeleteEntry(c *gin.Context) {
	_, err := a.deleteEntry.Execute(c.Request.Context(), inboundports.DeleteU1ListCommand{
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

func (a *HttpU1ListAdapter) handleError(c *gin.Context, err error) {
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
