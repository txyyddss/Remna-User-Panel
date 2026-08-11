# Router

- `index.ts` declares routes and installs the session guard.
- `guards.ts` resolves onboarding, administrator, and member access.
- `guards.test.ts` covers protected-route decisions.

`/payment-result` is an immersive provider-return landing route. It polls only the member's durable order state before showing a success result; failed or expired orders hand their ID to Home for an owner-scoped reissue fetch.

`/balance` remains a compatibility redirect to `/home?topUp=1`; Home owns the real funding sheet.
