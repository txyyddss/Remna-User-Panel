# Catalog domain

This package builds the live catalog, validates purchase selections against current Remnawave squads, and serves member purchase and dashboard workflows.

- `service.go` defines repository/provider contracts and coordinates quotes, purchases, process-local dashboard caching, and subscription revocation without durable bearer storage.
- `catalog.go` hydrates sparse local merchandising and typed squad profiles with live Remnawave squad identities and read-only node metadata.
- `automatic_renewal.go` owns member toggle/status projections, catalog blocking, due-cycle queue revalidation, one-successor commits, and failure notices. It keeps provider access behind the existing queued adapter boundary.
- `automatic_renewal_test.go` covers catalog blocking and enablement eligibility without running provider calls.
- `renewals.go` retains the internal legacy batch implementation only; manual renewal is no longer a public member flow.
- `cancellation.go` exposes the authenticated member operation for cancelling a queued purchase through the transactional store refund path.
- `dashboard.go` composes current balance and entitlement data with cached or live upstream statistics and handles subscription URL revocation.
- `nodes.go` resolves the queued upstream node preview for the exact combo and optional-squad selection used by a quote, and projects enabled, display-only node metadata into the catalog for Home's usage multiplier lookup. Node metadata is never persisted.
- `nodes_test.go` verifies quote node projection, filtering, ordering, selection deduplication, and provider failure handling.
- `usage.go` validates a member-selected UTC range and projects live, bounded per-node traffic through the optional queued provider reader.
- `rollover.go` validates the active purchase owner and composes the fresh aggregate rollover projection through the optional queued usage-snapshot reader; it never stores provider series.
- `service_test.go` verifies forwarding, validation errors, dashboard freshness, and subscription revocation behavior.
- `live_catalog_test.go` verifies missing live squads are rejected during catalog hydration and purchase validation.
- `usage_quote_test.go` verifies bounded node-usage ranges, quote selection validation, and purchase-history forwarding.
- `rollover_test.go` verifies projection ownership, active-term validation, upstream identity requirements, and queued snapshot ranges.
