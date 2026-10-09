package time

import (
	"time"

	"github.com/google/wire"
)

// Clock is the interface for time operations.
type Clock interface {
	Now() time.Time
}

// realClock implements Clock interface using real time.
type realClock struct{}

// NewRealClock creates a new realClock instance.
func NewRealClock() Clock {
	return &realClock{}
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

// WireSet holds the Wire providers for time.
var WireSet = wire.NewSet(
	NewRealClock,
)
