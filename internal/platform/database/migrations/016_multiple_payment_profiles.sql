PRAGMA foreign_keys = ON;

-- A provider may have several independent accounts. Keep existing rows and
-- remove only the provider-wide uniqueness constraint introduced in 015.
CREATE TABLE payment_profiles_next (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL CHECK (provider IN ('ezpay','bepusdt')),
  provider_name TEXT NOT NULL,
  enabled_channels TEXT NOT NULL DEFAULT '',
  endpoint TEXT NOT NULL,
  merchant_id TEXT NOT NULL DEFAULT '',
  credential_ciphertext TEXT NOT NULL DEFAULT '',
  acknowledgement TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0,1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

INSERT INTO payment_profiles_next(id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at
FROM payment_profiles;

DROP TABLE payment_profiles;
ALTER TABLE payment_profiles_next RENAME TO payment_profiles;
CREATE INDEX payment_profiles_provider_idx ON payment_profiles(provider, enabled, id);
