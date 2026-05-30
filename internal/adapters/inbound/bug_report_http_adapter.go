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

// HttpBugReportAdapter exposes bug report CRUD over HTTP.
type HttpBugReportAdapter struct {
	createBugReport inboundports.CreateBugReportPort
	getBugReport    inboundports.GetBugReportPort
	listBugReports  inboundports.ListBugReportsPort
	updateBugReport inboundports.UpdateBugReportPort
	deleteBugReport inboundports.DeleteBugReportPort
}

type createReportRequest struct {
	TransactionID string `json:"transaction_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
}

type updateReportRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// NewHttpBugReportAdapter builds an HttpReportAdapter.
func NewHttpBugReportAdapter(
	createBugReport inboundports.CreateBugReportPort,
	getBugReport inboundports.GetBugReportPort,
	listBugReports inboundports.ListBugReportsPort,
	updateBugReport inboundports.UpdateBugReportPort,
	deleteBugReport inboundports.DeleteBugReportPort,
) *HttpBugReportAdapter {
	return &HttpBugReportAdapter{
		createBugReport: createBugReport,
		getBugReport:    getBugReport,
		listBugReports:  listBugReports,
		updateBugReport: updateBugReport,
		deleteBugReport: deleteBugReport,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpBugReportAdapter) RegisterRoutes(r *gin.Engine) {
	bugReports := r.Group("/api/v1/bug-reports")
	bugReports.POST("", a.CreateBugReport)
	bugReports.GET("", a.ListBugReports)
	bugReports.GET("/:id", a.GetBugReport)
	bugReports.PUT("/:id", a.UpdateBugReport)
	bugReports.DELETE("/:id", a.DeleteBugReport)
}

// CreateBugReport handles POST /api/v1/bug-reports.
func (a *HttpBugReportAdapter) CreateBugReport(c *gin.Context) {
	requestBody, err := decodeJSONBody[createReportRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	transactionID, err := valueobjects.TransactionIDFromString(requestBody.TransactionID)
	if err != nil {
		a.handleError(c, err)
		return
	}

	title, err := valueobjects.NewBugReportTitle(requestBody.Title)
	if err != nil {
		a.handleError(c, err)
		return
	}

	description, err := valueobjects.NewBugReportDescription(requestBody.Description)
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.createBugReport.Execute(c.Request.Context(), inboundports.CreateBugReportCommand{
		TransactionID: transactionID,
		Title:         title,
		Description:   description,
		ActorUserID:   c.GetHeader(actorUserIDHeader),
		ActorGroupID:  c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: result})
}

// GetBugReport handles GET /api/v1/bug-reports/:id.
func (a *HttpBugReportAdapter) GetBugReport(c *gin.Context) {
	id, err := valueobjects.BugReportIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.getBugReport.Execute(c.Request.Context(), inboundports.GetBugReportQuery{
		ID:           id,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// ListBugReports handles GET /api/v1/bug-reports.
func (a *HttpBugReportAdapter) ListBugReports(c *gin.Context) {
	result, err := a.listBugReports.Execute(c.Request.Context(), inboundports.ListBugReportsQuery{
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// UpdateBugReport handles PUT /api/v1/bug-reports/:id.
func (a *HttpBugReportAdapter) UpdateBugReport(c *gin.Context) {
	id, err := valueobjects.BugReportIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	requestBody, err := decodeJSONBody[updateReportRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	title, err := valueobjects.NewBugReportTitle(requestBody.Title)
	if err != nil {
		a.handleError(c, err)
		return
	}

	description, err := valueobjects.NewBugReportDescription(requestBody.Description)
	if err != nil {
		a.handleError(c, err)
		return
	}

	status, err := valueobjects.BugReportStatusFromString(requestBody.Status)
	if err != nil {
		a.handleError(c, err)
		return
	}

	result, err := a.updateBugReport.Execute(c.Request.Context(), inboundports.UpdateBugReportCommand{
		ID:           id,
		Title:        title,
		Description:  description,
		Status:       status,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: result})
}

// DeleteBugReport handles DELETE /api/v1/bug-reports/:id.
func (a *HttpBugReportAdapter) DeleteBugReport(c *gin.Context) {
	id, err := valueobjects.BugReportIDFromString(c.Param("id"))
	if err != nil {
		a.handleError(c, err)
		return
	}

	_, err = a.deleteBugReport.Execute(c.Request.Context(), inboundports.DeleteBugReportCommand{
		ID:           id,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *HttpBugReportAdapter) handleError(c *gin.Context, err error) {
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
