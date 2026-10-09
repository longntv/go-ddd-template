package health

import "github.com/google/wire"

// WireSet holds the Wire providers for health handler.
var WireSet = wire.NewSet(
	NewHealthHandler,
)
