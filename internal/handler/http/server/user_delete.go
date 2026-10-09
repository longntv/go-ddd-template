package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// Delete handles DELETE /users/:id
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	input := &input.DeleteUser{
		ID: entity.UserID(id),
	}

	if err := h.deleteUser.Execute(c.Request.Context(), input); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// newUserDeleteHandler creates a new DeleteUser handler for Wire.
func newUserDeleteHandler(deleteUser usecase.DeleteUser) *Handler {
	return &Handler{deleteUser: deleteUser}
}
