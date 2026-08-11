# Catalog domain

This package builds the live catalog, validates purchase selections against current Remnawave squads, and serves member purchase and dashboard workflows.

- `service.go` defines repository/provider contracts and coordinates catalog hydration, quotes, purchases, process-local dashboard caching, and subscription revocation without durable bearer storage.
- `nodes.go` resolves the queued upstream node preview for the exact combo and optional-squad selection used by a quote.
- `nodes_test.go` verifies quote node projection, filtering, ordering, selection deduplication, and provider failure handling.
- `usage.go` validates a member-selected UTC range and projects live, bounded per-node traffic through the optional queued provider reader.
- `service_test.go` verifies forwarding, validation errors, dashboard freshness, and subscription revocation behavior.
- `live_catalog_test.go` verifies missing live squads are rejected during catalog hydration and purchase validation.
