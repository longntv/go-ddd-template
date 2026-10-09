package service

import (
	"context"
	"errors"

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
	eventPublisher gateway.EventPublisher,
) usecase.UpdateUser {
	return &updateUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		eventPublisher:      eventPublisher,
	}
}

// updateUser implements the UpdateUser use case.
type updateUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	eventPublisher      gateway.EventPublisher
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

	// Update user entity
	userEntity := &entity.User{
		ID:       in.ID,
		Name:     in.Name,
		Email:    in.Email,
		Password: in.Password,
	}
	userEntity.Update(in.Name, in.Email, in.Password)

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
	userEvent := event.NewUserEvent(
		event.UserUpdatedEvent,
		"go-ddd-template",
		updatedUser.ID.String(),
		&event.UserEventData{
			ID:    updatedUser.ID.String(),
			Name:  updatedUser.Name,
			Email: updatedUser.Email,
		},
	)
	if err := s.eventPublisher.Publish(ctx, userEvent); err != nil {
		// Log error but don't fail the request
		_ = err
	}

	return &output.UpdateUser{User: updatedUser}, nil
}
