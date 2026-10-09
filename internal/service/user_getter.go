package service

import (
	"context"
	"errors"

	"go-ddd-template/internal/domain/gateway"
	"go-ddd-template/internal/domain/model"
	"go-ddd-template/internal/usecase"
	"go-ddd-template/internal/usecase/input"
	"go-ddd-template/internal/usecase/output"
)

// NewGetUser creates a new GetUser use case.
func NewGetUser(
	userQueriesGateway gateway.UserQueriesGateway,
) usecase.GetUser {
	return &getUser{
		userQueriesGateway: userQueriesGateway,
	}
}

// getUser implements the GetUser use case.
type getUser struct {
	userQueriesGateway gateway.UserQueriesGateway
}

func (s *getUser) Execute(ctx context.Context, in *input.GetUser) (*output.GetUser, error) {
	userEntity, err := s.userQueriesGateway.Get(ctx, in.ID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, model.NewDomainError("USER_NOT_FOUND", "user not found", model.ErrUserNotFound)
		}
		return nil, model.NewDomainError("INTERNAL", "failed to get user", err)
	}

	return &output.GetUser{User: userEntity}, nil
}
