package datastore

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/domain/gateway"
)

const (
	// outboxMaxBackoff caps the wait before retrying a failed event.
	outboxMaxBackoff = 5 * time.Minute

	// outboxMaxErrorLen bounds the publish error stored in last_error.
	outboxMaxErrorLen = 1000

	// outboxPublishTimeout bounds one Publish call. A batch holds a database
	// transaction while it publishes, so a hanging broker must not keep it
	// open.
	outboxPublishTimeout = 10 * time.Second

	// outboxRelayLockKey is the Postgres advisory lock that lets only one
	// relay work at a time ("outbox" in ASCII).
	outboxRelayLockKey int64 = 0x6f7574626f78
)

// OutboxRelay publishes the events that EventOutbox saved and marks them
// published.
//
// Delivery is at least once: if the process dies after Publish but before
// the batch commits, those events are published again, so consumers must
// deduplicate by event ID.
//
// Every server replica runs a relay, but only one works at a time: a batch
// first takes a transaction-scoped advisory lock and does nothing if another
// relay holds it. Events are published in outbox order (seq), so the events
// of one aggregate, whose writes are serialized by its row lock, arrive in
// the order they happened. A failed event is retried with exponential
// backoff while later events go ahead, so that order holds only while
// publishing succeeds.
type OutboxRelay struct {
	db        *gorm.DB
	publisher gateway.EventPublisher
	logger    *zap.Logger

	pollInterval time.Duration
	batchSize    int
}

// NewOutboxRelay creates a new OutboxRelay. It fails when the poll interval
// or batch size is not positive.
func NewOutboxRelay(db *gorm.DB, publisher gateway.EventPublisher, cfg *appconfig.Config, zapLogger *zap.Logger) (*OutboxRelay, error) {
	if cfg.Outbox.PollInterval <= 0 {
		return nil, fmt.Errorf("outbox poll interval must be positive, got %s", cfg.Outbox.PollInterval)
	}
	if cfg.Outbox.BatchSize <= 0 {
		return nil, fmt.Errorf("outbox batch size must be positive, got %d", cfg.Outbox.BatchSize)
	}
	if db != nil {
		// The relay polls every PollInterval; log only its slow or failed SQL.
		db = db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Warn)})
	}
	return &OutboxRelay{
		db:           db,
		publisher:    publisher,
		logger:       zapLogger,
		pollInterval: cfg.Outbox.PollInterval,
		batchSize:    cfg.Outbox.BatchSize,
	}, nil
}

// Run publishes due events until ctx is done, then returns ctx's error. A
// failed batch is logged and retried on the next poll.
func (r *OutboxRelay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.pollInterval)
	defer ticker.Stop()

	for {
		// Drain: a full batch means more events may be due.
		for ctx.Err() == nil {
			n, err := r.RelayBatch(ctx)
			if err != nil {
				r.logger.Error("failed to relay outbox events", zap.Error(err))
				break
			}
			if n < r.batchSize {
				break
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// RelayBatch publishes up to one batch of due events in one transaction and
// returns how many it handled. Published events are marked published;
// failed ones are scheduled for a retry. It returns 0 when another relay is
// working.
//
// When ctx is cancelled (shutdown), it stops before the next event and
// commits what it has done, so published events are not published again.
// The error is non-nil only when the database fails; nothing in the batch is
// then recorded.
func (r *OutboxRelay) RelayBatch(ctx context.Context) (int, error) {
	if ctx.Err() != nil {
		return 0, nil
	}

	// The transaction outlives ctx so that shutdown can still commit.
	dbCtx := context.WithoutCancel(ctx)
	handled := 0
	err := r.db.WithContext(dbCtx).Transaction(func(tx *gorm.DB) error {
		var locked bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", outboxRelayLockKey).Scan(&locked).Error; err != nil {
			return fmt.Errorf("take outbox relay lock: %w", err)
		}
		if !locked {
			return nil // another relay is working
		}

		var rows []OutboxEventEntity
		err := tx.Where("published_at IS NULL AND next_attempt_at <= now()").
			Order("seq").
			Limit(r.batchSize).
			Find(&rows).Error
		if err != nil {
			return fmt.Errorf("claim outbox events: %w", err)
		}

		for i := range rows {
			if ctx.Err() != nil {
				break // shutting down: keep the rest for the next relay
			}
			done, err := r.relay(ctx, tx, &rows[i])
			if err != nil {
				return err
			}
			if done {
				handled++
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return handled, nil
}

// relay publishes one row and records the outcome in tx. It returns false,
// recording nothing, when the publish was cut short by shutdown.
func (r *OutboxRelay) relay(ctx context.Context, tx *gorm.DB, row *OutboxEventEntity) (bool, error) {
	publishCtx, cancel := context.WithTimeout(ctx, outboxPublishTimeout)
	pubErr := r.publisher.Publish(publishCtx, row.ToDomain())
	cancel()
	if pubErr != nil && ctx.Err() != nil {
		return false, nil // shutdown, not a failed attempt: leave the row as it was
	}

	updates := map[string]any{"published_at": gorm.Expr("now()")}
	if pubErr != nil {
		attempts := row.Attempts + 1
		r.logger.Error("failed to publish outbox event",
			zap.String("event_id", row.ID.String()),
			zap.String("event_type", row.Type),
			zap.String("subject", row.Subject),
			zap.Int("attempts", attempts),
			zap.Error(pubErr),
		)
		updates = map[string]any{
			"attempts":        attempts,
			"last_error":      truncate(pubErr.Error(), outboxMaxErrorLen),
			"next_attempt_at": gorm.Expr("now() + make_interval(secs => ?)", outboxBackoff(attempts).Seconds()),
		}
	}

	if err := tx.Model(&OutboxEventEntity{}).Where("id = ?", row.ID).Updates(updates).Error; err != nil {
		return false, fmt.Errorf("record outbox event %s: %w", row.ID, err)
	}
	return true, nil
}

// outboxBackoff returns the wait before retrying an event that has failed
// attempts times: 1s, 2s, 4s, ... capped at outboxMaxBackoff.
func outboxBackoff(attempts int) time.Duration {
	if attempts < 1 {
		return time.Second
	}
	if attempts > 20 { // far past the cap; avoids overflowing the shift
		return outboxMaxBackoff
	}
	return min(time.Second<<(attempts-1), outboxMaxBackoff)
}

// truncate shortens s to at most n bytes without splitting a UTF-8 character.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
