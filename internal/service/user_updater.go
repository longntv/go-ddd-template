package service

import (
	"context"
	"errors"
	"fmt"

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
	transactor gateway.Transactor,
	eventOutbox gateway.EventOutbox,
) usecase.UpdateUser {
	return &updateUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		passwordHasher:      passwordHasher,
		transactor:          transactor,
		eventOutbox:         eventOutbox,
	}
}

// updateUser implements the UpdateUser use case.
type updateUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	passwordHasher      gateway.PasswordHasher
	transactor          gateway.Transactor
	eventOutbox         gateway.EventOutbox
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

	// Save the change and its event in one transaction, so the event is
	// published (by the outbox relay) if and only if the change is saved.
	var updatedUser *entity.User
	err = s.transactor.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userCommandsGateway.Update(ctx, userEntity); err != nil {
			return fmt.Errorf("update user: %w", err)
		}

		user, err := s.userQueriesGateway.Get(ctx, in.ID)
		if err != nil {
			return fmt.Errorf("get updated user: %w", err)
		}
		updatedUser = user

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
		if err := s.eventOutbox.Add(ctx, evt); err != nil {
			return fmt.Errorf("add %s event to outbox: %w", evt.Type, err)
		}
		return nil
	})
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to update user", err)
	}

	return &output.UpdateUser{User: updatedUser}, nil
}
