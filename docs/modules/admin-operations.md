# Administration module

## Authorization and setup

The admin module is available only to a valid Telegram session whose Telegram ID exactly equals `ADMIN_TELEGRAM_ID`. The environment variable is the authorization source; request data, usernames, Telegram handles, and mutable database roles cannot grant access. The designated admin may enter setup before ordinary onboarding completes so a greenfield deployment can become ready.

Admin endpoints are grouped below `/api/v1/admin` and expose domain-safe operations for settings, combos, squad products, users, balance adjustments, entitlements, payments, refunds, backups, synchronization jobs, and audit events. There is no raw SQL console, arbitrary table editor, or endpoint that updates/deletes ledger and audit rows.

## Settings and secrets

The setting registry defines each known key's category, validator, sensitivity, and readiness impact. Unknown keys are rejected. Sensitive Remnawave and provider values are encrypted using AES-256-GCM with a new random nonce for each write and authenticated context that includes the setting key.

Reads return the setting key, display category, configured/encrypted state, an empty value for secrets, and update time; the UI renders the mask. Plaintext and ciphertext are never returned. Writing an empty secret keeps an existing value, while replacement requires a new explicit value. Every change appends an audit event with redacted metadata.

The fixed registry is:

| Setting | Purpose | Protection/default |
| --- | --- | --- |
| `telegram.group_chat_id` | Required onboarding group | Required, nonzero integer |
| `telegram.channel_chat_id` | Required onboarding channel | Required, nonzero integer |
| `telegram.webhook_secret` | Telegram webhook header authentication | Encrypted; generated at bootstrap |
| `remnawave.base_url` | Remnawave API origin | Required HTTPS URL |
| `remnawave.api_token` | Remnawave bearer credential | Required, encrypted |
| `billing.rate.cny_per_txb` | EZPay conversion rate | Required positive fixed decimal |
| `billing.rate.usd_per_txb` | BEPusdt fiat conversion rate | Required positive fixed decimal |
| `billing.rate.xtr_per_txb` | Stars conversion rate | Required positive fixed decimal |
| `billing.ezpay.enabled` | EZPay feature gate | `false` by default |
| `billing.ezpay.base_url` | EZPay origin | HTTPS URL when enabled |
| `billing.ezpay.merchant_id` | EZPay merchant identifier | Required when enabled |
| `billing.ezpay.key` | EZPay signing credential | Encrypted; required when enabled |
| `billing.ezpay.payment_type` | EZPay channel | `alipay` by default |
| `billing.bepusdt.enabled` | BEPusdt feature gate | `false` by default |
| `billing.bepusdt.base_url` | BEPusdt origin | HTTPS URL when enabled |
| `billing.bepusdt.api_token` | BEPusdt signing credential | Encrypted; required when enabled |
| `billing.bepusdt.trade_type` | BEPusdt transfer network | Required when enabled |
| `billing.bepusdt.ack` | Successful callback response body | `ok` by default; `success` supported |
| `billing.stars.enabled` | Telegram Stars feature gate | `true` by default |

Setup validation checks required presence and local semantics: nonzero Telegram chat IDs, an HTTPS Remnawave URL and token, positive parseable rates, enabled-provider completeness, and at least one active combo. Results contain no credentials. `/readyz` consumes this local projection without calling external providers; webhook/menu setup and normal integration calls surface connectivity or permission failures separately.

## Domain operations

- Combo changes affect future purchases only. Archive is soft and cannot erase historical snapshots.
- Squad import reconciles Remnawave records; local merchandising updates cannot invent an upstream squad UUID.
- The user list exposes safe identity, balance, and synchronization summaries; the single-user response is the safe identity projection. Balance changes and entitlement or payment history stay behind their dedicated domain endpoints. Subscription bearer URLs are omitted throughout.
- Balance changes append a signed integer delta with mandatory reason, actor, resulting balance, reference, and audit event.
- Entitlement cancellation records state and durable revocation; it does not directly edit Remnawave and then hope the database follows.
- Refund is a once-only command against a paid order and invokes standard debt reconciliation.
- Backup trigger is single-flight. Job retry is permitted only for a failed retryable job and cannot alter its payload.
- Audit events are append-only and returned newest-first with a bounded server-side limit. Secret fields are redacted before the event is persisted, not merely at response time.

## Failure behavior

All admin failures use the standard request-ID error envelope. Unauthorized and forbidden responses do not reveal whether a target resource exists. Validation returns safe field messages. Provider validation failure leaves the prior setting intact. Concurrent catalog edits use transaction/version conflict behavior rather than last-write corruption.

Commands return the status and durable representation documented in OpenAPI. Subsequent worker failure is visible on the resource/job and retryable after configuration is corrected. Admin HTTP timeouts do not imply rollback of an already committed command, so clients refetch by returned ID before retrying.

## Verification

- Authorization tests cover missing/expired sessions, a non-admin, forged roles, the exact admin ID, and bootstrap before onboarding.
- Secret tests confirm write-only responses, nonce uniqueness, key-bound authenticated data, masked audit records, and preservation after failed validation.
- CRUD tests cover future-only catalog edits, archive with historical references, squad reconciliation, bounded list results, and conflict handling.
- Money/admin tests verify mandatory reasons, one immutable adjustment/refund, debt reconciliation, and no direct ledger mutation route.
- Backup/job tests cover single-flight commands, failed-only retries, redacted job errors, and complete audit attribution.
