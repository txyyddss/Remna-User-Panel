ALTER TABLE combos ADD COLUMN rollover_min_remaining_bps INTEGER NOT NULL DEFAULT 0
  CHECK (rollover_min_remaining_bps BETWEEN 0 AND 10000);
ALTER TABLE combos ADD COLUMN rollover_max_txb_minor INTEGER NOT NULL DEFAULT 0
  CHECK (rollover_max_txb_minor >= 0);

ALTER TABLE purchases ADD COLUMN rollover_min_remaining_bps INTEGER NOT NULL DEFAULT 0
  CHECK (rollover_min_remaining_bps BETWEEN 0 AND 10000);
ALTER TABLE purchases ADD COLUMN rollover_max_txb_minor INTEGER NOT NULL DEFAULT 0
  CHECK (rollover_max_txb_minor >= 0);

CREATE TABLE purchase_rollovers (
  purchase_id TEXT PRIMARY KEY REFERENCES purchases(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('pending','processing','credited','zero','exception')),
  traffic_limit_bytes INTEGER NOT NULL CHECK (traffic_limit_bytes >= 0),
  used_traffic_bytes INTEGER,
  remaining_traffic_bytes INTEGER,
  minimum_remaining_bps INTEGER NOT NULL CHECK (minimum_remaining_bps BETWEEN 0 AND 10000),
  maximum_txb_minor INTEGER NOT NULL CHECK (maximum_txb_minor >= 0),
  net_paid_txb_minor INTEGER NOT NULL CHECK (net_paid_txb_minor >= 0),
  credited_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (credited_txb_minor >= 0),
  exception_code TEXT NOT NULL DEFAULT '',
  attempts INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT
);
CREATE INDEX purchase_rollovers_status_idx ON purchase_rollovers(status, updated_at);

ALTER TABLE payment_orders ADD COLUMN method_id TEXT NOT NULL DEFAULT '';
ALTER TABLE payment_orders ADD COLUMN provider_rail TEXT NOT NULL DEFAULT '';
ALTER TABLE payment_orders ADD COLUMN rate_direction TEXT NOT NULL DEFAULT 'currency_per_txb'
  CHECK (rate_direction IN ('currency_per_txb','txb_per_currency'));
ALTER TABLE payment_orders ADD COLUMN receiving_address TEXT;
ALTER TABLE payment_orders ADD COLUMN actual_crypto_amount TEXT;
ALTER TABLE payment_orders ADD COLUMN actual_crypto_currency TEXT;
ALTER TABLE payment_orders ADD COLUMN cancelled_at TEXT;
ALTER TABLE payment_orders ADD COLUMN cancel_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE payment_orders ADD COLUMN provider_cancel_status TEXT NOT NULL DEFAULT ''
  CHECK (provider_cancel_status IN ('','unsupported','requested','cancelled','failed'));
CREATE INDEX payment_orders_method_created_idx ON payment_orders(method_id, created_at DESC);

ALTER TABLE users ADD COLUMN recovery_reason TEXT NOT NULL DEFAULT '';

CREATE TABLE restore_jobs (
  id TEXT PRIMARY KEY,
  backup_run_id TEXT NOT NULL REFERENCES backup_runs(id),
  actor_user_id TEXT REFERENCES users(id),
  status TEXT NOT NULL CHECK (status IN ('staging','ready','applying','complete','failed','rolled_back')),
  staged_path TEXT NOT NULL,
  rescue_path TEXT NOT NULL,
  source_sha256 TEXT NOT NULL,
  error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT
);
CREATE INDEX restore_jobs_created_idx ON restore_jobs(created_at DESC);
