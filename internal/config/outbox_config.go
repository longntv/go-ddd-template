package config

import "time"

// OutboxConfig holds the outbox relay configuration.
type OutboxConfig struct {
	// PollInterval is how long the relay waits between checks for pending
	// events once it has published everything that was due.
	PollInterval time.Duration

	// BatchSize is the most events the relay publishes per transaction.
	BatchSize int
}
