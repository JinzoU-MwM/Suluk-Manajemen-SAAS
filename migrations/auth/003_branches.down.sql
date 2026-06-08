ALTER TABLE organizations DROP COLUMN IF EXISTS parent_org_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS branch_name;
ALTER TABLE organizations DROP COLUMN IF EXISTS is_branch;
