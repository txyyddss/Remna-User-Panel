# HTTP API transport

## Core transport

- `server.go` constructs middleware, public routes, authenticated routes, and the SPA fallback.
- The SPA fallback revalidates `index.html` while serving hashed assets with immutable caching to avoid stale WebView bundles.
- `request_auth.go` requires the per-session HMAC contract on authenticated browser requests.
- `request_auth_test.go` tests signed-route enforcement and session-cookie visibility rules.
- `request_validation.go` applies shared request validation, strict JSON decoding, and URL parameter normalization.
- `authentication.go` exchanges verified Telegram init data for session and request-signing cookies.
- `handlers.go` serves onboarding actions, dashboard, catalog, purchases, balance, and ledger history.
- `purchases.go` serves the authenticated queued-purchase cancellation and TXB refund endpoint.
- `dashboard_usage.go` validates the authenticated member's UTC traffic range and serves the bounded per-node usage projection.
- `billing_orders.go` creates, polls, and cancels member payment orders.
- `renewals.go` quotes and commits 1-6-term current-ride renewal batches.
- `onboarding.go` serves and administers localized onboarding content.
- `emby.go` serves member and administrator Emby account operations.
- `database_admin.go` exposes guarded database inspection, backup download, and restore transports.

## Community features

- `community.go` registers member and administrator Activity, coupon, and questionnaire routes.
- `community_shared.go` defines shared community settings and nullable JSON request fields.
- `community_helpers.go` contains idempotency, bounded multipart, decimal, collection, and error helpers.
- `activity_types.go` maps Activity domain games, draws, rewards, and history to responses.
- `activity_member.go` serves member Activity overview, check-in, bet, draw, and configuration behavior.
- `activity_admin.go` serves Activity settings and game administration.
- `activity_draw_admin.go` serves lucky-draw administration and prize mapping.
- `coupons_member.go` serves wallet, soft-discard, and redemption behavior plus coupon response mapping.
- `coupons_admin.go` serves coupon creation, updates, deactivation, and partial-input merging.
- `questionnaire_types.go` maps questionnaire, participation, import, and settlement responses.
- `questionnaire_member.go` serves active questionnaire, history, and participation behavior.
- `questionnaire_admin.go` serves questionnaire lifecycle and editable-input behavior.
- `questionnaire_imports.go` serves bounded CSV upload, analysis, settlement, and import lookup.

## Administrator features

- `admin.go` registers administrator routes and delegates community administration registration.
- `admin_settings.go` lists, creates, and updates deployment settings.
- `admin_payment_profiles.go` lists and updates one masked EZPay/BEPusdt provider profile with independently enabled channels.
- `admin_catalog.go` manages combos, squad products, nodes, and statistics windows.
- `admin_accounts.go` manages users, balance adjustments, and entitlements.
- `admin_payments.go` lists administrator payment and refund projections, applies refunds, and grants terminal-payment courtesy credits.
- `admin_operations.go` manages backups, durable jobs, and audit-event listings.
- `admin_errors.go` maps administrator domain and upstream failures to transport errors.

## Operational and provider protocols

- `operations.go` serves liveness and readiness probes.
- `operations_shared.go` contains bounded-context and request-ID helpers.
- `telegram_webhook.go` validates and processes Telegram membership and Stars payment updates.
- `telegram_commands.go` handles group-message rewards and administrator deduction commands.
- `payment_callbacks.go` handles EZPay/BEPusdt callbacks, short-lived payment-return capabilities, and navigation-only payment returns/status polling.

## Tests

- `server_test.go` covers strict decoding, onboarding access, and SPA delivery.
- `community_test.go` covers nullable inputs, multipart bounds, decimal parsing, reward mapping, and idempotency.
- `operations_commands_test.go` covers Telegram deduction command parsing.
- `payment_callbacks_test.go` verifies that navigation returns accept only the documented payment providers.

`README.md` is this direct-file ownership index. Unsigned routes remain limited to operational probes, Telegram authentication, provider-authenticated callbacks, the capability-limited payment return/status flow, and static assets.
