CREATE TRIGGER provider_operation_items_open_host_guard
BEFORE INSERT ON provider_operation_items
WHEN NEW.target_type='remnawave_host'
  AND EXISTS (
    SELECT 1 FROM provider_operations incoming
    WHERE incoming.id=NEW.operation_id AND incoming.kind='host_remark_update'
  )
  AND EXISTS (
    SELECT 1 FROM provider_operation_items existing_item
    JOIN provider_operations existing_operation
      ON existing_operation.id=existing_item.operation_id
    WHERE existing_item.target_type='remnawave_host'
      AND existing_item.target_id=NEW.target_id
      AND existing_operation.kind='host_remark_update'
      AND existing_operation.status IN ('queued','processing','pending_review','partial')
      AND existing_operation.id<>NEW.operation_id
  )
BEGIN
  SELECT RAISE(ABORT,'open host remark operation exists');
END;
