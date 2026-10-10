package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"
)

// NewUpdateUser creates a new UpdateUser use case.
func NewUpdateUser(
	userCommandsGateway gateway.UserCommandsGateway,
	userQueriesGateway gateway.UserQueriesGateway,
	passwordHasher gateway.PasswordHasher,
	eventPublisher gateway.EventPublisher,
	logger *zap.Logger,
) usecase.UpdateUser {
	return &updateUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		passwordHasher:      passwordHasher,
		eventPublisher:      eventPublisher,
		logger:              logger,
	}
}

// updateUser implements the UpdateUser use case.
type updateUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	passwordHasher      gateway.PasswordHasher
	eventPublisher      gateway.EventPublisher
	logger              *zap.Logger
}

func (s *updateUser) Execute(ctx context.Context, in *input.UpdateUser) (*output.UpdateUser, error) {
	// Check if user exists
	_, err := s.userQueriesGateway.Get(ctx, in.ID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.NewDomainError("USER_NOT_FOUND", "user not found", model.ErrUserNotFound)
		}
		return nil, model.NewDomainError("INTERNAL", "failed to get user", err)
	}

	// Hash the new password; the plain text never leaves this function.
	passwordHash, err := s.passwordHasher.Hash(in.Password)
	if err != nil {
		if errors.Is(err, model.ErrPasswordTooLong) {
			return nil, model.NewDomainError("INVALID_INPUT", "password must be at most 72 bytes", err)
		}
		return nil, model.NewDomainError("INTERNAL", "failed to hash password", err)
	}

	// Update user entity
	userEntity := &entity.User{ID: in.ID}
	userEntity.Update(in.Name, in.Email, passwordHash)

	// Update user
	if err := s.userCommandsGateway.Update(ctx, userEntity); err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to update user", err)
	}

	// Get updated user
	updatedUser, err := s.userQueriesGateway.Get(ctx, in.ID)
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to get updated user", err)
	}

	// Publish event
	evt := event.NewDomainEvent(
		event.UserUpdatedEvent,
		event.Source,
		updatedUser.ID.String(),
		&event.UserEventData{
			ID:    updatedUser.ID.String(),
			Name:  updatedUser.Name,
			Email: updatedUser.Email,
		},
	)
	publishBestEffort(ctx, s.eventPublisher, s.logger, evt)

	return &output.UpdateUser{User: updatedUser}, nil
}
