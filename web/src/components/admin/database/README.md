# Database admin

- `AdminDatabasePanel.vue` coordinates table selection, debounced queries, row views, and mutation requests.
- `DatabaseQueryControls.vue` owns typed search and filter input events.
- `DatabaseMobileRowCard.vue` presents a key-only row summary with touch-sized edit and delete actions below the desktop breakpoint.
- `DatabaseMobileRowCard.test.ts` verifies key-only rendering and the typed edit/delete events.
- `DatabaseRecordEditor.vue` preserves review-before-apply database mutations in a safe-area-aware mobile drawer.
- `DatabaseRecordEditor.test.ts` verifies sticky-footer form submission, review diffs, confirmation gating, and apply events.
- `types.ts` contains shared query-control contracts for the database feature.
