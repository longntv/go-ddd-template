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

// NewCreateUser creates a new CreateUser use case.
func NewCreateUser(
	userCommandsGateway gateway.UserCommandsGateway,
	userQueriesGateway gateway.UserQueriesGateway,
	passwordHasher gateway.PasswordHasher,
	transactor gateway.Transactor,
	eventOutbox gateway.EventOutbox,
) usecase.CreateUser {
	return &createUser{
		userCommandsGateway: userCommandsGateway,
		userQueriesGateway:  userQueriesGateway,
		passwordHasher:      passwordHasher,
		transactor:          transactor,
		eventOutbox:         eventOutbox,
	}
}

// createUser implements the CreateUser use case.
type createUser struct {
	userCommandsGateway gateway.UserCommandsGateway
	userQueriesGateway  gateway.UserQueriesGateway
	passwordHasher      gateway.PasswordHasher
	transactor          gateway.Transactor
	eventOutbox         gateway.EventOutbox
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

	// Hash the password; the plain text never leaves this function.
	passwordHash, err := s.passwordHasher.Hash(in.Password)
	if err != nil {
		if errors.Is(err, model.ErrPasswordTooLong) {
			return nil, model.NewDomainError("INVALID_INPUT", "password must be at most 72 bytes", err)
		}
		return nil, model.NewDomainError("INTERNAL", "failed to hash password", err)
	}

	// Create user entity
	userEntity := entity.NewUser(in.Name, in.Email, passwordHash)

	// Save the user and its event in one transaction, so the event is
	// published (by the outbox relay) if and only if the user is saved.
	var createdUser *entity.User
	err = s.transactor.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.userCommandsGateway.Create(ctx, userEntity); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		user, err := s.userQueriesGateway.GetByEmail(ctx, in.Email)
		if err != nil {
			return fmt.Errorf("get created user: %w", err)
		}
		createdUser = user

		evt := event.NewDomainEvent(
			event.UserCreatedEvent,
			event.Source,
			createdUser.ID.String(),
			&event.UserEventData{
				ID:    createdUser.ID.String(),
				Name:  createdUser.Name,
				Email: createdUser.Email,
			},
		)
		if err := s.eventOutbox.Add(ctx, evt); err != nil {
			return fmt.Errorf("add %s event to outbox: %w", evt.Type, err)
		}
		return nil
	})
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to create user", err)
	}

	return &output.CreateUser{User: createdUser}, nil
}
