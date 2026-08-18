# Statistics components

- `StatisticsPage.vue` is the thin page orchestrator for loading, stale, empty, and partial-error states.
- `StatisticsFreshness.vue` exposes independent Remnawave and database snapshot timestamps and stale state.
- `StatisticsFreshness.test.ts` covers independent partition timestamps and stale/current disclosure.
- `StatisticsOverview.vue` combines conversion, weekly growth, usage, spend, rollover, activity, squad, and database KPIs.
- `StatisticsNodes.vue` presents live node health, throughput, version, multiplier, and online-user cards.
- `StatisticsTrafficChart.vue` renders the seven-day per-node traffic stack with keyboard-accessible Nuxt UI tooltips.
- `StatisticsShareCharts.vue` composes subscription, combo, and payment proportions.
- `StatisticsDistribution.vue` switches between squad-first and combo-first normalized SVG stacked bars.
- `StatisticsDonut.vue` is the reusable SVG donut-with-text and legend primitive.
- `StatisticsPie.vue` renders accessible SVG pies with either outside labels or a compact label list.
- `StatisticsPaymentDonut.vue` renders the nested SVG EZPay and BEPusdt terminal-status rings.
- `StatisticsCharts.test.ts` covers donut, labeled-pie, and normalized stacked-bar semantics.
- `StatisticsPaymentDonut.test.ts` covers concentric ring order and the explicit no-terminal-payment state.
- `statisticsGeometry.ts` owns deterministic pie-path and ring-dash geometry.
- `statisticsFormat.ts` owns stable chart colors plus localized number, percentage, date, and byte formatting.
- `statisticsFormat.test.ts` covers UTC weekday buckets and invalid chart fact filtering.
