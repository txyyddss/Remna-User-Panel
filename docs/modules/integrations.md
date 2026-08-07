# External integrations module

## Boundary rule

Each integration package translates a documented external protocol into a small application-facing interface. Domain services own interfaces and business decisions; adapters own URLs, authentication, signing, transport timeouts, response decoding, and provider-specific errors. Every method that can block accepts `context.Context`. HTTP clients have explicit timeouts and bounded response bodies.

## Telegram

The Telegram adapter validates Mini App `initData`, creates and revokes join-request links, approves only an expected Telegram identity, reads canonical chat membership, configures the webhook and chat menu button, creates Stars invoice links, reconciles transactions, refunds Stars, and decodes updates. Its Bot API surface is `setWebhook`, `setChatMenuButton`, `createChatInviteLink` with `creates_join_request=true`, `approveChatJoinRequest`, `revokeChatInviteLink`, `getChatMember`, `createInvoiceLink` using XTR, `answerPreCheckoutQuery`, `getStarTransactions`, and `refundStarPayment`. Startup derives callback URLs from `PUBLIC_BASE_URL`, requests `message`, `chat_member`, `chat_join_request`, and `pre_checkout_query` updates, and verifies the secret-token header in constant time. Selecting the bot's Main Mini App remains a BotFather deployment step.

The adapter exposes raw Telegram IDs as integers internally but API boundaries serialize them as strings. It never consumes `initDataUnsafe`. Join links request approval, expire in 30 minutes, and are revoked after use. Successful payment fields, including `telegram_payment_charge_id`, are passed to billing for idempotent validation; the adapter does not credit balances.

## Remnawave

The Remnawave adapter follows the bundled v3.2.1 contract at `reference/Upstream/Remnawave/api.json`. Its bearer-authenticated surface is `POST /api/users`, `PATCH /api/users`, `GET /api/users/{id}`, `GET /api/users/by-username/{username}`, `POST /api/users/resolve`, `POST /api/users/{id}/actions/revoke`, `POST /api/users/{id}/actions/reset-traffic`, `GET /api/subscriptions/by-id/{id}`, `GET /api/bandwidth-stats/users/{id}?start&end&topNodesLimit`, and `GET /api/internal-squads`. It supports username/Telegram reconciliation, initial ACTIVE user creation, internal squad import, full access replacement, traffic reset/limit/reset-strategy application, statistics, subscription URL retrieval, and revoke with `revokeOnlyPasswords:false`.

Initial users use `2099-12-31T23:59:59Z`, `trafficLimitBytes=0`, `NO_RESET`, no external squad, and an empty internal-squad list. Error code `A019` has a typed duplicate-name meaning. Subscription URLs are bearer secrets: response structs may carry them to the authenticated dashboard, but log attributes, tracing, cache diagnostics, and admin list projections must omit them.

Recent statistics are cached briefly by user. If Remnawave fails and a prior value exists, the dashboard returns it with `statisticsStale:true`; without a prior value, statistics are absent, the warning remains visible, and local balance and entitlement data remain usable. Mutation failures are retried only through idempotent full-desired-state outbox commands.

## EZPay

EZPay signing follows the supplied PHP reference: omit `sign` and `sign_type`, remove empty values, sort remaining keys, concatenate `key=value` pairs with `&`, append the merchant key, and compute lowercase MD5. MD5 is retained only because it is the provider protocol and is never used for password storage or local integrity.

The checkout is a signed GET to `{base}/submit.php` containing `pid`, `type`, `notify_url`, `return_url`, `out_trade_no`, `name`, exact CNY `money`, `sign`, and `sign_type=MD5`. The signed GET notification must match local order, provider trade ID, payable amount, and `TRADE_SUCCESS`. The return route never settles the order.

## BEPusdt

BEPusdt creation posts JSON to `/api/v1/order/create-transaction` with local `order_id`, exact USD `amount`, `fiat=USD`, selected trade type, name, timeout, and canonical notify/redirect URLs. Its signature filters empty and `signature` values, sorts keys lexically, joins `key=value` with `&`, appends the API token without another separator, and computes lowercase MD5.

The adapter persists `trade_id`, `actual_amount`, receiving `token`, expiration, and `payment_url`. Notifications accept status `1`/`3` as non-crediting lifecycle updates and status `2` as a candidate settlement. Successful signed/replayed callbacks respond with the configured acknowledgement body, default `ok`, matching the supplied BEPusdt notification documentation.

## Failure and verification

- Non-2xx status, malformed JSON, provider-declared error, unexpected content type, oversized body, timeout, or cancelled context becomes a typed/redacted error.
- Retried requests use stable local order or user identifiers so delayed callbacks and reconciliation can target the original durable record.
- Adapter tests use `httptest.Server` and assert exact paths, methods, headers, query/body fields, signatures, fixed-decimal preservation, timeouts, redaction, and typed errors.
- Contract fixtures cover current reference responses plus missing/extra fields. Unknown JSON fields are tolerated only where upstream compatibility requires it; required identifiers and amounts remain strict.
