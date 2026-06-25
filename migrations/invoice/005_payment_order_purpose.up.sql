-- Distinguish a subscription purchase from a scan top-up so the Pakasir webhook
-- can route to the right post-payment action. Existing rows = 'subscription'.
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS purpose VARCHAR(30) NOT NULL DEFAULT 'subscription';
