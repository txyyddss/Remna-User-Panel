CREATE TABLE connection_ip_blocks (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  node_uuid TEXT NOT NULL,
  ip_digest TEXT NOT NULL CHECK (length(ip_digest) = 64),
  sealed_ip TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'blocking'
    CHECK (status IN ('blocking','active','unblocking','pending_review')),
  block_operation_id TEXT REFERENCES provider_operations(id) ON DELETE SET NULL,
  unblock_operation_id TEXT REFERENCES provider_operations(id) ON DELETE SET NULL,
  expiry_job_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  UNIQUE(user_id,node_uuid,ip_digest)
);

CREATE INDEX connection_ip_blocks_owner_expiry_idx
  ON connection_ip_blocks(user_id,expires_at,id);

CREATE UNIQUE INDEX connection_ip_blocks_expiry_job_idx
  ON connection_ip_blocks(expiry_job_id);
