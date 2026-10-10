-- Go migration hook converts legacy reports to the versioned compact envelope.
-- Accounting rows and historical terminal statuses remain unchanged.
DROP INDEX ip_lookup_cache;
CREATE INDEX ip_lookup_cache ON ip_lookup_reports(ip,julianday(completed_at) DESC)
WHERE status IN ('succeeded','partial');
