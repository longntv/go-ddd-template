package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"

	"github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/handler/http/middleware"
)

func TestCORS(t *testing.T) {
	t.Parallel()

	const (
		appOrigin   = "https://app.example.com"
		otherOrigin = "https://evil.example.com"
	)

	type request struct {
		method       string
		routeMethod  string // method the route is registered for; defaults to method
		origin       string
		preflightFor string // Access-Control-Request-Method; makes OPTIONS a preflight
	}
	type expected struct {
		Status       int
		AllowOrigin  string
		Credentials  string
		AllowMethods string
		AllowHeaders string
		Vary         string
		ReachedRoute bool
	}
	type testcase struct {
		allowed  []string
		request  request
		expected expected
	}

	tests := map[string]testcase{
		"no Origin header passes through without CORS headers": {
			allowed:  []string{appOrigin},
			request:  request{method: http.MethodGet},
			expected: expected{Status: http.StatusOK, Vary: "Origin", ReachedRoute: true},
		},
		"allowed origin is echoed with credentials": {
			allowed: []string{appOrigin},
			request: request{method: http.MethodGet, origin: appOrigin},
			expected: expected{
				Status: http.StatusOK, AllowOrigin: appOrigin, Credentials: "true", Vary: "Origin", ReachedRoute: true,
			},
		},
		"other origin gets no CORS headers": {
			allowed:  []string{appOrigin},
			request:  request{method: http.MethodGet, origin: otherOrigin},
			expected: expected{Status: http.StatusOK, Vary: "Origin", ReachedRoute: true},
		},
		"empty allow-list allows no origin": {
			allowed:  nil,
			request:  request{method: http.MethodGet, origin: appOrigin},
			expected: expected{Status: http.StatusOK, Vary: "Origin", ReachedRoute: true},
		},
		"wildcard allows any origin without credentials": {
			allowed:  []string{"*"},
			request:  request{method: http.MethodGet, origin: otherOrigin},
			expected: expected{Status: http.StatusOK, AllowOrigin: "*", ReachedRoute: true},
		},
		"preflight from allowed origin is answered with 204": {
			allowed: []string{appOrigin},
			request: request{method: http.MethodOptions, origin: appOrigin, preflightFor: http.MethodPut},
			expected: expected{
				Status: http.StatusNoContent, AllowOrigin: appOrigin, Credentials: "true",
				AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
				AllowHeaders: "Accept, Authorization, Cache-Control, Content-Type, X-Requested-With", Vary: "Origin",
			},
		},
		"preflight to a route without an OPTIONS handler is answered with 204": {
			allowed: []string{appOrigin},
			request: request{method: http.MethodOptions, routeMethod: http.MethodPut, origin: appOrigin, preflightFor: http.MethodPut},
			expected: expected{
				Status: http.StatusNoContent, AllowOrigin: appOrigin, Credentials: "true",
				AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
				AllowHeaders: "Accept, Authorization, Cache-Control, Content-Type, X-Requested-With", Vary: "Origin",
			},
		},
		"wildcard preflight is answered without credentials": {
			allowed: []string{"*"},
			request: request{method: http.MethodOptions, origin: otherOrigin, preflightFor: http.MethodPost},
			expected: expected{
				Status: http.StatusNoContent, AllowOrigin: "*",
				AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
				AllowHeaders: "Accept, Authorization, Cache-Control, Content-Type, X-Requested-With",
			},
		},
		"origin match is exact": {
			allowed:  []string{appOrigin},
			request:  request{method: http.MethodGet, origin: "https://app.example.com.evil.io"},
			expected: expected{Status: http.StatusOK, Vary: "Origin", ReachedRoute: true},
		},
		"preflight from other origin is rejected with 403": {
			allowed:  []string{appOrigin},
			request:  request{method: http.MethodOptions, origin: otherOrigin, preflightFor: http.MethodDelete},
			expected: expected{Status: http.StatusForbidden, Vary: "Origin"},
		},
		"OPTIONS without preflight header reaches the route": {
			allowed: []string{appOrigin},
			request: request{method: http.MethodOptions, origin: appOrigin},
			expected: expected{
				Status: http.StatusOK, AllowOrigin: appOrigin, Credentials: "true", Vary: "Origin", ReachedRoute: true,
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			reached := false
			r := gin.New()
			r.Use(middleware.CORS(middleware.CORSConfig{AllowedOrigins: tt.allowed}))
			routeMethod := tt.request.routeMethod
			if routeMethod == "" {
				routeMethod = tt.request.method
			}
			r.Handle(routeMethod, "/users", func(c *gin.Context) {
				reached = true
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.request.method, "/users", http.NoBody)
			if tt.request.origin != "" {
				req.Header.Set("Origin", tt.request.origin)
			}
			if tt.request.preflightFor != "" {
				req.Header.Set("Access-Control-Request-Method", tt.request.preflightFor)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			actual := expected{
				Status:       w.Code,
				AllowOrigin:  w.Header().Get("Access-Control-Allow-Origin"),
				Credentials:  w.Header().Get("Access-Control-Allow-Credentials"),
				AllowMethods: w.Header().Get("Access-Control-Allow-Methods"),
				AllowHeaders: w.Header().Get("Access-Control-Allow-Headers"),
				Vary:         w.Header().Get("Vary"),
				ReachedRoute: reached,
			}
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("CORS() response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestProvideCORSConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		origins []string
		wantErr bool
	}{
		"none":                  {origins: nil},
		"wildcard":              {origins: []string{"*"}},
		"origins with ports":    {origins: []string{"https://app.example.com", "http://localhost:3000"}},
		"wildcard with origins": {origins: []string{"https://app.example.com", "*"}, wantErr: true},
		"trailing slash":        {origins: []string{"https://app.example.com/"}, wantErr: true},
		"path":                  {origins: []string{"https://app.example.com/app"}, wantErr: true},
		"uppercase":             {origins: []string{"https://App.example.com"}, wantErr: true},
		"no scheme":             {origins: []string{"app.example.com"}, wantErr: true},
		"null origin":           {origins: []string{"null"}, wantErr: true},
		"non-http scheme":       {origins: []string{"file://app"}, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := &config.Config{App: config.AppConfig{CORSAllowedOrigins: tt.origins}}
			got, err := middleware.ProvideCORSConfig(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ProvideCORSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if diff := cmp.Diff(tt.origins, got.AllowedOrigins); diff != "" {
					t.Errorf("AllowedOrigins mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
