# Squad additions

The dialog loads immediately when mounted already open by the desktop sidebar and reloads if its active purchase changes. The regression in `SquadAdditionDialog.test.ts` protects this initial-open path.

Owns the two-step active-ride add-on dialog. `SquadAdditionDialog.vue` orchestrates selection, activation codes, checkout, and in-place confirmation; `SquadAdditionCheckout.vue` renders the price review and completion state. `SquadAdditionDialog.test.ts` protects the one-based flow to zero-based progress mapping. The module reuses the catalog squad picker and `useSquadAddition` for API state.
