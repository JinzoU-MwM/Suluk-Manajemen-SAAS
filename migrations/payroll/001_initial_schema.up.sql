CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    position VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'tetap',
    base_salary BIGINT NOT NULL DEFAULT 0,
    allowance BIGINT DEFAULT 0,
    bpjs_tk BIGINT DEFAULT 0,
    bpjs_kes BIGINT DEFAULT 0,
    pph21_rate NUMERIC(5,2) DEFAULT 0,
    phone VARCHAR(30) DEFAULT '',
    email VARCHAR(255) DEFAULT '',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_employees_org ON employees(org_id);

CREATE TABLE salary_slips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    period VARCHAR(7) NOT NULL,
    base_salary BIGINT NOT NULL DEFAULT 0,
    allowance BIGINT DEFAULT 0,
    deductions BIGINT DEFAULT 0,
    pph21_amount BIGINT DEFAULT 0,
    bpjs_amount BIGINT DEFAULT 0,
    advance_deduction BIGINT DEFAULT 0,
    net_salary BIGINT NOT NULL DEFAULT 0,
    package_id UUID,
    status VARCHAR(20) DEFAULT 'draft',
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_salary_slips_org ON salary_slips(org_id);
CREATE INDEX idx_salary_slips_employee ON salary_slips(employee_id);
CREATE INDEX idx_salary_slips_period ON salary_slips(period);

CREATE TABLE advances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    remaining BIGINT NOT NULL,
    reason TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_advances_org ON advances(org_id);
CREATE INDEX idx_advances_employee ON advances(employee_id);

CREATE TABLE advance_repayments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advance_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    salary_slip_id UUID,
    paid_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_advance_repayments_advance ON advance_repayments(advance_id);
