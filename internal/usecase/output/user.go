package output

import (
	"go-ddd-template/internal/domain/entity"
)

// CreateUser is the output for creating a user.
type CreateUser struct {
	User *entity.User
}

// GetUser is the output for getting a user.
type GetUser struct {
	User *entity.User
}

// ListUsers is the output for listing users.
type ListUsers struct {
	Users      []*entity.User
	TotalCount int
	Page       int
	Limit      int
}

// UpdateUser is the output for updating a user.
type UpdateUser struct {
	User *entity.User
}
