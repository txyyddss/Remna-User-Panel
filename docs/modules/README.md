# Module specifications

These documents describe the implemented module boundaries. They are maintenance contracts: each document states ownership, durable invariants, external interfaces, failure behavior, and the verification expected before release.

| Module | Specification |
| --- | --- |
| Runtime, persistence, scheduler, and backups | [platform-runtime.md](platform-runtime.md) |
| Telegram identity and onboarding | [accounts-onboarding.md](accounts-onboarding.md) |
| Catalog, purchases, and access terms | [catalog-entitlements.md](catalog-entitlements.md) |
| TXB ledger, top-ups, and refunds | [billing-payments.md](billing-payments.md) |
| Games, daily check-ins, and lucky draws | [activity.md](activity.md) |
| Coupon definitions, wallet grants, and purchase discounts | [coupons.md](coupons.md) |
| Questionnaire participation and CSV settlement | [questionnaires.md](questionnaires.md) |
| Emby account setup and restricted policy management | [emby.md](emby.md) |
| Telegram, Remnawave, EZPay, and BEPusdt adapters | [integrations.md](integrations.md) |
| Admin authorization and domain operations | [admin-operations.md](admin-operations.md) |
| Vue Mini App and generated API client | [web-ui.md](web-ui.md) |

Cross-module rules:

- HTTP request/response compatibility is defined by `api/openapi.yaml`; generated frontend types must not be hand-edited.
- All blocking repository and provider calls accept `context.Context`. Background work stops through the application context.
- SQLite is the source of truth for identity, money, purchases, callback deduplication, and pending work. Provider redirects and browser events are never authoritative.
- IDs are opaque strings, timestamps are RFC3339 UTC strings, TXB uses integer hundredths, provider decimals use fixed-decimal parsing, and traffic byte counts never pass through floating point.
- Secret and bearer values are redacted at the logging boundary. Sensitive dashboard settings are AES-256-GCM encrypted and write-only through the API.
- Domain APIs treat ledger entries, refunds, webhook receipts, Activity results, questionnaire awards, rollover results, and audit events as immutable; corrections use compensating records. The explicitly warned break-glass database editor can bypass domain hooks only after diff review, typed confirmation, rescue backup, and audit.
