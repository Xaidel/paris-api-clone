package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpGroupAdapter exposes group CRUD over HTTP.
type HttpGroupAdapter struct {
	createGroup inboundports.CreateGroupPort
	getGroup    inboundports.GetGroupPort
	listGroups  inboundports.ListGroupsPort
	updateGroup inboundports.UpdateGroupPort
	deleteGroup inboundports.DeleteGroupPort
}

type groupRequest struct {
	Name string `json:"name"`
}

// NewHttpGroupAdapter builds an HttpGroupAdapter.
func NewHttpGroupAdapter(
	createGroup inboundports.CreateGroupPort,
	getGroup inboundports.GetGroupPort,
	listGroups inboundports.ListGroupsPort,
	updateGroup inboundports.UpdateGroupPort,
	deleteGroup inboundports.DeleteGroupPort,
) *HttpGroupAdapter {
	return &HttpGroupAdapter{
		createGroup: createGroup,
		getGroup:    getGroup,
		listGroups:  listGroups,
		updateGroup: updateGroup,
		deleteGroup: deleteGroup,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpGroupAdapter) RegisterRoutes(r *gin.Engine) {
	groups := r.Group("/api/v1/groups")
	groups.POST("", a.CreateGroup)
	groups.GET("", a.ListGroups)
	groups.GET("/:id", a.GetGroup)
	groups.PUT("/:id", a.UpdateGroup)
	groups.DELETE("/:id", a.DeleteGroup)
}

// CreateGroup handles POST /api/v1/groups.
func (a *HttpGroupAdapter) CreateGroup(c *gin.Context) {
	requestBody, err := decodeJSONBody[groupRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createGroup.Execute(c.Request.Context(), inboundports.CreateGroupCommand{
		Name:         requestBody.Name,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: result})
}

// GetGroup handles GET /api/v1/groups/:id.
func (a *HttpGroupAdapter) GetGroup(c *gin.Context) {
	result, err := a.getGroup.Execute(c.Request.Context(), inboundports.GetGroupQuery{
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

// ListGroups handles GET /api/v1/groups.
func (a *HttpGroupAdapter) ListGroups(c *gin.Context) {
	result, err := a.listGroups.Execute(c.Request.Context(), inboundports.ListGroupsQuery{
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// UpdateGroup handles PUT /api/v1/groups/:id.
func (a *HttpGroupAdapter) UpdateGroup(c *gin.Context) {
	requestBody, err := decodeJSONBody[groupRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.updateGroup.Execute(c.Request.Context(), inboundports.UpdateGroupCommand{
		ID:           c.Param("id"),
		Name:         requestBody.Name,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteGroup handles DELETE /api/v1/groups/:id.
func (a *HttpGroupAdapter) DeleteGroup(c *gin.Context) {
	_, err := a.deleteGroup.Execute(c.Request.Context(), inboundports.DeleteGroupCommand{
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

func (a *HttpGroupAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var domainErr *domain.DomainError

	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: "internal server error"}})
	}
}
