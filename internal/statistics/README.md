# Product statistics

- `provider.go` defines the narrow queue-backed Remnawave metrics and host contracts.
- `host_rewriter.go` replaces the first standalone decimal token only for hosts linked to exactly one known node and isolates per-host patch failures.
- `host_rewriter_test.go` pins decimal-token boundaries without making provider calls.
- `service_test.go` pins seven-day created-user semantics, 30-day active-member usage, and all-or-stale refresh behavior.
- `service.go` refreshes remote/database partitions independently; `partitions.go` persists and restores last-good values.
- `usage_projection.go` averages each current non-admin member's live 30-day cadence and weighted-node usage projection.
- `node_cache.go` implements the shared on-demand ten-second node cache.
- `host_worker.go` maps queued host-multiplier updates onto provider-operation items.

The aggregate service keeps remote and database partitions independently timestamped so a partial refresh serves the last good partition as stale.
