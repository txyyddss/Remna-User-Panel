CREATE TABLE purchase_addon_adjustments (
  id TEXT PRIMARY KEY,
  purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
  idempotency_key TEXT NOT NULL CHECK (length(idempotency_key) BETWEEN 1 AND 128),
  request_fingerprint TEXT NOT NULL CHECK (length(request_fingerprint) BETWEEN 16 AND 128),
  charged_txb_minor INTEGER NOT NULL CHECK (charged_txb_minor >= 0),
  created_at TEXT NOT NULL,
  UNIQUE(purchase_id,idempotency_key)
);
CREATE INDEX purchase_addon_adjustments_purchase_created_idx
  ON purchase_addon_adjustments(purchase_id,created_at DESC,id);
