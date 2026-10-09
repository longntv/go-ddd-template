package cloudevents

import "github.com/google/wire"

// WireSet holds the Wire providers for CloudEvents handler.
var WireSet = wire.NewSet(
	NewHandler,
)
