PRAGMA foreign_keys = ON;

CREATE TABLE abuse_punishment_rules_v2 (
  action TEXT PRIMARY KEY CHECK(action IN ('warning','ip_ban','subscription_revoke','temporary_ban')),
  enabled INTEGER NOT NULL DEFAULT 0 CHECK(enabled IN (0,1)),
  incident_threshold INTEGER NOT NULL DEFAULT 1 CHECK(incident_threshold BETWEEN 1 AND 100000),
  duration_minutes INTEGER NOT NULL DEFAULT 0 CHECK(duration_minutes BETWEEN 0 AND 525600),
  all_nodes INTEGER NOT NULL DEFAULT 0 CHECK(all_nodes IN (0,1)),
  revision INTEGER NOT NULL DEFAULT 0 CHECK(revision >= 0),
  created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
INSERT INTO abuse_punishment_rules_v2
SELECT action,enabled,incident_threshold,
  CASE WHEN action IN ('ip_ban','temporary_ban') THEN duration_minutes ELSE 0 END,
  all_nodes,revision,created_at,updated_at
FROM abuse_punishment_rules;
DROP TABLE abuse_punishment_rules;
ALTER TABLE abuse_punishment_rules_v2 RENAME TO abuse_punishment_rules;

CREATE TABLE abuse_ip_ban_scans (
  record_id TEXT PRIMARY KEY REFERENCES abuse_records(id) ON DELETE CASCADE,
  scan_job_id TEXT NOT NULL,
  created_at TEXT NOT NULL
);
