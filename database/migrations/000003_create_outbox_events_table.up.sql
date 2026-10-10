-- Transactional outbox: services insert a row in the same transaction as the
-- change it describes; the outbox relay publishes pending rows to SNS and
-- marks them published. Published rows are kept; delete old ones on a
-- schedule, e.g. DELETE FROM outbox_events WHERE published_at < now() - interval '7 days'.
CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,                                  -- the event (CloudEvent) ID
    seq BIGINT GENERATED ALWAYS AS IDENTITY UNIQUE,       -- publish order
    type VARCHAR(255) NOT NULL,
    source VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL DEFAULT '',
    data JSONB NOT NULL,
    occurred_at TIMESTAMPTZ(3) NOT NULL,
    attempts INT NOT NULL DEFAULT 0,                      -- failed publish attempts
    next_attempt_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    last_error TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ(3)                           -- NULL until published
);

-- The relay only scans unpublished rows.
CREATE INDEX IF NOT EXISTS idx_outbox_events_pending ON outbox_events (seq) WHERE published_at IS NULL;
