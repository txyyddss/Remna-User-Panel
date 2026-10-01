-- The old core-change reward erased all custom terms. Recover only active
-- purchases with that exact empty-override signature and unambiguous evidence.
-- No balance changes, automatic-renewal enablement, or historical rerolls.
CREATE TEMP TABLE recover_custom_combo AS
WITH rewards AS MATERIALIZED (
  SELECT id,user_id,reward_kind,created_at,
    CASE WHEN json_valid(reward_payload) THEN reward_payload ELSE '{}' END payload
  FROM activity_draw_results
)
SELECT p.id purchase_id,p.user_id,r.id result_id,
  json_extract(r.payload,'$.squadUuids') squads,
  COALESCE(json_extract(r.payload,'$.renewalPriceMinor'),0) price,
  json_extract(r.payload,'$.trafficLimitBytes') traffic,
  COALESCE(json_extract(r.payload,'$.rolloverMinRemainingBps'),0) threshold
FROM purchases p JOIN rewards r ON r.user_id=p.user_id
WHERE p.status IN ('active','activating')
  AND julianday(p.valid_from)<=julianday('now') AND julianday(p.valid_until)>julianday('now')
  AND p.entitlement_squad_uuids IS NULL AND p.entitlement_addon_squad_uuids IS NULL
  AND p.entitlement_traffic_limit_bytes IS NULL AND p.reward_renewal_price_minor IS NULL
  AND p.reward_rollover_min_remaining_bps IS NULL AND p.reward_renewal_traffic_limit_bytes IS NULL
  AND p.reward_traffic_renewal=1
  AND r.reward_kind='entitlement_grant' AND json_extract(r.payload,'$.kind')='entitlement_grant'
  AND julianday(r.created_at)>=julianday(p.created_at)
  AND julianday(r.created_at)>=julianday(p.valid_from) AND julianday(r.created_at)<julianday(p.valid_until)
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
  -- Do not infer an order between custom prizes sharing a settlement timestamp.
  AND NOT EXISTS (SELECT 1 FROM rewards newer WHERE newer.user_id=p.user_id
    AND newer.reward_kind='entitlement_grant' AND newer.id<>r.id
    AND julianday(newer.created_at)>=julianday(r.created_at))
  AND EXISTS (SELECT 1 FROM rewards change WHERE change.user_id=p.user_id
    AND change.reward_kind='core_combo_switch' AND json_extract(change.payload,'$.kind')='core_combo_switch'
    AND json_extract(change.payload,'$.comboId')=p.combo_id
    AND julianday(change.created_at)>=julianday(r.created_at)
    AND julianday(change.created_at)<julianday(p.valid_until))
  -- Traffic changes and administrator edits require individual review; never
  -- overwrite their effects by guessing the intended current/renewal allowance.
  AND NOT EXISTS (SELECT 1 FROM rewards traffic WHERE traffic.user_id=p.user_id
    AND traffic.reward_kind='traffic_grant' AND julianday(traffic.created_at)>=julianday(r.created_at))
  AND NOT EXISTS (SELECT 1 FROM purchase_addon_adjustments addon WHERE addon.purchase_id=p.id
    AND julianday(addon.created_at)>=julianday(r.created_at))
  AND NOT EXISTS (SELECT 1 FROM audit_events a WHERE julianday(a.created_at)>=julianday(r.created_at)
    AND ((a.target_type='purchase' AND a.target_id=p.id)
      OR (a.target_type='user' AND a.target_id=p.user_id)
      OR (a.target_type='database_table' AND (a.target_id='purchases' OR a.target_id LIKE 'purchases:%'))))
  AND NOT EXISTS (SELECT 1 FROM purchases other WHERE other.id<>p.id AND other.user_id=p.user_id
    AND julianday(other.valid_from)<=julianday(r.created_at) AND julianday(other.valid_until)>julianday(r.created_at))
  AND NOT EXISTS (SELECT 1 FROM purchases successor WHERE successor.auto_renew_source_purchase_id=p.id)
  AND NOT EXISTS (SELECT 1 FROM purchase_rollovers rollover WHERE rollover.purchase_id=p.id AND rollover.status<>'pending');

UPDATE purchases AS p SET
  entitlement_squad_uuids=fix.squads,entitlement_addon_squad_uuids='[]',
  entitlement_traffic_limit_bytes=fix.traffic,reward_renewal_traffic_limit_bytes=fix.traffic,
  reward_renewal_price_minor=fix.price,reward_rollover_min_remaining_bps=fix.threshold,
  updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM recover_custom_combo fix WHERE p.id=fix.purchase_id;

UPDATE purchase_rollovers AS rollover SET traffic_limit_bytes=fix.traffic,
  minimum_remaining_bps=fix.threshold,updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM recover_custom_combo fix WHERE rollover.purchase_id=fix.purchase_id AND rollover.status='pending';

INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at)
SELECT lower(hex(randomblob(16))),NULL,'activity.custom_combo_recovered','purchase',purchase_id,
  json_object('sourceResultId',result_id,'migration','054_restore_custom_combo_after_core_reward.sql'),
  strftime('%Y-%m-%dT%H:%M:%fZ','now') FROM recover_custom_combo;

INSERT INTO outbox_jobs(id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at)
SELECT lower(hex(randomblob(16))),'remna_sync_user',json_object('userId',fix.user_id),'pending',0,
  strftime('%Y-%m-%dT%H:%M:%fZ','now'),'',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM (SELECT DISTINCT user_id FROM recover_custom_combo) fix
WHERE NOT EXISTS (SELECT 1 FROM outbox_jobs job WHERE job.kind='remna_sync_user'
  AND job.payload=json_object('userId',fix.user_id) AND job.status IN ('pending','processing'));

DROP TABLE recover_custom_combo;
