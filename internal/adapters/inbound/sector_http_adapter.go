package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpSectorAdapter exposes sector CRUD over HTTP.
type HttpSectorAdapter struct {
	createSector inboundports.CreateSectorPort
	getSector    inboundports.GetSectorPort
	listSectors  inboundports.ListSectorsPort
	updateSector inboundports.UpdateSectorPort
	deleteSector inboundports.DeleteSectorPort
}

type sectorRequest struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// NewHttpSectorAdapter builds an HttpSectorAdapter.
func NewHttpSectorAdapter(
	createSector inboundports.CreateSectorPort,
	getSector inboundports.GetSectorPort,
	listSectors inboundports.ListSectorsPort,
	updateSector inboundports.UpdateSectorPort,
	deleteSector inboundports.DeleteSectorPort,
) *HttpSectorAdapter {
	return &HttpSectorAdapter{
		createSector: createSector,
		getSector:    getSector,
		listSectors:  listSectors,
		updateSector: updateSector,
		deleteSector: deleteSector,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpSectorAdapter) RegisterRoutes(r *gin.Engine) {
	sectors := r.Group("/api/v1/sectors")
	sectors.POST("", a.CreateSector)
	sectors.GET("", a.ListSectors)
	sectors.GET("/:id", a.GetSector)
	sectors.PUT("/:id", a.UpdateSector)
	sectors.DELETE("/:id", a.DeleteSector)
}

// CreateSector handles POST /api/v1/sectors.
func (a *HttpSectorAdapter) CreateSector(c *gin.Context) {
	requestBody, err := decodeJSONBody[sectorRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createSector.Execute(c.Request.Context(), inboundports.CreateSectorCommand{
		Type:         requestBody.Type,
		Name:         requestBody.Name,
		Description:  requestBody.Description,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: result})
}

// GetSector handles GET /api/v1/sectors/:id.
func (a *HttpSectorAdapter) GetSector(c *gin.Context) {
	result, err := a.getSector.Execute(c.Request.Context(), inboundports.GetSectorQuery{
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

// ListSectors handles GET /api/v1/sectors.
func (a *HttpSectorAdapter) ListSectors(c *gin.Context) {
	result, err := a.listSectors.Execute(c.Request.Context(), inboundports.ListSectorsQuery{
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// UpdateSector handles PUT /api/v1/sectors/:id.
func (a *HttpSectorAdapter) UpdateSector(c *gin.Context) {
	requestBody, err := decodeJSONBody[sectorRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.updateSector.Execute(c.Request.Context(), inboundports.UpdateSectorCommand{
		ID:           c.Param("id"),
		Type:         requestBody.Type,
		Name:         requestBody.Name,
		Description:  requestBody.Description,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteSector handles DELETE /api/v1/sectors/:id.
func (a *HttpSectorAdapter) DeleteSector(c *gin.Context) {
	_, err := a.deleteSector.Execute(c.Request.Context(), inboundports.DeleteSectorCommand{
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

func (a *HttpSectorAdapter) handleError(c *gin.Context, err error) {
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
