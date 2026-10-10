package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/longntv/go-ddd-template/internal/config"
)

const (
	corsAllowMethods = "GET, POST, PUT, DELETE, OPTIONS"
	corsAllowHeaders = "Accept, Authorization, Cache-Control, Content-Type, X-Requested-With"
	corsMaxAge       = "600" // seconds a browser may cache a preflight result
)

// CORSConfig lists the browser origins allowed to call the API.
type CORSConfig struct {
	// AllowedOrigins are exact origins such as "https://app.example.com".
	// "*" allows any origin, but then without credentials: browsers reject
	// "Access-Control-Allow-Origin: *" combined with credentials.
	// Empty allows no cross-origin calls.
	AllowedOrigins []string
}

// ProvideCORSConfig reads the allowed origins from the application config and
// rejects values that would never match a browser Origin header, so a typo
// fails at startup instead of silently blocking the frontend.
func ProvideCORSConfig(cfg *config.Config) (CORSConfig, error) {
	origins := cfg.App.CORSAllowedOrigins
	if slices.Contains(origins, "*") {
		if len(origins) > 1 {
			return CORSConfig{}, errors.New(`CORS_ALLOWED_ORIGINS: "*" cannot be combined with explicit origins`)
		}
		return CORSConfig{AllowedOrigins: origins}, nil
	}
	for _, o := range origins {
		if !isOrigin(o) {
			return CORSConfig{}, fmt.Errorf("CORS_ALLOWED_ORIGINS: %q is not an origin like https://app.example.com", o)
		}
	}
	return CORSConfig{AllowedOrigins: origins}, nil
}

// isOrigin reports whether s is a serialized origin as browsers send it:
// lowercase http(s) scheme and host, optional port, nothing else.
func isOrigin(s string) bool {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	return u.String() == s && u.Path == "" && u.User == nil && u.RawQuery == "" && u.Fragment == "" &&
		s == strings.ToLower(s)
}

// CORS adds CORS headers for allowed origins and answers preflight requests.
// Requests from other origins get no CORS headers, so the browser blocks them;
// their preflight gets 403. Same-origin and non-browser requests (no Origin
// header) pass through untouched.
func CORS(cfg CORSConfig) gin.HandlerFunc {
	allowAny := slices.Contains(cfg.AllowedOrigins, "*")

	return func(c *gin.Context) {
		h := c.Writer.Header()
		if !allowAny {
			// The response depends on Origin, including whether it is sent at
			// all; keep shared caches from serving one origin's response to another.
			h.Add("Vary", "Origin")
		}

		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		preflight := c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") != ""

		switch {
		case allowAny:
			h.Set("Access-Control-Allow-Origin", "*")
		case slices.Contains(cfg.AllowedOrigins, origin):
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
		default:
			if preflight {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		if preflight {
			h.Set("Access-Control-Allow-Methods", corsAllowMethods)
			h.Set("Access-Control-Allow-Headers", corsAllowHeaders)
			h.Set("Access-Control-Max-Age", corsMaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
