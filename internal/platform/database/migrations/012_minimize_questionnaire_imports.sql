-- Settled imports need only their aggregate analysis and settlement report.
UPDATE questionnaire_imports
SET raw_csv=x'', sample_rows_json='[]'
WHERE status='settled';

-- Older builds left the import queued/processing when its durable job exhausted
-- retries. Align those rows with the explicit-retry lifecycle.
UPDATE questionnaire_imports AS import
SET status='failed',
    last_error=COALESCE((
      SELECT job.last_error FROM outbox_jobs job
      WHERE job.kind='questionnaire_settlement'
        AND job.status='failed'
        AND json_extract(job.payload,'$.importId')=import.id
      ORDER BY job.updated_at DESC LIMIT 1
    ),last_error),
    updated_at=COALESCE((
      SELECT job.updated_at FROM outbox_jobs job
      WHERE job.kind='questionnaire_settlement'
        AND job.status='failed'
        AND json_extract(job.payload,'$.importId')=import.id
      ORDER BY job.updated_at DESC LIMIT 1
    ),updated_at)
WHERE status IN ('queued','processing')
  AND EXISTS (
    SELECT 1 FROM outbox_jobs job
    WHERE job.kind='questionnaire_settlement'
      AND job.status='failed'
      AND json_extract(job.payload,'$.importId')=import.id
  );
