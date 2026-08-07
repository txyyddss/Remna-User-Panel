# Catalog and entitlements module

## Ownership and interfaces

This module owns combo definitions, locally merchandised Remnawave squad products, server-side pricing, purchases, active/queued access terms, and the outbox commands that make Remnawave match durable entitlement state. It consumes the balance transaction interface and a narrow access-synchronization interface; it does not call HTTP providers directly.

Users read `/api/v1/catalog`, create/list `/api/v1/purchases`, and see entitlement projections on `/api/v1/dashboard`. Admins manage combos, import/update squad products, inspect entitlements, and issue audited cancellation commands.

## Catalog and pricing invariants

- A combo has a positive traffic limit, `DAY`, `WEEK`, or `MONTH` reset cadence, a positive validity in days, a nonnegative TXB minor-unit price, and included internal squads.
- “Free nodes” are represented only by included Remnawave internal squads. Node IDs are not stored as a second access model.
- Squad products originate from Remnawave import. Local description, term price, visibility, and availability do not mutate the upstream squad.
- A public catalog contains only active combos whose included squads are all upstream-present, plus visible upstream-present add-ons. A stale selection is revalidated in the purchase transaction before any debit. Historical records are archived rather than deleted when purchases refer to them.
- The server reloads all selected records and calculates the final TXB price in one transaction. Browser totals are display-only.
- Every purchase stores immutable combo, squad, traffic, cadence, validity, and price snapshots so later catalog edits do not rewrite history.

## Term behavior

A user has at most one active combo and one effective account-wide traffic budget/expiry. First purchase starts immediately. A same-combo renewal extends from `max(now, currentEnd)`. Selecting a different combo or changing add-ons creates a queued term starting exactly at the current term end; v1 performs no proration or mid-term squad replacement.

The balance debit, purchase row, squad snapshot, and outbox operation commit atomically. Insufficient TXB, including a balance already below zero after debt reconciliation, returns HTTP 409 with `INSUFFICIENT_BALANCE`. No purchase is visible without its matching ledger debit and synchronization job.

When a new term activates, synchronization persists three phases: remove all squads to quiesce access, reset usage while access is quiesced, then replace the complete internal-squad list and apply the account-wide limit/reset strategy. A crash or ambiguous reset response repeats only a phase that is safe while no traffic can accrue; a retry after the final apply never resets again. At term expiry, synchronization removes all internal squads but leaves the upstream identity ACTIVE with the fixed 2099 upstream expiry. The local term—not Remnawave's user expiry—is authoritative.

## Failure behavior

- Invalid, hidden, archived, duplicate, or upstream-missing selections fail before any debit.
- SQLite contention follows the configured busy timeout and returns a retryable conflict without a partial ledger entry.
- Remnawave failure after commit leaves the purchase `activating` or `queued` and retries through the durable outbox. Access is never reported active merely because the browser request succeeded.
- Replayed activation/expiry jobs converge by replacing the full upstream squad set and desired limit; they do not accumulate squads.
- At a term boundary, the queued term is promoted and the expired term is closed in one database transaction before synchronization is attempted.
- Admin cancellation appends an audit event and a compensating operation. Historical snapshots are retained.

## Verification

- Pricing tests cover included squads, paid add-ons, duplicate IDs, archived records, integer overflow, and tampered client totals.
- Transition tables cover first purchase, early/late same-plan renewal, plan change, add-on change, exact boundary time, expiry, cancellation, and concurrent purchase attempts.
- Transaction tests inject failures between debit, purchase, squads, and outbox writes and assert all-or-nothing behavior.
- Fake Remnawave tests assert reset ordering, complete squad replacement, account-wide limit/cadence, idempotent retries, and expiry cleanup.
- Restart tests confirm queued boundaries and failed activation jobs resume from SQLite without duplicate debits.
