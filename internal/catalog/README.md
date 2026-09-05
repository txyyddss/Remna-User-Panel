# Catalog domain

This package builds the live catalog, validates purchase selections against current Remnawave squads, and serves member purchase and dashboard workflows.

- `service.go` defines repository/provider contracts and coordinates quotes, purchases, process-local dashboard caching, and subscription revocation without durable bearer storage.
- `catalog.go` hydrates sparse local merchandising and typed squad profiles with live Remnawave squad identities and one read-only node snapshot.
- `automatic_renewal.go` owns member toggle/status projections, catalog blocking, due-cycle queue revalidation, one-successor commits, and failure notices. It keeps provider access behind the existing queued adapter boundary.
- `automatic_renewal_test.go` covers catalog blocking and enablement eligibility without running provider calls.
- `renewals.go` retains the internal legacy batch implementation only; manual renewal is no longer a public member flow.
- `renewal_catalog.go` hydrates owned renewal selections independently of storefront visibility while verifying every retained squad against the queued live provider.
- `renewal_catalog_test.go` covers repriced hidden-squad enablement and processing, current-price balance checks, missing-upstream rejection, and storefront isolation.
- `cancellation.go` exposes the authenticated member operation for cancelling a queued purchase through the transactional store refund path.
- `squad_additions.go` validates current catalog visibility and delegates active-term squad quotes and commits to the transactional store without making upstream calls; `squad_additions_test.go` covers forwarding, visibility, onboarding, and repository failure paths.
- `catalog_for_user.go` derives current-holder stock facts for catalog responses without persisting them.
- `dashboard.go` composes current balance and entitlement data with cached or live upstream statistics and handles subscription URL revocation.
- `revoke_operations.go` stores credential-rotation intent with only a hash of the prior subscription URL.
- `revoke_worker.go` performs one queued rotation and reconciles ambiguous provider outcomes.
- `revoke_lifecycle.go` owns the single-item receipt lifecycle and dashboard-cache invalidation.
- `nodes.go` queries accessibility once per unique live squad, filters disabled nodes, attaches deterministic node groups to each product, and derives the authoritative quote union from that hydrated snapshot. Node assignments are never persisted.
- `node_selection.go` normalizes combo and add-on squad selection for quote node unioning.
- `catalog_node_groups_test.go`, `nodes_test.go`, `node_projection_test.go`, and `node_selection_test.go` verify grouping, filtering, deterministic ordering, node/selection deduplication, quote reuse, and provider failure handling.
- `usage.go` validates a member-selected UTC range and projects live, bounded per-node traffic through the optional queued provider reader.
- `rollover.go` validates the active purchase owner and composes the fresh aggregate rollover projection through the optional queued usage-snapshot reader; it never stores provider series.
- `service_test.go` verifies forwarding, validation errors, dashboard freshness, and subscription revocation behavior.
- `live_catalog_test.go` verifies missing live squads are rejected during catalog hydration and purchase validation.
- `usage_quote_test.go` verifies bounded node-usage ranges, quote selection validation, and purchase-history forwarding.
- `rollover_test.go` verifies projection ownership, active-term validation, upstream identity requirements, and queued snapshot ranges.
- `zero_coverage_test.go` covers queued-purchase cancellation, legacy renewal quote/commit flows, due automatic-renewal processing, and failure reasons.
