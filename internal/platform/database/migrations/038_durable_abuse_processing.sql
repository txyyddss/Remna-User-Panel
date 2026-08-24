PRAGMA foreign_keys = ON;

ALTER TABLE abuse_policy ADD COLUMN streak_seconds INTEGER NOT NULL DEFAULT 30
  CHECK(streak_seconds BETWEEN 1 AND 1800);
ALTER TABLE abuse_detector_state ADD COLUMN incident_emitted INTEGER NOT NULL DEFAULT 0
  CHECK(incident_emitted IN (0,1));
UPDATE abuse_detector_state
SET incident_emitted = CASE WHEN streak_seconds >= 30 THEN 1 ELSE 0 END,
    streak_seconds = MIN(streak_seconds, 30);

CREATE TABLE abuse_pending_log_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  node_uuid TEXT NOT NULL,
  event_second TEXT NOT NULL,
  domain TEXT NOT NULL,
  fingerprint TEXT NOT NULL,
  received_at TEXT NOT NULL,
  claim_token TEXT,
  claimed_at TEXT,
  UNIQUE(node_uuid,fingerprint)
);
CREATE INDEX abuse_pending_log_events_ready
  ON abuse_pending_log_events(claim_token,event_second,user_id,id);
CREATE INDEX abuse_pending_log_events_claimed
  ON abuse_pending_log_events(claimed_at) WHERE claim_token IS NOT NULL;

CREATE TABLE abuse_qps_rollups (
  window_at TEXT PRIMARY KEY,
  observation_count INTEGER NOT NULL CHECK(observation_count >= 0),
  qps_sum INTEGER NOT NULL CHECK(qps_sum >= 0),
  qps_min INTEGER NOT NULL CHECK(qps_min >= 0),
  qps_max INTEGER NOT NULL CHECK(qps_max >= 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE abuse_incident_facts (
  incident_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  incident_bucket_at TEXT NOT NULL,
  selected_action TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(user_id,incident_bucket_at)
);
CREATE INDEX abuse_incident_facts_member
  ON abuse_incident_facts(user_id,created_at DESC);
INSERT OR IGNORE INTO abuse_incident_facts(
  incident_id,user_id,incident_bucket_at,selected_action,created_at
)
SELECT id,user_id,incident_bucket_at,selected_action,created_at
FROM abuse_records WHERE deleted_at IS NULL;

ALTER TABLE abuse_records ADD COLUMN punishment_completed_at TEXT;
UPDATE abuse_records
SET punishment_completed_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE NOT EXISTS (
  SELECT 1 FROM outbox_jobs
  WHERE kind='abuse_punishment'
    AND json_extract(payload,'$.recordId')=abuse_records.id
    AND status IN ('pending','processing','failed')
);

WITH per_second AS (
  SELECT user_id,bucket_at,SUM(qps) AS qps
  FROM abuse_qps_samples
  WHERE bucket_at >= strftime('%Y-%m-%dT%H:%M:%fZ','now','-24 hours')
  GROUP BY user_id,bucket_at
), windowed AS (
  SELECT substr(bucket_at,1,14) ||
    CASE WHEN CAST(substr(bucket_at,15,2) AS INTEGER) < 30
      THEN '00:00Z' ELSE '30:00Z' END AS window_at,
    COUNT(*) AS observation_count,SUM(qps) AS qps_sum,
    MIN(qps) AS qps_min,MAX(qps) AS qps_max
  FROM per_second GROUP BY window_at
)
INSERT INTO abuse_qps_rollups(
  window_at,observation_count,qps_sum,qps_min,qps_max,created_at,updated_at
)
SELECT window_at,observation_count,qps_sum,qps_min,qps_max,
  strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM windowed;
