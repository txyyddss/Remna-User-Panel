# Outbox payload helpers

- `payload.go` extracts typed target identifiers and validates immutable successful-payment announcement snapshots.
- `kinds.go` owns shared job-kind constants used by persistence and handlers, including durable payment announcements.
