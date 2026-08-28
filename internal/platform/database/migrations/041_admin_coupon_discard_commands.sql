PRAGMA foreign_keys = ON;

CREATE TABLE admin_coupon_discard_commands (
  actor_user_id TEXT NOT NULL REFERENCES users(id), idempotency_key TEXT NOT NULL,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  grant_id TEXT NOT NULL REFERENCES coupon_grants(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL, PRIMARY KEY(actor_user_id, idempotency_key)
);
