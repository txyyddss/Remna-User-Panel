# Database admin

- `AdminDatabasePanel.vue` coordinates table selection, debounced queries, row views, and mutation requests.
- `DatabaseTablePicker.vue` provides the searchable phone select and preserves the searchable desktop table list.
- `DatabaseTablePicker.test.ts` covers searchable table selection on both responsive presentations.
- `DatabaseQueryControls.vue` owns typed search plus desktop-live and phone-staged filter behavior.
- `DatabaseQueryControls.test.ts` covers staged Apply, Cancel, Clear, the five-filter limit, and body scrolling.
- `DatabaseFilterFields.vue` renders reusable typed column, operator, value, and removal controls.
- `DatabaseMobileRowCard.vue` renders structured keys, safe schema-ordered previews, expansion, and labeled row actions.
- `DatabaseMobileRowCard.test.ts` covers three-field previews, sensitive exclusion, expansion, and typed actions.
- `DatabaseRecordEditor.vue` coordinates distinct Draft and Review stages in a fixed-chrome, safe-area-aware drawer.
- `DatabaseRecordEditor.test.ts` covers draft submission, stage switching, confirmation gating, and delete review.
- `DatabaseRecordFields.vue` renders editable, nullable, BLOB, and protected record metadata.
- `DatabaseMutationReviewPanel.vue` renders exact before/after values, backup guidance, and typed confirmation.
- `types.ts` contains shared readonly and query-control contracts for the database feature.
