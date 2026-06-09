-- Performance indexes (audit P8): cover columns that are filtered/joined on hot
-- paths but were previously unindexed.

-- ListPendingRegistrations joins registration_links and filters by group_id.
CREATE INDEX IF NOT EXISTS idx_registration_links_group ON registration_links(group_id);

-- ListNotes filters WHERE org_id = $1 AND jamaah_id = $2; only jamaah_id was indexed.
CREATE INDEX IF NOT EXISTS idx_notes_org_jamaah ON jamaah_notes(org_id, jamaah_id);
