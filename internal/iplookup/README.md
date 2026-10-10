# IP Lookup
- `types.go` defines safe configuration, signed quotes, normalized evidence, reports and member receipts.
- `config.go` validates disabled-by-default configuration, credential setting names and safe error codes.
- `registry.go` documents provider order, capabilities and credential requirements.
- `address.go` canonicalizes public addresses and rejects reserved, scoped or non-IP inputs.
- `service.go` signs owner-bound quotes and delegates atomic billing and owner-scoped reads.
- `verdict.go` aggregates source-attributed facts and conservative residential/mobile verdicts.
- `worker.go` resumes persisted provider stages in the existing operation dispatcher.
- `provider_http.go` uses fixed official endpoints, bounded queued HTTP calls and redacted errors.
- `provider_parse.go` normalizes AbuseIPDB/IPAPI responses and preserves unknown fields.
- `provider_insights.go` normalizes MaxMind, IPQS and IP2Location responses, including subscription gaps.
- `provider_test.go` covers provider response variants, public-IP validation and verdict coverage for hosted CI.
- `worker_test.go` covers early stopping, recovery and provider outages for hosted CI.

## Responsibilities and flow

`Service` accepts only canonical public IPv4/IPv6 addresses and signs two-minute,
owner-bound pricing quotes. SQLite owns immutable TXB ledger entries, allowances
bound to activated purchase IDs, paid checks, and versioned shared reports. Fresh
ordinary submissions consume quota or TXB. Cache hits are free. Refresh always
uses its independent fee. Operation polling never charges.

The existing provider-operation outbox worker executes enabled providers in
registry order. Each adapter enters its own bounded `upstreamqueue.Queue` before
HTTP execution. Every attempt marker and parsed result is persisted; recovery
marks interrupted stages unavailable rather than repeating a credit-bearing call.
Provider URLs and credentials never enter logs, receipts, or report snapshots.

Reports have no automatic expiry. Cache hits preserve their original policy,
provider selection and verdict. Fresh partial reports become the latest cache;
any contacted-provider outage restores each cost once. All-provider failures also preserve the previous cache.
Provider errors continue; positive per-IP risks stop later checks. Network-wide
scores are informational. A suitable result requires complete risk coverage and
residential or mobile ISP evidence; it does not promise service-specific access.

## Extension and API references

Add a descriptor to `Registry`, supply a queued adapter implementing `Provider`,
and extend the HTTP parser and localization/contract provider identifiers. Billing,
quota and aggregation do not depend on the provider's wire schema. Document its
available capabilities and omitted-field semantics before adding normalization.

- [AbuseIPDB check](https://docs.abuseipdb.com/): 90-day report window, no verbose comments.
- [Scamalytics v3](https://docs.scamalytics.com/ip-fraud-risk-api/v3/): EU node `api12`, account username in the path, per-IP fraud score and Essential/Premium enrichment. ISP-wide scores are contextual and Premium placeholders stay unknown.
- [IPAPI](https://ipapi.is/developers.html): keyed detection flags and truthy VPN variants.
- [MaxMind Insights](https://dev.maxmind.com/geoip/docs/web-services/responses/): current anonymizer and deprecated traits compatibility; omitted false flags.
- [IPQS](https://www.ipqualityscore.com/documentation/proxy-detection-api/response-parameters): strictness 3; unavailable connection type stays unknown.
- [IP2Location](https://www.ip2location.io/ip2location-documentation): tier-specific risk fields stay unknown when omitted.

Configuration starts disabled with fees and quotas unset. The standalone admin
interface commits settings, vaulted secrets, quotas and audit atomically. A zero
fee explicitly grants a free paid-path lookup; a zero combo quota grants no checks.
Migration grants one included check to each member with a live active purchase.
Later quotas are snapshotted only at first successful activation. Traffic resets,
extensions and changes within the same purchase cannot reset the allowance.

- `provider_http_test.go` covers queued fixed-endpoint contracts, authentication, quota errors and credential redaction for hosted CI.

## Browser evidence, 2026-10-10

Chrome DevTools MCP 0.26.0 reviewed the actual production components inside
`AppShell` using the committed constructed browser fixture. Mobile was 390 x 844
with touch emulation; desktop was 1280 x 900. Both had viewport-width documents
and 44 px lookup/refresh action buttons. The review exercised initial/empty,
loading, disabled, unavailable, suitable, rejected and partial reports, provider
expansion, cached lookups, paid refresh, included quota, restored allowance after
all-provider failure, insufficient balance, and double-submit suppression.

The standalone administrator form was checked on both layouts, including exact
2.50/3.50 TXB saves, provider toggles, five/zero combo quotas, blank write-only
credentials, and save feedback. The actual Home Around TX link and desktop
sidebar expose IP Lookup without a combo while hiding other gated tools.
Simplified Chinese rendering was visually inspected. Final console inspection
had no warnings/errors; all 940 requests in the final fixture navigation had
successful statuses. API/provider outcomes were constructed, not live-provider
or deployed-service evidence. Local automated suites were not executed.

The existing core-combo editor also exposed its current quota on both viewports and emitted an explicit zero-quota payload through its real save handler.

- `provider_scamalytics.go` normalizes Scamalytics v3, enrichment flags and per-IP versus ISP-wide scores.

Scamalytics executes after AbuseIPDB and before IPAPI. Existing five-provider report snapshots remain readable, and pending runs match configuration by provider ID rather than array position. Scamalytics low risk passes its score check; medium, high and very high fail. Explicit abuse, hosting, VPN, proxy and Tor detections still fail regardless of a low score classification. Other providers retain their existing per-IP score policy. A contacted-provider error refunds partial or rejected reports as well as total failures; missing optional premium fields remain unknown rather than being considered an outage.

The revised browser review confirmed free cache views preserve the full allowance, partial API outages restore cost, and Scamalytics low risk displays as pass. Its account username and credential controls were added to the standalone settings page.
