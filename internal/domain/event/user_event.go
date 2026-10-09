package event

import (
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

const (
	// UserEventTypePrefix is the prefix for all user-related events.
	UserEventTypePrefix = "com.go-ddd-template.user"

	// Event types
	UserCreatedEvent = UserEventTypePrefix + ".created"
	UserUpdatedEvent = UserEventTypePrefix + ".updated"
	UserDeletedEvent = UserEventTypePrefix + ".deleted"
)

// UserEvent represents a user-related domain event.
type UserEvent struct {
	ID        uuid.UUID
	Type      string
	Source    string
	Subject   string
	Data      interface{}
	Timestamp time.Time
}

// NewUserEvent creates a new UserEvent.
func NewUserEvent(eventType, source, subject string, data interface{}) *UserEvent {
	return &UserEvent{
		ID:        uuid.New(),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ToCloudEvent converts UserEvent to CloudEvents SDK event.
func (e *UserEvent) ToCloudEvent() event.Event {
	cloudevent := event.New()
	cloudevent.SetID(e.ID.String())
	cloudevent.SetType(e.Type)
	cloudevent.SetSource(e.Source)
	cloudevent.SetSubject(e.Subject)
	cloudevent.SetTime(e.Timestamp)
	cloudevent.SetDataContentType("application/json")
	_ = cloudevent.SetData("application/json", e.Data)
	return cloudevent
}

// UserEventData represents the data payload for user events.
type UserEventData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
