package input

import (
	"go-ddd-template/internal/domain/entity"
)

// CreateUser holds the input for creating a user.
type CreateUser struct {
	Name     string `validate:"required,min=1,max=255"`
	Email    string `validate:"required,email,max=255"`
	Password string `validate:"required,min=8"`
}

// GetUser holds the input for getting a user.
type GetUser struct {
	ID entity.UserID `validate:"required"`
}

// ListUsers holds the input for listing users.
type ListUsers struct {
	Page  int `validate:"required,min=1"`
	Limit int `validate:"required,min=1,max=100"`
}

// UpdateUser holds the input for updating a user.
type UpdateUser struct {
	ID       entity.UserID `validate:"required"`
	Name     string        `validate:"required,min=1,max=255"`
	Email    string        `validate:"required,email,max=255"`
	Password string        `validate:"required,min=8"`
}

// DeleteUser holds the input for deleting a user.
type DeleteUser struct {
	ID entity.UserID `validate:"required"`
}
