package service

import (
	"context"
	"errors"
	"fmt"

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
	transactor gateway.Transactor,
	eventOutbox gateway.EventOutbox,
) usecase.DeleteUser {
	return &deleteUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		transactor:          transactor,
		eventOutbox:         eventOutbox,
	}
}

// deleteUser implements the DeleteUser use case.
type deleteUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	transactor          gateway.Transactor
	eventOutbox         gateway.EventOutbox
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

	// Delete the user and save its event in one transaction, so the event is
	// published (by the outbox relay) if and only if the user is deleted.
	err = s.transactor.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userCommandsGateway.Delete(ctx, userEntity.ID); err != nil {
			return fmt.Errorf("delete user: %w", err)
		}

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
		if err := s.eventOutbox.Add(ctx, evt); err != nil {
			return fmt.Errorf("add %s event to outbox: %w", evt.Type, err)
		}
		return nil
	})
	if err != nil {
		return model.NewDomainError("INTERNAL", "failed to delete user", err)
	}

	return nil
}
