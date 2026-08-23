# Provider operations

This package defines provider-neutral command, item, replay, and connection-scan
metadata. SQLite persists request fingerprints and sanitized lifecycle facts;
the outbox performs external calls. Raw connection IPs and signed connection
handles remain short-lived service data and are never stored in these records.

- `types.go` defines operation states, commands, receipts, items, and results.
- `connections.go` defines metadata-only connection scan inputs and progress.
- `validation.go` canonicalizes commands and bounded JSON result objects.
- `dispatcher.go` routes the one shared outbox lane to kind-specific handlers.
- `admin_kinds.go` defines administrator command kind names shared at composition, including reviewed node compensation.
- `command_kinds.go` defines subscription, Emby, questionnaire, retry, and payment-refund command names.
- `dispatcher_test.go` covers kind routing without provider calls.
