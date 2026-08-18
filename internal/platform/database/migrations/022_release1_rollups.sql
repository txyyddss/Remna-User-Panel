PRAGMA foreign_keys = ON;

CREATE TABLE maintenance_runs (
  id TEXT PRIMARY KEY,
  local_date TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL CHECK (status IN ('running','succeeded','failed')),
  lease_owner TEXT NOT NULL,
  lease_expires_at TEXT NOT NULL,
  backup_run_id TEXT REFERENCES backup_runs(id) ON DELETE SET NULL,
  cleanup_counts_json TEXT NOT NULL DEFAULT '{}'
    CHECK (json_valid(cleanup_counts_json) AND json_type(cleanup_counts_json)='object'),
  error_code TEXT NOT NULL DEFAULT '',
  started_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT
);
CREATE INDEX maintenance_runs_status_lease_idx
  ON maintenance_runs(status,lease_expires_at);

CREATE TABLE payment_status_rollups (
  local_date TEXT NOT NULL,
  provider TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('paid','expired','cancelled','failed','refunded')),
  order_count INTEGER NOT NULL DEFAULT 0 CHECK (order_count >= 0),
  txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (txb_minor >= 0),
  updated_at TEXT NOT NULL,
  PRIMARY KEY(local_date,provider,status)
);

CREATE TABLE activity_daily_rollups (
  local_date TEXT PRIMARY KEY,
  checkin_count INTEGER NOT NULL DEFAULT 0 CHECK (checkin_count >= 0),
  checkin_reward_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (checkin_reward_txb_minor >= 0),
  group_message_count INTEGER NOT NULL DEFAULT 0 CHECK (group_message_count >= 0),
  group_message_reward_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (group_message_reward_txb_minor >= 0),
  updated_at TEXT NOT NULL
);

CREATE TABLE rollover_member_daily_rollups (
  local_date TEXT NOT NULL,
  user_id TEXT NOT NULL,
  settlement_count INTEGER NOT NULL DEFAULT 0 CHECK (settlement_count >= 0),
  credited_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (credited_txb_minor >= 0),
  allocated_traffic_bytes INTEGER NOT NULL DEFAULT 0 CHECK (allocated_traffic_bytes >= 0),
  used_traffic_bytes INTEGER NOT NULL DEFAULT 0 CHECK (used_traffic_bytes >= 0),
  updated_at TEXT NOT NULL,
  PRIMARY KEY(local_date,user_id)
);
CREATE INDEX rollover_member_daily_user_date_idx
  ON rollover_member_daily_rollups(user_id,local_date);

CREATE TABLE purchase_history_tombstones (
  purchase_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  combo_id TEXT NOT NULL,
  status TEXT NOT NULL,
  charged_txb_minor INTEGER NOT NULL CHECK (charged_txb_minor >= 0),
  gross_txb_minor INTEGER NOT NULL CHECK (gross_txb_minor >= 0),
  core_gross_txb_minor INTEGER NOT NULL CHECK (core_gross_txb_minor >= 0),
  coupon_discount_txb_minor INTEGER NOT NULL CHECK (coupon_discount_txb_minor >= 0),
  addon_count INTEGER NOT NULL DEFAULT 0 CHECK (addon_count >= 0),
  traffic_limit_bytes INTEGER NOT NULL CHECK (traffic_limit_bytes > 0),
  reset_strategy TEXT NOT NULL CHECK (reset_strategy IN ('DAY','WEEK','MONTH_ROLLING')),
  valid_from TEXT NOT NULL,
  valid_until TEXT NOT NULL,
  renewal_batch_id TEXT NOT NULL DEFAULT '',
  renewal_index INTEGER,
  auto_renew_source_purchase_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  removed_at TEXT NOT NULL
);
CREATE INDEX purchase_history_tombstones_user_created_idx
  ON purchase_history_tombstones(user_id,created_at DESC,purchase_id);
CREATE INDEX purchase_history_tombstones_combo_created_idx
  ON purchase_history_tombstones(combo_id,created_at DESC,purchase_id);

CREATE TABLE statistics_snapshots (
  partition TEXT PRIMARY KEY CHECK (length(partition) BETWEEN 1 AND 80),
  payload_json TEXT NOT NULL
    CHECK (json_valid(payload_json) AND json_type(payload_json)='object'),
  generated_at TEXT NOT NULL
);
