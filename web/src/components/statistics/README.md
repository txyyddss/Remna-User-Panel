# Statistics components

- `StatisticsPage.vue` is the thin page orchestrator for loading, stale, empty, and partial-error states.
- `StatisticsFreshness.vue` exposes independent Remnawave and database snapshot timestamps and stale state.
- `StatisticsFreshness.test.ts` covers independent partition timestamps and stale/current disclosure.
- `StatisticsOverview.vue` centers the weekly added-user total in the user donut, with its color legend below the chart, and combines usage, spend, rollover, activity, squad, and database KPIs.
- `StatisticsNodes.vue` presents live node health, throughput, version, multiplier, and online-user cards.
- `StatisticsTrafficChart.vue` renders the seven-day per-node traffic stack with keyboard-accessible Nuxt UI tooltips.
- `StatisticsShareCharts.vue` composes member and payment proportions.
- `StatisticsDistribution.vue` switches between squad-first and combo-first normalized SVG stacked bars.
- `StatisticsDonut.vue` is the reusable SVG donut-with-text and legend primitive.
- `StatisticsPie.vue` renders accessible SVG pies with either outside labels or a compact label list.
- `StatisticsConcentricDonut.vue` owns the reusable outer/inner ring geometry and grouped legends.
- `StatisticsMembershipDonut.vue` combines subscription states and active combo choices with the active-member count at its center.
- `StatisticsPaymentDonut.vue` renders EZPay outside BEPusdt and excludes failed or refunded payment facts.
- `StatisticsCharts.test.ts` covers donut, labeled-pie, and normalized stacked-bar semantics.
- `StatisticsPaymentDonut.test.ts` covers concentric payment-ring order, excluded payment states, and the explicit no-payment state.
- `statisticsGeometry.ts` owns deterministic pie-path and ring-dash geometry.
- `statisticsFormat.ts` owns stable chart colors plus localized number, percentage, date, and byte formatting.
- `statisticsFormat.test.ts` covers UTC weekday buckets and invalid chart fact filtering.
