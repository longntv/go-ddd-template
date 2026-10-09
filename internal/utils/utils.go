package utils

import (
	"github.com/google/wire"

	"go-ddd-template/internal/utils/time"
	"go-ddd-template/internal/utils/validator"
)

// WireSet holds the Wire providers for utils.
var WireSet = wire.NewSet(
	time.WireSet,
	validator.WireSet,
)
