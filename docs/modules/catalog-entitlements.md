# Catalog and entitlements module

## Ownership and interfaces

This module owns live combo definitions, sparse per-Remnawave-squad merchandising overrides, server-side quotes/pricing, coupon-aware purchases, active/queued access terms, rollover settlement, statistics, and durable commands that make Remnawave match local entitlement state. It consumes the transactional balance interface and a narrow access-synchronization interface; it does not call HTTP providers directly.

Members read `/api/v1/catalog`, quote at `/api/v1/purchases/quote`, create/list `/api/v1/purchases`, see entitlement projections on `/api/v1/dashboard`, and can read bounded per-node traffic at `/api/v1/dashboard/node-usage` with inclusive UTC `start`/`end` dates. The active ride's rollover detail is fetched on demand from `/api/v1/purchases/{id}/rollover`; it requires the authenticated owner, returns only aggregate cadence windows, and never persists raw provider statistics. The owner-only `GET` and `PUT /api/v1/purchases/{id}/auto-renewal` resource returns the authoritative one-cycle price, eligibility, dates, and toggle state. Administrators manage combos and sparse squad overrides, inspect combo/squad statistics, inspect entitlements, and issue audited cancellation commands. Remnawave owns accessible-node resolution.

## Catalog and pricing invariants

- Combos use a strict minimum-remaining rollover threshold and have no configured TXB rollover cap. Forecasts use actual usage divided by elapsed term time, projected across the full term, and report the maximum allowable usage plus required total/daily reduction when ineligible.
- Activation-required squads remain visible with a localized badge. Checkout prompts sequentially for every selected gated squad, including combo-included squads; raw codes stay memory-only and the database stores only bcrypt hashes. Previously validated renewal access is reused without another prompt.

- A combo has a positive traffic limit, `DAY`, `WEEK`, or `MONTH` reset cadence, positive validity in days, a nonnegative TXB minor-unit price, included internal squads, and a strict rollover threshold without a TXB credit cap.
- "Free nodes" are represented only by included Remnawave internal squads. Node IDs are not stored as a second access model.
- Squad identity/name/availability comes from every live Remnawave list call. `squad_product_overrides` stores only edited Markdown, typed profile metadata, term price, and visibility keyed by upstream UUID; restoring defaults removes the row.
- A public catalog contains only active combos whose included squads are all upstream-present, plus visible upstream-present add-ons. A stale selection is revalidated in the purchase transaction before any debit. Historical records are archived rather than deleted when purchases refer to them.
- The public catalog exposes enabled node display metadata (UUID, name, country, and consumption multiplier) for Home's usage distribution bar. The purchase quote's accessible-node projection additionally carries the live provider name/favicon URL; all such metadata is read-only, display-only, and never persisted locally.
- Included squads are selected explicitly in combo administration. They are disabled, gray, and labelled `Included` in member add-on selection; a combo change immediately prunes newly included squads from the paid selection.
- Only paid add-ons whose authoritative `stockRemaining` is zero are shown as disabled gray `Full` choices in the member catalog. Bundled combos and included squads remain selectable and rely on the existing server-side validation.
- Combo and squad descriptions use Markdown with raw HTML disabled and sanitized link protocols. The closed `[text]{color=... size=...}` directive accepts only the documented six color and four size tokens and emits classes; unsupported directives remain text. Administration uses the same renderer for preview.
- The server reloads all selected records and calculates the final gross and coupon-adjusted net TXB price in one transaction. Browser totals are display-only.
- Included upstream squad UUIDs are validated JSON on `combos`; paid add-ons alone are retained in `purchase_addons`. A purchase retains charged/gross/discount facts, coupon relationship, validity, status, user, idempotency, and one live `combo_id`. Combo edits intentionally change active, queued, and historical behavior/projections and enqueue one canonical user-sync job per affected member. Referenced combos can only be hidden.

## Purchase and term behavior

A member has at most one active combo and one effective account-wide traffic budget/expiry. A first purchase starts immediately and defaults automatic renewal on. Selecting a different combo or changing add-ons creates a queued term starting exactly at the current term end and defaults automatic renewal off; there is no proration or mid-term squad replacement. While an active purchase has automatic renewal enabled, catalog quote and purchase requests are rejected to prevent conflicting terms.

The read-only quote and the creation transaction both revalidate the live combo, coupon/grant, selected add-ons, active term, effective date, and upstream squad identities. Creation then atomically commits coupon use, balance debit, purchase/add-on facts, pending extension-credit consumption, ledger entry, and canonical outbox operation. Insufficient TXB returns `409 INSUFFICIENT_BALANCE`; no purchase is visible without its matching debit and synchronization job.

Combo and internal-squad statistics accept a bounded date range, IANA timezone, and daily/weekly bucket. They report unique buyers, purchase count, charged/discount/add-on totals, series, and included-versus-add-on distribution from authoritative purchase facts plus current live combo references.

When a term activates, synchronization persists three phases: remove all squads to quiesce access, reset usage while access is quiesced, then replace the complete internal-squad list, apply the account-wide limit/reset strategy, and set Remnawave `expireAt` to the term's local `validUntil`. A crash or ambiguous reset response repeats only a phase safe while no traffic can accrue; a retry after final apply never resets again. At final expiry, synchronization sets the upstream identity DISABLED, clears squads and traffic entitlement, and preserves the disabled-user far-future expiry; only a later active term restores ACTIVE. The local term remains authoritative.

Automatic renewal creates exactly one queued successor from the current ride's combo and paid add-ons. The scheduler runs it before expiry transitions, refreshes provider-dependent availability through the queue, and atomically debits the one-cycle current price, creates the successor, and links it to the source purchase. The unique source-successor link makes retries idempotent; normal rollover then quiesces the expiring term and activates the queued successor in the existing outbox order. Existing purchases remain opt-in after migration. Enabling validates balance, absence of a queued successor, accessible nodes, and paid add-on stock. Due renewal skips included-squad reservation limits but rechecks paid add-on stock. If it cannot charge because of balance, combo/node availability, or a paid add-on becoming full, it charges nothing, disables automatic renewal, records a stable failure code for the dashboard, and normal expiry removes access.

An attached recurring discount is persisted when the original purchase is made (and legacy rows are backfilled only from an already-linked recurring grant). Automatic renewal keeps that relationship even if the member later discards the wallet grant. It uses only the grant definition's latest discount mode, value, and percent cap; it deliberately ignores later coupon active state, expiry, quota, eligibility, and kind changes, and writes no coupon-use record. One-time coupons never attach.

Members may cancel only their own queued purchase through `POST /api/v1/purchases/{id}/cancel`. The local transaction requires `status='queued'`, marks the purchase cancelled, credits the snapshotted charged TXB amount, and appends one `purchase_cancellation` ledger entry. Because the entitlement has not reached Remnawave, this path performs no provider call and no upstream job is needed; cancellation also releases the local stock reservation while retaining immutable purchase history.

## Rollover ordering and formula

Expiry first queues `rollover_finalize` and blocks renewal activation. The worker quiesces old access before fetching the upstream reset strategy, last reset timestamp, and bounded daily sparkline usage. It retains only aggregate allocation, used bytes, eligible unused bytes, and the algorithm version.

- A zero traffic limit awards zero.
- Remaining percentage must be strictly greater than the snapshotted `rolloverMinRemainingBps` threshold.
- An eligible award is `floor(netPaidMinor * remainingBytes / limitBytes)` with no configured TXB cap; strict minimum-remaining threshold checks still apply.
- The traffic inputs, result, old-term expiry, optional ledger credit, and next activation command commit atomically. Only then may reset/activation run.
- Transient Remnawave failures retry without resetting traffic. A confirmed missing user records a zero-credit exception instead of assuming all traffic was unused.
- Purchase ID is the rollover credit's unique semantic reference; replay cannot credit twice.
- Cadence-aware settlement derives `DAY`, `WEEK`, `MONTH`, and `MONTH_ROLLING` intervals from the term and reset metadata, uses Remnawave's authoritative current-period used counter for the newest interval, uses bounded daily buckets for historical intervals, prorates partial intervals, excludes intervals below threshold from `eligibleUnusedAllowance`, and applies `netPaid * eligibleUnusedAllowance / totalAllowance` without a rollover cap. `totalAllowance` still includes every prorated interval.
- The live rollover projection forecasts from actual usage over elapsed term time, reports maximum allowable usage, maximum daily usage when the current total is still eligible but the forecast exceeds the limit, and total/daily reduction when needed. It calculates money from the immutable `charged_txb_minor` fact. When automatic renewal is disabled it returns a localized warning state without requesting provider usage.

## Failure behavior

- Invalid, hidden, archived, duplicate, or upstream-missing selections fail before debit.
- A failed due automatic renewal leaves no successor or debit, records a member-visible failure reason, and lets the ordinary expiry path disable access.
- SQLite contention follows the configured busy timeout and returns a retryable conflict without a partial ledger entry.
- Remnawave failure after commit leaves the purchase activating/queued or rollover pending and retries through the durable outbox. Access is never reported active merely because the browser request succeeded.
- Replayed activation/expiry jobs converge by replacing the full upstream squad set and desired limit; they do not accumulate squads.
- A queued term cannot be promoted until rollover inputs/result and old-term closure have committed.
- Administrator cancellation appends an audit event and compensating operation. Historical snapshots remain immutable.

## Verification

- Pricing tests cover included squads, paid add-ons, coupon selection, duplicate IDs, archived records, integer overflow, and tampered client totals.
- Transition tables cover first purchase, early/late same-plan renewal, plan/add-on changes, exact boundary, expiry, cancellation, pending extensions, and concurrent purchase attempts.
- Transaction tests inject failures between coupon use, debit, purchase, squads, rollover, ledger, and outbox writes and assert all-or-nothing behavior.
- Fake Remnawave tests assert rollover quiesce/read/credit ordering before reset, strict-threshold/no-cap math, confirmed 404 versus transient failure, complete squad replacement, account-wide limit/cadence, idempotent retries, and expiry cleanup.
- Restart tests confirm queued boundaries and failed rollover/activation jobs resume from SQLite without duplicate debits or credits.
- Automatic-renewal tests cover migration defaults and legacy backfill, owner toggle eligibility, concurrent/due retry idempotency, failure-to-expiry behavior, paid-only stock rechecks, and the recurring-coupon attachment policy.
