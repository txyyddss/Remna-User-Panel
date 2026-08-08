CREATE TABLE database_admin_reviews (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action TEXT NOT NULL CHECK (action IN ('insert','update','delete')),
  table_name TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  consumed_at TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX database_admin_reviews_expiry_idx
  ON database_admin_reviews(expires_at, consumed_at);
