# Common components

- `ConfirmDialog.vue` provides the shared Nuxt UI confirmation modal.
- `CountryFlag.vue` renders bundled Iconify flags for common nodes and an ISO regional-indicator flag fallback for every other valid country code.
- `InlineNotice.vue`, `OperationStatusNotice.vue`, `SkeletonBlock.vue`, and `StatusBadge.vue` provide status feedback, including durable operation polling failures.
- `LanguageSwitcher.vue` controls the active locale.
- `SwitchField.vue`, `TxbAmountField.vue`, and `MarkdownEditorField.vue` provide domain form fields.
- `MarkdownContent.vue` safely renders allowlisted Markdown.
- `MarkdownContent.test.ts` covers Markdown sanitization and rendering.
