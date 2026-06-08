CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(30) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    address TEXT DEFAULT '',
    commission_rate NUMERIC(5,2) DEFAULT 5.0,
    bank_name VARCHAR(100) DEFAULT '',
    bank_account_number VARCHAR(50) DEFAULT '',
    bank_account_name VARCHAR(255) DEFAULT '',
    notes TEXT DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE agent_commissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    agent_id UUID NOT NULL REFERENCES agents(id),
    jamaah_id UUID,
    invoice_id UUID,
    package_id UUID,
    jamaah_name VARCHAR(255) DEFAULT '',
    package_name VARCHAR(255) DEFAULT '',
    commission_amount BIGINT NOT NULL DEFAULT 0,
    commission_rate NUMERIC(5,2) DEFAULT 5.0,
    status VARCHAR(20) DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agents_org ON agents(org_id);
CREATE INDEX idx_agent_commissions_org ON agent_commissions(org_id);
CREATE INDEX idx_agent_commissions_agent ON agent_commissions(agent_id);
CREATE INDEX idx_agent_commissions_status ON agent_commissions(status);
