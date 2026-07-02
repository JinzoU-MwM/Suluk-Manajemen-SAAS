-- Records which refund policy (if any) applied when a refund was initiated,
-- so finance can audit "why was this refund only 50%?" instead of just
-- trusting whatever refund_pct the requester typed in. ON DELETE SET NULL:
-- deleting a policy later must never block or cascade into historical
-- refund rows.
ALTER TABLE refunds ADD COLUMN policy_id UUID REFERENCES refund_policies(id) ON DELETE SET NULL;
