package event

const (
	// UserEventTypePrefix is the prefix for all user-related events.
	UserEventTypePrefix = "com.go-ddd-template.user"

	// Event types
	UserCreatedEvent = UserEventTypePrefix + ".created"
	UserUpdatedEvent = UserEventTypePrefix + ".updated"
	UserDeletedEvent = UserEventTypePrefix + ".deleted"
)

// UserEventData is the payload of every user event. It never carries the
// password hash.
type UserEventData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
