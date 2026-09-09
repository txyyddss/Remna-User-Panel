# Activity admin

Lucky-draw prize drafts use memory-only client keys for Motion presence. The
persisted prize ID remains the API identity; the client key is never sent to or
stored by the backend.

- `AdminActivitySettings.vue` edits the calendar timezone, daily check-in reward range, and group-message reward configuration from Settings; its typed save API is coordinated by the parent Settings action.
- `AdminLuckyDrawEditor.vue` validates weighted prize configurations.
