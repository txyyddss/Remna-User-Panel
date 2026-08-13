# Squad profiles

## Ownership

This module owns local, customer-facing metadata for live Remnawave internal
squads. Remnawave remains authoritative for squad UUIDs, names, and presence;
this module never edits upstream squad or node configuration.

## Profile contract

Each saved profile is one of `broadband`, `china_optimized`, or
`international_network`. Broadband stores ISP, positive whole-number Mbps,
static/dynamic mode, and free-form detailed location. The other two types store
carrier details, nullable unlimited-or-Mbps port speed, and an ISO alpha-2
country code. International profiles also require at least one upstream carrier.

The existing local `description` column remains additional Markdown. Generated
facts are projected from the typed profile at runtime, so localized labels are
not duplicated in SQLite or the API payload.

## Boundaries and failure behavior

The Go profile package normalizes whitespace, country casing, carrier lists, and
irrelevant fields before persistence. Invalid or incomplete admin writes return
`INVALID_SQUAD_PROFILE`. Legacy overrides may have a null profile and are
configured through the editor before they can be saved again.

All upstream reads continue through the existing Remnawave queue. Database
migration `017_squad_profiles.sql` is additive and leaves deployed migrations
immutable. Vue uses Nuxt UI controls, locale-owned text, and a compact profile
summary for both admin and member catalog surfaces.

## Verification

Contract generation, structure/README audits, locale audits, typechecking,
linting, Go formatting, and diff inspection are required. Local runtime tests
are intentionally omitted for this repository workflow.
