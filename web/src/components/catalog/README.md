# Catalog components

- `CatalogPage.vue` owns the five-step customer purchase journey, user-scoped draft restoration, guarded transitions, and quote restoration when returning to the node or review steps.
- `CatalogFlowProgress.vue` and `CatalogFlowControls.vue` present the current step and navigation controls; completed catalog steps use a check indicator while current and future steps retain their own icons. Controls stay in normal document flow on mobile.
- `catalogFlowProgress.ts` contains the pure completed-step icon rule used by the progress component.
- `CatalogFlowProgress.test.ts` verifies one-based catalog steps map to the stepper index and completed steps use check indicators.
- `ComboOption.vue` and `SquadSelector.vue` collect the core-combo and optional-squad selections.
- Both squad surfaces use the member presentation of `../squad-profile/SquadProfileSummary.vue` for a named, profile-icon-led, color-coded summary with localized generated facts and extra Markdown.
- `CatalogNodes.vue` presents the server-projected accessible nodes with a country flag and traffic multiplier.
- `CatalogCouponStep.vue` selects an eligible wallet coupon or redeems a new code.
- `CatalogConfirmation.vue` presents the localized post-purchase summary and emits the Home navigation action.
- `CatalogConfirmation.test.ts` verifies purchase details and the Home action.
- `CatalogCheckout.vue` combines authoritative review and idempotent purchase confirmation; payment funding remains in the balance sheet. Returning to review restores a missing quote before confirmation and does not leave a completed or failed quote in a calculating state.
- `CatalogPaymentStep.vue` is retained as the route-safe payment handoff component for legacy links while the catalog review owns purchase confirmation.
- `CatalogPage.test.ts` verifies confirmed purchases do not trigger quote restoration and returning to accessible nodes restores the missing quote.
- `ComboOption.test.ts` verifies stable plan values and selection events.
- `CatalogPaymentStep.test.ts` verifies the add-balance action stays in Vue Router history.

Coupon purchase discounts are described as price reductions, including recurring
discounts, rather than as balance additions.
