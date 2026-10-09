//go:build integration

package datastore_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
)

// Users from testdata/fixtures/users.yml.
var (
	fixtureAlice = &entity.User{
		ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:         "Alice",
		Email:        "alice@example.com",
		PasswordHash: "$2a$04$Y5C21koavab7w1PK4ItXPeqiA.c.bs6Mf9OGCwZFUaTYgE8hWmNVC",
		CreatedAt:    time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
	}
	fixtureBob = &entity.User{
		ID:           uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Name:         "Bob",
		Email:        "bob@example.com",
		PasswordHash: "$2a$04$3SoNKU2fO2RQ6NZEi/z7lOAlIDQBGZKwluW6FLoGVE9A5y6CEpgX6",
		CreatedAt:    time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 2, 10, 0, 0, 0, time.UTC),
	}
	fixtureUserCount = 3
)

// equalTime compares timestamps by instant; Postgres returns them without the
// UTC location the fixtures were written in.
var equalTime = cmp.Comparer(func(a, b time.Time) bool { return a.Equal(b) })

func Test_userReader_Get(t *testing.T) {
	t.Parallel()

	type testcase struct {
		id       entity.UserID
		expected *entity.User
		wantErr  error
	}

	tests := map[string]testcase{
		"found": {
			id:       fixtureAlice.ID,
			expected: fixtureAlice,
		},
		"not found": {
			id:      uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			wantErr: model.ErrUserNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := datastore.NewUserReader(readDB).Get(context.Background(), tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("userReader.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.expected, actual, equalTime); diff != "" {
				t.Errorf("userReader.Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userReader_GetByEmail(t *testing.T) {
	t.Parallel()

	type testcase struct {
		email    string
		expected *entity.User
		wantErr  error
	}

	tests := map[string]testcase{
		"found":     {email: "bob@example.com", expected: fixtureBob},
		"not found": {email: "nobody@example.com", wantErr: model.ErrUserNotFound},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := datastore.NewUserReader(readDB).GetByEmail(context.Background(), tt.email)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("userReader.GetByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.expected, actual, equalTime); diff != "" {
				t.Errorf("userReader.GetByEmail() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userReader_List(t *testing.T) {
	t.Parallel()

	type testcase struct {
		limit, offset int
		wantLen       int
	}

	tests := map[string]testcase{
		"first page":        {limit: 2, offset: 0, wantLen: 2},
		"last partial page": {limit: 2, offset: 2, wantLen: 1},
		"past the end":      {limit: 2, offset: 10, wantLen: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			users, total, err := datastore.NewUserReader(readDB).List(context.Background(), tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("userReader.List() error = %v", err)
			}
			if total != fixtureUserCount {
				t.Errorf("userReader.List() total = %d, want %d", total, fixtureUserCount)
			}
			if len(users) != tt.wantLen {
				t.Errorf("userReader.List() returned %d users, want %d", len(users), tt.wantLen)
			}
		})
	}
}

func Test_userReader_Exists(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		email    string
		expected bool
	}{
		"existing email": {email: "alice@example.com", expected: true},
		"unknown email":  {email: "nobody@example.com", expected: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, err := datastore.NewUserReader(readDB).Exists(context.Background(), tt.email)
			if err != nil {
				t.Fatalf("userReader.Exists() error = %v", err)
			}
			if actual != tt.expected {
				t.Errorf("userReader.Exists() = %v, want %v", actual, tt.expected)
			}
		})
	}
}
