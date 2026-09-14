# Abuse detector

- `types.go` defines the safe persisted and transport-facing detector values,
  including the action-specific duration rule used by administration.
- `parser.go` accepts only Xray domain connection accepts for the administrator-selected outbound-tag list and fingerprints lines without retaining them.
- `outbound_tags.go` normalizes the bounded tag list and rejects malformed or duplicate entries before persistence.
- `ingestion.go` bounds report parsing, resolves only reported remote identities, and persists normalized events without applying policy.
- `service.go` owns service construction plus compiled RE2 and token helpers.
- `processing.go` snapshots policy and runs durable grace-delayed batches.
- `evaluation.go` carries configurable cross-task streak boundaries and emits one incident per continuous streak.
- `evaluation_aggregate.go` deduplicates cross-node events and builds reason/second QPS buckets plus compact 30-minute rollups.
- `admin.go` exposes validated member and administrator detector operations.
- `node_keys.go` provisions, encrypts, copies, and rotates node agent keys.
- `repository.go` declares the persistence boundary used by the detector.
- `parser_test.go` and `outbound_tags_test.go` cover configured-tag acceptance plus malformed, duplicate, IP, other-tag, and error rejection.
- `service_test.go` covers inclusive QPS-limit qualification; database processing tests cover configurable streak continuation, one-shot incidents, gap reset, policy changes, claims, and rollups.
