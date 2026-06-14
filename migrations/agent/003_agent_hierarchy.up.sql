-- Agent hierarchy + tiered ("berjenjang") commissions (Phase 3B).
-- Agents gain a self-referential parent + depth/type; commissions gain the tier
-- they belong to and a link back to the originating (seller) commission; a new
-- commission_tiers table holds the per-org override rates for each upline level.

ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES agents(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS level     SMALLINT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS type      VARCHAR(16) NOT NULL DEFAULT 'agent';

CREATE INDEX IF NOT EXISTS idx_agents_parent ON agents(parent_id);

ALTER TABLE agent_commissions
    ADD COLUMN IF NOT EXISTS tier_level           SMALLINT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS source_commission_id UUID REFERENCES agent_commissions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_agent_commissions_source ON agent_commissions(source_commission_id);
CREATE INDEX IF NOT EXISTS idx_agent_commissions_tier ON agent_commissions(org_id, tier_level);

-- Per-org override rates. level = distance from the seller: 1 = seller (full,
-- not stored here), 2 = direct upline, 3 = next, ... rate_pct applies to the
-- same sale base the seller's commission was computed from. Absent rows fall
-- back to the code defaults (L2=2%, L3=1%).
CREATE TABLE IF NOT EXISTS commission_tiers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL,
    level      SMALLINT NOT NULL,
    rate_pct   NUMERIC(5,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, level)
);
