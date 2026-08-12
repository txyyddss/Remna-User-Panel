# Dashboard components

- `UserHome.vue` composes the ordered Home experience and its request states; it resolves active squad UUIDs against the current catalog without persisting duplicated squad data.
- `BalanceHero.vue` opens the real funding sheet from Home and refetches a provider-returned reissue order before prepopulating it; `SubscriptionPanel.vue` displays and copies the active subscription URL.
- `UsagePanel.vue` shows current-term traffic as an accessible SVG graph and opens the date-bounded per-node detail; `TrafficUsageDetails.vue` owns its UTC range controls and node/day breakdown. `EntitlementSummary.vue` presents the active ride, reset cadence, and renewal dialog.
- `ComingSoonLinks.vue` navigates to Questionnaire and Emby under Around TX through Vue Router actions without native document links, preserving Telegram WebView context.
- `ComingSoonLinks.test.ts` verifies both member-tool actions keep the launch document URL unchanged.
- `UsagePanel.test.ts` verifies stale upstream data disclosure.
- `EntitlementSummary.test.ts` verifies the catalog action stays in Vue Router history.
