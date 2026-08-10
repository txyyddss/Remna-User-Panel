-- Provider responses are represented by the narrow payment-order columns used
-- for display and callback validation. Raw provider documents are never needed
-- for settlement and must not remain durable.
UPDATE payment_orders
SET provider_payload='{}',
    payment_url=CASE
      WHEN status IN ('paid','expired','failed','refunded') OR cancelled_at IS NOT NULL THEN NULL
      ELSE payment_url
    END,
    qr_payload=CASE
      WHEN status IN ('paid','failed','refunded') THEN NULL
      WHEN (status='expired' OR cancelled_at IS NOT NULL) AND receiving_address IS NOT NULL THEN NULL
      ELSE qr_payload
    END;
