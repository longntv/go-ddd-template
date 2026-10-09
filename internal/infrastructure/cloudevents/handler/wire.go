package handler

import (
	"github.com/google/wire"
)

// WireSet holds the Wire providers for CloudEvents handler mux.
var WireSet = wire.NewSet(
	NewMux,
)
