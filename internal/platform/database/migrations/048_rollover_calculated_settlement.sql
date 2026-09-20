CREATE TABLE purchase_rollovers_next (
  purchase_id TEXT PRIMARY KEY REFERENCES purchases(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('pending','processing','calculated','credited','zero','exception')),
  traffic_limit_bytes INTEGER NOT NULL CHECK (traffic_limit_bytes >= 0),
  used_traffic_bytes INTEGER,
  remaining_traffic_bytes INTEGER,
  minimum_remaining_bps INTEGER NOT NULL CHECK (minimum_remaining_bps BETWEEN 0 AND 10000),
  net_paid_txb_minor INTEGER NOT NULL CHECK (net_paid_txb_minor >= 0),
  credited_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (credited_txb_minor >= 0),
  exception_code TEXT NOT NULL DEFAULT '',
  attempts INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  completed_at TEXT,
  allocated_traffic_bytes INTEGER,
  eligible_unused_bytes INTEGER,
  algorithm_version TEXT NOT NULL DEFAULT ''
);

INSERT INTO purchase_rollovers_next(
  purchase_id,status,traffic_limit_bytes,used_traffic_bytes,remaining_traffic_bytes,
  minimum_remaining_bps,net_paid_txb_minor,credited_txb_minor,exception_code,attempts,
  created_at,updated_at,completed_at,allocated_traffic_bytes,eligible_unused_bytes,algorithm_version
)
SELECT
  purchase_id,status,traffic_limit_bytes,used_traffic_bytes,remaining_traffic_bytes,
  minimum_remaining_bps,net_paid_txb_minor,credited_txb_minor,exception_code,attempts,
  created_at,updated_at,completed_at,allocated_traffic_bytes,eligible_unused_bytes,algorithm_version
FROM purchase_rollovers;

DROP TABLE purchase_rollovers;
ALTER TABLE purchase_rollovers_next RENAME TO purchase_rollovers;
CREATE INDEX purchase_rollovers_status_idx ON purchase_rollovers(status,updated_at);
