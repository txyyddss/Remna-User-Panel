PRAGMA foreign_keys = ON;

CREATE TABLE node_compensation_config (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0,1)),
  threshold_minutes INTEGER CHECK (threshold_minutes BETWEEN 1 AND 5256000),
  multiplier_bps INTEGER CHECK (multiplier_bps BETWEEN 100 AND 1000000),
  revision INTEGER NOT NULL DEFAULT 0 CHECK (revision >= 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK (enabled = 0 OR (threshold_minutes IS NOT NULL AND multiplier_bps IS NOT NULL))
);

INSERT INTO node_compensation_config(id,enabled,threshold_minutes,multiplier_bps,revision,created_at,updated_at)
VALUES(1,0,NULL,NULL,0,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE node_compensation_events (
  id TEXT PRIMARY KEY,
  node_uuid TEXT NOT NULL,
  node_name TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('observing','pending_review','queued','dismissed','ineligible')),
  offline_observed_at TEXT NOT NULL,
  recovered_observed_at TEXT,
  observed_duration_seconds INTEGER CHECK (observed_duration_seconds IS NULL OR observed_duration_seconds >= 0),
  threshold_minutes INTEGER NOT NULL CHECK (threshold_minutes BETWEEN 1 AND 5256000),
  multiplier_bps INTEGER NOT NULL CHECK (multiplier_bps BETWEEN 100 AND 1000000),
  proposed_extension_minutes INTEGER CHECK (proposed_extension_minutes IS NULL OR proposed_extension_minutes BETWEEN 0 AND 5256000),
  final_extension_minutes INTEGER CHECK (final_extension_minutes IS NULL OR final_extension_minutes BETWEEN 1 AND 5256000),
  capped INTEGER NOT NULL DEFAULT 0 CHECK (capped IN (0,1)),
  frozen_recipient_count INTEGER NOT NULL DEFAULT 0 CHECK (frozen_recipient_count >= 0),
  eligible_recipient_count INTEGER CHECK (eligible_recipient_count IS NULL OR eligible_recipient_count >= 0),
  skipped_recipient_count INTEGER CHECK (skipped_recipient_count IS NULL OR skipped_recipient_count >= 0),
  ineligible_reason TEXT CHECK (ineligible_reason IS NULL OR ineligible_reason IN ('node_disabled','below_threshold','no_recipients','computed_zero')),
  reviewed_by TEXT REFERENCES users(id),
  reviewed_at TEXT,
  review_reason TEXT,
  review_idempotency_key TEXT,
  review_fingerprint TEXT,
  provider_operation_id TEXT REFERENCES provider_operations(id),
  revision INTEGER NOT NULL DEFAULT 0 CHECK (revision >= 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX node_compensation_one_observing_per_node
  ON node_compensation_events(node_uuid) WHERE status='observing';
CREATE UNIQUE INDEX node_compensation_review_keys
  ON node_compensation_events(reviewed_by,review_idempotency_key)
  WHERE reviewed_by IS NOT NULL AND review_idempotency_key IS NOT NULL;
CREATE INDEX node_compensation_events_status_cursor
  ON node_compensation_events(status,created_at DESC,id DESC);

CREATE TABLE node_compensation_event_squads (
  event_id TEXT NOT NULL REFERENCES node_compensation_events(id) ON DELETE CASCADE,
  squad_uuid TEXT NOT NULL,
  squad_name TEXT NOT NULL,
  PRIMARY KEY(event_id,squad_uuid)
);

CREATE TABLE node_compensation_event_recipients (
  event_id TEXT NOT NULL REFERENCES node_compensation_events(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES users(id),
  PRIMARY KEY(event_id,user_id)
);

CREATE TABLE node_compensation_node_state (
  node_uuid TEXT PRIMARY KEY,
  node_name TEXT NOT NULL,
  is_connected INTEGER NOT NULL CHECK (is_connected IN (0,1)),
  is_disabled INTEGER NOT NULL CHECK (is_disabled IN (0,1)),
  open_event_id TEXT REFERENCES node_compensation_events(id),
  last_observed_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
