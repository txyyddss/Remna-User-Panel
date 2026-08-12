PRAGMA foreign_keys = ON;

ALTER TABLE squad_product_overrides ADD COLUMN stock_limit INTEGER
  CHECK (stock_limit IS NULL OR stock_limit >= 0);
ALTER TABLE purchases ADD COLUMN renewal_batch_id TEXT;
ALTER TABLE purchases ADD COLUMN renewal_index INTEGER;
CREATE TABLE renewal_batches (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source_purchase_id TEXT NOT NULL REFERENCES purchases(id),
  idempotency_key TEXT NOT NULL,
  request_fingerprint TEXT NOT NULL,
  term_count INTEGER NOT NULL CHECK (term_count BETWEEN 1 AND 6),
  charged_txb_minor INTEGER NOT NULL CHECK (charged_txb_minor >= 0),
  created_at TEXT NOT NULL,
  UNIQUE(user_id, idempotency_key)
);
CREATE UNIQUE INDEX purchases_renewal_term_idx ON purchases(renewal_batch_id, renewal_index)
  WHERE renewal_batch_id IS NOT NULL;
ALTER TABLE purchase_rollovers ADD COLUMN allocated_traffic_bytes INTEGER;
ALTER TABLE purchase_rollovers ADD COLUMN eligible_unused_bytes INTEGER;
ALTER TABLE purchase_rollovers ADD COLUMN algorithm_version TEXT NOT NULL DEFAULT '';
CREATE TABLE payment_rail_profiles (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL CHECK (provider IN ('ezpay','bepusdt')),
  rail TEXT NOT NULL,
  channel_name TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  merchant_id TEXT NOT NULL DEFAULT '',
  credential_ciphertext TEXT NOT NULL DEFAULT '',
  acknowledgement TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 0 CHECK (enabled IN (0,1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(provider, rail)
);
CREATE INDEX payment_rail_profiles_enabled_idx ON payment_rail_profiles(provider, enabled, rail);

INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'ezpay:alipay','ezpay','alipay','Alipay',COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.base_url'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.merchant_id'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.key'),''),'',CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.methods'),'')||',',',alipay,')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now');
INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'ezpay:wxpay','ezpay','wxpay','WeChat Pay',COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.base_url'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.merchant_id'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.key'),''),'',CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.methods'),'')||',',',wxpay,')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now');
INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'ezpay:qqpay','ezpay','qqpay','QQ Wallet',COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.base_url'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.merchant_id'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.key'),''),'',CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.methods'),'')||',',',qqpay,')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now');
INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'ezpay:bank','ezpay','bank','UnionPay',COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.base_url'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.merchant_id'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.key'),''),'',CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.methods'),'')||',',',bank,')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now');
INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'ezpay:jdpay','ezpay','jdpay','JD Pay',COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.base_url'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.merchant_id'),''),COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.key'),''),'',CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.ezpay.methods'),'')||',',',jdpay,')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now');

INSERT OR IGNORE INTO payment_rail_profiles(id,provider,rail,channel_name,endpoint,merchant_id,credential_ciphertext,acknowledgement,enabled,created_at,updated_at)
SELECT 'bepusdt:'||rail,'bepusdt',rail,channel_name,COALESCE((SELECT value FROM settings WHERE key='billing.bepusdt.base_url'),''),'',COALESCE((SELECT value FROM settings WHERE key='billing.bepusdt.api_token'),''),COALESCE((SELECT value FROM settings WHERE key='billing.bepusdt.ack'),''),CASE WHEN COALESCE((SELECT value FROM settings WHERE key='billing.bepusdt.enabled'),'')='true' AND instr(','||COALESCE((SELECT value FROM settings WHERE key='billing.bepusdt.methods'),'')||',',','||rail||',')>0 THEN 1 ELSE 0 END,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM (SELECT 'usdt.trc20' AS rail, 'USDT TRC20' AS channel_name UNION ALL SELECT 'usdt.erc20', 'USDT ERC20' UNION ALL SELECT 'usdt.polygon', 'USDT Polygon' UNION ALL SELECT 'usdt.bep20', 'USDT BEP20' UNION ALL SELECT 'usdt.aptos', 'USDT Aptos' UNION ALL SELECT 'usdt.solana', 'USDT Solana' UNION ALL SELECT 'usdt.xlayer', 'USDT X-Layer' UNION ALL SELECT 'usdt.arbitrum', 'USDT Arbitrum One' UNION ALL SELECT 'usdt.plasma', 'USDT Plasma' UNION ALL SELECT 'usdt.ton', 'USDT TON');
