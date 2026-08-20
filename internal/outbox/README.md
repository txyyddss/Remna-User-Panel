# Outbox payload helpers

- `payload.go` extracts typed target identifiers and validates immutable
  successful-payment announcement snapshots. The optional `providerName`
  field preserves the administrator label while legacy queued payloads retain
  the provider fallback.
- `payload_test.go` covers new and legacy payment-announcement JSON payloads.
- `kinds.go` owns shared job-kind constants used by persistence and handlers, including durable payment announcements.
