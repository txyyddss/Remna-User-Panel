# IP Lookup
- `useIPLookup.ts` owns quotes, idempotent submission, account isolation and receipt polling.
- `IPLookupPage.vue` composes member loading, input, errors and report states.
- `IPLookupForm.vue` keeps the address and 44px lookup control on one row, with the current quota or TXB price below.
- `IPLookupReport.vue` presents only the compact frozen verdict, refusal evidence, aggregate facts, field sources, database participation and four MaxMind fields. Absent values are hidden; zero remains visible.
- `reportPresentation.ts` derives localized, nullable aggregate rows without retaining provider payloads.
- `IPLocationMap.vue` and `map.ts` embed the selected coordinate pair through OpenStreetMap's native iframe, preserve its attribution and reuse Telegram-aware external links. Invalid pairs are hidden; map loading never changes lookup billing.
- `presentation.ts` maps safe errors, provider/subdatabase names and source labels to translations.

The feature uses Vue Composition API, existing dark-theme controls and English/Chinese localization. Route views only compose feature components. Async results are scoped to the current member, and scope disposal cancels future polling and quote updates. Browser fixtures exercise these real components through Chrome DevTools MCP. No local automated test suite is required or permitted.

- `useIPLookup.test.ts` covers quote-only reads, duplicate suppression, ambiguous submission replay and polling cleanup for hosted CI.

Quote reads automatically display exact-IP or IPv4 /24 cached reports without creating a paid receipt or consuming quota. A /24 reuse labels both the requested address and the original cached address. Cached addresses have no refresh action. Provider API errors stay in the cached completed report and do not trigger refunds; application/queue failures remain distinct operational errors.

The map uses only a fixed HTTPS OpenStreetMap origin with a bounded nondegenerate extent, a lazy iframe, a localized accessible title and no referrer. The HTTP CSP permits that frame origin without expanding parent script, image or connection origins. Backend coordinates and country/city are selected together; the frontend does not mix database locations.

- `presentation.test.ts` covers coordinate edge cases, hidden nullable rows, zero values and source names for hosted CI.
- `useIPLookup.test.ts` also covers automatic subnet cache display, stale address/account response isolation and receipt rehydration preserving the original fresh status and charge/quota usage when it fills the address input.

## Browser evidence, 2026-10-10

Chrome DevTools MCP 0.26.0 with Chrome 154 reviewed the real components through
the constructed fixture at 320/390px phone widths and 1280px desktop width.
Input/actions stayed inline with 44px controls and viewport-width documents.
Exact/subnet previews issued one quote and zero submissions while retaining the
full allowance. Fresh refusals, partial/all-source outages, empty/loading,
disabled/application errors, insufficient balance, null rows, zero coordinates,
and English/Chinese content were inspected. Fresh receipts retained quota usage
after address hydration. No console warnings/errors or failed requests appeared
in the completed reviews. Provider outcomes were constructed, not live API proof.

The native OpenStreetMap frame loaded with its marker and attribution; final
map sizing was inspected using ordinary viewport resizing because DevTools
emulation temporarily mis-sized the cross-origin canvas. No map lifecycle or
provider-request workaround was added to production code.
