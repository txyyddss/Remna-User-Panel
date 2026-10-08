CREATE INDEX node_compensation_events_created_idx ON node_compensation_events(created_at);
CREATE INDEX outbox_failed_retention_idx ON outbox_jobs(updated_at,id) WHERE status='failed';
