PRAGMA foreign_keys = ON;

ALTER TABLE combos DROP COLUMN rollover_max_txb_minor;
ALTER TABLE purchase_rollovers DROP COLUMN maximum_txb_minor;

ALTER TABLE squad_product_overrides ADD COLUMN activation_required INTEGER NOT NULL DEFAULT 0
  CHECK (activation_required IN (0,1));
ALTER TABLE squad_product_overrides ADD COLUMN activation_code_hash TEXT;
