package entity_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"go-ddd-template/internal/domain/entity"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	u := entity.NewUser("Alice", "alice@example.com", "password123")

	if u.ID == uuid.Nil {
		t.Error("NewUser() ID is nil, want a generated UUID")
	}
	if u.Name != "Alice" || u.Email != "alice@example.com" || u.Password != "password123" {
		t.Errorf("NewUser() = %+v, want fields from arguments", u)
	}
	if u.CreatedAt.IsZero() || !u.CreatedAt.Equal(u.UpdatedAt) {
		t.Errorf("NewUser() CreatedAt = %v, UpdatedAt = %v, want equal non-zero times", u.CreatedAt, u.UpdatedAt)
	}
}

func TestUser_Update(t *testing.T) {
	t.Parallel()

	created := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	u := &entity.User{ID: uuid.New(), Name: "Alice", Email: "alice@example.com", CreatedAt: created, UpdatedAt: created}

	u.Update("Alice B", "alice.b@example.com", "newpassword")

	if u.Name != "Alice B" || u.Email != "alice.b@example.com" || u.Password != "newpassword" {
		t.Errorf("Update() = %+v, want fields from arguments", u)
	}
	if !u.CreatedAt.Equal(created) {
		t.Errorf("Update() changed CreatedAt to %v", u.CreatedAt)
	}
	if !u.UpdatedAt.After(created) {
		t.Errorf("Update() UpdatedAt = %v, want after %v", u.UpdatedAt, created)
	}
}
