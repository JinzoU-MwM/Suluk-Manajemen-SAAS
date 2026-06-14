-- jamaah_tabungan: Savings (Tabungan Umroh/Haji) module.
CREATE TABLE savings_accounts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID NOT NULL,
    jamaah_id         UUID NOT NULL,
    jamaah_name       VARCHAR(255) NOT NULL DEFAULT '',
    target_package_id UUID,
    target_amount     BIGINT NOT NULL DEFAULT 0,
    balance           BIGINT NOT NULL DEFAULT 0,
    status            VARCHAR(12) NOT NULL DEFAULT 'aktif', -- aktif | converted | closed
    notes             TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_balance_nonneg CHECK (balance >= 0)
);
CREATE INDEX idx_savings_org ON savings_accounts(org_id);
CREATE INDEX idx_savings_jamaah ON savings_accounts(org_id, jamaah_id);

CREATE TABLE savings_deposits (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  UUID NOT NULL REFERENCES savings_accounts(id) ON DELETE CASCADE,
    org_id      UUID NOT NULL,
    amount      BIGINT NOT NULL,                 -- positive
    direction   VARCHAR(8) NOT NULL DEFAULT 'in', -- in (setor) | out (konversi/tarik)
    type        VARCHAR(12) NOT NULL DEFAULT 'setor', -- setor | konversi | tarik
    method      VARCHAR(20) NOT NULL DEFAULT 'tunai',
    reference   VARCHAR(100) NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_deposits_account ON savings_deposits(account_id);
CREATE INDEX idx_deposits_org_date ON savings_deposits(org_id, created_at);
