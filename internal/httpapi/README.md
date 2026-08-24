# HTTP API transport
- `activity_member_part2.go` continues the focused implementation from its original package module.
- `database_admin_part2.go` continues the focused implementation from its original package module.
- `handlers_part2.go` continues the focused implementation from its original package module.
- `server_part2.go` continues the focused implementation from its original package module.

## Core transport

- `server.go` constructs middleware, public routes, authenticated routes, and the SPA fallback.
- `responses.go` owns compact JSON response serialization and request-correlated API errors.
- `abuse_agent.go` accepts bounded authenticated node log reports.
- `abuse_member.go` exposes privacy-safe member incident history.
- `abuse_admin.go` serves detector rules, records, and node-key controls; `abuse_admin_policy.go` owns the revisioned policy contract and cached-client streak compatibility.
- The SPA fallback revalidates `index.html` while serving hashed assets with immutable caching to avoid stale WebView bundles.
- `request_auth.go` requires the per-session HMAC contract on authenticated browser requests.
- `request_auth_test.go` tests signed-route enforcement and session-cookie visibility rules.
- `auth_rate_limit.go` bounds Telegram session exchange attempts by normalized client IP.
- `affiliates.go` exposes member referral pages and optimistic administrator tier replacement.
- `affiliate_commands.go` accepts exact positive-decimal private referral parameters.
- `affiliate_commands_test.go` covers accepted and rejected referral parameters.
- `auth_rate_limit_test.go` covers client-IP isolation, refill, and forwarded-address handling.
- `request_validation.go` applies shared request validation, strict JSON decoding, and URL parameter normalization.
- `authentication.go` exchanges verified Telegram init data for session and request-signing cookies.
- `handlers.go` serves onboarding actions, dashboard, catalog, purchases, balance, and ledger history.
- `purchases.go` serves the authenticated queued-purchase cancellation and TXB refund endpoint.
- `member_purchase_operations.go` serves paid reset/refund quotes, idempotent mutations, and owner-scoped receipts.
- `traffic_reset_automation.go` serves the onboarded member's account-wide automatic-reset preference and strict boolean update.
- `member_connections.go` preserves scan and drop-route compatibility while the signed-handle command now queues a three-day block followed by disconnect.
- `member_ip_blocks.go` lists owner-only active blocks, queues owner unblocks, and maps ownership mismatches to not found.
- `member_routes.go` mounts the member connection, reset automation, reset, refund, and operation resources.
- `dashboard_usage.go` validates the authenticated member's UTC traffic range and serves the bounded per-node usage projection.
- `statistics.go` serves the cached aggregate snapshot, on-demand shared
  ten-second node snapshot, and image-only process-local node Geocheck result; first-load provider failures are logged with the request ID.
- `rollover.go` serves the authenticated owner's on-demand active-purchase rollover projection, maps ownership, inactive-term, and upstream failures to stable API errors, and correlates unexpected provider failures with the request ID.
- `billing_orders.go` creates, polls, and cancels member payment orders.
- `automatic_renewal.go` serves the owner-only automatic-renewal status and toggle resource and logs unexpected provider failures with request, user, and purchase correlation; manual renewal routes are not public.
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
- `admin_accounts.go` manages user lists, aggregate user detail, balance adjustments, and admin-as-actor block removal.
- `admin_inventory_page.go` validates shared cursor, limit, search, and status filters for administrative inventories.
- `admin_user_detail.go` maps the non-duplicated aggregate profile and its active IP blocks to the public response.
- `admin_user_commands.go` serves optimistic entitlement edits, entitlement refunds, and no-charge combo replacements.
- `admin_bulk_extensions.go` normalizes minute durations or deprecated day inputs for inclusive-OR previews and idempotent bulk-extension jobs.
- `admin_bulk_extensions_test.go` covers minute input, legacy-day normalization, and ambiguous duration rejection.
- `admin_compensation.go` serves revisioned policy, recipient-safe cursor history, and idempotent approve/dismiss reviews.
- `admin_operation_resolution.go` records idempotent audited resolutions for pending-review and partial operations without provider retry.
- `admin_payments.go` lists administrator payment and refund projections, applies refunds, and grants terminal-payment courtesy credits.
- `admin_operations.go` manages backups, durable jobs, and audit-event listings.
- `admin_errors.go` maps administrator domain and upstream failures to transport errors.

## Operational and provider protocols

- `purchase_addons.go` serves owner-scoped active-term squad quote and idempotent commit routes with stable selection, stock, queued-term, balance, and activation errors.

- `operations.go` serves liveness and readiness probes.
- `operations_shared.go` contains bounded-context and request-ID helpers.
- `telegram_webhook.go` validates and dispatches Telegram membership and Stars payment updates.
- `telegram_commands.go` dispatches localized slash commands before group-message rewards and preserves the unadvertised administrator deduction command.
- `telegram_command_replies.go` composes subscription/combo replies from
  existing catalog services and reads the cached statistics average for check-in copy.
- `telegram_membership.go` refreshes membership and independently welcomes
  genuine non-bot joins in the configured group with a localized safe mention.
- `payment_callbacks.go` handles EZPay/BEPusdt callbacks, short-lived payment-return capabilities, and navigation-only payment returns/receipt polling.

## Tests

- `server_test.go` covers strict decoding, onboarding access, and SPA delivery.
- `community_test.go` covers nullable inputs, multipart bounds, decimal parsing, reward mapping, and idempotency.
- `admin_catalog_test.go` covers strict decoding of typed squad profile writes.
- `operations_commands_test.go` covers Telegram deduction command parsing.
- `payment_callbacks_test.go` verifies that navigation returns accept only the documented payment providers.
- `admin_user_detail_test.go` covers owner-ID restoration in nested aggregate records.
- `member_ip_blocks_test.go` covers member isolation, administrator actor attribution, and open-operation conflicts.
- `statistics_geocheck_test.go` covers cached image-only and unavailable node Geocheck responses.
- `telegram_membership_test.go` covers joins, rejoins, promotions, departures,
  bots, missing configuration, and unrelated chats through a pure transition predicate.

`README.md` is this direct-file ownership index. Unsigned routes remain limited to operational probes, Telegram authentication, provider-authenticated callbacks, the capability-limited payment return/status flow, and static assets.
- `handlers_part2.go` contains the remaining route handler implementations.
- `server_part2.go` contains authenticated request-context, static asset, and transport helper implementations.
