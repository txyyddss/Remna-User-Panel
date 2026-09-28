-- Draw jobs carry a numeric revision. The former string-only TargetID decoder
-- rejected these payloads before any Telegram edit or settlement could run.
-- Restore exhausted attempts and long backoffs only for this specific failure.
UPDATE outbox_jobs
SET status='pending', attempts=0, last_error='',
    available_at=strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE status IN ('failed','pending')
  AND kind IN ('draw_telegram_update','draw_raffle_settle')
  AND instr(last_error, 'decode ' || kind || ' job payload: json: cannot unmarshal number into Go value of type string') > 0
  AND CASE WHEN json_valid(payload) THEN
    json_type(payload,'$.drawId')='text'
    AND length(trim(json_extract(payload,'$.drawId'))) > 0
    AND json_type(payload,'$.revision')='integer'
  ELSE 0 END
  AND (status='pending' OR (
    id=(SELECT MIN(candidate.id) FROM outbox_jobs candidate
      WHERE candidate.kind=outbox_jobs.kind AND candidate.payload=outbox_jobs.payload
        AND candidate.status='failed'
        AND instr(candidate.last_error, 'decode ' || candidate.kind || ' job payload: json: cannot unmarshal number into Go value of type string') > 0)
    AND NOT EXISTS (SELECT 1 FROM outbox_jobs active
      WHERE active.kind=outbox_jobs.kind AND active.payload=outbox_jobs.payload
        AND active.status IN ('pending','processing'))
  ));
