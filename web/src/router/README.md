# Router

- `index.ts` declares routes and installs the session guard.
- `guards.ts` resolves onboarding, administrator, and member access.
- `guards.test.ts` covers protected-route decisions.

Telegram sessions use Vue Router's in-memory history so internal links do not
reload the Mini App at a new WebView URL and lose Telegram's launch context.
Regular browser sessions use HTML5 history and keep shareable route URLs.

`/payment-result` is an immersive provider-return landing route. It polls only the member's durable order state before showing a success result; failed or expired orders hand their ID to Home for an owner-scoped reissue fetch.

`/balance` remains a compatibility redirect to `/home?topUp=1`; Home owns the real funding sheet.
