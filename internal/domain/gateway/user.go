package gateway

import (
	"context"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/user.go -source=user.go

// Binder defines the interface for binding database to context.
type Binder interface {
	Bind(ctx context.Context) context.Context
}

// UserQueriesGateway defines the interface for querying users (CQRS read side).
type UserQueriesGateway interface {
	// Get retrieves a user by ID.
	Get(ctx context.Context, id entity.UserID) (*entity.User, error)

	// GetByEmail retrieves a user by email.
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// List retrieves a list of users with pagination.
	List(ctx context.Context, limit, offset int) ([]*entity.User, int, error)

	// Exists checks if a user exists by email.
	Exists(ctx context.Context, email string) (bool, error)
}

// UserCommandsGateway defines the interface for modifying users (CQRS write side).
type UserCommandsGateway interface {
	// Create creates a new user.
	Create(ctx context.Context, user *entity.User) error

	// Update updates an existing user.
	Update(ctx context.Context, user *entity.User) error

	// Delete deletes a user by ID.
	Delete(ctx context.Context, id entity.UserID) error
}
