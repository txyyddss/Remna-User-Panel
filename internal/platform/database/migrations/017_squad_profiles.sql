ALTER TABLE squad_product_overrides
  ADD COLUMN profile_json TEXT
  CHECK (profile_json IS NULL OR json_valid(profile_json));
