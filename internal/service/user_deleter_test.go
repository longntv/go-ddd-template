package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

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
		mockCommands *mockgateway.MockUserCommandsGateway
		mockQueries  *mockgateway.MockUserQueriesGateway
		mockTx       *mockgateway.MockTransactor
		mockOutbox   *mockgateway.MockEventOutbox
	}
	type args struct {
		ctx context.Context
		in  *input.DeleteUser
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		wantErrCode string
	}

	tests := map[string]testcase{
		"successfully delete user and add event to outbox in one transaction": {
			prepare: func(a *args, f *fields) {
				gomock.InOrder(
					f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil),
					expectTx(f.mockTx, a.ctx),
					f.mockCommands.EXPECT().Delete(txCtx, userID).Return(nil),
					expectOutboxAdd(t, f.mockOutbox, event.UserDeletedEvent, user),
				)
			},
			args: args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
		},
		"outbox Add returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil).Times(1)
				expectTx(f.mockTx, a.ctx)
				f.mockCommands.EXPECT().Delete(txCtx, userID).Return(nil).Times(1)
				f.mockOutbox.EXPECT().Add(txCtx, gomock.Any()).Return(errors.New("database connection error")).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantErrCode: "INTERNAL",
		},
		"commit fails": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil).Times(1)
				expectTxCommitFails(f.mockTx, a.ctx, errors.New("database connection error"))
				f.mockCommands.EXPECT().Delete(txCtx, userID).Return(nil).Times(1)
				f.mockOutbox.EXPECT().Add(txCtx, gomock.Any()).Return(nil).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantErrCode: "INTERNAL",
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
				expectTx(f.mockTx, a.ctx)
				f.mockCommands.EXPECT().Delete(txCtx, userID).Return(errors.New("database connection error")).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.DeleteUser{ID: userID}},
			wantErrCode: "INTERNAL",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			f := &fields{
				mockCommands: mockgateway.NewMockUserCommandsGateway(ctrl),
				mockQueries:  mockgateway.NewMockUserQueriesGateway(ctrl),
				mockTx:       mockgateway.NewMockTransactor(ctrl),
				mockOutbox:   mockgateway.NewMockEventOutbox(ctrl),
			}
			if tt.prepare != nil {
				tt.prepare(&tt.args, f)
			}

			uc := NewDeleteUser(f.mockCommands, f.mockQueries, f.mockTx, f.mockOutbox)
			err := uc.Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
		})
	}
}
