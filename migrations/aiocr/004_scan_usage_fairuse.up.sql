-- One-shot-per-month marker so the fair-use WARN fires at most once per org/month.
ALTER TABLE scan_usage ADD COLUMN IF NOT EXISTS fairuse_alerted_at TIMESTAMPTZ;
