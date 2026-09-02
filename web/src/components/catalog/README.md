# Catalog components

- `CatalogPage.vue` owns the four-step squads, combo, coupon, and review journey, its versioned user-scoped step restoration, authoritative quote refresh after core-combo selection, and squad-composition preload. `CatalogSquadStep.vue` presents the prepared ordering and owns the shared node Geocheck flow.
- `useCatalogSquadPresentation.ts` starts non-blocking composition-statistics loading with the catalog page and returns featured and ordered squad IDs from the selected combo or the global composition fallback.
- `CatalogSquadStep.test.ts` covers prepared ranking handoff and the catalog node Geocheck target.
- `CatalogFlowProgress.vue` and `CatalogFlowControls.vue` present the current step and navigation controls; completed catalog steps use a check indicator while current and future steps retain their own icons. The compact progress rail fits narrow phones without forced horizontal expansion, and controls stay in normal document flow.
- `catalogFlowProgress.ts` contains the pure completed-step icon rule used by the progress component.
- `CatalogFlowProgress.test.ts` verifies one-based catalog steps map to the stepper index and completed steps use check indicators.
- `CatalogComboPricingTable.vue` lays core combos out as responsive selectable cards, preserving Markdown descriptions and server-projected values while combining traffic with reset cadence, presenting the price period as `per N days`, showing the included-squad count, and shortening rollover eligibility copy. `ComboOption.vue` is the focused option primitive covered by its existing values and selection test.
- `SquadSelector.vue` groups prepared optional squads into non-empty International Network, Broadband, and China Optimized sections. `SquadPricingTable.vue` provides the responsive group grid, while `SquadPricingCard.vue` presents each member profile with compact facts, node carousel, price, selection, activation, Full, unlimited-stock, and crown-only Featured states.
- `catalogSquadPresentation.ts` orders selectable add-ons independently inside each squad type, uses aggregate global composition before a combo is selected, moves Included and Full squads not already held by the current user to the end of their type, and returns every tied positive leader per type. `catalogSquadPresentation.test.ts` covers selected-combo ordering, global fallback, type-local leaders, exclusions, and missing data.
- `SquadNodeBlocks.vue` renders exactly one anonymous node control with its country flag above the multiplier. It auto-advances every two seconds, supports mobile swipes and desktop switch buttons, and opens the exact shared Geocheck result from the whole block without exposing node or provider names.
- `CatalogCouponStep.vue` selects an eligible wallet coupon or redeems a new code while its Continue action waits for a quote matching the current selection.
- `CatalogConfirmation.vue` presents the localized post-purchase summary and emits the Home navigation action.
- `CatalogConfirmation.test.ts` verifies purchase details and the Home action.
- `CatalogCheckout.vue` combines authoritative review and idempotent purchase confirmation; payment funding remains in the balance sheet. Returning to review restores a missing quote before confirmation and does not leave a completed or failed quote in a calculating state.
- `SquadActivationDialog.vue` prompts sequentially for every selected gated squad, including combo-included squads; raw codes remain memory-only until one purchase request.
- `CatalogPaymentStep.vue` is retained as the route-safe payment handoff component for legacy links while the catalog review owns purchase confirmation.
- `CatalogPage.test.ts` verifies confirmed purchases do not trigger quote restoration, squad-step exit refreshes the quote, empty node unions block progress, and Coupon-step continuation stays disabled until a usable quote returns.
- `SquadNodeBlocks.test.ts` verifies one-node rendering, privacy-safe labels, decimal lowercase-`x` multipliers, desktop switching, and exact Geocheck events.
- `SquadSelector.test.ts` verifies non-empty type grouping, prepared order handoff, legacy-profile exclusion, and event forwarding.
- `ComboOption.test.ts` verifies stable plan values, hidden included-squad detail, and selection events.
- `CatalogPaymentStep.test.ts` verifies the add-balance action stays in Vue Router history.

Coupon purchase discounts are described as price reductions, including recurring
discounts, rather than as balance additions.
Combo, squad, and coupon feedback is emitted only when the selection changes; step navigation is soft and purchase confirmation is rigid.
