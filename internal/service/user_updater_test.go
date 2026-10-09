package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"

	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

func Test_updateUser_Execute(t *testing.T) {
	t.Parallel()

	var (
		userID   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
		existing = &entity.User{ID: userID, Name: "Alice", Email: "alice@example.com"}
		updated  = &entity.User{ID: userID, Name: "Alice B", Email: "alice.b@example.com"}
		in       = &input.UpdateUser{ID: userID, Name: "Alice B", Email: "alice.b@example.com", Password: "newpassword"}
		errDB    = errors.New("database connection error")
	)

	type fields struct {
		mockCommands  *mockgateway.MockUserCommandsGateway
		mockQueries   *mockgateway.MockUserQueriesGateway
		mockHasher    *mockgateway.MockPasswordHasher
		mockPublisher *mockgateway.MockEventPublisher
	}
	type args struct {
		ctx context.Context
		in  *input.UpdateUser
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		expected    *output.UpdateUser
		wantErrCode string
	}

	tests := map[string]testcase{
		"successfully update user and publish event": {
			prepare: func(a *args, f *fields) {
				gomock.InOrder(
					f.mockQueries.EXPECT().Get(a.ctx, userID).Return(existing, nil),
					f.mockHasher.EXPECT().Hash(in.Password).Return(fakeHash, nil),
					f.mockCommands.EXPECT().
						Update(a.ctx, gomock.Any()).
						DoAndReturn(func(_ context.Context, u *entity.User) error {
							want := &entity.User{ID: userID, Name: in.Name, Email: in.Email, PasswordHash: fakeHash}
							if diff := cmp.Diff(want, u, ignoreTimestamps); diff != "" {
								t.Errorf("Update() user mismatch (-want +got):\n%s", diff)
							}
							return nil
						}),
					f.mockQueries.EXPECT().Get(a.ctx, userID).Return(updated, nil),
					f.mockPublisher.EXPECT().
						Publish(a.ctx, gomock.Any()).
						DoAndReturn(func(_ context.Context, evt *event.UserEvent) error {
							if evt.Type != event.UserUpdatedEvent {
								t.Errorf("Publish() event type = %s, want %s", evt.Type, event.UserUpdatedEvent)
							}
							return nil
						}),
				)
			},
			args:     args{ctx: context.Background(), in: in},
			expected: &output.UpdateUser{User: updated},
		},
		"user not found": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(nil, model.ErrUserNotFound).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "USER_NOT_FOUND",
		},
		"Update returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(existing, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return(fakeHash, nil).Times(1)
				f.mockCommands.EXPECT().Update(a.ctx, gomock.Any()).Return(errDB).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
		"Hash returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(existing, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return("", errors.New("hash password: boom")).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INTERNAL",
		},
		"Hash rejects password as too long": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(existing, nil).Times(1)
				f.mockHasher.EXPECT().Hash(in.Password).Return("", model.ErrPasswordTooLong).Times(1)
			},
			args:        args{ctx: context.Background(), in: in},
			wantErrCode: "INVALID_INPUT",
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

			uc := NewUpdateUser(f.mockCommands, f.mockQueries, f.mockHasher, f.mockPublisher)
			actual, err := uc.Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("updateUser.Execute() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
