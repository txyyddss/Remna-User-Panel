PRAGMA foreign_keys = ON;

-- Remnawave distinguishes a rolling 30-day cadence from calendar months.
ALTER TABLE combos ADD COLUMN reset_strategy_next TEXT NOT NULL DEFAULT 'MONTH_ROLLING'
  CHECK (reset_strategy_next IN ('DAY','WEEK','MONTH_ROLLING'));
UPDATE combos
SET reset_strategy_next=CASE reset_strategy
  WHEN 'MONTH' THEN 'MONTH_ROLLING'
  ELSE reset_strategy
END;
ALTER TABLE combos DROP COLUMN reset_strategy;
ALTER TABLE combos RENAME COLUMN reset_strategy_next TO reset_strategy;

-- Paid traffic resets use the immutable core combo price at purchase time.
-- Add-ons and coupon discounts are intentionally excluded from this basis.
ALTER TABLE purchases ADD COLUMN core_gross_txb_minor INTEGER NOT NULL DEFAULT 0
  CHECK (core_gross_txb_minor >= 0);
UPDATE purchases
SET core_gross_txb_minor=MAX(
  0,
  COALESCE(gross_price_txb_minor, charged_txb_minor + coupon_discount_txb_minor)
  - COALESCE((
      SELECT SUM(charged_txb_minor)
      FROM purchase_addons
      WHERE purchase_id=purchases.id
    ), 0)
);

CREATE TABLE txb_limits (
  singleton INTEGER PRIMARY KEY CHECK (singleton=1),
  minimum_txb_minor INTEGER NOT NULL CHECK (minimum_txb_minor > 0),
  maximum_txb_minor INTEGER NOT NULL CHECK (maximum_txb_minor >= minimum_txb_minor),
  updated_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  updated_at TEXT NOT NULL
);
INSERT INTO txb_limits(singleton,minimum_txb_minor,maximum_txb_minor,updated_at)
VALUES(1,100,10000000000,strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE provider_operations (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL REFERENCES users(id),
  owner_user_id TEXT REFERENCES users(id),
  kind TEXT NOT NULL CHECK (length(kind) BETWEEN 1 AND 80),
  idempotency_key TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
  request_fingerprint TEXT NOT NULL CHECK (length(request_fingerprint) BETWEEN 16 AND 128),
  status TEXT NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','processing','succeeded','failed','compensated','pending_review','partial')),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
  attempt_started_at TEXT,
  provider_reference TEXT NOT NULL DEFAULT '',
  error_code TEXT NOT NULL DEFAULT '',
  result_json TEXT NOT NULL DEFAULT '{}'
    CHECK (json_valid(result_json) AND json_type(result_json)='object'),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  UNIQUE(actor_user_id,kind,idempotency_key)
);
CREATE INDEX provider_operations_owner_created_idx
  ON provider_operations(owner_user_id,created_at DESC,id DESC);
CREATE INDEX provider_operations_status_updated_idx
  ON provider_operations(status,updated_at,id);

CREATE TABLE provider_operation_items (
  operation_id TEXT NOT NULL REFERENCES provider_operations(id) ON DELETE CASCADE,
  item_key TEXT NOT NULL CHECK (length(item_key) BETWEEN 1 AND 160),
  target_type TEXT NOT NULL CHECK (length(target_type) BETWEEN 1 AND 80),
  target_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','processing','succeeded','failed','compensated','pending_review')),
  provider_reference TEXT NOT NULL DEFAULT '',
  error_code TEXT NOT NULL DEFAULT '',
  result_json TEXT NOT NULL DEFAULT '{}'
    CHECK (json_valid(result_json) AND json_type(result_json)='object'),
  attempt_started_at TEXT,
  completed_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY(operation_id,item_key)
);
CREATE INDEX provider_operation_items_status_idx
  ON provider_operation_items(operation_id,status,item_key);

CREATE TABLE provider_operation_replays (
  id TEXT PRIMARY KEY,
  operation_id TEXT NOT NULL REFERENCES provider_operations(id) ON DELETE CASCADE,
  actor_user_id TEXT NOT NULL REFERENCES users(id),
  idempotency_key TEXT NOT NULL,
  request_fingerprint TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(operation_id,idempotency_key,request_fingerprint)
);
CREATE UNIQUE INDEX provider_operation_replays_resolution_key_idx
  ON provider_operation_replays(operation_id,actor_user_id,idempotency_key);

-- Raw connection IPs and signed drop handles are never persisted here.
CREATE TABLE connection_scans (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  idempotency_key TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
  request_fingerprint TEXT NOT NULL CHECK (length(request_fingerprint) BETWEEN 16 AND 128),
  provider_job_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','processing','succeeded','failed')),
  progress_percent REAL NOT NULL DEFAULT 0
    CHECK (progress_percent >= 0 AND progress_percent <= 100),
  error_code TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  expires_at TEXT NOT NULL,
  UNIQUE(user_id,idempotency_key)
);
CREATE INDEX connection_scans_owner_expiry_idx
  ON connection_scans(user_id,expires_at DESC,id);

ALTER TABLE backup_runs ADD COLUMN source TEXT NOT NULL DEFAULT 'generated';
ALTER TABLE backup_runs ADD COLUMN actor_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE backup_runs ADD COLUMN idempotency_key TEXT NOT NULL DEFAULT '';
ALTER TABLE backup_runs ADD COLUMN request_fingerprint TEXT NOT NULL DEFAULT '';
ALTER TABLE backup_runs ADD COLUMN sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE backup_runs ADD COLUMN original_filename TEXT NOT NULL DEFAULT '';
ALTER TABLE backup_runs ADD COLUMN reason TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX backup_runs_upload_idempotency_idx
  ON backup_runs(actor_user_id,idempotency_key)
  WHERE source='upload' AND actor_user_id IS NOT NULL AND idempotency_key<>'';
