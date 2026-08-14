PRAGMA foreign_keys = ON;

ALTER TABLE purchases ADD COLUMN auto_renew_enabled INTEGER NOT NULL DEFAULT 0
  CHECK (auto_renew_enabled IN (0,1));
ALTER TABLE purchases ADD COLUMN auto_renew_failure_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE purchases ADD COLUMN auto_renew_failed_at TEXT;
ALTER TABLE purchases ADD COLUMN recurring_discount_attached INTEGER NOT NULL DEFAULT 0
  CHECK (recurring_discount_attached IN (0,1));
ALTER TABLE purchases ADD COLUMN auto_renew_source_purchase_id TEXT
  REFERENCES purchases(id);

CREATE INDEX purchases_auto_renew_due_idx
  ON purchases(auto_renew_enabled,status,valid_until);
CREATE UNIQUE INDEX purchases_auto_renew_successor_idx
  ON purchases(auto_renew_source_purchase_id)
  WHERE auto_renew_source_purchase_id IS NOT NULL;

-- Legacy rows have no immutable coupon-kind snapshot. Preserve the currently
-- linked recurring grant as the best available attachment without enabling it.
UPDATE purchases
SET recurring_discount_attached=1
WHERE coupon_grant_id IN (
  SELECT coupon_grants.id
  FROM coupon_grants
  JOIN coupon_definitions ON coupon_definitions.id=coupon_grants.coupon_id
  WHERE coupon_definitions.kind='purchase_recurring'
);
