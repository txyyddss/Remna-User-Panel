ALTER TABLE squad_product_overrides ADD COLUMN geocheck_disabled INTEGER NOT NULL DEFAULT 0 CHECK (geocheck_disabled IN (0, 1));
