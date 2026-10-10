package event

import (
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

// Source identifies this service as the producer of its events.
const Source = "go-ddd-template"

// DomainEvent is something that happened to an aggregate, published to other
// services. Every aggregate uses this type; the aggregate's event file only
// defines its type constants and data payload.
type DomainEvent struct {
	ID        uuid.UUID
	Type      string // e.g. UserCreatedEvent
	Source    string
	Subject   string // ID of the aggregate the event is about
	Data      any    // JSON-serialisable payload, e.g. *UserEventData
	Timestamp time.Time
}

// NewDomainEvent creates an event with a fresh ID and the current time.
func NewDomainEvent(eventType, source, subject string, data any) *DomainEvent {
	return &DomainEvent{
		ID:        uuid.New(),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ToCloudEvent converts the event to a CloudEvents SDK event. It fails when
// Data cannot be encoded as JSON.
func (e *DomainEvent) ToCloudEvent() (event.Event, error) {
	ce := event.New()
	ce.SetID(e.ID.String())
	ce.SetType(e.Type)
	ce.SetSource(e.Source)
	ce.SetSubject(e.Subject)
	ce.SetTime(e.Timestamp)
	if err := ce.SetData(event.ApplicationJSON, e.Data); err != nil {
		return event.Event{}, fmt.Errorf("encode %s event data: %w", e.Type, err)
	}
	return ce, nil
}
