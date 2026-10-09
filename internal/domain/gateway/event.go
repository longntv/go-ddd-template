package gateway

import (
	"context"

	"github.com/longntv/go-ddd-template/internal/domain/event"
)

//go:generate go run go.uber.org/mock/mockgen@latest -destination=mock/event.go -source=event.go

// EventPublisher defines the interface for publishing domain events.
type EventPublisher interface {
	// Publish publishes a domain event.
	Publish(ctx context.Context, evt *event.UserEvent) error
}
