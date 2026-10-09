package service

import (
	"context"

	"go-ddd-template/internal/domain/gateway"
	"go-ddd-template/internal/domain/model"
	"go-ddd-template/internal/usecase"
	"go-ddd-template/internal/usecase/input"
	"go-ddd-template/internal/usecase/output"
)

// NewListUsers creates a new ListUsers use case.
func NewListUsers(
	userQueriesGateway gateway.UserQueriesGateway,
) usecase.ListUsers {
	return &listUsers{
		userQueriesGateway: userQueriesGateway,
	}
}

// listUsers implements the ListUsers use case.
type listUsers struct {
	userQueriesGateway gateway.UserQueriesGateway
}

func (s *listUsers) Execute(ctx context.Context, in *input.ListUsers) (*output.ListUsers, error) {
	offset := (in.Page - 1) * in.Limit

	users, total, err := s.userQueriesGateway.List(ctx, in.Limit, offset)
	if err != nil {
		return nil, model.NewDomainError("INTERNAL", "failed to list users", err)
	}

	return &output.ListUsers{
		Users:      users,
		TotalCount: total,
		Page:       in.Page,
		Limit:      in.Limit,
	}, nil
}
