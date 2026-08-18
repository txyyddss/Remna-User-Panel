# HTTP API transport
- `activity_member_part2.go` continues the focused implementation from its original package module.
- `database_admin_part2.go` continues the focused implementation from its original package module.
- `handlers_part2.go` continues the focused implementation from its original package module.
- `server_part2.go` continues the focused implementation from its original package module.

## Core transport

- `server.go` constructs middleware, public routes, authenticated routes, and the SPA fallback.
- The SPA fallback revalidates `index.html` while serving hashed assets with immutable caching to avoid stale WebView bundles.
- `request_auth.go` requires the per-session HMAC contract on authenticated browser requests.
- `request_auth_test.go` tests signed-route enforcement and session-cookie visibility rules.
- `request_validation.go` applies shared request validation, strict JSON decoding, and URL parameter normalization.
- `authentication.go` exchanges verified Telegram init data for session and request-signing cookies.
- `handlers.go` serves onboarding actions, dashboard, catalog, purchases, balance, and ledger history.
- `purchases.go` serves the authenticated queued-purchase cancellation and TXB refund endpoint.
- `member_purchase_operations.go` serves paid reset/refund quotes, idempotent mutations, and owner-scoped receipts.
- `member_connections.go` serves metadata-only scan creation/polling and signed-handle drop commands.
- `member_routes.go` mounts the member connection, reset, refund, and operation resources.
- `dashboard_usage.go` validates the authenticated member's UTC traffic range and serves the bounded per-node usage projection.
- `statistics.go` serves the cached aggregate snapshot and on-demand shared
  ten-second node snapshot.
- `rollover.go` serves the authenticated owner's on-demand active-purchase rollover projection and maps ownership, inactive-term, and upstream failures to stable API errors.
- `billing_orders.go` creates, polls, and cancels member payment orders.
- `automatic_renewal.go` serves the owner-only automatic-renewal status and toggle resource; manual renewal routes are not public.
- `onboarding.go` serves and administers localized onboarding content.
- `emby.go` serves member and administrator Emby account projections.
- `emby_commands.go` validates durable member setup/update and administrator retry commands.
- `database_admin.go` exposes guarded database inspection, backup download, restore, and streamed-upload transports.
- `backup_upload.go` parses one bounded multipart SQLite stream and publishes it only after hash and schema verification.
- `request_stream_upload.go` recognizes the exact streamed-upload route while retaining origin, session, administrator, and idempotency enforcement.

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
- `admin_billing_limits.go` validates exact TXB minor-unit bounds before committing the singleton range and audit record atomically.
- `admin_payment_profiles.go` lists, creates, and updates masked EZPay/BEPusdt account profiles with independently enabled channels.
- `admin_catalog.go` manages combos, squad products, and statistics windows; node accessibility is read-only in member catalog quotes.
- `admin_statistics_window.go` validates administrator statistics date ranges and buckets.
- `admin_accounts.go` manages user lists, aggregate user detail, and balance adjustments.
- `admin_user_detail.go` maps the non-duplicated aggregate profile to the public response.
- `admin_user_commands.go` serves optimistic entitlement edits, entitlement refunds, and no-charge combo replacements.
- `admin_bulk_extensions.go` serves inclusive-OR previews and idempotent bulk-extension jobs.
- `admin_operation_resolution.go` records idempotent audited resolutions for pending-review and partial operations without provider retry.
- `admin_payments.go` lists administrator payment and refund projections, applies refunds, and grants terminal-payment courtesy credits.
- `admin_operations.go` manages backups, durable jobs, and audit-event listings.
- `admin_errors.go` maps administrator domain and upstream failures to transport errors.

## Operational and provider protocols

- `operations.go` serves liveness and readiness probes.
- `operations_shared.go` contains bounded-context and request-ID helpers.
- `telegram_webhook.go` validates and processes Telegram membership and Stars payment updates.
- `telegram_commands.go` dispatches localized slash commands before group-message rewards and preserves the unadvertised administrator deduction command.
- `telegram_command_replies.go` composes subscription and combo replies from existing catalog domain services.
- `payment_callbacks.go` handles EZPay/BEPusdt callbacks, short-lived payment-return capabilities, and navigation-only payment returns/receipt polling.

## Tests

- `server_test.go` covers strict decoding, onboarding access, and SPA delivery.
- `community_test.go` covers nullable inputs, multipart bounds, decimal parsing, reward mapping, and idempotency.
- `admin_catalog_test.go` covers strict decoding of typed squad profile writes.
- `operations_commands_test.go` covers Telegram deduction command parsing.
- `payment_callbacks_test.go` verifies that navigation returns accept only the documented payment providers.
- `admin_user_detail_test.go` covers owner-ID restoration in nested aggregate records.

`README.md` is this direct-file ownership index. Unsigned routes remain limited to operational probes, Telegram authentication, provider-authenticated callbacks, the capability-limited payment return/status flow, and static assets.
- `handlers_part2.go` contains the remaining route handler implementations.
- `server_part2.go` contains static asset and transport helper implementations.
