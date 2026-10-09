package service

import "github.com/google/wire"

// WireSet holds the Wire providers for user service.
var WireSet = wire.NewSet(
	NewCreateUser,
	NewGetUser,
	NewListUsers,
	NewUpdateUser,
	NewDeleteUser,
)
