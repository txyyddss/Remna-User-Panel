# Product statistics

- `provider.go` defines the narrow queue-backed Remnawave metrics and host contracts.
- `host_rewriter.go` replaces the first standalone decimal multiplier marked with a lowercase `x` only for hosts linked to exactly one known node, preserving the marker and isolating per-host patch failures.
- `host_rewriter_test.go` pins decimal-token boundaries without making provider calls.
- `service_test.go` pins seven-day created-user semantics and best-effort 30-day active-member usage, including unavailable or missing upstream identities.
- `service.go` refreshes remote/database partitions independently; `partitions.go` persists and restores last-good values.
- `squad_names.go` resolves live queued Remnawave squad names independently of aggregate refreshes and retains the last resolved labels through the local distribution snapshot; `squad_names_test.go` covers mapping and cached-label isolation.
- `usage_projection.go` averages each current non-admin member's live 30-day cadence and weighted-node usage projection, skipping only missing local or deleted upstream identities.
- `node_cache.go` implements the shared on-demand ten-second node cache from the documented live node collection.
- `host_worker.go` maps queued host-multiplier updates onto provider-operation items.

The aggregate service keeps remote and database partitions independently timestamped so a partial refresh serves the last good partition as stale.
