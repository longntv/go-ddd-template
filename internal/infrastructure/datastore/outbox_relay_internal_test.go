package datastore

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
)

func Test_outboxBackoff(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		attempts int
		expected time.Duration
	}{
		"first failure waits 1s":      {attempts: 1, expected: time.Second},
		"doubles per failure":         {attempts: 4, expected: 8 * time.Second},
		"last step below the cap":     {attempts: 9, expected: 256 * time.Second},
		"capped at the maximum":       {attempts: 10, expected: outboxMaxBackoff},
		"huge counts do not overflow": {attempts: 1000, expected: outboxMaxBackoff},
		"zero is treated as first":    {attempts: 0, expected: time.Second},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := outboxBackoff(tt.attempts); got != tt.expected {
				t.Errorf("outboxBackoff(%d) = %s, want %s", tt.attempts, got, tt.expected)
			}
		})
	}
}

func Test_truncate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		s        string
		n        int
		expected string
	}{
		"short string unchanged":              {s: "sns down", n: 10, expected: "sns down"},
		"cut at n bytes":                      {s: "abcdef", n: 3, expected: "abc"},
		"never splits a multi-byte character": {s: "ééé", n: 3, expected: "é"}, // é is 2 bytes
		"long error bounded":                  {s: strings.Repeat("x", 2000), n: outboxMaxErrorLen, expected: strings.Repeat("x", outboxMaxErrorLen)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := truncate(tt.s, tt.n); got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.expected)
			}
		})
	}
}

func TestNewOutboxRelay(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		cfg     appconfig.OutboxConfig
		wantErr bool
	}{
		"valid config":           {cfg: appconfig.OutboxConfig{PollInterval: time.Second, BatchSize: 100}},
		"zero poll interval":     {cfg: appconfig.OutboxConfig{BatchSize: 100}, wantErr: true},
		"negative poll interval": {cfg: appconfig.OutboxConfig{PollInterval: -time.Second, BatchSize: 100}, wantErr: true},
		"zero batch size":        {cfg: appconfig.OutboxConfig{PollInterval: time.Second}, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := NewOutboxRelay(nil, nil, &appconfig.Config{Outbox: tt.cfg}, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOutboxRelay() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOutboxEventEntity_ToDomain(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		data     string
		expected any
	}{
		"returns the stored JSON as is":     {data: `{"id": "1"}`, expected: json.RawMessage(`{"id": "1"}`)},
		"event saved without data has none": {data: "null", expected: nil},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := (&OutboxEventEntity{Data: []byte(tt.data)}).ToDomain().Data
			if diff := cmp.Diff(tt.expected, got); diff != "" {
				t.Errorf("ToDomain().Data mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
