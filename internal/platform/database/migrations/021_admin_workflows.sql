PRAGMA foreign_keys = ON;

-- Advanced entitlement edits are isolated from the live catalog and never
-- rewrite immutable purchase pricing or ledger evidence.
ALTER TABLE purchases ADD COLUMN entitlement_traffic_limit_bytes INTEGER
  CHECK (entitlement_traffic_limit_bytes IS NULL OR entitlement_traffic_limit_bytes > 0);
ALTER TABLE purchases ADD COLUMN entitlement_reset_strategy TEXT
  CHECK (entitlement_reset_strategy IS NULL OR entitlement_reset_strategy IN ('DAY','WEEK','MONTH_ROLLING'));
ALTER TABLE purchases ADD COLUMN entitlement_squad_uuids TEXT
  CHECK (entitlement_squad_uuids IS NULL OR (json_valid(entitlement_squad_uuids) AND json_type(entitlement_squad_uuids)='array'));
ALTER TABLE purchases ADD COLUMN entitlement_addon_squad_uuids TEXT
  CHECK (entitlement_addon_squad_uuids IS NULL OR (json_valid(entitlement_addon_squad_uuids) AND json_type(entitlement_addon_squad_uuids)='array'));

-- Upload intake is a filesystem/database publication saga. Only a publishing
-- row may be completed by startup reconciliation after a process interruption.
CREATE TABLE backup_uploads (
  id TEXT PRIMARY KEY,
  backup_run_id TEXT NOT NULL UNIQUE REFERENCES backup_runs(id) ON DELETE CASCADE,
  actor_user_id TEXT NOT NULL REFERENCES users(id),
  idempotency_key TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
  expected_sha256 TEXT NOT NULL DEFAULT '',
  actual_sha256 TEXT NOT NULL DEFAULT '',
  temporary_path TEXT NOT NULL,
  final_path TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
  status TEXT NOT NULL CHECK (status IN ('receiving','validating','publishing','complete','failed')),
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  UNIQUE(actor_user_id,idempotency_key)
);
CREATE INDEX backup_uploads_status_updated_idx ON backup_uploads(status,updated_at,id);
