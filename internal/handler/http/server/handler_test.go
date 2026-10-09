package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/model"
	"github.com/longntv/go-ddd-template/internal/handler/http/server"
	"github.com/longntv/go-ddd-template/internal/usecase/input"
	"github.com/longntv/go-ddd-template/internal/usecase/output"

	mockusecase "github.com/longntv/go-ddd-template/internal/usecase/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockUseCases holds one mock per use case injected into the handler.
type mockUseCases struct {
	create *mockusecase.MockCreateUser
	get    *mockusecase.MockGetUser
	list   *mockusecase.MockListUsers
	update *mockusecase.MockUpdateUser
	delete *mockusecase.MockDeleteUser
}

// newTestRouter builds a gin engine with the user routes backed by mocks.
func newTestRouter(t *testing.T) (*gin.Engine, *mockUseCases) {
	t.Helper()

	ctrl := gomock.NewController(t)
	m := &mockUseCases{
		create: mockusecase.NewMockCreateUser(ctrl),
		get:    mockusecase.NewMockGetUser(ctrl),
		list:   mockusecase.NewMockListUsers(ctrl),
		update: mockusecase.NewMockUpdateUser(ctrl),
		delete: mockusecase.NewMockDeleteUser(ctrl),
	}
	h := server.NewHandler(m.create, m.get, m.list, m.update, m.delete)

	r := gin.New()
	users := r.Group("/api/v1/users")
	users.POST("", h.Create)
	users.GET("", h.List)
	users.GET("/:id", h.Get)
	users.PUT("/:id", h.Update)
	users.DELETE("/:id", h.Delete)

	return r, m
}

func TestHandler(t *testing.T) {
	t.Parallel()

	var (
		userID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
		ts     = time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
		user   = &entity.User{ID: userID, Name: "Alice", Email: "alice@example.com", CreatedAt: ts, UpdatedAt: ts}
	)

	type request struct {
		method string
		path   string
		body   string
	}
	type expected struct {
		status int
		body   map[string]any // nil skips the body check
	}
	type testcase struct {
		prepare  func(m *mockUseCases)
		request  request
		expected expected
	}

	tests := map[string]testcase{
		"POST /users creates user": {
			prepare: func(m *mockUseCases) {
				m.create.EXPECT().
					Execute(gomock.Any(), &input.CreateUser{Name: "Alice", Email: "alice@example.com", Password: "password123"}).
					Return(&output.CreateUser{User: user}, nil).
					Times(1)
			},
			request: request{http.MethodPost, "/api/v1/users", `{"name":"Alice","email":"alice@example.com","password":"password123"}`},
			expected: expected{
				status: http.StatusCreated,
				body: map[string]any{
					"id": userID.String(), "name": "Alice", "email": "alice@example.com",
					"created_at": "2025-01-01T10:00:00Z", "updated_at": "2025-01-01T10:00:00Z",
				},
			},
		},
		"POST /users rejects invalid email before reaching the use case": {
			request:  request{http.MethodPost, "/api/v1/users", `{"name":"Alice","email":"not-an-email","password":"password123"}`},
			expected: expected{status: http.StatusBadRequest},
		},
		"POST /users maps USER_EXISTS to 409": {
			prepare: func(m *mockUseCases) {
				m.create.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, model.NewDomainError("USER_EXISTS", "user with this email already exists", model.ErrUserAlreadyExists)).
					Times(1)
			},
			request:  request{http.MethodPost, "/api/v1/users", `{"name":"Alice","email":"alice@example.com","password":"password123"}`},
			expected: expected{status: http.StatusConflict},
		},
		"GET /users/:id returns user": {
			prepare: func(m *mockUseCases) {
				m.get.EXPECT().Execute(gomock.Any(), &input.GetUser{ID: userID}).Return(&output.GetUser{User: user}, nil).Times(1)
			},
			request:  request{http.MethodGet, "/api/v1/users/" + userID.String(), ""},
			expected: expected{status: http.StatusOK},
		},
		"GET /users/:id maps USER_NOT_FOUND to 404": {
			prepare: func(m *mockUseCases) {
				m.get.EXPECT().
					Execute(gomock.Any(), gomock.Any()).
					Return(nil, model.NewDomainError("USER_NOT_FOUND", "user not found", model.ErrUserNotFound)).
					Times(1)
			},
			request:  request{http.MethodGet, "/api/v1/users/" + userID.String(), ""},
			expected: expected{status: http.StatusNotFound, body: map[string]any{"error": "USER_NOT_FOUND: user not found"}},
		},
		"GET /users/:id rejects malformed id": {
			request:  request{http.MethodGet, "/api/v1/users/not-a-uuid", ""},
			expected: expected{status: http.StatusBadRequest, body: map[string]any{"error": "invalid user ID"}},
		},
		"GET /users/:id maps unexpected error to 500": {
			prepare: func(m *mockUseCases) {
				m.get.EXPECT().Execute(gomock.Any(), gomock.Any()).Return(nil, model.NewDomainError("INTERNAL", "failed to get user", nil)).Times(1)
			},
			request:  request{http.MethodGet, "/api/v1/users/" + userID.String(), ""},
			expected: expected{status: http.StatusInternalServerError, body: map[string]any{"error": "internal server error"}},
		},
		"GET /users clamps out-of-range limit to 10": {
			prepare: func(m *mockUseCases) {
				m.list.EXPECT().
					Execute(gomock.Any(), &input.ListUsers{Page: 2, Limit: 10}).
					Return(&output.ListUsers{Users: []*entity.User{user}, TotalCount: 11, Page: 2, Limit: 10}, nil).
					Times(1)
			},
			request:  request{http.MethodGet, "/api/v1/users?page=2&limit=500", ""},
			expected: expected{status: http.StatusOK},
		},
		"PUT /users/:id updates user": {
			prepare: func(m *mockUseCases) {
				m.update.EXPECT().
					Execute(gomock.Any(), &input.UpdateUser{ID: userID, Name: "Alice", Email: "alice@example.com", Password: "password123"}).
					Return(&output.UpdateUser{User: user}, nil).
					Times(1)
			},
			request:  request{http.MethodPut, "/api/v1/users/" + userID.String(), `{"name":"Alice","email":"alice@example.com","password":"password123"}`},
			expected: expected{status: http.StatusOK},
		},
		"DELETE /users/:id returns 204": {
			prepare: func(m *mockUseCases) {
				m.delete.EXPECT().Execute(gomock.Any(), &input.DeleteUser{ID: userID}).Return(nil).Times(1)
			},
			request:  request{http.MethodDelete, "/api/v1/users/" + userID.String(), ""},
			expected: expected{status: http.StatusNoContent},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r, m := newTestRouter(t)
			if tt.prepare != nil {
				tt.prepare(m)
			}

			req := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(tt.request.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.expected.status {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.expected.status, rec.Body.String())
			}
			if tt.expected.body == nil {
				return
			}

			var actual map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &actual); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if diff := cmp.Diff(tt.expected.body, actual); diff != "" {
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
