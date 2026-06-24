-- Per-org monthly AI-scan usage counter (Phase 4a: metering).
-- One row per (org, year, month); incremented atomically per scanned document.
CREATE TABLE IF NOT EXISTS scan_usage (
    org_id            UUID NOT NULL,
    year              INT  NOT NULL,
    month             INT  NOT NULL,
    documents_scanned INT  NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, year, month)
);

CREATE INDEX IF NOT EXISTS idx_scan_usage_org ON scan_usage (org_id);
