-- Profile customization fields on users (phone already exists).
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS city               TEXT,
  ADD COLUMN IF NOT EXISTS bio                TEXT,
  ADD COLUMN IF NOT EXISTS avatar_color       TEXT    NOT NULL DEFAULT 'blue',
  ADD COLUMN IF NOT EXISTS notify_usage_limit BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS notify_expiry      BOOLEAN NOT NULL DEFAULT TRUE;
