package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
)

// NewDeleteUser creates a new DeleteUser use case.
func NewDeleteUser(
	userCommandsGateway gateway.UserCommandsGateway,
	userQueriesGateway gateway.UserQueriesGateway,
	eventPublisher gateway.EventPublisher,
	logger *zap.Logger,
) usecase.DeleteUser {
	return &deleteUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		eventPublisher:      eventPublisher,
		logger:              logger,
	}
}

// deleteUser implements the DeleteUser use case.
type deleteUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	eventPublisher      gateway.EventPublisher
	logger              *zap.Logger
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
	evt := event.NewDomainEvent(
		event.UserDeletedEvent,
		event.Source,
		userEntity.ID.String(),
		&event.UserEventData{
			ID:    userEntity.ID.String(),
			Name:  userEntity.Name,
			Email: userEntity.Email,
		},
	)
	publishBestEffort(ctx, s.eventPublisher, s.logger, evt)

	return nil
}
