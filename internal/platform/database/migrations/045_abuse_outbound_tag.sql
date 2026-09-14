ALTER TABLE abuse_policy ADD COLUMN outbound_tag TEXT NOT NULL DEFAULT 'direct'
  CHECK(length(outbound_tag) BETWEEN 1 AND 120);
