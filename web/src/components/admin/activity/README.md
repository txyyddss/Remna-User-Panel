# Activity admin

Lucky-draw prize drafts use memory-only client keys for Motion presence. The
persisted prize ID remains the API identity; the client key is never sent to or
stored by the backend.

- `AdminActivitySettings.vue` edits the calendar timezone, daily check-in reward range, and group-message reward configuration from Settings; its typed save API is coordinated by the parent Settings action.
- `AdminLuckyDrawEditor.vue` composes the mobile-first paid draw form and exact probability or stock totals.
- `AdminDrawPrizeRow.vue` presents prize inputs and catalog selectors through typed change events.
- `drawDraft.ts` maps saved configuration to editor state and serializes quantized rewards.
- `AdminDrawForecast.vue` reads current-catalog income, expense, and break-even estimates.
