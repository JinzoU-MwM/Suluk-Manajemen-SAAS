ALTER TABLE pricing_tiers
    DROP COLUMN IF EXISTS reserved_seats,
    DROP COLUMN IF EXISTS quota_seats;
