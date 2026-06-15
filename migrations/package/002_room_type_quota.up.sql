-- Room-type quota lock (Phase 4A). Per-room-type seat caps on pricing tiers, in
-- addition to the package-wide total_seats. Opt-in: quota_seats = 0 means no
-- per-type cap (only the total applies), so existing packages are unaffected.
ALTER TABLE pricing_tiers
    ADD COLUMN IF NOT EXISTS quota_seats    INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reserved_seats INT NOT NULL DEFAULT 0;
