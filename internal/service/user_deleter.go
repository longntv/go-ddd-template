package service

import (
	"context"
	"errors"

	"go-ddd-template/internal/domain/event"
	"go-ddd-template/internal/domain/gateway"
	"go-ddd-template/internal/domain/model"
	"go-ddd-template/internal/usecase"
	"go-ddd-template/internal/usecase/input"
)

// NewDeleteUser creates a new DeleteUser use case.
func NewDeleteUser(
	userCommandsGateway gateway.UserCommandsGateway,
	userQueriesGateway gateway.UserQueriesGateway,
	eventPublisher gateway.EventPublisher,
) usecase.DeleteUser {
	return &deleteUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		eventPublisher:      eventPublisher,
	}
}

// deleteUser implements the DeleteUser use case.
type deleteUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	eventPublisher      gateway.EventPublisher
}

func (s *deleteUser) Execute(ctx context.Context, in *input.DeleteUser) error {
	// Check if user exists
	userEntity, err := s.userQueriesGateway.Get(ctx, in.ID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return model.NewDomainError("USER_NOT_FOUND", "user not found", model.ErrUserNotFound)
		}
		return model.NewDomainError("INTERNAL", "failed to get user", err)
	}

	// Delete user
	if err := s.userCommandsGateway.Delete(ctx, userEntity.ID); err != nil {
		return model.NewDomainError("INTERNAL", "failed to delete user", err)
	}

	// Publish event
	userEvent := event.NewUserEvent(
		event.UserDeletedEvent,
		"go-ddd-template",
		userEntity.ID.String(),
		&event.UserEventData{
			ID:    userEntity.ID.String(),
			Name:  userEntity.Name,
			Email: userEntity.Email,
		},
	)
	if err := s.eventPublisher.Publish(ctx, userEvent); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	return nil
}
