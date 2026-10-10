# IP Lookup
- `useAdminIPLookup.ts` loads and saves atomic admin settings without retaining saved credentials.
- `AdminIPLookupPage.vue` composes the standalone settings form, validation and save feedback.
- `ProviderSettings.vue` edits provider switches, MaxMind account ID and write-only credential changes.
- `ComboQuotas.vue` edits the next-term included quota for each core combo.

The feature uses Vue Composition API, existing dark-theme controls and English/Chinese localization. Route views only compose feature components. Async results are scoped to the current member, and scope disposal cancels future polling and quote updates. Browser fixtures exercise these real components through Chrome DevTools MCP. No local automated test suite is required or permitted.
