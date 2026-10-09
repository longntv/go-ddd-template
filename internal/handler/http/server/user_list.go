package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// List handles GET /users
func (h *Handler) List(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 || limitInt > 100 {
		limitInt = 10
	}

	input := &input.ListUsers{
		Page:  pageInt,
		Limit: limitInt,
	}

	output, err := h.listUsers.Execute(c.Request.Context(), input)
	if err != nil {
		handleError(c, err)
		return
	}

	users := make([]gin.H, 0, len(output.Users))
	for _, u := range output.Users {
		users = append(users, gin.H{
			"id":         u.ID.String(),
			"name":       u.Name,
			"email":      u.Email,
			"created_at": u.CreatedAt,
			"updated_at": u.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"users":       users,
		"total_count": output.TotalCount,
		"page":        output.Page,
		"limit":       output.Limit,
	})
}
