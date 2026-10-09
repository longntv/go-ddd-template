package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// Update handles PUT /users/:id
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var in struct {
		Name     string `json:"name" binding:"required,min=1,max=255"`
		Email    string `json:"email" binding:"required,email,max=255"`
		Password string `json:"password" binding:"required,min=8,max=72"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := &input.UpdateUser{
		ID:       id,
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
	}

	output, err := h.updateUser.Execute(c.Request.Context(), input)
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
