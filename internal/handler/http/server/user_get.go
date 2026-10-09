package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// Get handles GET /users/:id
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	input := &input.GetUser{
		ID: entity.UserID(id),
	}

	output, err := h.getUser.Execute(c.Request.Context(), input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         output.User.ID.String(),
		"name":       output.User.Name,
		"email":      output.User.Email,
		"created_at": output.User.CreatedAt,
		"updated_at": output.User.UpdatedAt,
	})
}

// newUserGetHandler creates a new GetUser handler for Wire.
func newUserGetHandler(getUser usecase.GetUser) *Handler {
	return &Handler{getUser: getUser}
}
