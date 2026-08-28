# Frontend API

- `purchaseAddons.ts` owns protected active-ride squad-addition quote and idempotent commit requests.

- `adminBilling.ts` owns the atomic administrator update for global Add TXB bounds.
- `abuse.ts` owns typed privacy-safe member and administrator detector resources, including configurable streaks, the revisioned warning-record cooldown, and WebView-safe idempotency keys for mutations.
- `abuse.test.ts` verifies detector mutations keep route-owned IDs out of strict JSON bodies and submit the complete streak policy.
- `adminOperations.ts` owns aggregate user profiles, facet filtering, coupon wallet actions, exact entitlement refunds, queued provider-account actions, transient connection scans, bulk extensions, operation resolution, and streamed backup upload.
- `client.ts` exposes typed member and administrator API operations, including idempotent payment, subscription-revoke, refund, and outbox-retry commands, connection scans, renewal batches, payment profiles, dashboard usage, and cached product statistics.
- `client.test.ts` covers client request behavior.
- `compensation.ts` owns revisioned configuration, cursor history, and idempotent approve/dismiss requests.
- `features.ts` exposes feature-specific endpoints and contract types, including durable Emby and questionnaire commands, the browser-safe payment-return receipt projection, and member coupon soft-discard.
- `features.test.ts` verifies feature request construction.
- `generated.ts` contains the compact generated OpenAPI contract.
- `http.ts` provides authenticated HTTP and error handling primitives.
- `memberOperations.ts` exposes generated-type connection scans, signed-handle blocks, active block listing/removal, the account-wide reset automation resource, reset/refund quotes, and durable operation receipts.
- `request-signing.ts` signs every protected request with Web Crypto or the audited pure-JS HMAC/SHA-256 fallback; it never sends an unsigned fallback.
- `request-signing.test.ts` covers signing, nonce behavior, and missing-Web-Crypto behavior.
- `openapiContract.test.ts` compile-checks reset automation, required squad-node groups, nullable provider fallback, predicted-rollover ownership, and the required abuse streak contract.
- `types.ts` exports stable aliases over generated schema types, including payment-operation envelopes, connection scans, durable operation receipts, reset automation, statistics snapshots, member mutation quotes, and typed squad profile read/write unions.
