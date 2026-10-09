package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase"
)

// Handler handles user HTTP requests.
type Handler struct {
	createUser usecase.CreateUser
	getUser    usecase.GetUser
	listUsers  usecase.ListUsers
	updateUser usecase.UpdateUser
	deleteUser usecase.DeleteUser
}

// WireSet holds the Wire providers for server handlers.
var WireSet = wire.NewSet(
	NewHandler,
)

// NewHandler creates a new Handler with all dependencies.
func NewHandler(
	createUser usecase.CreateUser,
	getUser usecase.GetUser,
	listUsers usecase.ListUsers,
	updateUser usecase.UpdateUser,
	deleteUser usecase.DeleteUser,
) *Handler {
	return &Handler{
		createUser: createUser,
		getUser:    getUser,
		listUsers:  listUsers,
		updateUser: updateUser,
		deleteUser: deleteUser,
	}
}

func handleError(c *gin.Context, err error) {
	var domainErr *model.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case "USER_NOT_FOUND":
			c.JSON(http.StatusNotFound, gin.H{"error": domainErr.Error()})
		case "USER_EXISTS":
			c.JSON(http.StatusConflict, gin.H{"error": domainErr.Error()})
		case "INVALID_INPUT":
			c.JSON(http.StatusBadRequest, gin.H{"error": domainErr.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
