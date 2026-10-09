package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"go-ddd-template/internal/domain/entity"
	"go-ddd-template/internal/domain/model"
	"go-ddd-template/internal/usecase/input"
	"go-ddd-template/internal/usecase/output"

	mockgateway "go-ddd-template/internal/domain/gateway/mock"
)

func Test_getUser_Execute(t *testing.T) {
	t.Parallel()

	var (
		userID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
		user   = &entity.User{ID: userID, Name: "Alice", Email: "alice@example.com"}
	)

	type fields struct {
		mockQueries *mockgateway.MockUserQueriesGateway
	}
	type args struct {
		ctx context.Context
		in  *input.GetUser
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		expected    *output.GetUser
		wantErrCode string
	}

	tests := map[string]testcase{
		"successfully get user": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(user, nil).Times(1)
			},
			args:     args{ctx: context.Background(), in: &input.GetUser{ID: userID}},
			expected: &output.GetUser{User: user},
		},
		"user not found": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(nil, model.ErrUserNotFound).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.GetUser{ID: userID}},
			wantErrCode: "USER_NOT_FOUND",
		},
		"Get returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().Get(a.ctx, userID).Return(nil, errors.New("database connection error")).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.GetUser{ID: userID}},
			wantErrCode: "INTERNAL",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			f := &fields{mockQueries: mockgateway.NewMockUserQueriesGateway(ctrl)}
			if tt.prepare != nil {
				tt.prepare(&tt.args, f)
			}

			actual, err := NewGetUser(f.mockQueries).Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("getUser.Execute() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
