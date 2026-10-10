package config

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

// Not parallel: the cases use t.Setenv, which modifies process environment.

func Test_getEnvAsInt32(t *testing.T) {
	tests := map[string]struct {
		value    string // "" leaves the variable unset
		expected int32
	}{
		"unset uses default":        {value: "", expected: 7},
		"valid number":              {value: "20", expected: 20},
		"not a number uses default": {value: "abc", expected: 7},
		"overflow uses default":     {value: "2147483648", expected: 7},
		"max int32":                 {value: "2147483647", expected: 2147483647},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const key = "TEST_GET_ENV_AS_INT32"
			t.Setenv(key, tt.value)

			if got := getEnvAsInt32(key, 7); got != tt.expected {
				t.Errorf("getEnvAsInt32(%q) = %d, want %d", tt.value, got, tt.expected)
			}
		})
	}
}

func TestLoad_SQSLimits(t *testing.T) {
	tests := map[string]struct {
		maxMessages, waitTime         string
		wantMaxMessages, wantWaitTime int32
	}{
		"defaults":          {wantMaxMessages: 10, wantWaitTime: 20},
		"within range":      {maxMessages: "5", waitTime: "0", wantMaxMessages: 5, wantWaitTime: 0},
		"above SQS maximum": {maxMessages: "11", waitTime: "60", wantMaxMessages: 10, wantWaitTime: 20},
		"below SQS minimum": {maxMessages: "0", waitTime: "-1", wantMaxMessages: 1, wantWaitTime: 0},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("SQS_MAX_MESSAGES", tt.maxMessages)
			t.Setenv("SQS_WAIT_TIME_SECONDS", tt.waitTime)

			cfg := Load()

			if cfg.SQS.MaxMessages != tt.wantMaxMessages || cfg.SQS.WaitTimeSeconds != tt.wantWaitTime {
				t.Errorf("SQS limits = (%d, %d), want (%d, %d)",
					cfg.SQS.MaxMessages, cfg.SQS.WaitTimeSeconds, tt.wantMaxMessages, tt.wantWaitTime)
			}
		})
	}
}

func Test_getEnvAsList(t *testing.T) {
	tests := map[string]struct {
		value    string
		expected []string
	}{
		"blank is nil":            {value: "", expected: nil},
		"single origin":           {value: "https://app.example.com", expected: []string{"https://app.example.com"}},
		"trims and drops empties": {value: " https://a.example.com, ,https://b.example.com ,", expected: []string{"https://a.example.com", "https://b.example.com"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const key = "TEST_GET_ENV_AS_LIST"
			t.Setenv(key, tt.value)

			if diff := cmp.Diff(tt.expected, getEnvAsList(key)); diff != "" {
				t.Errorf("getEnvAsList(%q) mismatch (-want +got):\n%s", tt.value, diff)
			}
		})
	}
}

func Test_getEnvAsDuration(t *testing.T) {
	tests := map[string]struct {
		value    string // "" leaves the variable unset
		expected time.Duration
	}{
		"unset uses default":       {value: "", expected: time.Second},
		"valid duration":           {value: "250ms", expected: 250 * time.Millisecond},
		"bare number uses default": {value: "5", expected: time.Second},
		"invalid uses default":     {value: "soon", expected: time.Second},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			const key = "TEST_GET_ENV_AS_DURATION"
			t.Setenv(key, tt.value)

			if got := getEnvAsDuration(key, time.Second); got != tt.expected {
				t.Errorf("getEnvAsDuration(%q) = %s, want %s", tt.value, got, tt.expected)
			}
		})
	}
}
