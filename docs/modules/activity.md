# Activity module

## Ownership and interfaces

Activity owns administrator-defined betting games, one daily check-in per configured local day, weighted lucky draws, immutable outcomes, and pending subscription-extension credits. Member operations are `GET /api/v1/activity` plus the idempotent `POST /api/v1/activity/check-ins`, `/bets`, and `/draws` actions. Administrator operations manage Activity settings, games, and complete lucky-draw prize lists.

The HTTP layer accepts TXB as decimal-string minor units and basis-point integers. The domain and database use `int64`; no Activity calculation uses binary floating point. The `Idempotency-Key` header is required for a bet or draw and is scoped to the member action.

## Financial and randomness invariants

- A game snapshots its name, icon, win chance, stake bounds, and total-return multiplier into every immutable result.
- The stake is debited before the injected cryptographic random source selects an integer in `[0,10000)`. A win credits `floor(stakeMinor * returnMultiplierBps / 10000)` as total return.
- A check-in key is the member plus local `YYYY-MM-DD` in the administrator-configured IANA timezone. Replays return the original result and never append another ledger credit.
- A lucky draw validates its total positive weight and available stock in the same write transaction that charges the fee and applies its reward.
- Before drawing, available balance must cover the fee plus the largest possible negative TXB prize. This prevents the random result from producing unapproved debt.
- Prize effects are a closed `Reward` union: `none`, signed `txb_delta`, `coupon_grant`, or positive `subscription_extension`. The selected definition is snapshotted in the result.
- A subscription extension moves an active term's expiry and every queued term by the same number of days. When no term exists, a durable extension credit is consumed by the next purchase.

Every debit and credit is committed with its ledger entry and a semantic reference derived from the immutable result. Failed validation, insufficient funds, exhausted stock, or reward application failure rolls back the result, stock, balance, and ledger together.

## Product and safety behavior

The member interface presents probability and cost plainly, never uses streak pressure, near-miss animation, confetti, or loss-chasing copy, and respects reduced motion. Disabled games and draws cannot be invoked through the API even if a stale client still renders them. Activity history is user-scoped and bounded.

## Verification

- Table-driven tests inject deterministic randomness for loss, win, weight boundaries, and invalid upper bounds.
- SQLite tests cover insufficient funds, fee-plus-deduction coverage, stock races, idempotent replays, ledger atomicity, and extension application.
- Calendar tests cover the configured timezone at UTC/local day boundaries.
- Vue tests cover human-major TXB conversion, loading/error/empty states, visible cost and odds, reduced motion, keyboard use, and narrow layouts.
