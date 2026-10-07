# Constructed browser fixtures

- `LuckyAudit.vue` mounts the member Activity page in Nuxt UI without authentication.
- `lucky-audit.ts` supplies a local Activity overview and controllable draw response through `window.__audit`.
- `lucky-audit.html` opens the fixture at `/fixtures/lucky-audit.html` while the Vite development server runs.
- `LuckyAdminAudit.vue`, `lucky-admin-audit.ts`, and `lucky-admin-audit.html` mount the admin activity panel with constructed instant, raffle, catalog, and forecast data. Query parameters `slow`, `empty`, and `error` exercise loading and failure views.

These files are for local visual audits. The production Vite build starts from the root `index.html` and does not include this page.

- `activity-boost-audit.ts` mounts the real activity page with controllable constructed boost, loading, error, reward and check-in states for Chrome DevTools MCP validation. Exposes `window.__boostAudit` and a locale switch for manual inspection.
- `AbuseAudit.vue`, `abuse-audit.ts`, and `abuse-audit.html` mount the detector admin panel with isolated constructed API responses. Query parameters `slow`, `empty`, and `error` exercise loading, empty and failure states; `window.__abuseAudit` records form saves and rule actions for manual Chrome DevTools MCP inspection.
- The same fixture's `operation` mode exercises simultaneous operation receipts,
  polling errors and retry controls; `reduced` constructs reduced-motion preference.
