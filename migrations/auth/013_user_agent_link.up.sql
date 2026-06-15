-- B2B agent portal (Phase 3C): link a login user to an agent record so the JWT
-- can carry an agent_id and the portal can scope to that agent's subtree. The
-- agent lives in agent-service's DB, so this is an unenforced cross-service id.
ALTER TABLE users ADD COLUMN IF NOT EXISTS agent_id UUID;

CREATE INDEX IF NOT EXISTS idx_users_agent ON users(agent_id);
