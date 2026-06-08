-- jamaah_auth: Notification system (catch-up)
-- Migration 009: idempotently ensure the notifications table exists.
--
-- 002_notifications was added with a version (2) that several databases had
-- already passed, so `migrate up` never applied it there and the table was
-- missing in production. This catch-up runs at a version above the current head
-- and uses IF NOT EXISTS so it is a no-op on databases that already created the
-- table via 002, and creates it everywhere else.

CREATE TABLE IF NOT EXISTS notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    severity    VARCHAR(20) NOT NULL DEFAULT 'info' CHECK (severity IN ('error','warning','info')),
    title       VARCHAR(255) NOT NULL,
    message     TEXT NOT NULL DEFAULT '',
    group_id    VARCHAR(100),
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_org ON notifications(org_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, is_read, created_at DESC);
