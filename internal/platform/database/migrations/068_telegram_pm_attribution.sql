-- Pending outbound text/caption is encrypted and bound to its operation identity.
-- Settled/expired ciphertext is erased; a tiny tombstone prevents legacy downgrade
-- if a terminal operation is explicitly retried. Operation pruning cascades it.
CREATE TABLE telegram_pm_payloads (
    operation_id TEXT PRIMARY KEY REFERENCES provider_operations(id) ON DELETE CASCADE,
    encrypted_payload TEXT NOT NULL,
    expires_at INTEGER NOT NULL
);
CREATE INDEX telegram_pm_payload_expiry_idx ON telegram_pm_payloads(expires_at);

-- Existing unsent outbound copies retain their contents and get a footer reply.
INSERT INTO provider_operation_items(operation_id,item_key,target_type,target_id,created_at,updated_at)
SELECT item.operation_id,'footer','pm_outbound_footer',item.target_id,item.created_at,item.updated_at
FROM provider_operation_items item JOIN provider_operations operation ON operation.id=item.operation_id
WHERE operation.kind='telegram_pm_relay' AND operation.status IN ('queued','processing')
AND item.item_key='relay' AND item.target_type='pm_outbound_message';
