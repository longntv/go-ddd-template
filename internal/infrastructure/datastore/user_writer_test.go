//go:build integration

package datastore_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-ddd-template/internal/domain/entity"
	"go-ddd-template/internal/domain/model"
	"go-ddd-template/internal/infrastructure/datastore"
	"go-ddd-template/internal/testutil"
)

// Writer tests get a fresh database per test case (testutil.InitDB) so writes
// never leak into other cases, and can still run in parallel.

func Test_userWriter_Create(t *testing.T) {
	t.Parallel()

	type testcase struct {
		user    *entity.User
		wantErr bool
	}

	tests := map[string]testcase{
		"creates a new user": {
			user: &entity.User{ID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), Name: "Dave", Email: "dave@example.com", Password: "password-dave"},
		},
		"rejects a duplicate email": {
			user:    &entity.User{ID: uuid.MustParse("55555555-5555-5555-5555-555555555555"), Name: "Alice 2", Email: "alice@example.com", Password: "password"},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, _ := testutil.InitDB(t)
			ctx := context.Background()

			err := datastore.NewUserWriter(db).Create(ctx, tt.user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("userWriter.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			actual, err := datastore.NewUserReader(db).Get(ctx, tt.user.ID)
			if err != nil {
				t.Fatalf("read back created user: %v", err)
			}
			if diff := cmp.Diff(tt.user, actual, cmpopts.IgnoreFields(entity.User{}, "CreatedAt", "UpdatedAt")); diff != "" {
				t.Errorf("created user mismatch (-want +got):\n%s", diff)
			}
			if actual.CreatedAt.IsZero() {
				t.Error("created user has zero CreatedAt, want database default")
			}
		})
	}
}

func Test_userWriter_Update(t *testing.T) {
	t.Parallel()

	type testcase struct {
		user    *entity.User
		wantErr error
	}

	tests := map[string]testcase{
		"updates an existing user": {
			user: &entity.User{ID: fixtureAlice.ID, Name: "Alice B", Email: "alice.b@example.com", Password: "new-password"},
		},
		"returns not found for unknown id": {
			user:    &entity.User{ID: uuid.MustParse("99999999-9999-9999-9999-999999999999"), Name: "X", Email: "x@example.com", Password: "password"},
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, _ := testutil.InitDB(t)
			ctx := context.Background()

			err := datastore.NewUserWriter(db).Update(ctx, tt.user)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("userWriter.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			actual, err := datastore.NewUserReader(db).Get(ctx, tt.user.ID)
			if err != nil {
				t.Fatalf("read back updated user: %v", err)
			}
			if diff := cmp.Diff(tt.user, actual, cmpopts.IgnoreFields(entity.User{}, "CreatedAt", "UpdatedAt")); diff != "" {
				t.Errorf("updated user mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userWriter_Delete(t *testing.T) {
	t.Parallel()

	type testcase struct {
		id      entity.UserID
		wantErr error
	}

	tests := map[string]testcase{
		"deletes an existing user": {
			id: fixtureBob.ID,
		},
		"returns not found for unknown id": {
			id:      uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			wantErr: gorm.ErrRecordNotFound,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, _ := testutil.InitDB(t)
			ctx := context.Background()

			err := datastore.NewUserWriter(db).Delete(ctx, tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("userWriter.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			if _, err := datastore.NewUserReader(db).Get(ctx, tt.id); !errors.Is(err, model.ErrUserNotFound) {
				t.Errorf("Get() after delete error = %v, want %v", err, model.ErrUserNotFound)
			}
		})
	}
}
