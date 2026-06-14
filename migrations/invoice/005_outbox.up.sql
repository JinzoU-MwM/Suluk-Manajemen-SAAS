-- Transactional outbox for the Integration Bus. Domain writes insert an event
-- row in the SAME transaction; a relay worker publishes unsent rows to NATS and
-- stamps published_at. Guarantees no event is lost on a crash between commit and
-- publish (dual-write problem).
CREATE TABLE outbox (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    aggregate_type VARCHAR(40) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type     VARCHAR(60) NOT NULL,
    payload        JSONB NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ,
    attempts       INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Relay scans unpublished rows oldest-first.
CREATE INDEX idx_outbox_unpublished ON outbox(occurred_at) WHERE published_at IS NULL;
