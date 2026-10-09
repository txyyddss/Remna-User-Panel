# Administrator host connectivity

This settings module configures an existing Remnawave account and presents authenticated proxy connectivity attempts. It performs no browser-side probes and receives no proxy credentials. All data uses the administrator-only client in `web/src/api/connectivity.ts`.

- `AdminConnectivitySettings.vue` composes the form, latest results and history; its exposed `save` and `loading` join the Settings page’s existing Save action. Failed or invalid connectivity saves suppress the page’s success notice.
- `ConnectivityConfigForm.vue` emits typed configuration patches and exact-username actions using existing Nuxt fields and `SwitchField`. Only the resolved numeric account reference is persisted.
- `ConnectivityResults.vue` presents batch progress, explicit stale/setup errors and mobile-first flat host rows.
- `ConnectivityAttemptSummary.vue` reuses `StatusBadge` for status, response timing, HTTP status and sanitized diagnostic messages.
- `ConnectivityHistory.vue` presents a host filter, 24-hour attempt list and cursor continuation. Its nonempty all-host sentinel remains a UI value and is omitted from API queries.
- `useConnectivityMonitor.ts` owns initial loading, two-second active/15-second idle polling, manual requests, superseded-response guards and abort/timer disposal.
- `useConnectivitySettings.ts` merges refreshed settings without overwriting edits, resolves exact usernames and saves one atomic `connectivity.config` value. Run now is unavailable while configuration or account selection is unsaved.
- `useConnectivityHistory.ts` owns filter-bound pages, aborts obsolete requests and deduplicates appended attempts. Completion of a batch refreshes the current filter.
- `config.ts` provides typed defaults, Zod validation and configuration comparison; `presentation.ts` provides localized safe diagnostics and shared status tones.
- `useConnectivitySettings.test.ts` covers preserved drafts, exact account resolution and clearing in-flight selection. `useConnectivityMonitor.test.ts` covers stale responses and polling disposal. `useConnectivityHistory.test.ts` covers filter changes and cursor continuation. These suites run in hosted CI.
- `ConnectivityResults.test.ts` covers restored identities without endpoint metadata and actionable subscription failures without duplicate empty-state prompts.

Retry controls use `maxRetries` (0–10, default 10 additional attempts) and `retryIntervalSeconds` (1–60, default 1). The timeout applies to every attempt. Existing typed draft merging preserves retry edits across polling; either field marks the form dirty and blocks Run now until saved. Zero retries is valid. A host stays Checking through its retries and publishes one final result. `config.test.ts` and the settings tests cover bounds, independent retry edits, polling and atomic saves in hosted CI.

English and Chinese strings live in `web/locales/*/host-connectivity.json`. Future UI capabilities should extend these focused components and the existing typed client; this module currently covers only connectivity configuration, checks and retained results.
