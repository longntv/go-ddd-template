//go:build integration

package http_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"go-ddd-template/internal/testutil"
	"go-ddd-template/test/integration/registry"

	mockgateway "go-ddd-template/internal/domain/gateway/mock"
)

// HTTPTestHelper runs requests against the real router, services and
// datastore backed by a fresh database. Only the event publisher is mocked.
type HTTPTestHelper struct {
	router        *gin.Engine
	gormDB        *gorm.DB
	mockPublisher *mockgateway.MockEventPublisher
}

// NewHTTPTestHelper creates a helper with its own cloned database.
func NewHTTPTestHelper(t *testing.T) *HTTPTestHelper {
	t.Helper()

	gormDB, _ := testutil.InitDB(t)
	mockPublisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))

	router, err := registry.InitializeServer(gormDB, mockPublisher)
	if err != nil {
		t.Fatalf("initialize server: %v", err)
	}

	return &HTTPTestHelper{router: router, gormDB: gormDB, mockPublisher: mockPublisher}
}

// Do sends a JSON request and returns the status code and decoded body
// (nil for an empty body).
func (h *HTTPTestHelper) Do(t *testing.T, method, path, body string) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.router.ServeHTTP(rec, req)

	if rec.Body.Len() == 0 {
		return rec.Code, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode %s %s response %q: %v", method, path, rec.Body.String(), err)
	}
	return rec.Code, decoded
}

// CountUsers returns the number of rows in the users table.
func (h *HTTPTestHelper) CountUsers(t *testing.T) int64 {
	t.Helper()

	var count int64
	if err := h.gormDB.Table("users").Count(&count).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

// expectStatus fails the test when got differs from want.
func expectStatus(t *testing.T, got, want int, body map[string]any) {
	t.Helper()

	if got != want {
		t.Fatalf("status = %d, want %d (body: %v)", got, want, body)
	}
}
