-- Jemaah self-service portal (Phase 6): link a login user to a jamaah profile.
-- The profile lives in jamaah-service's DB, so this is an unenforced
-- cross-service id (portal queries still scope by org_id).
ALTER TABLE users ADD COLUMN IF NOT EXISTS jamaah_id UUID;
CREATE INDEX IF NOT EXISTS idx_users_jamaah ON users(jamaah_id);
