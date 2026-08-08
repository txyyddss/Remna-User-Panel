# Catalog and entitlements module

## Ownership and interfaces

This module owns combo definitions, locally merchandised Remnawave squad products, server-side pricing, coupon-aware purchases, active/queued access terms, rollover settlement, and durable commands that make Remnawave match local entitlement state. It consumes the transactional balance interface and a narrow access-synchronization interface; it does not call HTTP providers directly.

Members read `/api/v1/catalog`, create/list `/api/v1/purchases`, and see entitlement projections on `/api/v1/dashboard`. Administrators manage combos, import/update squad products, inspect entitlements, and issue audited cancellation commands.

## Catalog and pricing invariants

- A combo has a positive traffic limit, `DAY`, `WEEK`, or `MONTH` reset cadence, positive validity in days, a nonnegative TXB minor-unit price, included internal squads, and rollover threshold/cap settings.
- "Free nodes" are represented only by included Remnawave internal squads. Node IDs are not stored as a second access model.
- Squad products originate from Remnawave import. Local description, term price, visibility, and availability do not mutate the upstream squad.
- A public catalog contains only active combos whose included squads are all upstream-present, plus visible upstream-present add-ons. A stale selection is revalidated in the purchase transaction before any debit. Historical records are archived rather than deleted when purchases refer to them.
- Included squads are selected explicitly in combo administration. They are disabled, gray, and labelled `Included` in member add-on selection; a combo change immediately prunes newly included squads from the paid selection.
- Combo and squad descriptions use Markdown with raw HTML disabled and sanitized link protocols. Administration uses the same renderer for preview.
- The server reloads all selected records and calculates the final gross and coupon-adjusted net TXB price in one transaction. Browser totals are display-only.
- Every purchase stores immutable combo, squad, traffic, cadence, validity, gross/net price, coupon effect, rollover threshold, and rollover cap snapshots so later catalog edits do not rewrite history.

## Purchase and term behavior

A member has at most one active combo and one effective account-wide traffic budget/expiry. A first purchase starts immediately. A same-combo renewal extends from `max(now,currentEnd)`. Selecting a different combo or changing add-ons creates a queued term starting exactly at the current term end; there is no proration or mid-term squad replacement.

The selected coupon check/consumption, balance debit, purchase row, squad snapshot, rollover snapshot, pending extension-credit consumption, ledger entry, and outbox operation commit atomically. Insufficient TXB, including a balance already below zero after debt reconciliation, returns `409 INSUFFICIENT_BALANCE`. No purchase is visible without its matching ledger debit and synchronization job.

When a term activates, synchronization persists three phases: remove all squads to quiesce access, reset usage while access is quiesced, then replace the complete internal-squad list and apply the account-wide limit/reset strategy. A crash or ambiguous reset response repeats only a phase safe while no traffic can accrue; a retry after final apply never resets again. At expiry, synchronization removes all internal squads but leaves the upstream identity ACTIVE with the fixed 2099 upstream expiry. The local term, not Remnawave's user expiry, is authoritative.

## Rollover ordering and formula

Expiry first queues `rollover_finalize` and blocks renewal activation. The worker quiesces old access before fetching authoritative `trafficLimitBytes` and `usedTrafficBytes`.

- A zero traffic limit awards zero.
- Remaining percentage must be strictly greater than the snapshotted `rolloverMinRemainingBps` threshold.
- An eligible award is `floor(netPaidMinor * remainingBytes / limitBytes)`, capped at snapshotted `rolloverMaxTxbMinor`.
- The traffic inputs, result, old-term expiry, optional ledger credit, and next activation command commit atomically. Only then may reset/activation run.
- Transient Remnawave failures retry without resetting traffic. A confirmed missing user records a zero-credit exception instead of assuming all traffic was unused.
- Purchase ID is the rollover credit's unique semantic reference; replay cannot credit twice.

## Failure behavior

- Invalid, hidden, archived, duplicate, or upstream-missing selections fail before debit.
- SQLite contention follows the configured busy timeout and returns a retryable conflict without a partial ledger entry.
- Remnawave failure after commit leaves the purchase activating/queued or rollover pending and retries through the durable outbox. Access is never reported active merely because the browser request succeeded.
- Replayed activation/expiry jobs converge by replacing the full upstream squad set and desired limit; they do not accumulate squads.
- A queued term cannot be promoted until rollover inputs/result and old-term closure have committed.
- Administrator cancellation appends an audit event and compensating operation. Historical snapshots remain immutable.

## Verification

- Pricing tests cover included squads, paid add-ons, coupon selection, duplicate IDs, archived records, integer overflow, and tampered client totals.
- Transition tables cover first purchase, early/late same-plan renewal, plan/add-on changes, exact boundary, expiry, cancellation, pending extensions, and concurrent purchase attempts.
- Transaction tests inject failures between coupon use, debit, purchase, squads, rollover, ledger, and outbox writes and assert all-or-nothing behavior.
- Fake Remnawave tests assert rollover quiesce/read/credit ordering before reset, strict threshold/cap math, confirmed 404 versus transient failure, complete squad replacement, account-wide limit/cadence, idempotent retries, and expiry cleanup.
- Restart tests confirm queued boundaries and failed rollover/activation jobs resume from SQLite without duplicate debits or credits.
