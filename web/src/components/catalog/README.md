# Catalog components

- `CatalogPage.vue` owns the six-step customer purchase journey and its guarded transitions.
- `CatalogFlowProgress.vue` and `CatalogFlowControls.vue` present the current step and navigation controls.
- `ComboOption.vue` and `SquadSelector.vue` collect the core-combo and optional-squad selections.
- `CatalogNodes.vue` presents the server-projected accessible nodes with a country flag and traffic multiplier.
- `CatalogCouponStep.vue` selects an eligible wallet coupon or redeems a new code.
- `CatalogCheckout.vue` and `CatalogPaymentStep.vue` separate review from the idempotent purchase confirmation.
- `ComboOption.test.ts` verifies stable plan values and selection events.
