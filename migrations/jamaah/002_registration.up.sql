CREATE TABLE registration_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    group_id UUID,
    package_id UUID,
    token VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_registration_links_token ON registration_links(token);
CREATE INDEX idx_registration_links_org ON registration_links(org_id);

CREATE TABLE pending_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    registration_link_id UUID REFERENCES registration_links(id),
    phone_number VARCHAR(30) NOT NULL,
    name VARCHAR(255) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    ktp_file_url TEXT DEFAULT '',
    passport_file_url TEXT DEFAULT '',
    visa_file_url TEXT DEFAULT '',
    notes TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'pending',
    jamaah_id UUID,
    reviewed_by UUID,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_pending_registrations_org ON pending_registrations(org_id);
CREATE INDEX idx_pending_registrations_link ON pending_registrations(registration_link_id);
CREATE INDEX idx_pending_registrations_status ON pending_registrations(status);
