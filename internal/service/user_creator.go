package service

import (
	"context"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"
)

// NewCreateUser creates a new CreateUser use case.
func NewCreateUser(
	userCommandsGateway gateway.UserCommandsGateway,
	userQueriesGateway gateway.UserQueriesGateway,
	eventPublisher gateway.EventPublisher,
) usecase.CreateUser {
	return &createUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		eventPublisher:      eventPublisher,
	}
}

// createUser implements the CreateUser use case.
type createUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	eventPublisher      gateway.EventPublisher
}

func (s *createUser) Execute(ctx context.Context, in *input.CreateUser) (*output.CreateUser, error) {
	// Check if user already exists
	exists, err := s.userQueriesGateway.Exists(ctx, in.Email)
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to check user existence", err)
	}
	if exists {
		return nil, model.NewDomainError("USER_EXISTS", "user with this email already exists", model.ErrUserAlreadyExists)
	}

	// Create user entity
	userEntity := entity.NewUser(in.Name, in.Email, in.Password)

	// Create user
	if err := s.userCommandsGateway.Create(ctx, userEntity); err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to create user", err)
	}

	// Get created user
	createdUser, err := s.userQueriesGateway.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to get created user", err)
	}

	// Publish event
	userEvent := event.NewUserEvent(
		event.UserCreatedEvent,
		"go-ddd-template",
		createdUser.ID.String(),
		&event.UserEventData{
			ID:    createdUser.ID.String(),
			Name:  createdUser.Name,
			Email: createdUser.Email,
		},
	)
	if err := s.eventPublisher.Publish(ctx, userEvent); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	return &output.CreateUser{User: createdUser}, nil
}
