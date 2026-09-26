# Constructed browser fixtures

- `LuckyAudit.vue` mounts the member Activity page in Nuxt UI without authentication.
- `lucky-audit.ts` supplies a local Activity overview and controllable draw response through `window.__audit`.
- `lucky-audit.html` opens the fixture at `/fixtures/lucky-audit.html` while the Vite development server runs.
- `LuckyAdminAudit.vue`, `lucky-admin-audit.ts`, and `lucky-admin-audit.html` mount the admin activity panel with constructed instant, raffle, catalog, and forecast data. Query parameters `slow`, `empty`, and `error` exercise loading and failure views.

These files are for local visual audits. The production Vite build starts from the root `index.html` and does not include this page.
