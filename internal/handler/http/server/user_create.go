package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// Create handles POST /users
func (h *Handler) Create(c *gin.Context) {
	var in struct {
		Name     string `json:"name" binding:"required,min=1,max=255"`
		Email    string `json:"email" binding:"required,email,max=255"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &input.CreateUser{
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
	}

	output, err := h.createUser.Execute(c.Request.Context(), input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         output.User.ID.String(),
		"name":       output.User.Name,
		"email":      output.User.Email,
		"created_at": output.User.CreatedAt,
		"updated_at": output.User.UpdatedAt,
	})
}

// newUserCreateHandler creates a new CreateUser handler for Wire.
func newUserCreateHandler(createUser usecase.CreateUser) *Handler {
	return &Handler{createUser: createUser}
}
