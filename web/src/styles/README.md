# Visual system

`main.css` is the ordered entry point for these cascade layers:

- `foundation-01.css` defines tokens, document defaults, and typography.
- `foundation-02.css` defines shared controls, surfaces, and utility patterns.
- `shell-01.css` styles the application shell and primary navigation.
- `shell-02.css` styles page framing, notices, and session states.
- `shell-03.css` styles the fullscreen-only localized greeting and its safe-area reserve.
- `onboarding-01.css` styles onboarding structure and progress.
- `onboarding-02.css` styles onboarding forms and agreements.
- `onboarding-03.css` styles onboarding completion and supporting states.
- `session-01.css` styles the non-blocking CSS car animation used by the authentication loading screen.
- `dashboard-01.css` styles dashboard summaries and account details.
- `dashboard-02.css` styles dashboard lists, actions, and secondary states.
- `home-01.css` styles the compact balance, 44px Home actions, narrow-phone subscription layout, and traffic surfaces.
- `home-02.css` styles traffic details, ride facts, and Around TX links on Home.
- `home-03.css` styles the compact renewal term slider and separated quote total/date block.
- `home-04.css` styles the accessible Your ride rollover flip card, its concise current-term detail, size-safe face swap, and reduced-motion fallback.
- `connections-01.css` styles the connections page heading, node/IP rows, empty states, and selected connection confirmation detail.
- `connections-02.css` styles the stable connection scan visualization, progress presentation, tablet split layout, and reduced-motion fallback.
- `connections-03.css` styles active block rows, exact-expiry metadata, and compact unblock confirmation targets.
- `statistics-01.css` styles the independently divided statistics sections, centered refresh control, KPI panels, dense icon-based node rows, and compact states.
- `statistics-02.css` styles the responsive generic SVG donut, labeled pies, traffic tooltips, legends, and normalized stacked bars.
- `statistics-03.css` constrains the node Geocheck modal to Telegram safe areas and styles its low-latency touch zoom canvas and fixed icon controls.
- `affiliates-01.css` styles flat member metrics, tier progress, referral rows, and tablet expansion.
- `catalog-01.css` styles catalog browsing and plan presentation.
- `catalog-02.css` styles catalog selection and checkout details.
- `billing-01.css` styles funding controls.
- `billing-02.css` styles provider-payment details.
- `billing-03.css` styles crypto currency/network selection, address copy, QR loading, and the fixed payment countdown.
- `feedback-01.css` styles feedback, empty, and loading states.
- `admin-01.css` styles administrative navigation and section layouts.
- `admin-02.css` styles administrative forms, lists, and data surfaces.
- `admin-03.css` styles administrative drawers and review workflows.
- `admin-04.css` styles aggregate user profiles, bulk previews, operation resolution, and streamed backup upload.
- `admin-05.css` styles active IP blocks and mobile unblock controls in aggregate user profiles.
- `overlays-01.css` resolves Telegram safe-area bounds while retaining Nuxt UI's native centered position for teleported modals and responsive drawer/slideover layouts.
- `motion-01.css` defines short transitions and state motion.
- `responsive-01.css` owns shared motion keyframes used by compact surfaces.
- `responsive-02.css` expands the phone-first shell into a sticky-rail desktop layout, keeps Home in its centered balance-first column, and restores wide route, dialog, catalog, settings, and administrative list arrangements.
- `responsive-03.css` applies compact administrative rows below 640px and owns reduced-motion/transparency fallbacks.

The system is dark-only, flat, gradient-free, and uses borders instead of decorative shadows.
