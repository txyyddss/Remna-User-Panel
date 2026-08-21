# Catalog components

- `CatalogPage.vue` owns the four-step combo, squads, coupon, and review journey, its versioned user-scoped step restoration, and authoritative quote refresh when leaving squads or restoring later steps. `CatalogSquadStep.vue` keeps squad selection usable while it opportunistically loads composition statistics and owns the shared node Geocheck flow.
- `CatalogSquadStep.test.ts` covers non-blocking statistics failure, composition ranking handoff, and the catalog node Geocheck target.
- `CatalogFlowProgress.vue` and `CatalogFlowControls.vue` present the current step and navigation controls; completed catalog steps use a check indicator while current and future steps retain their own icons. Controls stay in normal document flow on mobile.
- `catalogFlowProgress.ts` contains the pure completed-step icon rule used by the progress component.
- `CatalogFlowProgress.test.ts` verifies one-based catalog steps map to the stepper index and completed steps use check indicators.
- `ComboOption.vue` collects the core-combo selection and summarizes included squads by count without displaying their squad or node detail. `SquadSelector.vue` collects optional-squad selections, shows success-tone Featured labels without changing catalog order, and keeps nested node actions outside checkbox behavior; bundled squads remain server-validated selections.
- `catalogFeaturedSquads.ts` ranks selectable add-ons against the selected combo's squad composition, excludes Included and Full squads, and returns every tied positive leader. `catalogFeaturedSquads.test.ts` covers ties, exclusions, and missing or non-positive data.
- Standalone optional-squad cards use the member presentation of `../squad-profile/SquadProfileSummary.vue` for a named, profile-icon-led summary with localized generated facts and extra Markdown. They carry the profile color across their outer surface; customer-facing occupancy is rendered as an exact percentage-only fact without exposing stock counts, and only full paid add-ons are greyed, disabled, and labeled with localized Full copy.
- `SquadNodeBlocks.vue` presents each optional squad's live server-projected nodes with country flag, name, decimal lowercase-`x` multiplier, provider fallback, and the shared explicit Geocheck action in a phone-first compact grid.
- `CatalogCouponStep.vue` selects an eligible wallet coupon or redeems a new code while its Continue action waits for a quote matching the current selection.
- `CatalogConfirmation.vue` presents the localized post-purchase summary and emits the Home navigation action.
- `CatalogConfirmation.test.ts` verifies purchase details and the Home action.
- `CatalogCheckout.vue` combines authoritative review and idempotent purchase confirmation; payment funding remains in the balance sheet. Returning to review restores a missing quote before confirmation and does not leave a completed or failed quote in a calculating state.
- `SquadActivationDialog.vue` prompts sequentially for every selected gated squad, including combo-included squads; raw codes remain memory-only until one purchase request.
- `CatalogPaymentStep.vue` is retained as the route-safe payment handoff component for legacy links while the catalog review owns purchase confirmation.
- `CatalogPage.test.ts` verifies confirmed purchases do not trigger quote restoration, squad-step exit refreshes the quote, empty node unions block progress, and Coupon-step continuation stays disabled until a usable quote returns.
- `SquadNodeBlocks.test.ts` verifies deterministic live node content, provider fallback, lowercase-`x` multiplier rendering, one Geocheck action per node, and exact node events.
- `SquadSelector.test.ts` verifies full paid add-ons receive the unavailable treatment, bounded occupancy exposes only a whole percentage, and node Geocheck never toggles squad selection.
- `ComboOption.test.ts` verifies stable plan values, hidden included-squad detail, and selection events.
- `CatalogPaymentStep.test.ts` verifies the add-balance action stays in Vue Router history.

Coupon purchase discounts are described as price reductions, including recurring
discounts, rather than as balance additions.
Combo, squad, and coupon feedback is emitted only when the selection changes; step navigation is soft and purchase confirmation is rigid.
