//go:build integration

package http_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/longntv/go-ddd-template/internal/domain/event"
)

const (
	aliceID   = "11111111-1111-1111-1111-111111111111" // testdata/fixtures/users.yml
	unknownID = "99999999-9999-9999-9999-999999999999"
)

// expectPublished expects exactly one published event of the given type.
func expectPublished(t *testing.T, h *HTTPTestHelper, eventType string) {
	t.Helper()

	h.mockPublisher.EXPECT().
		Publish(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, evt *event.DomainEvent) error {
			if evt.Type != eventType {
				t.Errorf("published event type = %s, want %s", evt.Type, eventType)
			}
			return nil
		}).
		Times(1)
}

func Test_Integration_CreateUser(t *testing.T) {
	t.Parallel()

	type testcase struct {
		body           string
		prepare        func(t *testing.T, h *HTTPTestHelper)
		wantStatus     int
		wantUserCount  int64
		wantBodyFields map[string]any
	}

	tests := map[string]testcase{
		"creates user, stores it and publishes UserCreated": {
			body: `{"name":"Dave","email":"dave@example.com","password":"password123"}`,
			prepare: func(t *testing.T, h *HTTPTestHelper) {
				expectPublished(t, h, event.UserCreatedEvent)
			},
			wantStatus:     http.StatusCreated,
			wantUserCount:  4,
			wantBodyFields: map[string]any{"name": "Dave", "email": "dave@example.com"},
		},
		"duplicate email returns 409 and stores nothing": {
			body:          `{"name":"Alice 2","email":"alice@example.com","password":"password123"}`,
			wantStatus:    http.StatusConflict,
			wantUserCount: 3,
		},
		"invalid body returns 400": {
			body:          `{"name":"Dave","email":"dave@example.com","password":"short"}`,
			wantStatus:    http.StatusBadRequest,
			wantUserCount: 3,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := NewHTTPTestHelper(t)
			if tt.prepare != nil {
				tt.prepare(t, h)
			}

			status, body := h.Do(t, http.MethodPost, "/api/v1/users", tt.body)

			expectStatus(t, status, tt.wantStatus, body)
			for k, want := range tt.wantBodyFields {
				if diff := cmp.Diff(want, body[k]); diff != "" {
					t.Errorf("body[%q] mismatch (-want +got):\n%s", k, diff)
				}
			}
			if got := h.CountUsers(t); got != tt.wantUserCount {
				t.Errorf("users in db = %d, want %d", got, tt.wantUserCount)
			}
		})
	}
}

func Test_Integration_PasswordIsStoredHashed(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		method, path, body, email, password string
		event                               string
		wantStatus                          int
	}{
		"on create": {
			method: http.MethodPost, path: "/api/v1/users", email: "gina@example.com", password: "gina-password",
			body:  `{"name":"Gina","email":"gina@example.com","password":"gina-password"}`,
			event: event.UserCreatedEvent, wantStatus: http.StatusCreated,
		},
		"on update": {
			method: http.MethodPut, path: "/api/v1/users/" + aliceID, email: "alice@example.com", password: "new-alice-password",
			body:  `{"name":"Alice","email":"alice@example.com","password":"new-alice-password"}`,
			event: event.UserUpdatedEvent, wantStatus: http.StatusOK,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := NewHTTPTestHelper(t)
			expectPublished(t, h, tt.event)

			status, body := h.Do(t, tt.method, tt.path, tt.body)
			expectStatus(t, status, tt.wantStatus, body)

			if _, ok := body["password"]; ok {
				t.Error("response contains a password field")
			}
			hash := h.PasswordHash(t, tt.email)
			if hash == tt.password {
				t.Fatal("password stored in plain text")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password)); err != nil {
				t.Errorf("stored hash does not match the password: %v", err)
			}
		})
	}
}

// Test_Integration_RejectsPasswordsOverBcryptByteLimit covers passwords that pass
// the 72-character binding but exceed bcrypt's 72-byte limit: they must get
// 400, not 500, and nothing must be stored or published.
func Test_Integration_RejectsPasswordsOverBcryptByteLimit(t *testing.T) {
	t.Parallel()

	password := strings.Repeat("é", 40) // 40 characters, 80 bytes
	tests := map[string]struct {
		method, path, body string
	}{
		"on create": {
			method: http.MethodPost, path: "/api/v1/users",
			body: `{"name":"Hana","email":"hana@example.com","password":"` + password + `"}`,
		},
		"on update": {
			method: http.MethodPut, path: "/api/v1/users/" + aliceID,
			body: `{"name":"Alice","email":"alice@example.com","password":"` + password + `"}`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := NewHTTPTestHelper(t)
			usersBefore := h.CountUsers(t)
			aliceHashBefore := h.PasswordHash(t, "alice@example.com")

			status, body := h.Do(t, tt.method, tt.path, tt.body)
			expectStatus(t, status, http.StatusBadRequest, body)

			if got := h.CountUsers(t); got != usersBefore {
				t.Errorf("CountUsers() = %d, want %d", got, usersBefore)
			}
			if got := h.PasswordHash(t, "alice@example.com"); got != aliceHashBefore {
				t.Error("alice's password hash changed")
			}
		})
	}
}

func Test_Integration_GetUser(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		id         string
		wantStatus int
		wantBody   map[string]any
	}{
		"returns fixture user": {
			id:         aliceID,
			wantStatus: http.StatusOK,
			wantBody: map[string]any{
				"id": aliceID, "name": "Alice", "email": "alice@example.com",
				"created_at": "2025-01-01T10:00:00Z", "updated_at": "2025-01-01T10:00:00Z",
			},
		},
		"unknown id returns 404": {
			id:         unknownID,
			wantStatus: http.StatusNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := NewHTTPTestHelper(t)
			status, body := h.Do(t, http.MethodGet, "/api/v1/users/"+tt.id, "")

			expectStatus(t, status, tt.wantStatus, body)
			if tt.wantBody == nil {
				return
			}
			if diff := cmp.Diff(tt.wantBody, body); diff != "" {
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_Integration_ListUsers(t *testing.T) {
	t.Parallel()

	h := NewHTTPTestHelper(t)
	status, body := h.Do(t, http.MethodGet, "/api/v1/users?page=2&limit=2", "")

	expectStatus(t, status, http.StatusOK, body)
	if body["total_count"] != float64(3) {
		t.Errorf("total_count = %v, want 3", body["total_count"])
	}
	if users, _ := body["users"].([]any); len(users) != 1 {
		t.Errorf("page 2 returned %d users, want 1", len(users))
	}
}

func Test_Integration_UpdateUser(t *testing.T) {
	t.Parallel()

	h := NewHTTPTestHelper(t)
	expectPublished(t, h, event.UserUpdatedEvent)

	status, body := h.Do(t, http.MethodPut, "/api/v1/users/"+aliceID,
		`{"name":"Alice B","email":"alice.b@example.com","password":"password123"}`)
	expectStatus(t, status, http.StatusOK, body)

	// Read back through the API to prove the change was persisted.
	status, body = h.Do(t, http.MethodGet, "/api/v1/users/"+aliceID, "")
	expectStatus(t, status, http.StatusOK, body)
	if body["name"] != "Alice B" || body["email"] != "alice.b@example.com" {
		t.Errorf("after update got name=%v email=%v", body["name"], body["email"])
	}
}

func Test_Integration_DeleteUser(t *testing.T) {
	t.Parallel()

	h := NewHTTPTestHelper(t)
	expectPublished(t, h, event.UserDeletedEvent)

	status, body := h.Do(t, http.MethodDelete, "/api/v1/users/"+aliceID, "")
	expectStatus(t, status, http.StatusNoContent, body)

	status, body = h.Do(t, http.MethodGet, "/api/v1/users/"+aliceID, "")
	expectStatus(t, status, http.StatusNotFound, body)

	if got := h.CountUsers(t); got != 2 {
		t.Errorf("users in db = %d, want 2", got)
	}
}
