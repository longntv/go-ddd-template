package testutil

import (
	"os"
	"strconv"
)

// TestEnv holds the environment variables required for testing.
type TestEnv struct {
	DBHost string // the host of the database
	DBPort int    // the port of the database
	DBUser string // the user for the database
	DBPass string // the password for the database
	DBName string // the prefix for test database names
}

// LoadEnv loads the test environment, falling back to the values used by
// docker/docker-compose.yaml.
func LoadEnv() *TestEnv {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		panic("invalid DB_PORT: " + err.Error())
	}

	return &TestEnv{
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: port,
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASSWORD", "postgres"),
		DBName: getEnv("TEST_DB_NAME_PREFIX", "go_ddd_template_test"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
