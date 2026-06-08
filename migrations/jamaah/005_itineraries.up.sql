CREATE TABLE itineraries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    group_id UUID NOT NULL,
    day_number INTEGER NOT NULL DEFAULT 1,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    location VARCHAR(255) DEFAULT '',
    start_time TIME,
    end_time TIME,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_itineraries_org ON itineraries(org_id);
CREATE INDEX idx_itineraries_group ON itineraries(group_id);
