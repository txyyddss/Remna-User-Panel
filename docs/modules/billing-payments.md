# Billing and payments module

## Ownership and interfaces

Billing owns TXB balances, the immutable ledger, fixed-decimal exchange rates, durable payment orders, provider event deduplication, successful top-up credit, manual adjustments, refunds, and debt enforcement. Provider adapters create external checkouts and validate callbacks; billing decides whether durable business state may change.

User operations are `GET /api/v1/balance`, `GET /api/v1/ledger`, `POST /api/v1/payments/orders`, order polling by ID, and user-owned cancellation. Payment creation receives a stable `methodId` such as `ezpay:alipay`, `bepusdt:usdt.trc20`, or `stars`; the provider name is never sufficient to select a rail. Administrative operations append balance adjustments and refunds and inspect payments/refunds. No endpoint edits a balance, ledger row, or provider event in place.

## Amount invariants

- One TXB is 100 integer minor units. Database and API arithmetic uses integers for TXB.
- Required rates are `txb_per_cny`, `txb_per_usd`, and `txb_per_xtr`. Each is a validated positive fixed decimal. Legacy inverse values are not reciprocated automatically; a method remains unavailable until an administrator explicitly enters its new-direction rate.
- The server computes `ceil(requestedTxbMinor / txbPerCurrency)` at the provider's currency precision and stores the requested TXB, method, provider, rail, payable value, rate value, and rate-direction/version snapshots. Existing pending legacy orders retain `currency_per_txb` snapshots and validate without reinterpretation.
- Successful settlement credits exactly the requested `txbMinor`, not a value recomputed from a later rate.
- Each provider transaction/charge ID is globally unique where the provider guarantees uniqueness. A webhook dedupe record and its ledger credit commit together.

## Provider order lifecycle

Orders move from `creating` to `pending`, then to `paid`, `expired`, `failed`, `cancelled`, or `refunded`. Cancellation stops member polling immediately but is not authoritative against money already collected: a later verified paid event may transition a locally cancelled order to paid and credit it once. Provider setup failure preserves a diagnostic order without exposing secret request data. The client receives the exact payable value plus independently nullable payment URL, QR payload, receiving address, actual crypto amount/currency, and expiry fields.

EZPay exposes five ordered administrator-selectable rails: Alipay, WeChat Pay, QQ Wallet, bank/UnionPay, and JD Pay. BEPusdt exposes ten ordered USDT rails: TRC20, ERC20, Polygon, BEP20, Aptos, Solana, X-Layer, Arbitrum, Plasma, and TON. A saved rail must still be enabled when a new order is created. The signed EZPay callback `type` must equal the order's rail.

EZPay checkout uses a signed CNY URL and signed GET notification. BEPusdt calls `/api/v1/order/create-transaction` using USD fiat and records the returned `actual_amount`, token/address, trade ID, and payment URL separately from fiat payable amount. A callback token carrying address semantics must exactly match that stored checkout address; the reference-compatible literal `USDT` token is accepted as a currency marker. Signed callbacks require the documented signature. Unsigned v1.19 direct-transaction notifications require an unguessable per-order HMAC capability embedded in the callback path; access logs redact that segment. Stars creates a single-price XTR invoice link, opened with `Telegram.WebApp.openInvoice`. QR images are generated locally only from an explicit server-returned QR payload.

`POST /api/v1/payments/orders/{id}/cancel` is owner-scoped and idempotent. For BEPusdt, the server best-effort calls the signed `/api/v1/order/cancel-transaction` endpoint when a trade ID exists and stores a safe provider-cancellation status. Local cancellation succeeds even if the provider does not support cancellation or is temporarily unavailable.

Redirects and invoice-close UI events are navigation signals only. Settlement requires verified EZPay `TRADE_SUCCESS`, signed BEPusdt status `2`, or Telegram `successful_payment`. Order polling closes the payment sheet only after persisted state is `paid`.

## Refund and debt behavior

A refund appends one reversal against one paid order and cannot be applied twice. If reversal makes the TXB balance negative, reconciliation cancels queued purchases first, then active purchases newest-first, appending compensating ledger records and durable Remnawave revocation jobs. Any debt left after cancellable value is exhausted remains visible and blocks new purchases; it never rewrites previous entries.

Manual adjustments require an administrator, a nonzero bounded delta, and a non-empty reason. A negative resulting balance blocks subsequent purchases; payment refunds additionally run the queued-first/active-newest entitlement cancellation policy described above.

## Failure behavior

- Unknown order, wrong provider, amount mismatch, user/payload mismatch, invalid signature, and non-success state cannot credit funds.
- Callback replay returns the provider's success acknowledgement after confirming the stored dedupe result, but does not append another credit.
- A callback arriving before the checkout HTTP response can still settle the already-durable `creating` order.
- Provider redirects arriving first leave the order pending. Browser polling is bounded and may resume after navigation or restart.
- EZPay returns `success` only for a valid accepted/replayed notification. BEPusdt returns configurable acknowledgement text (`ok` or `success`, default `ok`) to support both bundled reference variants; invalid notifications deliberately do not receive it.
- Stars reconciliation queries durable Telegram charge/order state after lost updates without trusting the client.
- Before inserting a new order, the database retains at most 199 existing orders by pruning the oldest terminal records and their dependent webhook/refund evidence while preserving all ledger credits. If 200 non-prunable orders remain, creation returns a retryable capacity conflict rather than deleting payable evidence.

## Verification

- Table-driven decimal tests cover all three currencies, non-terminating rates, round-up boundaries, huge inputs, zero/negative requests, and snapshot stability.
- Concurrency tests settle and refund the same order from many goroutines under the race detector and assert one credit/reversal.
- Callback tests cover signatures, reordered parameters, empty-value rules, EZPay subtype tampering, fiat/crypto/recipient mismatches, both BEPusdt callback shapes and acknowledgements, status transitions, replay, and callback-before-response.
- Method fixtures cover every enabled EZPay/BEPusdt rail in administrator order, cancellation signatures, cancellation/paid races, and secret/capability redaction.
- Refund tests cover queued-first/active-newest cancellation, partial debt recovery, remaining debt, retry after Remnawave failure, and immutable history.
- Handler tests ensure users cannot access another order or ledger, redirects never credit, and provider/bearer secrets are absent from responses and logs.
