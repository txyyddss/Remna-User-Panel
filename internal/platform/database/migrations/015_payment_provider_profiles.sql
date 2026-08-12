PRAGMA foreign_keys = ON;

CREATE TABLE payment_profiles (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL UNIQUE CHECK (provider IN ('ezpay','bepusdt')),
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

INSERT INTO payment_profiles(id,provider,provider_name,enabled_channels,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT provider, provider, provider,
  COALESCE(group_concat(CASE WHEN enabled=1 THEN rail END, ','), ''),
  COALESCE(MAX(NULLIF(endpoint,'')), ''),
  COALESCE(MAX(NULLIF(merchant_id,'')), ''),
  COALESCE(MAX(NULLIF(credential_ciphertext,'')), ''),
  COALESCE(MAX(NULLIF(acknowledgement,'')), ''),
  MAX(enabled), MIN(created_at), MAX(updated_at)
FROM payment_rail_profiles
GROUP BY provider;

DROP INDEX IF EXISTS payment_rail_profiles_enabled_idx;
DROP TABLE payment_rail_profiles;
