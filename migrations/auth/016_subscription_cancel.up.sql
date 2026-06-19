-- Cancel-at-period-end flag: a paid sub flagged true keeps its tier until
-- expires_at, then the existing auto-expire drops the org to Gratis.
ALTER TABLE subscriptions
  ADD COLUMN IF NOT EXISTS cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE;
