package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"

	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

const fakeHash = "$2a$10$hashed-password"

func Test_createUser_Execute(t *testing.T) {
	t.Parallel()

	var (
		in = &input.CreateUser{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		}
		createdUser = &entity.User{
			ID:           uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			Name:         "Alice",
			Email:        "alice@example.com",
			PasswordHash: fakeHash,
			CreatedAt:    time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		}
		errDB = errors.New("database connection error")
	)

	type fields struct {
		mockCommands  *mockgateway.MockUserCommandsGateway
		mockQueries   *mockgateway.MockUserQueriesGateway
		mockHasher    *mockgateway.MockPasswordHasher
		mockPublisher *mockgateway.MockEventPublisher
	}
	type args struct {
		ctx context.Context
		in  *input.CreateUser
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		expected    *output.CreateUser
		wantErrCode string
	}

	tests := map[string]testcase{
		"successfully create user and publish event": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().
					Exists(a.ctx, in.Email).
					Return(false, nil).
					Times(1)

				f.mockHasher.EXPECT().
					Hash(in.Password).
					Return(fakeHash, nil).
					Times(1)

				f.mockCommands.EXPECT().
					Create(a.ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, u *entity.User) error {
						if u.Name != in.Name || u.Email != in.Email {
							t.Errorf("Create() got user %+v, want name/email from input", u)
						}
						if u.PasswordHash != fakeHash {
							t.Errorf("Create() PasswordHash = %q, want the hasher output, never the plain text", u.PasswordHash)
						}
						return nil
					}).
					Times(1)

				f.mockQueries.EXPECT().
					GetByEmail(a.ctx, in.Email).
					Return(createdUser, nil).
					Times(1)

				f.mockPublisher.EXPECT().
					Publish(a.ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, evt *event.UserEvent) error {
						if evt.Type != event.UserCreatedEvent {
							t.Errorf("Publish() event type = %s, want %s", evt.Type, event.UserCreatedEvent)
						}
						if evt.Subject != createdUser.ID.String() {
							t.Errorf("Publish() subject = %s, want %s", evt.Subject, createdUser.ID)
						}
						return nil
					}).
					Times(1)
			},
			args:     args{ctx: context.Background(), in: in},
			expected: &output.CreateUser{User: createdUser},
		},
		"publish failure does not fail the request": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return(fakeHash, nil).Times(1)
				f.mockCommands.EXPECT().Create(a.ctx, gomock.Any()).Return(nil).Times(1)
				f.mockQueries.EXPECT().GetByEmail(a.ctx, in.Email).Return(createdUser, nil).Times(1)
				f.mockPublisher.EXPECT().Publish(a.ctx, gomock.Any()).Return(errors.New("sns unavailable")).Times(1)
			},
			args:     args{ctx: context.Background(), in: in},
			expected: &output.CreateUser{User: createdUser},
		},
		"email already exists": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(true, nil).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "USER_EXISTS",
		},
		"Exists returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, errDB).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
		"Hash returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return("", errors.New("hash password: boom")).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
		"Hash rejects password as too long": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return("", model.ErrPasswordTooLong).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INVALID_INPUT",
		},
		"Create returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return(fakeHash, nil).Times(1)
				f.mockCommands.EXPECT().Create(a.ctx, gomock.Any()).Return(errDB).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
		"GetByEmail after create returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Exists(a.ctx, in.Email).Return(false, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return(fakeHash, nil).Times(1)
				f.mockCommands.EXPECT().Create(a.ctx, gomock.Any()).Return(nil).Times(1)
				f.mockQueries.EXPECT().GetByEmail(a.ctx, in.Email).Return(nil, errDB).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			f := &fields{
				mockCommands:  mockgateway.NewMockUserCommandsGateway(ctrl),
				mockQueries:   mockgateway.NewMockUserQueriesGateway(ctrl),
				mockHasher:    mockgateway.NewMockPasswordHasher(ctrl),
				mockPublisher: mockgateway.NewMockEventPublisher(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&tt.args, f)
			}

			uc := NewCreateUser(f.mockCommands, f.mockQueries, f.mockHasher, f.mockPublisher)
			actual, err := uc.Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("createUser.Execute() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// assertDomainErrorCode checks that err is a *model.DomainError with the
// given code, or nil when wantCode is empty.
func assertDomainErrorCode(t *testing.T, err error, wantCode string) {
	t.Helper()

	if wantCode == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}

	var domainErr *model.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("error = %v, want *model.DomainError with code %s", err, wantCode)
	}
	if domainErr.Code != wantCode {
		t.Errorf("error code = %s, want %s", domainErr.Code, wantCode)
	}
}

// ignoreTimestamps ignores fields set from the wall clock.
var ignoreTimestamps = cmpopts.IgnoreFields(entity.User{}, "CreatedAt", "UpdatedAt")
