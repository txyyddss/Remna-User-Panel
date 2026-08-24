CREATE INDEX IF NOT EXISTS abuse_qps_samples_pending
  ON abuse_qps_samples(user_id,reason_name,bucket_at,node_uuid);
