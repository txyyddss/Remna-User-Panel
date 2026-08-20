# Outbox payload helpers

- `payload.go` extracts typed target identifiers and validates immutable
  successful-payment announcement snapshots. The optional `providerName`
  field preserves the administrator label while legacy queued payloads retain
  the provider fallback.
- `payload_test.go` covers new and legacy payment-announcement JSON payloads.
- `kinds.go` owns shared job-kind constants used by persistence and handlers, including durable payment announcements.
- `user_notification.go` validates immutable, locale-aware private-chat event
  snapshots used by the durable user-notification worker.
- `user_notification_test.go` covers canonical payload encode/decode behavior.
