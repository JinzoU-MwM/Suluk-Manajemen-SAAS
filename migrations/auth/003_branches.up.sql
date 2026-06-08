ALTER TABLE organizations ADD COLUMN IF NOT EXISTS parent_org_id UUID;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS branch_name VARCHAR(255) DEFAULT '';
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS is_branch BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_orgs_parent ON organizations(parent_org_id);
