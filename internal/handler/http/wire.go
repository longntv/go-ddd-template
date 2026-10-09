package http

import "github.com/google/wire"

// WireSet holds the Wire providers for HTTP router.
var WireSet = wire.NewSet(
	Router,
)
