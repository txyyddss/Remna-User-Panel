# External integrations module

## Boundary rule

Each integration package translates a documented external protocol into a small application-facing interface. Domain services own interfaces and business decisions; adapters own URLs, authentication, signing, transport timeouts, response decoding, and provider-specific errors. Every method that can block accepts `context.Context`. HTTP clients have explicit timeouts and bounded response bodies.

## Telegram

The Telegram adapter validates Mini App `initData`, creates and revokes join-request links, approves only an expected Telegram identity, reads canonical chat membership, configures the webhook and chat menu button, creates Stars invoice links, reconciles transactions, refunds Stars, and decodes updates. Its Bot API surface is `setWebhook`, `setChatMenuButton`, `createChatInviteLink` with `creates_join_request=true`, `approveChatJoinRequest`, `revokeChatInviteLink`, `getChatMember`, `createInvoiceLink` using XTR, `answerPreCheckoutQuery`, `getStarTransactions`, and `refundStarPayment`. Startup derives callback URLs from `PUBLIC_BASE_URL`, requests `message`, `chat_member`, `chat_join_request`, and `pre_checkout_query` updates, and verifies the secret-token header in constant time. Selecting the bot's Main Mini App remains a BotFather deployment step.

The adapter exposes raw Telegram IDs as integers internally but API boundaries serialize them as strings. It never consumes `initDataUnsafe`. Join links request approval, expire in 30 minutes, and are revoked after use. Successful payment fields, including `telegram_payment_charge_id`, are passed to billing for idempotent validation; the adapter does not credit balances.

## Remnawave

The Remnawave adapter follows the bundled v3.2.1 contract at `reference/Upstream/Remnawave/api.json`. Its bearer-authenticated surface is `POST /api/users`, `PATCH /api/users`, `GET /api/users/{id}`, `GET /api/users/by-username/{username}`, `POST /api/users/resolve`, `POST /api/users/{id}/actions/revoke`, `POST /api/users/{id}/actions/reset-traffic`, `GET /api/subscriptions/by-id/{id}`, `GET /api/bandwidth-stats/users/{id}?start&end&topNodesLimit`, and `GET /api/internal-squads`. It supports username/Telegram reconciliation, initial ACTIVE user creation, internal squad import, full access replacement, traffic reset/limit/reset-strategy application, statistics, subscription URL retrieval, and revoke with `revokeOnlyPasswords:false`.

Initial users use `2099-12-31T23:59:59Z`, `trafficLimitBytes=0`, `NO_RESET`, no external squad, and an empty internal-squad list. Error code `A019` has a typed duplicate-name meaning. Subscription URLs are bearer secrets: response structs may carry them to the authenticated dashboard, but log attributes, tracing, cache diagnostics, and admin list projections must omit them.

Recent statistics are cached briefly by user. If Remnawave fails and a prior value exists, the dashboard returns it with `statisticsStale:true`; without a prior value, statistics are absent, the warning remains visible, and local balance and entitlement data remain usable. Authentication distinguishes a confirmed `GET /api/users/{id}` 404 from timeout/5xx; only 404 enters missing-user recovery. Rollover quiesces access, then reads authoritative traffic, and never resets traffic before the durable rollover result. Mutation failures are retried only through kind-routed, idempotent full-desired-state outbox commands.

## EZPay

EZPay signing follows the supplied PHP reference: omit `sign` and `sign_type`, remove empty values, sort remaining keys, concatenate `key=value` pairs with `&`, append the merchant key, and compute lowercase MD5. MD5 is retained only because it is the provider protocol and is never used for password storage or local integrity.

The checkout is a signed GET to `{base}/submit.php` containing `pid`, one of the documented `alipay`, `wxpay`, `qqpay`, `bank`, or `jdpay` types, `notify_url`, `return_url`, `out_trade_no`, `name`, exact CNY `money`, `sign`, and `sign_type=MD5`. The signed GET notification must match local order, stored rail/type, provider trade ID, payable amount, and `TRADE_SUCCESS`. The return route never settles the order.

## BEPusdt

BEPusdt creation posts JSON to `/api/v1/order/create-transaction` with local `order_id`, exact USD `amount`, `fiat=USD`, one of the ten configured USDT trade types, name, timeout, and canonical notify/redirect URLs. Its signature filters empty and `signature` values, sorts keys lexically, joins `key=value` with `&`, appends the API token without another separator, and computes lowercase MD5.

The adapter persists fiat payable amount/currency, `trade_id`, crypto `actual_amount`/currency, receiving `token`/address, expiration, and `payment_url` as separate fields. Signed notifications require the provider signature. The v1.19 unsigned direct-notification shape is accepted only at a callback URL containing an HMAC capability derived from the API token and local order ID; that path segment is redacted from access logs. Notifications accept status `1`/`3` as non-crediting lifecycle updates and status `2` as a candidate settlement. Order ID, USD amount, actual USDT amount, currency/token semantics, address, and stored trade ID are checked before credit.

The bundled BEPusdt references differ on acknowledgement text: the general API document says `success`, while the endpoint-specific v1.19 notification document specifies HTTP 200 and `ok` for paid status. The adapter therefore keeps an administrator-configurable `ok`/`success` response, defaulting to `ok`, with fixtures for both. User cancellation calls the documented signed `/api/v1/order/cancel-transaction` endpoint when a direct transaction has a trade ID; a later authoritative paid notification still wins the race and credits once.

## Emby

The Emby adapter uses an encrypted base URL/token and sends `X-Emby-Token` only to that configured origin. It calls the bundled user, password, policy, library, and parental-rating endpoints with bounded responses and typed status errors. Candidate creation is reconciled only by the exact locally persisted name after ambiguous transport failure.

Policy updates never forward a browser document. The adapter fetches the complete current user policy and writes the domain's hardened overlay, preserving remote-access and unrelated fields while disabling hidden-login exposure, remote control, transcoding/remux/conversion, and downloads. Password byte slices are never logged or included in provider error bodies.

## Failure and verification

- Non-2xx status, malformed JSON, provider-declared error, unexpected content type, oversized body, timeout, or cancelled context becomes a typed/redacted error.
- Retried requests use stable local order or user identifiers so delayed callbacks and reconciliation can target the original durable record.
- Adapter tests use `httptest.Server` and assert exact paths, methods, headers, query/body fields, signatures, fixed-decimal preservation, timeouts, redaction, and typed errors.
- Contract fixtures cover every enabled EZPay/BEPusdt rail, both BEPusdt callback/ack variants, cancellation races, Emby policy enforcement, Remnawave 404 versus 5xx, current reference responses, and missing/extra fields. Unknown JSON fields are tolerated only where upstream compatibility requires it; required identifiers and amounts remain strict.
