package cloudevents_test

import (
	"context"
	"testing"

	"github.com/longntv/go-ddd-template/internal/domain/event"
	handlercloudevents "github.com/longntv/go-ddd-template/internal/handler/cloudevents"
	"github.com/longntv/go-ddd-template/internal/infrastructure/cloudevents"
)

// TestProvideConfiguredMux_RoutesDomainEvents guards the contract between the
// publisher and the subscriber: every event type the domain emits must have a
// registered handler, or the subscriber rejects it.
func TestProvideConfiguredMux_RoutesDomainEvents(t *testing.T) {
	t.Parallel()

	mux := cloudevents.ProvideConfiguredMux(handlercloudevents.NewHandler())

	tests := map[string]struct {
		eventType string
		wantErr   bool
	}{
		"user created":       {eventType: event.UserCreatedEvent},
		"user updated":       {eventType: event.UserUpdatedEvent},
		"user deleted":       {eventType: event.UserDeletedEvent},
		"unknown event type": {eventType: "com.go-ddd-template.unknown", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			evt := event.NewUserEvent(tt.eventType, "go-ddd-template", "subject", &event.UserEventData{
				ID:    "11111111-1111-1111-1111-111111111111",
				Name:  "Alice",
				Email: "alice@example.com",
			}).ToCloudEvent()

			err := mux.Handle(context.Background(), &evt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Mux.Handle(%s) error = %v, wantErr %v", tt.eventType, err, tt.wantErr)
			}
		})
	}
}
