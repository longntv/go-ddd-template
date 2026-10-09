package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"

	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

func Test_listUsers_Execute(t *testing.T) {
	t.Parallel()

	users := []*entity.User{
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Name: "Alice"},
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Name: "Bob"},
	}

	type fields struct {
		mockQueries *mockgateway.MockUserQueriesGateway
	}
	type args struct {
		ctx context.Context
		in  *input.ListUsers
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		expected    *output.ListUsers
		wantErrCode string
	}

	tests := map[string]testcase{
		"first page uses offset 0": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().List(a.ctx, 10, 0).Return(users, 2, nil).Times(1)
			},
			args:     args{ctx: context.Background(), in: &input.ListUsers{Page: 1, Limit: 10}},
			expected: &output.ListUsers{Users: users, TotalCount: 2, Page: 1, Limit: 10},
		},
		"third page computes offset from page and limit": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().List(a.ctx, 5, 10).Return([]*entity.User{}, 2, nil).Times(1)
			},
			args:     args{ctx: context.Background(), in: &input.ListUsers{Page: 3, Limit: 5}},
			expected: &output.ListUsers{Users: []*entity.User{}, TotalCount: 2, Page: 3, Limit: 5},
		},
		"List returns error": {
			prepare: func(a *args, f *fields) {
				f.mockQueries.EXPECT().List(a.ctx, 10, 0).Return(nil, 0, errors.New("database connection error")).Times(1)
			},
			args:        args{ctx: context.Background(), in: &input.ListUsers{Page: 1, Limit: 10}},
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

			actual, err := NewListUsers(f.mockQueries).Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("listUsers.Execute() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
