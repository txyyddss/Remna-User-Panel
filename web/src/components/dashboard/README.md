# Dashboard components

- `UserHome.vue` composes the ordered Home experience and its request states; it resolves active squad UUIDs against the current catalog without persisting duplicated squad data.
- `BalanceHero.vue` opens the real funding sheet from Home and refetches a provider-returned reissue order before prepopulating it; `SubscriptionPanel.vue` displays and copies the active subscription URL.
- `UsagePanel.vue` keeps the current-term limit meter and exposes a localized switch for including node traffic multipliers in the interactive `TrafficUsageBar.vue` distribution. Nodes below 5% are grouped into the other-traffic segment, and the former sparkline/detail popover is not rendered. `TrafficUsageBar.vue` keeps labels hidden until hover, focus, or activation, then exposes node name, country, multiplier, bytes, and share. `TrafficUsageDetails.vue` and `TrafficNodeChart.vue` remain the date-bounded detail building blocks for future dashboard entry points. `EntitlementSummary.vue` presents the active ride, reset cadence, a term slider, a separated renewal total/date block, and owner-only queued cancellation with refund feedback.
- `ComingSoonLinks.vue` navigates to Questionnaire and Emby under Around TX through Vue Router actions without native document links, preserving Telegram WebView context.
- `ComingSoonLinks.test.ts` verifies both member-tool actions keep the launch document URL unchanged.
- `UsagePanel.test.ts` verifies stale upstream data disclosure.
- `EntitlementSummary.test.ts` verifies the catalog action stays in Vue Router history.
