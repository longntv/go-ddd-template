package cloudevents

import (
	"context"
	"encoding/json"
	"log"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Handler handles CloudEvents.
type Handler struct{}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// HandleUserCreated handles UserCreated events.
func (h *Handler) HandleUserCreated(ctx context.Context, evt *event.Event) error {
	var data struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(evt.Data(), &data); err != nil {
		return err
	}

	log.Printf("[UserCreated] User ID: %s, Name: %s, Email: %s", data.ID, data.Name, data.Email)
	return nil
}

// HandleUserUpdated handles UserUpdated events.
func (h *Handler) HandleUserUpdated(ctx context.Context, evt *event.Event) error {
	var data struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(evt.Data(), &data); err != nil {
		return err
	}

	log.Printf("[UserUpdated] User ID: %s, Name: %s, Email: %s", data.ID, data.Name, data.Email)
	return nil
}

// HandleUserDeleted handles UserDeleted events.
func (h *Handler) HandleUserDeleted(ctx context.Context, evt *event.Event) error {
	var data struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(evt.Data(), &data); err != nil {
		return err
	}

	log.Printf("[UserDeleted] User ID: %s, Name: %s, Email: %s", data.ID, data.Name, data.Email)
	return nil
}
