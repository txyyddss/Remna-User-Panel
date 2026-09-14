ALTER TABLE abuse_policy ADD COLUMN outbound_tags TEXT NOT NULL DEFAULT 'direct'
  CHECK(length(outbound_tags) BETWEEN 1 AND 3024);
UPDATE abuse_policy SET outbound_tags=outbound_tag WHERE id=1;
