package adapters

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpAuditEventAdapter exposes audit event reads over HTTP.
type HttpAuditEventAdapter struct {
	listAuditEvents inboundports.ListAuditEventsPort
	getAuditEvent   inboundports.GetAuditEventPort
}

// NewHttpAuditEventAdapter builds an HttpAuditEventAdapter.
func NewHttpAuditEventAdapter(listAuditEvents inboundports.ListAuditEventsPort, getAuditEvent inboundports.GetAuditEventPort) *HttpAuditEventAdapter {
	return &HttpAuditEventAdapter{listAuditEvents: listAuditEvents, getAuditEvent: getAuditEvent}
}

// RegisterRoutes attaches audit handlers to the Gin engine.
func (a *HttpAuditEventAdapter) RegisterRoutes(r *gin.Engine) {
	audit := r.Group("/api/v1/audit/events")
	audit.GET("", a.ListAuditEvents)
	audit.GET("/:id", a.GetAuditEvent)
}

// ListAuditEvents handles GET /api/v1/audit/events.
func (a *HttpAuditEventAdapter) ListAuditEvents(c *gin.Context) {
	query, err := buildListAuditEventsQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.listAuditEvents.Execute(c.Request.Context(), query)
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// GetAuditEvent handles GET /api/v1/audit/events/:id.
func (a *HttpAuditEventAdapter) GetAuditEvent(c *gin.Context) {
	result, err := a.getAuditEvent.Execute(c.Request.Context(), inboundports.GetAuditEventQuery{ID: c.Param("id")})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

func (a *HttpAuditEventAdapter) handleError(c *gin.Context, err error) {
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

func buildListAuditEventsQuery(c *gin.Context) (inboundports.ListAuditEventsQuery, error) {
	startedAt, err := parseOptionalTime(c.Query("start_date"))
	if err != nil {
		return inboundports.ListAuditEventsQuery{}, err
	}

	endedAt, err := parseOptionalTime(c.Query("end_date"))
	if err != nil {
		return inboundports.ListAuditEventsQuery{}, err
	}

	return inboundports.ListAuditEventsQuery{
		EventOwner: c.Query("event_owner"),
		EventType:  c.Query("event_type"),
		SessionID:  c.Query("session_id"),
		UserID:     c.Query("user_id"),
		StartedAt:  startedAt,
		EndedAt:    endedAt,
	}, nil
}
