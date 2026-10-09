package service

import (
	"context"
	"errors"

	"github.com/longntv/go-ddd-template/internal/domain/gateway"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"
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
