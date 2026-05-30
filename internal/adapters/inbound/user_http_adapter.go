package adapters

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// HttpUserAdapter exposes user CRUD over HTTP.
type HttpUserAdapter struct {
	createUser inboundports.CreateUserPort
	getUser    inboundports.GetUserPort
	listUsers  inboundports.ListUsersPort
	updateUser inboundports.UpdateUserPort
	deleteUser inboundports.DeleteUserPort
}

// userResponse is the stable HTTP DTO returned by the user endpoints.
type userResponse struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"firstname"`
	MiddleName *string   `json:"middlename"`
	LastName   string    `json:"lastname"`
	GroupID    string    `json:"group_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// createUserRequest is the JSON payload accepted by POST /users.
type createUserRequest struct {
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	FirstName  string  `json:"firstname"`
	MiddleName *string `json:"middlename"`
	LastName   string  `json:"lastname"`
	GroupID    string  `json:"group_id"`
}

// updateUserRequest is the JSON payload accepted by PUT /users/:id.
type updateUserRequest struct {
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	FirstName  string  `json:"firstname"`
	MiddleName *string `json:"middlename"`
	LastName   string  `json:"lastname"`
	GroupID    string  `json:"group_id"`
}

// listUsersResponse wraps the collection payload returned by GET /users.
type listUsersResponse struct {
	Users []userResponse `json:"users"`
}

// NewHttpUserAdapter builds an HttpUserAdapter.
func NewHttpUserAdapter(
	createUser inboundports.CreateUserPort,
	getUser inboundports.GetUserPort,
	listUsers inboundports.ListUsersPort,
	updateUser inboundports.UpdateUserPort,
	deleteUser inboundports.DeleteUserPort,
) *HttpUserAdapter {
	return &HttpUserAdapter{
		createUser: createUser,
		getUser:    getUser,
		listUsers:  listUsers,
		updateUser: updateUser,
		deleteUser: deleteUser,
	}
}

// RegisterRoutes attaches handlers to the Gin engine.
func (a *HttpUserAdapter) RegisterRoutes(r *gin.Engine) {
	users := r.Group("/api/v1/users")
	users.POST("", a.CreateUser)
	users.GET("", a.ListUsers)
	users.GET("/:id", a.GetUser)
	users.PUT("/:id", a.UpdateUser)
	users.DELETE("/:id", a.DeleteUser)
}

// CreateUser handles POST /api/v1/users.
func (a *HttpUserAdapter) CreateUser(c *gin.Context) {
	requestBody, err := decodeJSONBody[createUserRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.createUser.Execute(c.Request.Context(), inboundports.CreateUserCommand{
		Username:     requestBody.Username,
		Password:     requestBody.Password,
		FirstName:    requestBody.FirstName,
		MiddleName:   requestBody.MiddleName,
		LastName:     requestBody.LastName,
		GroupID:      requestBody.GroupID,
		ActorUserID:  c.GetHeader(actorUserIDHeader),
		ActorGroupID: c.GetHeader(actorGroupIDHeader),
	})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, apiResponse{Data: toUserResponse(result)})
}

// GetUser handles GET /api/v1/users/:id.
func (a *HttpUserAdapter) GetUser(c *gin.Context) {
	result, err := a.getUser.Execute(c.Request.Context(), inboundports.GetUserQuery{ID: c.Param("id"), ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toUserResponse(result)})
}

// ListUsers handles GET /api/v1/users.
func (a *HttpUserAdapter) ListUsers(c *gin.Context) {
	result, err := a.listUsers.Execute(c.Request.Context(), inboundports.ListUsersQuery{ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	users := make([]userResponse, 0, len(result.Users))
	for _, user := range result.Users {
		users = append(users, toUserResponse(user))
	}

	c.JSON(http.StatusOK, apiResponse{Data: listUsersResponse{Users: users}})
}

// UpdateUser handles PUT /api/v1/users/:id.
func (a *HttpUserAdapter) UpdateUser(c *gin.Context) {
	requestBody, err := decodeJSONBody[updateUserRequest](c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, apiResponse{Error: &errorResponse{Code: "bad_request", Message: err.Error()}})
		return
	}

	result, err := a.updateUser.Execute(c.Request.Context(), inboundports.UpdateUserCommand{ID: c.Param("id"), Username: requestBody.Username, Password: requestBody.Password, FirstName: requestBody.FirstName, MiddleName: requestBody.MiddleName, LastName: requestBody.LastName, GroupID: requestBody.GroupID, ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, apiResponse{Data: toUserResponse(result)})
}

// DeleteUser handles DELETE /api/v1/users/:id.
func (a *HttpUserAdapter) DeleteUser(c *gin.Context) {
	_, err := a.deleteUser.Execute(c.Request.Context(), inboundports.DeleteUserCommand{ID: c.Param("id"), ActorUserID: c.GetHeader(actorUserIDHeader), ActorGroupID: c.GetHeader(actorGroupIDHeader)})
	if err != nil {
		a.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (a *HttpUserAdapter) handleError(c *gin.Context, err error) {
	var notFoundErr *usecases.NotFoundError
	var domainErr *domain.DomainError

	// Map application and domain errors to transport-specific status codes here so
	// the use case layer stays HTTP-agnostic.
	switch {
	case errors.As(err, &notFoundErr):
		c.JSON(http.StatusNotFound, apiResponse{Error: &errorResponse{Code: "not_found", Message: err.Error()}})
	case errors.As(err, &domainErr):
		c.JSON(http.StatusUnprocessableEntity, apiResponse{Error: &errorResponse{Code: domainErr.Code, Message: domainErr.Message}})
	default:
		c.JSON(http.StatusInternalServerError, apiResponse{Error: &errorResponse{Code: "internal_error", Message: "internal server error"}})
	}
}

func toUserResponse(result ports.UserResult) userResponse {
	// Keep the HTTP DTO mapping in one place so field naming stays consistent
	// across create, get, list, and update responses.
	return userResponse{ID: result.ID, Username: result.Username, FirstName: result.FirstName, MiddleName: result.MiddleName, LastName: result.LastName, GroupID: result.GroupID, CreatedAt: result.CreatedAt, UpdatedAt: result.UpdatedAt}
}
