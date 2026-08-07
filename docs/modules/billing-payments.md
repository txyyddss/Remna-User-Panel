# Billing and payments module

## Ownership and interfaces

Billing owns TXB balances, the immutable ledger, fixed-decimal exchange rates, durable payment orders, provider event deduplication, successful top-up credit, manual adjustments, refunds, and debt enforcement. Provider adapters create external checkouts and validate callbacks; billing decides whether durable business state may change.

User operations are `GET /api/v1/balance`, `GET /api/v1/ledger`, `POST /api/v1/payments/orders`, and order polling by ID. Administrative operations append balance adjustments and refunds and inspect payments/refunds. No endpoint edits a balance, ledger row, or provider event in place.

## Amount invariants

- One TXB is 100 integer minor units. Database and API arithmetic uses integers for TXB.
- CNY/TXB, USD/TXB, and XTR/TXB rates are validated fixed decimals stored as configuration snapshots. Binary floating point is forbidden in price or equality checks.
- The server converts the requested positive `txbMinor`, rounds the payable amount upward to the provider's precision, and stores both the requested TXB and rate/payable snapshots.
- Successful settlement credits exactly the requested `txbMinor`, not a value recomputed from a later rate.
- Each provider transaction/charge ID is globally unique where the provider guarantees uniqueness. A webhook dedupe record and its ledger credit commit together.

## Provider order lifecycle

Orders move from `creating` to `pending`, then to one terminal state: `paid`, `expired`, `failed`, or `refunded`. Provider setup failure preserves a diagnostic order without exposing secret request data. The client receives the exact payable value, link, QR payload, expiry, and BEPusdt receiving address when applicable.

EZPay checkout uses a signed CNY URL and signed GET notification. BEPusdt calls `/api/v1/order/create-transaction` using USD fiat and records the returned `actual_amount`, token/address, trade ID, and payment URL. A callback token carrying address semantics must exactly match that stored checkout address; the reference-compatible literal `USDT` token is accepted as a currency marker. Stars creates a single-price XTR invoice link, opened with `Telegram.WebApp.openInvoice`. QR images are generated locally by the frontend from the returned payload.

Redirects and invoice-close UI events are navigation signals only. Settlement requires verified EZPay `TRADE_SUCCESS`, signed BEPusdt status `2`, or Telegram `successful_payment`. Order polling closes the payment sheet only after persisted state is `paid`.

## Refund and debt behavior

A refund appends one reversal against one paid order and cannot be applied twice. If reversal makes the TXB balance negative, reconciliation cancels queued purchases first, then active purchases newest-first, appending compensating ledger records and durable Remnawave revocation jobs. Any debt left after cancellable value is exhausted remains visible and blocks new purchases; it never rewrites previous entries.

Manual adjustments require an administrator, a nonzero bounded delta, and a non-empty reason. A negative resulting balance blocks subsequent purchases; payment refunds additionally run the queued-first/active-newest entitlement cancellation policy described above.

## Failure behavior

- Unknown order, wrong provider, amount mismatch, user/payload mismatch, invalid signature, and non-success state cannot credit funds.
- Callback replay returns the provider's success acknowledgement after confirming the stored dedupe result, but does not append another credit.
- A callback arriving before the checkout HTTP response can still settle the already-durable `creating` order.
- Provider redirects arriving first leave the order pending. Browser polling is bounded and may resume after navigation or restart.
- EZPay returns `success` only for a valid accepted/replayed notification. BEPusdt returns configurable acknowledgement text, default `ok`; invalid notifications deliberately do not receive it.
- Stars reconciliation queries durable Telegram charge/order state after lost updates without trusting the client.

## Verification

- Table-driven decimal tests cover all three currencies, non-terminating rates, round-up boundaries, huge inputs, zero/negative requests, and snapshot stability.
- Concurrency tests settle and refund the same order from many goroutines under the race detector and assert one credit/reversal.
- Callback tests cover signatures, reordered parameters, empty-value rules, amount and recipient mismatches, status transitions, replay, and callback-before-response.
- Refund tests cover queued-first/active-newest cancellation, partial debt recovery, remaining debt, retry after Remnawave failure, and immutable history.
- Handler tests ensure users cannot access another order or ledger, redirects never credit, and provider/bearer secrets are absent from responses and logs.
