-- jamaah_auth: Notification system
-- Migration 002: notifications

CREATE TABLE notifications (
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

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_org ON notifications(org_id, is_read);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read, created_at DESC);
