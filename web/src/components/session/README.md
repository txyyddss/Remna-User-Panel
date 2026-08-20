# Session components

- `AuthGate.vue` explains Telegram authentication failures and offers retry.
- `LoadingScreen.vue` presents the authenticated-session loading state with a CSS-only car and road animation. The decorative scene is hidden from assistive technology while the localized loading status remains available through the live region.
- `BrowserCapabilityGate.vue` gives unsupported Telegram WebViews a localized recovery screen before any protected action can throw.
- `AppErrorBoundary.vue` catches descendant render failures and provides a
  localized full-app reload path without exposing exception details.
