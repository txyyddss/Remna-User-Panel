-- Follow up 054 without changing its deployed behavior. Raffle timestamps are
-- shared; settled tickets retain the exact created_at,id application order.
CREATE TEMP TABLE custom_recovery_rewards AS
SELECT r.id,r.user_id,r.draw_id,r.reward_kind,r.created_at,
  CASE WHEN json_valid(r.reward_payload) THEN r.reward_payload ELSE '{}' END payload,
  substr(r.created_at,1,19) || substr(
    CASE WHEN substr(r.created_at,20,1)='.' THEN substr(r.created_at,21,length(r.created_at)-21) ELSE '' END
    || '000000000',1,9) stamp_key,
  t.id ticket_id,t.created_at ticket_created_at,
  CASE WHEN r.idempotency_key LIKE 'raffle:%' AND t.id IS NULL THEN 1 ELSE 0 END ambiguous
FROM activity_draw_results r LEFT JOIN activity_raffle_tickets t
  ON t.result_id=r.id AND t.user_id=r.user_id AND t.draw_id=r.draw_id
  AND t.status='settled' AND r.idempotency_key='raffle:' || t.id
  AND (SELECT COUNT(*) FROM activity_raffle_tickets duplicate WHERE duplicate.result_id=r.id)=1
WHERE r.reward_kind IN ('entitlement_grant','core_combo_switch','traffic_grant');
ALTER TABLE custom_recovery_rewards ADD COLUMN sort_key TEXT;
UPDATE custom_recovery_rewards SET sort_key=stamp_key || ':' || COALESCE(ticket_created_at,'') || ':' || COALESCE(ticket_id,'');
CREATE INDEX custom_recovery_reward_order ON custom_recovery_rewards(user_id,sort_key);
UPDATE custom_recovery_rewards AS r SET ambiguous=1
WHERE EXISTS (SELECT 1 FROM custom_recovery_rewards other WHERE other.user_id=r.user_id
  AND other.id<>r.id AND other.stamp_key=r.stamp_key
  AND (other.draw_id<>r.draw_id OR other.ticket_id IS NULL OR r.ticket_id IS NULL));

CREATE TEMP TABLE custom_recovery_candidates AS
SELECT p.id purchase_id,p.user_id,r.id result_id,r.payload,
  r.sort_key custom_order,change.sort_key core_order,c.traffic_limit_bytes core_traffic,
  p.entitlement_traffic_limit_bytes stored_traffic,p.reward_renewal_traffic_limit_bytes stored_renewal_traffic
FROM purchases p JOIN combos c ON c.id=p.combo_id
JOIN custom_recovery_rewards r ON r.user_id=p.user_id AND r.reward_kind='entitlement_grant'
JOIN custom_recovery_rewards change ON change.user_id=p.user_id AND change.reward_kind='core_combo_switch'
WHERE p.status IN ('active','activating')
  AND julianday(p.valid_from)<=julianday('now') AND julianday(p.valid_until)>julianday('now')
  AND p.entitlement_squad_uuids IS NULL AND p.entitlement_addon_squad_uuids IS NULL
  AND p.reward_renewal_price_minor IS NULL AND p.reward_rollover_min_remaining_bps IS NULL
  AND julianday(r.created_at)>=julianday(p.created_at)
  AND julianday(r.created_at)>=julianday(p.valid_from) AND julianday(r.created_at)<julianday(p.valid_until)
  AND r.sort_key=(SELECT MAX(latest.sort_key) FROM custom_recovery_rewards latest
    WHERE latest.user_id=p.user_id AND latest.reward_kind='entitlement_grant')
  AND change.sort_key=(SELECT MAX(latest.sort_key) FROM custom_recovery_rewards latest
    WHERE latest.user_id=p.user_id AND latest.reward_kind='core_combo_switch')
  AND change.sort_key>r.sort_key AND julianday(change.created_at)<julianday(p.valid_until)
  AND json_extract(change.payload,'$.kind')='core_combo_switch'
  AND json_extract(change.payload,'$.comboId')=p.combo_id
  AND NOT EXISTS (SELECT 1 FROM custom_recovery_rewards e WHERE e.user_id=p.user_id
    AND e.sort_key>=r.sort_key AND (e.ambiguous=1 OR julianday(e.created_at) IS NULL))
  AND json_extract(r.payload,'$.kind')='entitlement_grant'
  AND json_type(r.payload,'$.squadUuids')='array'
  AND json_array_length(r.payload,'$.squadUuids') BETWEEN 1 AND 100
  AND NOT EXISTS (SELECT 1 FROM json_each(r.payload,'$.squadUuids') s WHERE s.type<>'text' OR trim(s.value)='')
  AND (SELECT COUNT(DISTINCT trim(value)) FROM json_each(r.payload,'$.squadUuids'))=json_array_length(r.payload,'$.squadUuids')
  AND COALESCE(json_type(r.payload,'$.renewalPriceMinor'),'integer')='integer'
  AND COALESCE(json_extract(r.payload,'$.renewalPriceMinor'),0) BETWEEN 0 AND 10000000000
  AND json_type(r.payload,'$.trafficLimitBytes')='integer'
  AND json_extract(r.payload,'$.trafficLimitBytes') BETWEEN 1 AND 1073741824000000
  AND COALESCE(json_type(r.payload,'$.rolloverMinRemainingBps'),'integer')='integer'
  AND COALESCE(json_extract(r.payload,'$.rolloverMinRemainingBps'),0) BETWEEN 0 AND 10000
  AND NOT EXISTS (SELECT 1 FROM custom_recovery_rewards e WHERE e.user_id=p.user_id AND e.sort_key>r.sort_key
    AND e.reward_kind='traffic_grant' AND NOT (
      COALESCE(json_extract(e.payload,'$.kind')='traffic_grant',0)
      AND COALESCE(json_type(e.payload,'$.resolvedValue'),'missing')='integer'
      AND json_extract(e.payload,'$.resolvedValue') BETWEEN -1000000 AND 1000000
      AND COALESCE(json_type(e.payload,'$.includeInRenewal'),'false') IN ('true','false')))
  AND NOT EXISTS (SELECT 1 FROM purchase_addon_adjustments addon WHERE addon.purchase_id=p.id
    AND julianday(addon.created_at)>=julianday(r.created_at))
  AND NOT EXISTS (SELECT 1 FROM audit_events a WHERE julianday(a.created_at)>=julianday(r.created_at)
    AND ((a.target_type='purchase' AND a.target_id=p.id)
      OR (a.target_type='database_table' AND (a.target_id='purchases' OR a.target_id LIKE 'purchases:%'))))
  AND NOT EXISTS (SELECT 1 FROM purchases other WHERE other.id<>p.id AND other.user_id=p.user_id
    AND julianday(other.valid_from)<=julianday(r.created_at) AND julianday(other.valid_until)>julianday(r.created_at))
  AND NOT EXISTS (SELECT 1 FROM purchases successor WHERE successor.auto_renew_source_purchase_id=p.id)
  AND NOT EXISTS (SELECT 1 FROM purchase_rollovers rollover WHERE rollover.purchase_id=p.id AND rollover.status<>'pending');

CREATE TEMP TABLE custom_recovery_terms AS
WITH deltas AS (
  SELECT candidate.purchase_id,
    COALESCE(SUM(json_extract(e.payload,'$.resolvedValue')),0) current_delta,
    COALESCE(SUM(CASE WHEN json_extract(e.payload,'$.includeInRenewal')=1 THEN json_extract(e.payload,'$.resolvedValue') ELSE 0 END),0) renewal_delta,
    COUNT(CASE WHEN e.sort_key>candidate.core_order THEN 1 END) post_core_count,
    COALESCE(SUM(CASE WHEN e.sort_key>candidate.core_order THEN json_extract(e.payload,'$.resolvedValue') ELSE 0 END),0) post_core_delta,
    COUNT(CASE WHEN e.sort_key>candidate.core_order AND json_extract(e.payload,'$.includeInRenewal')=1 THEN 1 END) post_core_recurring_count,
    COALESCE(SUM(CASE WHEN e.sort_key>candidate.core_order AND json_extract(e.payload,'$.includeInRenewal')=1 THEN json_extract(e.payload,'$.resolvedValue') ELSE 0 END),0) post_core_recurring_delta
  FROM custom_recovery_candidates candidate LEFT JOIN custom_recovery_rewards e
    ON e.user_id=candidate.user_id AND e.sort_key>candidate.custom_order AND e.reward_kind='traffic_grant'
  GROUP BY candidate.purchase_id
)
SELECT candidate.purchase_id,candidate.user_id,candidate.result_id,
  json_extract(candidate.payload,'$.squadUuids') squads,
  COALESCE(json_extract(candidate.payload,'$.renewalPriceMinor'),0) price,
  COALESCE(json_extract(candidate.payload,'$.rolloverMinRemainingBps'),0) threshold,
  json_extract(candidate.payload,'$.trafficLimitBytes') + deltas.current_delta*1073741824 traffic,
  json_extract(candidate.payload,'$.trafficLimitBytes') + deltas.renewal_delta*1073741824 renewal_traffic
FROM custom_recovery_candidates candidate JOIN deltas ON deltas.purchase_id=candidate.purchase_id
-- Verify that existing traffic matches the old core-reset behavior before
-- rebuilding it. This protects unexplained/manual partial changes.
WHERE ((deltas.post_core_count=0 AND candidate.stored_traffic IS NULL)
    OR (deltas.post_core_count>0 AND candidate.stored_traffic=candidate.core_traffic+deltas.post_core_delta*1073741824))
  AND ((deltas.post_core_recurring_count=0 AND candidate.stored_renewal_traffic IS NULL)
    OR (deltas.post_core_recurring_count>0 AND candidate.stored_renewal_traffic=candidate.core_traffic+deltas.post_core_recurring_delta*1073741824));
DELETE FROM custom_recovery_terms WHERE typeof(traffic)<>'integer' OR typeof(renewal_traffic)<>'integer'
  OR traffic<=0 OR renewal_traffic<=0;

UPDATE purchases AS p SET entitlement_squad_uuids=fix.squads,entitlement_addon_squad_uuids='[]',
  entitlement_traffic_limit_bytes=fix.traffic,reward_renewal_traffic_limit_bytes=fix.renewal_traffic,
  reward_traffic_renewal=CASE WHEN fix.traffic=fix.renewal_traffic THEN 1 ELSE 0 END,
  reward_renewal_price_minor=fix.price,reward_rollover_min_remaining_bps=fix.threshold,
  updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM custom_recovery_terms fix WHERE p.id=fix.purchase_id;
UPDATE purchase_rollovers AS rollover SET traffic_limit_bytes=fix.traffic,
  minimum_remaining_bps=fix.threshold,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM custom_recovery_terms fix WHERE rollover.purchase_id=fix.purchase_id AND rollover.status='pending';
INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at)
SELECT lower(hex(randomblob(16))),NULL,'activity.custom_combo_recovered','purchase',purchase_id,
  json_object('sourceResultId',result_id,'migration','055_restore_ordered_custom_combo_rewards.sql'),
  strftime('%Y-%m-%dT%H:%M:%fZ','now') FROM custom_recovery_terms;
INSERT INTO outbox_jobs(id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at)
SELECT lower(hex(randomblob(16))),'remna_sync_user',json_object('userId',fix.user_id),'pending',0,
  strftime('%Y-%m-%dT%H:%M:%fZ','now'),'',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM (SELECT DISTINCT user_id FROM custom_recovery_terms) fix
WHERE NOT EXISTS (SELECT 1 FROM outbox_jobs job WHERE job.kind='remna_sync_user'
  AND job.payload=json_object('userId',fix.user_id) AND job.status IN ('pending','processing'));
DROP TABLE custom_recovery_terms;
DROP TABLE custom_recovery_candidates;
DROP TABLE custom_recovery_rewards;
