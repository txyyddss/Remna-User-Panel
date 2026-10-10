# IP Lookup
- `useIPLookup.ts` owns quotes, idempotent submission, account isolation and receipt polling.
- `IPLookupPage.vue` composes member loading, input, errors and report states.
- `IPLookupForm.vue` shows active-term allowance and the exact check/refresh cost.
- `IPLookupReport.vue` presents the frozen verdict, cache provenance and aggregated network facts.
- `ProviderDetails.vue` shows localized per-provider facts, unknown signals, scores and coverage.
- `presentation.ts` maps safe errors, report reasons, provider names and field labels to translations.

The feature uses Vue Composition API, existing dark-theme controls and English/Chinese localization. Route views only compose feature components. Async results are scoped to the current member, and scope disposal cancels future polling and quote updates. Browser fixtures exercise these real components through Chrome DevTools MCP. No local automated test suite is required or permitted.

- `useIPLookup.test.ts` covers quote-only reads, duplicate suppression, ambiguous submission replay and polling cleanup for hosted CI.

Cached result views are free. Refresh remains a separate priced action. Reports show cost restoration when a contacted API fails, while cached original verdicts and coverage are preserved. The provider list includes Scamalytics immediately after AbuseIPDB.
