package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/usecase/input"

	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

func Test_deleteUser_Execute(t *testing.T) {
	t.Parallel()

	var (
		userID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
		user   = &entity.User{ID: userID, Name: "Alice", Email: "alice@example.com"}
	)

	type fields struct {
		mockCommands  *mockgateway.MockUserCommandsGateway
		mockQueries   *mockgateway.MockUserQueriesGateway
		mockPublisher *mockgateway.MockEventPublisher
	}
	type args struct {
		ctx context.Context
		in  *input.DeleteUser
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		wantErrCode string
		// wantPublishErrLogged is the event type whose failed Publish must be logged; "" = no log.
		wantPublishErrLogged string
	}

	tests := map[string]testcase{
		"successfully delete user and publish event": {
			prepare: func(a *args, f *fields) {
				gomock.InOrder(
					f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil),
					f.mockCommands.EXPECT().Delete(a.ctx, userID).Return(nil),
					f.mockPublisher.EXPECT().
						Publish(a.ctx, gomock.Any()).
						DoAndReturn(func(_ context.Context, evt *event.UserEvent) error {
							if evt.Type != event.UserDeletedEvent {
								t.Errorf("Publish() event type = %s, want %s", evt.Type, event.UserDeletedEvent)
							}
							return nil
						}),
				)
			},
			args: args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
		},
		"publish failure does not fail the request": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil).Times(1)
				f.mockCommands.EXPECT().Delete(a.ctx, userID).Return(nil).Times(1)
				f.mockPublisher.EXPECT().Publish(a.ctx, gomock.Any()).Return(errors.New("sns unavailable")).Times(1)
			},
			args:                 args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantPublishErrLogged: event.UserDeletedEvent,
		},
		"user not found": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(nil, model.ErrUserNotFound).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantErrCode: "USER_NOT_FOUND",
		},
		"Delete returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil).Times(1)
				f.mockCommands.EXPECT().Delete(a.ctx, userID).Return(errors.New("database connection error")).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantErrCode: "INTERNAL",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logCore, logs := observer.New(zap.ErrorLevel)
			f := &fields{
				mockCommands:  mockgateway.NewMockUserCommandsGateway(ctrl),
				mockQueries:   mockgateway.NewMockUserQueriesGateway(ctrl),
				mockPublisher: mockgateway.NewMockEventPublisher(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&tt.args, f)
			}

			uc := NewDeleteUser(f.mockCommands, f.mockQueries, f.mockPublisher, zap.New(logCore))
			err := uc.Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			assertPublishErrLogged(t, logs, tt.wantPublishErrLogged)
		})
	}
}
