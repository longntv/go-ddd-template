package usecase

import (
	"context"

	"go-ddd-template/internal/usecase/input"
	"go-ddd-template/internal/usecase/output"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=mock/usecase.go -source=usecase.go

// CreateUser defines the use case for creating a user.
type CreateUser interface {
	Execute(ctx context.Context, in *input.CreateUser) (*output.CreateUser, error)
}

// GetUser defines the use case for getting a user.
type GetUser interface {
	Execute(ctx context.Context, in *input.GetUser) (*output.GetUser, error)
}

// ListUsers defines the use case for listing users.
type ListUsers interface {
	Execute(ctx context.Context, in *input.ListUsers) (*output.ListUsers, error)
}

// UpdateUser defines the use case for updating a user.
type UpdateUser interface {
	Execute(ctx context.Context, in *input.UpdateUser) (*output.UpdateUser, error)
}

// DeleteUser defines the use case for deleting a user.
type DeleteUser interface {
	Execute(ctx context.Context, in *input.DeleteUser) error
}
