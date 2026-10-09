package handler

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Mux routes CloudEvents to their handlers.
type Mux struct {
	handlers map[string]Handler
}

// Handler handles a CloudEvent.
type Handler interface {
	Handle(ctx context.Context, evt *event.Event) error
}

// NewMux creates a new Mux.
func NewMux() *Mux {
	return &Mux{
		handlers: make(map[string]Handler),
	}
}

// Register registers a handler for an event type.
func (m *Mux) Register(eventType string, handler Handler) {
	m.handlers[eventType] = handler
}

// Handle routes a CloudEvent to its handler.
func (m *Mux) Handle(ctx context.Context, evt *event.Event) error {
	eventType := evt.Type()

	handler, ok := m.handlers[eventType]
	if !ok {
		return fmt.Errorf("no handler registered for event type: %s", eventType)
	}

	return handler.Handle(ctx, evt)
}
