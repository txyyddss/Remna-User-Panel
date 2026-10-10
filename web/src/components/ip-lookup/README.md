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

## Live enrichment

`useLiveIPDetails.ts` fetches once per requested-address/account/report identity,
uses no-store requests, cancels obsolete work, and owns explicit refresh.
`IPLiveDetails.vue` composes independent source sections beside the frozen
reputation report. It never submits paid checks or changes the verdict.

- `IPBGPTopology.vue` renders a native SVG with curved edges, keyboard/pointer pan, zoom/reset, twelve initial direct neighbors, expansion and route highlighting. Original AS paths and collector exchange context remain available in the path list.
- `bgpGraph.ts` derives bounded, deterministic layered placement and source-based visibility; `useGraphViewport.ts` owns transform interaction without persistent state.
- `IPASNDetails.vue` selects multiple origins, displays nullable registration/routing metrics and their dated provenance.
- `IPTrafficBars.vue` preserves Cloudflare percentages, Other devices, source status, and actual seven-day windows.
- `LiveSourceNote.vue` presents independent source availability and observation times.
- `useLiveIPDetails.test.ts` and `bgpGraph.test.ts` cover requested-IP isolation, cancelled/late responses, refresh, multiple origins and bounded graph visibility in hosted CI.

Only ASN names can be cached by the backend. Live frontend data is scope-local
and excluded from browser response snapshots by the existing IP Lookup cache
policy. Long refusal reasons use a fixed icon column and wrapping text.

## Browser evidence for live enrichment

Chrome DevTools MCP 1.10.1 with isolated Chrome 154 reviewed the real components
at 320/390px mobile and 1280px desktop widths: initial/loading/populated/empty,
partial/restricted/unconfigured sections, request errors, IPv6 and multiple
origins, zero counts/ratios, long multiple refusals, English/Chinese, graph
expansion/path highlighting/zoom/refresh, and Cloudflare replace/clear controls.
Document widths matched viewports, and final console and network reviews were
clear. Cached/subnet previews performed one live request for the requested IP
and zero paid submissions; manual refresh added one live request. Provider
outcomes in these fixtures were constructed, not paid-provider live proof.

Separate read-only backend probes fetched public IPv4/IPv6 topology, actual RIR
registration, dated CAIDA degrees, and PTR records; source timeouts remained
explicitly unavailable. InternetDB is queried live but its observations are
weekly and do not certify current port reachability. No Telegram messages were
sent by the browser review or the offline PM fixtures.
