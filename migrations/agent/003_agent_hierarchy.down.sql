DROP TABLE IF EXISTS commission_tiers;

DROP INDEX IF EXISTS idx_agent_commissions_tier;
DROP INDEX IF EXISTS idx_agent_commissions_source;
ALTER TABLE agent_commissions
    DROP COLUMN IF EXISTS source_commission_id,
    DROP COLUMN IF EXISTS tier_level;

DROP INDEX IF EXISTS idx_agents_parent;
ALTER TABLE agents
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS level,
    DROP COLUMN IF EXISTS parent_id;
