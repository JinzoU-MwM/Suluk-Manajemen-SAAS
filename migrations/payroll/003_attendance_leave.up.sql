-- HR attendance + leave (Phase 5C). Daily presence per employee and a leave
-- (cuti) request workflow. Unpaid days (absen / tanpa_gaji) feed salary-slip
-- deductions.
CREATE TABLE attendance (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'hadir', -- hadir|absen|izin|sakit|cuti|libur|tanpa_gaji
    notes       TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, date)
);
CREATE INDEX idx_attendance_emp_date ON attendance(org_id, employee_id, date);

CREATE TABLE leave_requests (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    type        VARCHAR(16) NOT NULL DEFAULT 'tahunan', -- tahunan|sakit|izin|tanpa_gaji
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    days        INT NOT NULL DEFAULT 1,
    reason      TEXT NOT NULL DEFAULT '',
    status      VARCHAR(12) NOT NULL DEFAULT 'pending', -- pending|approved|rejected
    decided_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_leave_org_status ON leave_requests(org_id, status);
CREATE INDEX idx_leave_employee ON leave_requests(employee_id);
