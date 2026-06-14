-- jamaah_accounting: Double-entry Accounting Engine (Standar Accurate)
-- Migration 001: COA, journals, journal lines, processed-events dedup.
-- All money is BIGINT IDR (no cents), matching the invoice/payment convention.

CREATE TABLE chart_of_accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    code           VARCHAR(20) NOT NULL,
    name           VARCHAR(120) NOT NULL,
    type           VARCHAR(12) NOT NULL,          -- asset | liability | equity | revenue | expense
    normal_balance VARCHAR(6) NOT NULL,           -- debit | credit
    parent_id      UUID REFERENCES chart_of_accounts(id) ON DELETE SET NULL,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, code)
);
CREATE INDEX idx_coa_org ON chart_of_accounts(org_id);

CREATE TABLE journals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    journal_no      VARCHAR(40) NOT NULL,
    journal_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    source_module   VARCHAR(30) NOT NULL,         -- invoice | vendor | payroll | agent | tabungan | manual | opening
    source_ref_id   UUID,                          -- id of the originating business row
    source_event_id VARCHAR(64),                   -- outbox/event id; dedup key (NULL for manual)
    description     TEXT NOT NULL DEFAULT '',
    status          VARCHAR(10) NOT NULL DEFAULT 'posted', -- draft | posted | void
    created_by      UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, journal_no)
);
-- One journal per (org, source_event_id) — DB-level idempotency backstop.
CREATE UNIQUE INDEX uq_journal_source_event
    ON journals(org_id, source_event_id) WHERE source_event_id IS NOT NULL;
CREATE INDEX idx_journals_org_date ON journals(org_id, journal_date);
CREATE INDEX idx_journals_source ON journals(org_id, source_module, source_ref_id);

CREATE TABLE journal_lines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id  UUID NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL,
    account_id  UUID NOT NULL REFERENCES chart_of_accounts(id),
    debit       BIGINT NOT NULL DEFAULT 0,
    credit      BIGINT NOT NULL DEFAULT 0,
    memo        VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_line_nonneg CHECK (debit >= 0 AND credit >= 0),
    CONSTRAINT chk_line_one_side CHECK ((debit = 0) <> (credit = 0))
);
CREATE INDEX idx_lines_journal ON journal_lines(journal_id);
CREATE INDEX idx_lines_account ON journal_lines(org_id, account_id);

-- Consumer dedup: an event is posted at most once (post-once), independent of
-- the per-journal source_event_id unique index (defense in depth).
CREATE TABLE processed_events (
    event_id     VARCHAR(64) PRIMARY KEY,
    org_id       UUID NOT NULL,
    event_type   VARCHAR(60) NOT NULL,
    journal_id   UUID,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
