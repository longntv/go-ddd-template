package event_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/longntv/go-ddd-template/internal/domain/event"
)

func TestDomainEvent_ToCloudEvent(t *testing.T) {
	t.Parallel()

	type expected struct {
		ID, Type, Source, Subject, DataContentType, Data string
		Time                                             time.Time
	}

	tests := map[string]struct {
		data    any
		wantErr bool
		want    func(e *event.DomainEvent) expected
	}{
		"copies the envelope and encodes data as JSON": {
			data: &event.UserEventData{ID: "11111111-1111-1111-1111-111111111111", Name: "Alice", Email: "alice@example.com"},
			want: func(e *event.DomainEvent) expected {
				return expected{
					ID: e.ID.String(), Type: event.UserCreatedEvent, Source: event.Source,
					Subject: "11111111-1111-1111-1111-111111111111", DataContentType: "application/json",
					Data: `{"id":"11111111-1111-1111-1111-111111111111","name":"Alice","email":"alice@example.com"}`,
					Time: e.Timestamp,
				}
			},
		},
		"data that cannot be encoded returns an error": {
			data:    make(chan int),
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := event.NewDomainEvent(event.UserCreatedEvent, event.Source, "11111111-1111-1111-1111-111111111111", tt.data)
			ce, err := e.ToCloudEvent()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ToCloudEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			actual := expected{
				ID: ce.ID(), Type: ce.Type(), Source: ce.Source(), Subject: ce.Subject(),
				DataContentType: ce.DataContentType(), Data: string(ce.Data()), Time: ce.Time(),
			}
			if diff := cmp.Diff(tt.want(e), actual); diff != "" {
				t.Errorf("ToCloudEvent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewDomainEvent(t *testing.T) {
	t.Parallel()

	data := &event.UserEventData{ID: "11111111-1111-1111-1111-111111111111"}
	tests := map[string]struct {
		eventType, subject string
		data               any
	}{
		"with data":    {eventType: event.UserCreatedEvent, subject: data.ID, data: data},
		"without data": {eventType: event.UserDeletedEvent, subject: data.ID},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			got := event.NewDomainEvent(tt.eventType, event.Source, tt.subject, tt.data)
			other := event.NewDomainEvent(tt.eventType, event.Source, tt.subject, tt.data)

			want := &event.DomainEvent{Type: tt.eventType, Source: event.Source, Subject: tt.subject, Data: tt.data}
			if diff := cmp.Diff(want, got, cmpopts.IgnoreFields(event.DomainEvent{}, "ID", "Timestamp")); diff != "" {
				t.Errorf("NewDomainEvent() mismatch (-want +got):\n%s", diff)
			}
			if got.ID == uuid.Nil || got.ID == other.ID {
				t.Errorf("NewDomainEvent() ID = %s, want a fresh non-nil ID per event", got.ID)
			}
			if got.Timestamp.Before(before) {
				t.Errorf("NewDomainEvent() Timestamp = %v, want the creation time", got.Timestamp)
			}
		})
	}
}
