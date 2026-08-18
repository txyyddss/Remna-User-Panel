PRAGMA foreign_keys = ON;

-- Provider callbacks outlive the seven-day payment-detail window. These
-- compact rows keep replay detection durable without retaining checkout data.
CREATE TABLE payment_callback_tombstones (
  provider TEXT NOT NULL,
  dedupe_key TEXT NOT NULL,
  order_id TEXT NOT NULL,
  payload_hash TEXT NOT NULL,
  received_at TEXT NOT NULL,
  processed_at TEXT NOT NULL,
  PRIMARY KEY(provider,dedupe_key)
);
CREATE INDEX payment_callback_tombstones_order_idx
  ON payment_callback_tombstones(order_id,provider);
