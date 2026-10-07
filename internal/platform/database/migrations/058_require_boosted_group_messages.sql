-- Old unfinished progress cannot prove boost eligibility. Preserve paid windows,
-- immutable ledger entries, and message identities so deliveries cannot replay.
UPDATE activity_group_message_events
SET counted=0
WHERE EXISTS (
  SELECT 1 FROM activity_group_message_windows AS progress
  WHERE progress.user_id=activity_group_message_events.user_id
    AND progress.local_date=activity_group_message_events.local_date
    AND progress.rewarded_at IS NULL
);

UPDATE activity_group_message_windows
SET message_count=0
WHERE rewarded_at IS NULL;
