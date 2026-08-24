PRAGMA foreign_keys = ON;

CREATE TABLE abuse_policy (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  global_enabled INTEGER NOT NULL DEFAULT 0 CHECK (global_enabled IN (0,1)),
  global_limit INTEGER NOT NULL DEFAULT 0 CHECK (global_limit BETWEEN 0 AND 100000),
  warning_validity_days INTEGER NOT NULL DEFAULT 7 CHECK (warning_validity_days BETWEEN 1 AND 365),
  revision INTEGER NOT NULL DEFAULT 0 CHECK (revision >= 0),
  created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
INSERT INTO abuse_policy(id,created_at,updated_at) VALUES(1,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE abuse_node_credentials (
  node_uuid TEXT PRIMARY KEY, node_name TEXT NOT NULL, key_digest TEXT NOT NULL UNIQUE,
  sealed_key TEXT NOT NULL, last_report_at TEXT, rotated_at TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE abuse_domain_rules (
  id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, expression TEXT NOT NULL, qps_limit INTEGER NOT NULL CHECK(qps_limit BETWEEN 1 AND 100000),
  enabled INTEGER NOT NULL DEFAULT 1 CHECK(enabled IN (0,1)), revision INTEGER NOT NULL DEFAULT 0 CHECK(revision >= 0), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE TABLE abuse_whitelist (remna_user_id TEXT PRIMARY KEY, created_at TEXT NOT NULL);
CREATE TABLE abuse_punishment_rules (
  action TEXT PRIMARY KEY CHECK(action IN ('warning','ip_ban','subscription_revoke','temporary_ban')),
  enabled INTEGER NOT NULL DEFAULT 0 CHECK(enabled IN (0,1)), incident_threshold INTEGER NOT NULL DEFAULT 1 CHECK(incident_threshold BETWEEN 1 AND 100000),
  duration_minutes INTEGER NOT NULL DEFAULT 60 CHECK(duration_minutes BETWEEN 1 AND 525600),
  all_nodes INTEGER NOT NULL DEFAULT 0 CHECK(all_nodes IN (0,1)), revision INTEGER NOT NULL DEFAULT 0 CHECK(revision >= 0), created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
INSERT INTO abuse_punishment_rules(action,created_at,updated_at) VALUES
 ('warning',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')),
 ('ip_ban',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')),
 ('subscription_revoke',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')),
 ('temporary_ban',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'));

CREATE TABLE abuse_log_fingerprints (node_uuid TEXT NOT NULL, fingerprint TEXT NOT NULL, observed_at TEXT NOT NULL, PRIMARY KEY(node_uuid,fingerprint));
CREATE TABLE abuse_qps_samples (
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, node_uuid TEXT NOT NULL, bucket_at TEXT NOT NULL,
  reason_name TEXT NOT NULL, qps_limit INTEGER NOT NULL CHECK(qps_limit BETWEEN 1 AND 100000), qps INTEGER NOT NULL DEFAULT 0 CHECK(qps >= 0),
  PRIMARY KEY(user_id,node_uuid,bucket_at,reason_name)
);
CREATE INDEX abuse_qps_samples_ready ON abuse_qps_samples(bucket_at,user_id,reason_name);
CREATE TABLE abuse_detector_state (
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, reason_name TEXT NOT NULL, last_bucket_at TEXT NOT NULL,
  streak_seconds INTEGER NOT NULL CHECK(streak_seconds >= 0), PRIMARY KEY(user_id,reason_name)
);
CREATE TABLE abuse_records (
  id TEXT PRIMARY KEY, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, incident_bucket_at TEXT NOT NULL,
  measured_qps INTEGER NOT NULL CHECK(measured_qps >= 0), qps_limit INTEGER NOT NULL CHECK(qps_limit >= 1), selected_action TEXT NOT NULL,
  expires_at TEXT, deleted_at TEXT, created_at TEXT NOT NULL, UNIQUE(user_id,incident_bucket_at)
);
CREATE INDEX abuse_records_member ON abuse_records(user_id,created_at DESC,id DESC);
CREATE TABLE abuse_record_reasons (record_id TEXT NOT NULL REFERENCES abuse_records(id) ON DELETE CASCADE, name TEXT NOT NULL, PRIMARY KEY(record_id,name));
CREATE TABLE abuse_record_nodes (record_id TEXT NOT NULL REFERENCES abuse_records(id) ON DELETE CASCADE, node_uuid TEXT NOT NULL, PRIMARY KEY(record_id,node_uuid));
CREATE TABLE abuse_temp_bans (
  record_id TEXT PRIMARY KEY REFERENCES abuse_records(id) ON DELETE CASCADE, user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL, restore_queued_at TEXT, restored_at TEXT, created_at TEXT NOT NULL
);
CREATE TABLE abuse_notification_deliveries (
  record_id TEXT NOT NULL REFERENCES abuse_records(id) ON DELETE CASCADE, recipient_telegram_id INTEGER NOT NULL,
  kind TEXT NOT NULL, delivered_at TEXT, created_at TEXT NOT NULL, PRIMARY KEY(record_id,recipient_telegram_id,kind)
);
