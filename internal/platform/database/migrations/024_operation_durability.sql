ALTER TABLE restore_jobs ADD COLUMN request_actor_id TEXT NOT NULL DEFAULT '';
ALTER TABLE restore_jobs ADD COLUMN idempotency_key TEXT NOT NULL DEFAULT ''
  CHECK (idempotency_key = '' OR length(idempotency_key) BETWEEN 1 AND 128);
ALTER TABLE restore_jobs ADD COLUMN request_fingerprint TEXT NOT NULL DEFAULT ''
  CHECK (request_fingerprint = '' OR length(request_fingerprint) BETWEEN 16 AND 128);

UPDATE restore_jobs
SET request_actor_id = COALESCE(actor_user_id, '')
WHERE request_actor_id = '';

CREATE UNIQUE INDEX restore_jobs_request_key_idx
  ON restore_jobs(request_actor_id,idempotency_key)
  WHERE idempotency_key <> '';

CREATE TABLE connection_scans_new (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  idempotency_key TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
  request_fingerprint TEXT NOT NULL CHECK (length(request_fingerprint) BETWEEN 16 AND 128),
  provider_job_id TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'queued'
    CHECK (status IN ('queued','processing','succeeded','failed','pending_review')),
  progress_percent REAL NOT NULL DEFAULT 0
    CHECK (progress_percent >= 0 AND progress_percent <= 100),
  error_code TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  expires_at TEXT NOT NULL,
  UNIQUE(user_id,idempotency_key)
);

INSERT INTO connection_scans_new(
  id,user_id,idempotency_key,request_fingerprint,provider_job_id,status,
  progress_percent,error_code,created_at,updated_at,completed_at,expires_at
)
SELECT
  id,user_id,idempotency_key,request_fingerprint,provider_job_id,status,
  progress_percent,error_code,created_at,updated_at,completed_at,expires_at
FROM connection_scans;

DROP TABLE connection_scans;
ALTER TABLE connection_scans_new RENAME TO connection_scans;

CREATE INDEX connection_scans_owner_expiry_idx
  ON connection_scans(user_id,expires_at DESC,id);
