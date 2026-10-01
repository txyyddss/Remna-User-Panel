# Custom combo and core-change rewards

## Responsibility and data flow

Instant settlement and community raffle settlement both call
`applyDrawEntitlementTx`. Its core-change branch delegates to
`applyDrawCoreComboTx` in `activity_reward_combo.go` within the same transaction.
Provider updates continue through `remna_sync_user` in the existing outbox.
No new external API or persistent tables are introduced.

A purchase with non-null `reward_renewal_price_minor` has custom terms, including
when that price is zero. A core-change prize updates only its base combo ID.
Custom squads, price, traffic, rollover and traffic reward lifetimes remain
intact. The current expiry is unchanged; subsequent renewals use the new base
combo's cadence. A later full custom prize replaces the custom terms normally.
An ordinary core change still adopts the target combo's default terms.

## Existing affected purchases

The previous core-change handler set every custom override to NULL. Renewal
then inspected only the target base combo, which could have no included squads.
Fixing renewal validation alone could not restore the erased terms.

Migration `054_restore_custom_combo_after_core_reward.sql` uses immutable
`activity_draw_results` to recover current purchases with the exact erased-field
signature. It requires one unambiguous latest custom award within the purchase
window and a subsequent core-change reward matching the current base combo.
A single custom award and a core change may share a raffle settlement timestamp:
the final erased-field signature establishes that the core change ran last.

Recovery skips expired terms, overlapping purchase windows, existing renewal
successors, rollovers already being processed/settled, malformed reward payloads,
custom prizes tied at the latest timestamp, later traffic rewards, paid squad
additions, and audited purchase/user/database edits. These cases need individual
review through the existing administration tools; the migration does not infer
missing effect order or overwrite newer choices.

For eligible rows it restores custom squads, renewal price, current/renewal
traffic and rollover threshold; updates pending rollover metadata; records
`activity.custom_combo_recovered` with the source result ID; and queues one
existing provider-sync job per user. An already pending/processing identical job
is reused. The temporary candidate table is dropped within the migration.

The migration preserves the selected base combo, current expiry, purchase
charges, balance, ledger, coupons, historical draw outcomes and auto-renewal
switch. It neither repeats a draw nor automatically authorizes recurring charges.
Both the schema migration record and restored-field signature prevent reapplication.

## Validation and deployment

`activity_reward_combo_test.go` checks both reward orders, zero-price custom
terms, ordinary core behavior, and separate temporary/recurring traffic values.
`internal/catalog/automatic_renewal_reward_test.go` checks custom-then-core-change
enablement, charging and successor renewal. The migration tests check eligible
restoration, exclusions, unchanged balances/switches and safe reapplication.
Tests run in hosted CI; local automated tests are prohibited by `AGENTS.md`.

Deploy/restart the new image to apply migration 054. Review
`activity.custom_combo_recovered` audit entries and the corresponding
`remna_sync_user` jobs, then reopen the renewal dialog. CI success confirms the
constructed cases, not that a particular production purchase was repaired.
