# Constructed browser fixtures

`SettingsAudit.vue`, `settings-audit.ts` and `settings-audit.html` mount the actual
settings page with isolated squad/activation receipts and existing preferences.
`single`, `inactive`, `noqueued`, `slow`, `error`, and `pollerror` exercise eligible,
loading, empty and recovery states. `window.__controlsAudit` records mutations;
`window.__controlsSnapshot()` exposes resulting term dates for inspection.
The Settings fixture also supplies two optional squads with one node each and
different popularity shares for Add squads geometry, sorting and checkout audits.
`addonempty`, `addonslow` and `addonerror` exercise popup state handling.
`refund` mounts retained purchase/refund actions and verifies the manual reset
control and dialog are absent while server-quoted refunds remain available.
`CommunityAudit.vue`, `community-audit.ts` and `community-audit.html` exercise the
full-width community layout with `locked`, `joined`, `slow` and `error` states.

- `LuckyAudit.vue` mounts the member Activity page in Nuxt UI without authentication.
- `lucky-audit.ts` supplies a local Activity overview and controllable draw response through `window.__audit`.
- `lucky-audit.html` opens the fixture at `/fixtures/lucky-audit.html` while the Vite development server runs.
- `LuckyAdminAudit.vue`, `lucky-admin-audit.ts`, and `lucky-admin-audit.html` mount the admin activity panel with constructed instant, raffle, catalog, and forecast data. Query parameters `slow`, `empty`, and `error` exercise loading and failure views.

These files are for local visual audits. The production Vite build starts from the root `index.html` and does not include this page.

- `activity-boost-audit.ts` mounts the real activity page with controllable constructed boost, loading, error, reward and check-in states for Chrome DevTools MCP validation. Exposes `window.__boostAudit` and a locale switch for manual inspection.
- `activity-boost-audit.html` opens that fixture directly, including the zero-boost
  card and its spacing against neighboring activity sections.
- `AbuseAudit.vue`, `abuse-audit.ts`, and `abuse-audit.html` mount the detector admin panel with isolated constructed API responses. Query parameters `slow`, `empty`, and `error` exercise loading, empty and failure states; `window.__abuseAudit` records form saves and rule actions for manual Chrome DevTools MCP inspection.
- The same fixture's `operation` mode exercises simultaneous operation receipts,
  polling errors and retry controls; `reduced` constructs reduced-motion preference.
