PRAGMA foreign_keys = ON;

CREATE TABLE coupon_grant_discards (
  grant_id TEXT PRIMARY KEY REFERENCES coupon_grants(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  discarded_at TEXT NOT NULL
);
CREATE INDEX coupon_grant_discards_user_idx ON coupon_grant_discards(user_id, discarded_at DESC);

CREATE TABLE courtesy_credits (
  id TEXT PRIMARY KEY,
  payment_order_id TEXT NOT NULL UNIQUE REFERENCES payment_orders(id) ON DELETE CASCADE,
  actor_user_id TEXT NOT NULL REFERENCES users(id),
  txb_minor INTEGER NOT NULL CHECK (txb_minor > 0),
  ledger_entry_id TEXT NOT NULL UNIQUE REFERENCES ledger_entries(id),
  reason TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX courtesy_credits_created_idx ON courtesy_credits(created_at DESC);
