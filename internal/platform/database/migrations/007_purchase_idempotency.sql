ALTER TABLE purchases ADD COLUMN idempotency_key TEXT;
ALTER TABLE purchases ADD COLUMN request_fingerprint TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX purchases_user_idempotency_idx
  ON purchases(user_id,idempotency_key)
  WHERE idempotency_key IS NOT NULL;
