# Catalog components

- `CatalogPage.vue` owns the five-step customer purchase journey, user-scoped draft restoration, and guarded transitions.
- `CatalogFlowProgress.vue` and `CatalogFlowControls.vue` present the current step and navigation controls; completed catalog steps use a check indicator while current and future steps retain their own icons. Controls stay in normal document flow on mobile.
- `catalogFlowProgress.ts` contains the pure completed-step icon rule used by the progress component.
- `CatalogFlowProgress.test.ts` verifies one-based catalog steps map to the stepper index and completed steps use check indicators.
- `ComboOption.vue` and `SquadSelector.vue` collect the core-combo and optional-squad selections.
- `CatalogNodes.vue` presents the server-projected accessible nodes with a country flag and traffic multiplier.
- `CatalogCouponStep.vue` selects an eligible wallet coupon or redeems a new code.
- `CatalogCheckout.vue` combines authoritative review and idempotent purchase confirmation; payment funding remains in the balance sheet. Returning to review restores a missing quote before confirmation.
- `CatalogPaymentStep.vue` is retained as the route-safe payment handoff component for legacy links while the catalog review owns purchase confirmation.
- `ComboOption.test.ts` verifies stable plan values and selection events.
- `CatalogPaymentStep.test.ts` verifies the add-balance action stays in Vue Router history.

Coupon purchase discounts are described as price reductions, including recurring
discounts, rather than as balance additions.
