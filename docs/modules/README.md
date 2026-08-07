# Module specifications

These documents describe the implemented v1 module boundaries. They are contracts for maintenance: each document states ownership, durable invariants, external interfaces, failure behavior, and the verification expected before release.

| Module | Specification |
| --- | --- |
| Runtime, persistence, scheduler, and backups | [platform-runtime.md](platform-runtime.md) |
| Telegram identity and onboarding | [accounts-onboarding.md](accounts-onboarding.md) |
| Catalog, purchases, and access terms | [catalog-entitlements.md](catalog-entitlements.md) |
| TXB ledger, top-ups, and refunds | [billing-payments.md](billing-payments.md) |
| Telegram, Remnawave, EZPay, and BEPusdt adapters | [integrations.md](integrations.md) |
| Admin authorization and domain operations | [admin-operations.md](admin-operations.md) |
| Vue Mini App and generated API client | [web-ui.md](web-ui.md) |

Cross-module rules:

- HTTP request/response compatibility is defined by `api/openapi.yaml`; generated frontend types must not be hand-edited.
- All blocking repository and provider calls accept `context.Context`. Background work stops through the application context.
- SQLite is the source of truth for identity, money, purchases, callback deduplication, and pending work. Provider redirects and browser events are never authoritative.
- IDs are opaque strings, timestamps are RFC3339 UTC strings, TXB uses integer hundredths, provider decimals use fixed-decimal parsing, and traffic byte counts never pass through floating point.
- Secret and bearer values are redacted at the logging boundary. Sensitive dashboard settings are AES-256-GCM encrypted and write-only through the API.
- Ledger entries, refunds, webhook receipts, and audit events are immutable. Corrections use compensating records.

