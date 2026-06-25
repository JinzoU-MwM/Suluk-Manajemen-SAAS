-- Per-purchase ledger of bought scan top-ups (Phase 4b). order_id is the
-- invoice payment_orders.id; UNIQUE makes crediting idempotent under Pakasir's
-- at-least-once webhook retries. purchased_this_month = SUM(scans) for the month.
CREATE TABLE IF NOT EXISTS scan_topups (
    order_id   UUID PRIMARY KEY,
    org_id     UUID NOT NULL,
    year       INT  NOT NULL,
    month      INT  NOT NULL,
    scans      INT  NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_topups_org_month ON scan_topups (org_id, year, month);
