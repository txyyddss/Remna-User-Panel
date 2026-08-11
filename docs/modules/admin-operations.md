# Administration module

## Authorization and route groups

Administration is available only to a valid Telegram session whose Telegram ID exactly equals `ADMIN_TELEGRAM_ID`. The environment variable is the authorization source; request data, usernames, handles, and mutable database roles cannot grant access. The designated administrator may configure a greenfield deployment before completing ordinary member onboarding.

All operations live below `/api/v1/admin`. The Vue router lazy-groups existing URLs into commerce, community, accounts, and system navigation without breaking bookmarks. Domain-specific endpoints remain preferred for settings, catalog, users, balances, entitlements, payments/refunds, Activity, coupons, questionnaires, Emby, backups, outbox jobs, and audits. A separate schema-aware editor exists for exceptional recovery work; it exposes no raw SQL.

## Settings and secrets

The fixed registry defines each known key's validator, sensitivity, display category, and readiness impact. Unknown keys are rejected. Sensitive values are encrypted with AES-256-GCM using a random nonce and the setting key as authenticated context. Reads expose an empty value plus configured/encrypted metadata; plaintext and ciphertext are never returned. An empty secret update preserves the existing value.

| Setting | Contract |
| --- | --- |
| `telegram.group_chat_id`, `telegram.channel_chat_id` | Required nonzero integer chat IDs |
| `telegram.webhook_secret` | Encrypted, URL-safe; generated at bootstrap |
| `remnawave.base_url`, `remnawave.api_token` | Required HTTPS origin and encrypted bearer token |
| `billing.rate.txb_per_cny`, `txb_per_usd`, `txb_per_xtr` | Required positive fixed decimals; legacy inverse rates are not converted |
| `billing.ezpay.enabled`, `.base_url`, `.merchant_id`, `.key`, `.methods` | Boolean gate, HTTPS origin, merchant ID, encrypted key, ordered five-rail list |
| `billing.bepusdt.enabled`, `.base_url`, `.api_token`, `.methods`, `.ack` | Boolean gate, HTTPS origin, encrypted token, ordered ten-rail list, `ok`/`success` ack |
| `billing.stars.enabled` | Boolean Stars gate |
| `emby.base_url`, `emby.api_token`, `emby.setup_price_txb` | Encrypted HTTPS origin/token and human-major setup price |
| `activity.timezone`, `activity.daily_reward_min_txb`, `activity.daily_reward_max_txb` | IANA timezone and inclusive nonnegative daily reward range; minimum may not exceed maximum |
| `activity.group_message_threshold`, `activity.group_message_reward_txb` | Nonnegative message threshold and human-major TXB reward; either zero disables the reward |

All TXB setting/editor fields accept human-major decimals with at most two fractional digits. `150` therefore stores or resolves to `15000` minor units when used financially. Admin API domain records continue to serialize money as decimal-string minor units.

Readiness checks required settings, enabled-provider completeness, and at least one active combo without calling external providers. Missing new-direction rates keep the corresponding payment method unavailable. Provider connectivity and permissions surface through their normal operations instead of leaking secrets through readiness.

## Domain operations and audit retention

- Combo records are live: changes affect active, queued, and historical purchases and enqueue deduplicated user synchronization. Referenced combos may be hidden but not hard-deleted.
- Squad merchandising is a sparse override over the live Remnawave list; default values remove the override. Node assignments are revalidated and re-fetched upstream and are never persisted locally.
- Balance adjustment requires a bounded nonzero signed amount and reason and appends one ledger entry plus audit event.
- Telegram `/deduct <amount>` is accepted only from the configured administrator in the configured group as a reply to a known human sender. Amounts are positive human-major TXB values; the atomic debit rejects insufficient balance, uses a deterministic quoted-message reference for replay safety, and appends `telegram.balance_deduct` audit metadata.
- Entitlement cancellation and payment refund append durable compensating commands; they never mutate provider state first and hope persistence follows.
- A terminal failed/expired payment may receive one locally funded courtesy credit only with a 3-500 byte reason. Its ledger entry, dedicated idempotency record, and audit event commit together; it never changes the original order to provider-paid or calls a provider.
- Activity games/draws, coupons, questionnaires/imports, and Emby retries use their module services so validation, transactions, and idempotency remain centralized.
- Job retry is allowed only for an eligible failed job and cannot change kind or typed payload. Pending, done, and failed jobs may be deleted; processing jobs return `409`.
- Every audit insertion transactionally retains the newest 200 events. Sensitive values are redacted before persistence.

## Schema-aware database editor

The editor lists every application table except `schema_migrations` and SQLite internals. Table and column identifiers come only from `sqlite_schema`, must pass a strict identifier allowlist, and are always quoted. Records use a declared primary key or `_rowid_` only when addressable. Cursor values are sealed by the vault; pages are bounded.

Wire values are typed: `null`, boolean, text, decimal-string integer/numeric/real, or `{blobBase64}`. Integers never cross JavaScript as numbers. The query endpoint accepts debounced broad search plus up to five allowlisted column filters using `eq`, `ne`, `contains`, `starts_with`, `gt`, `gte`, `lt`, `lte`, `is_null`, or `not_null`; all values are bound and cursors include the query fingerprint. Sensitive columns are masked and excluded from broad search.

Every insert/update/delete is a two-step command:

1. `POST /database/mutations/review` loads the current row, validates types/nullability, checks its optimistic `recordHash`, and returns a redacted before/after diff, exact `reviewHash`, warning, and required `EDIT <table>` confirmation.
2. `POST /database/mutations` must reproduce the exact reviewed mutation, reason, record hash, review hash, and confirmation before the review expires. The server creates a verified rescue backup, rechecks the row/hash, performs one transaction, consumes the review, and appends a structured redacted audit.

Review tokens are single-use. Concurrent changes produce a conflict and leave the review unconsumed so the administrator can refresh and review again. All tables are marked high risk because direct edits bypass domain synchronization hooks; the UI keeps that warning visible and uses a mobile drawer for record/diff work.

## Backup download and staged restore

Backup download accepts only an opaque stored run ID and streams a verified snapshot as `application/vnd.sqlite3`; browser paths cannot select files. List responses expose a basename and decimal-string size, not an absolute server path.

Restore requires a 4–500 character reason and exact `RESTORE <backup filename>` confirmation. It verifies the stored snapshot's checksum, SQLite integrity, foreign keys, and migration compatibility, creates a rescue backup, copies a staged file beside the live database, writes a durable marker and audit, returns `202`, then requests graceful shutdown.

The next startup processes the marker before opening SQLite, atomically swaps the candidate, verifies the result, and restores the original on failure. After migrations, it records completion/failure in `restore_jobs` and the audit stream. The UI reconnects and requires a fresh authenticated session.

## Failure behavior and verification

All failures use the request-ID envelope. Unauthorized responses do not reveal target existence. Validation gives safe field messages. Provider/setup failure leaves prior settings intact. HTTP timeout never proves a committed command rolled back; clients refetch the returned resource before retrying.

- Authorization tests cover missing/expired sessions, non-admin identity, forged roles, exact administrator ID, and bootstrap setup.
- Settings tests cover human-major TXB conversion, ordered rail validation, write-only secret rotation, nonce uniqueness, and redacted audits.
- Database tests cover identifier injection, composite-key/rowid cursors, values beyond JavaScript's safe integer range, null/boolean/blob editing, optimistic conflicts, review replay, rescue backups, encrypted setting replacement, foreign keys, and redacted structured logs.
- Restore tests cover source-ID confinement, confirmation mismatch, integrity/foreign-key/migration rejection, atomic success, simulated swap failure rollback, status import, and reauthentication after restart.
