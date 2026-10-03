-- Custom awards did not snapshot their reset strategy. A later core change
-- therefore changed cadence even after 054/055 restored the other custom terms.
-- Fill only absent overrides; explicit cadence choices remain authoritative.
CREATE TEMP TABLE custom_cadence_grants AS
SELECT r.id,r.user_id,r.draw_id,r.created_at,
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
WHERE r.reward_kind='entitlement_grant';
ALTER TABLE custom_cadence_grants ADD COLUMN sort_key TEXT;
UPDATE custom_cadence_grants SET sort_key=stamp_key || ':' || COALESCE(ticket_created_at,'') || ':' || COALESCE(ticket_id,'');
CREATE INDEX custom_cadence_grant_order ON custom_cadence_grants(user_id,sort_key);
UPDATE custom_cadence_grants AS r SET ambiguous=1
WHERE EXISTS (SELECT 1 FROM custom_cadence_grants other WHERE other.user_id=r.user_id
  AND other.id<>r.id AND other.stamp_key=r.stamp_key
  AND (other.draw_id<>r.draw_id OR other.ticket_id IS NULL OR r.ticket_id IS NULL));

-- Follow automatic-renewal references so a live successor can recover the
-- original award's cadence without rewriting its expired predecessors.
CREATE TEMP TABLE custom_cadence_lineage AS
WITH RECURSIVE lineage(purchase_id,ancestor_id) AS (
  SELECT id,id FROM purchases WHERE status IN ('active','activating')
    AND reward_renewal_price_minor IS NOT NULL AND entitlement_reset_strategy IS NULL
    AND julianday(valid_from)<=julianday('now') AND julianday(valid_until)>julianday('now')
  UNION
  SELECT lineage.purchase_id,parent.id FROM lineage
    JOIN purchases child ON child.id=lineage.ancestor_id
    JOIN purchases parent ON parent.id=child.auto_renew_source_purchase_id AND parent.user_id=child.user_id
)
SELECT * FROM lineage;

CREATE TEMP TABLE custom_cadence_repairs AS
SELECT p.id purchase_id,p.user_id,r.id result_id,c.id source_combo_id,c.reset_strategy
FROM purchases p JOIN custom_cadence_grants r ON r.user_id=p.user_id
JOIN combos c ON c.id=json_extract(r.payload,'$.comboId')
WHERE p.id IN (SELECT purchase_id FROM custom_cadence_lineage)
  AND r.ambiguous=0 AND r.sort_key=(SELECT MAX(latest.sort_key) FROM custom_cadence_grants latest WHERE latest.user_id=p.user_id)
  AND json_extract(r.payload,'$.kind')='entitlement_grant'
  AND json_type(r.payload,'$.comboId')='text' AND json_type(r.payload,'$.squadUuids')='array'
  AND json(p.entitlement_squad_uuids)=json_extract(r.payload,'$.squadUuids')
  AND p.reward_renewal_price_minor=COALESCE(json_extract(r.payload,'$.renewalPriceMinor'),0)
  AND p.reward_rollover_min_remaining_bps=COALESCE(json_extract(r.payload,'$.rolloverMinRemainingBps'),0)
  AND c.reset_strategy IN ('DAY','WEEK','MONTH_ROLLING')
  AND EXISTS (SELECT 1 FROM custom_cadence_lineage link JOIN purchases ancestor ON ancestor.id=link.ancestor_id
    WHERE link.purchase_id=p.id AND julianday(r.created_at)>=julianday(ancestor.created_at)
      AND julianday(r.created_at)>=julianday(ancestor.valid_from) AND julianday(r.created_at)<julianday(ancestor.valid_until))
  AND NOT EXISTS (SELECT 1 FROM audit_events a WHERE julianday(a.created_at)>=julianday(r.created_at)
    AND ((a.target_type='purchase' AND a.target_id IN (SELECT ancestor_id FROM custom_cadence_lineage WHERE purchase_id=p.id)
      AND a.action<>'activity.custom_combo_recovered')
      OR (a.target_type='database_table' AND (a.target_id='purchases' OR a.target_id LIKE 'purchases:%'))))
  AND NOT EXISTS (SELECT 1 FROM purchase_rollovers rollover WHERE rollover.purchase_id=p.id AND rollover.status<>'pending');

UPDATE purchases AS p SET entitlement_reset_strategy=fix.reset_strategy,
  updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM custom_cadence_repairs fix WHERE p.id=fix.purchase_id;
INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at)
SELECT lower(hex(randomblob(16))),NULL,'activity.custom_combo_cadence_recovered','purchase',purchase_id,
  json_object('sourceResultId',result_id,'sourceComboId',source_combo_id,'resetStrategy',reset_strategy,
    'migration','056_restore_custom_combo_reset_cadence.sql'),
  strftime('%Y-%m-%dT%H:%M:%fZ','now') FROM custom_cadence_repairs;
INSERT INTO outbox_jobs(id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at)
SELECT lower(hex(randomblob(16))),'remna_sync_user',json_object('userId',fix.user_id),'pending',0,
  strftime('%Y-%m-%dT%H:%M:%fZ','now'),'',strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')
FROM (SELECT DISTINCT user_id FROM custom_cadence_repairs) fix
WHERE NOT EXISTS (SELECT 1 FROM outbox_jobs job WHERE job.kind='remna_sync_user'
  AND job.payload=json_object('userId',fix.user_id) AND job.status IN ('pending','processing'));
DROP TABLE custom_cadence_repairs;
DROP TABLE custom_cadence_lineage;
DROP TABLE custom_cadence_grants;
