package config

// AppConfig holds the application configuration.
type AppConfig struct {
	Env      string
	Port     string
	LogLevel string

	// CORSAllowedOrigins are the browser origins allowed to call the API
	// ("https://app.example.com"). "*" allows any origin without credentials.
	// Empty disables CORS, so browsers block cross-origin calls.
	CORSAllowedOrigins []string
}
