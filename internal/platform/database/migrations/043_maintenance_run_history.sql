CREATE TABLE maintenance_runs_v2 (
  id TEXT PRIMARY KEY,
  local_date TEXT NOT NULL,
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

INSERT INTO maintenance_runs_v2(id,local_date,status,lease_owner,lease_expires_at,backup_run_id,
  cleanup_counts_json,error_code,started_at,updated_at,completed_at)
SELECT id,local_date,status,lease_owner,lease_expires_at,backup_run_id,
  cleanup_counts_json,error_code,started_at,updated_at,completed_at
FROM maintenance_runs;

DROP TABLE maintenance_runs;
ALTER TABLE maintenance_runs_v2 RENAME TO maintenance_runs;

CREATE INDEX maintenance_runs_local_date_idx
  ON maintenance_runs(local_date,started_at DESC,id DESC);
CREATE INDEX maintenance_runs_status_lease_idx
  ON maintenance_runs(status,lease_expires_at);
