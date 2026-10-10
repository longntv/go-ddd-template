package datastore

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

// errNoTransaction is returned by eventOutbox.Add outside Transactor.RunInTx.
var errNoTransaction = errors.New("outbox: Add must be called inside Transactor.RunInTx")

// eventOutbox implements gateway.EventOutbox.
type eventOutbox struct {
	db *gorm.DB
}

// NewEventOutbox creates a new EventOutbox.
func NewEventOutbox(db *gorm.DB) gateway.EventOutbox {
	return &eventOutbox{db: db}
}

// Add inserts evt as a pending outbox row in the caller's transaction. It
// refuses to run without one: the event could then be saved without the
// change it describes, or the other way round.
func (o *eventOutbox) Add(ctx context.Context, evt *event.DomainEvent) error {
	if _, ok := txFromContext(ctx); !ok {
		return errNoTransaction
	}
	row, err := newOutboxEventEntity(evt)
	if err != nil {
		return err
	}
	return conn(ctx, o.db).Create(row).Error
}
