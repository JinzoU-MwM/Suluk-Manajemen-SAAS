CREATE TABLE member_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    package_id UUID NOT NULL,
    member_id UUID NOT NULL,
    nama TEXT NOT NULL DEFAULT '',
    gender TEXT NOT NULL DEFAULT '',
    baju_size VARCHAR(5) DEFAULT '',
    family_id VARCHAR(10) DEFAULT '',
    is_equipment_received BOOLEAN DEFAULT FALSE,
    received_items TEXT[] DEFAULT '{}',
    received_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(package_id, member_id)
);

CREATE INDEX idx_member_equipment_org ON member_equipment(org_id);
CREATE INDEX idx_member_equipment_package ON member_equipment(package_id);
CREATE INDEX idx_member_equipment_received ON member_equipment(is_equipment_received);

CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'perlengkapan',
    unit TEXT DEFAULT 'pcs',
    stock INTEGER DEFAULT 0,
    min_stock INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_inventory_items_org ON inventory_items(org_id);
