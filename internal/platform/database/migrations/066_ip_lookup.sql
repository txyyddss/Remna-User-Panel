ALTER TABLE combos ADD COLUMN ip_lookup_quota INTEGER CHECK(ip_lookup_quota IS NULL OR ip_lookup_quota BETWEEN 0 AND 1000000);

CREATE TABLE ip_lookup_allowances (
    purchase_id TEXT PRIMARY KEY REFERENCES purchases(id) ON DELETE CASCADE,
    total INTEGER NOT NULL CHECK(total >= 0),
    used INTEGER NOT NULL DEFAULT 0 CHECK(used >= 0 AND used <= total)
);

-- One migration gift per member with a live term, even with anomalous overlap.
INSERT INTO ip_lookup_allowances(purchase_id,total)
SELECT id,1 FROM (
    SELECT id,ROW_NUMBER() OVER(PARTITION BY user_id ORDER BY valid_from DESC,id) AS position
    FROM purchases WHERE status='active'
      AND julianday(valid_from)<=julianday('now') AND julianday(valid_until)>julianday('now')
) WHERE position=1;

CREATE TABLE ip_lookup_reports (
    id TEXT PRIMARY KEY,
    ip TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('processing','succeeded','partial','failed')),
    config_json TEXT NOT NULL,
    report_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    completed_at TEXT
);
CREATE UNIQUE INDEX ip_lookup_inflight ON ip_lookup_reports(ip) WHERE status='processing';
CREATE INDEX ip_lookup_cache ON ip_lookup_reports(ip,completed_at DESC,id) WHERE status IN ('succeeded','partial');

CREATE TABLE ip_lookup_checks (
	 id TEXT NOT NULL UNIQUE,
    operation_id TEXT PRIMARY KEY REFERENCES provider_operations(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    report_id TEXT NOT NULL REFERENCES ip_lookup_reports(id),
    allowance_id TEXT REFERENCES ip_lookup_allowances(purchase_id) ON DELETE SET NULL,
    charge_minor INTEGER NOT NULL CHECK(charge_minor >= 0),
    cached INTEGER NOT NULL CHECK(cached IN (0,1)),
    refunded INTEGER NOT NULL DEFAULT 0 CHECK(refunded IN (0,1))
);
CREATE INDEX ip_lookup_check_report ON ip_lookup_checks(report_id);
