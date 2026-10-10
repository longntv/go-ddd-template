package gateway

import (
	"context"

	"github.com/longntv/go-ddd-template/internal/domain/event"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/event.go -source=event.go

// EventOutbox records domain events to be published once the transaction that
// records them commits (transactional outbox). Services use this, not
// EventPublisher: an event is then published if and only if the change it
// describes is saved.
type EventOutbox interface {
	// Add records evt. It must be called with the ctx of Transactor.RunInTx,
	// in the same transaction as the change evt describes.
	Add(ctx context.Context, evt *event.DomainEvent) error
}

// EventPublisher sends a domain event to other services. Only the outbox
// relay calls it; delivery is at least once, so consumers must deduplicate
// by event ID.
type EventPublisher interface {
	// Publish publishes a domain event.
	Publish(ctx context.Context, evt *event.DomainEvent) error
}
