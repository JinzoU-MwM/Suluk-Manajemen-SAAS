-- Phase 6: stock monitoring + auto-deduct on departure.

CREATE TABLE IF NOT EXISTS package_kit_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    package_id     UUID NOT NULL,
    item_id        UUID NOT NULL,
    qty_per_jamaah INTEGER NOT NULL CHECK (qty_per_jamaah > 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, package_id, item_id)
);
CREATE INDEX IF NOT EXISTS idx_package_kit_items_pkg ON package_kit_items(org_id, package_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    item_id     UUID NOT NULL,
    delta       INTEGER NOT NULL,
    reason      TEXT NOT NULL,         -- initial | restock | adjustment | departure
    note        TEXT NOT NULL DEFAULT '',
    group_id    UUID,
    package_id  UUID,
    created_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_stock_movements_item ON stock_movements(org_id, item_id, created_at DESC);

-- One departure deduction per (group, item) ⇒ a re-delivered group.departed is a no-op.
CREATE UNIQUE INDEX IF NOT EXISTS uniq_departure_movement
    ON stock_movements(group_id, item_id) WHERE reason = 'departure';
