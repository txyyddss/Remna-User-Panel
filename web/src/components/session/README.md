# Session components

- `AuthGate.vue` explains Telegram authentication failures, offers a light-feedback retry, and uses a local Motion status entrance without delaying the session handoff.
- `LoadingScreen.vue` presents the localized TX wordmark: outlined letters fill in, then the Chinese word slides out as Carpool enters. A scoped ResizeObserver measures the two words so the desktop mark remains centered during replacement and font/viewport changes. The decorative lettering is hidden from assistive technology; a localized status remains in the live region.
- `SessionEntrance.vue` accepts the session `status` and a default page slot. It mounts resolved content behind the loading cover and uses native Vue transitions to reveal the page as the cover fades. It owns no authentication or navigation decisions.
- `useLoadingSequence.ts` derives loading/cover visibility from session status. Successful fast loads finish the 1.9-second introduction; slow loads retain the settled wordmark until ready. Errors bypass the introduction hold. Reduced motion skips the hold; timers and media listeners are disposed with the component.
- `BrowserCapabilityGate.vue` gives unsupported Telegram WebViews a localized recovery screen before any protected action can throw, with an opacity-only entrance.
- `AppErrorBoundary.vue` catches descendant render failures and provides a
  localized full-app reload path with retry feedback and without exposing exception details.

The responsive wordmark, letter timing, reduced-motion fallback, and continuous loader keyframes live in `../../styles/session-01.css`. `SessionEntrance.vue` is the single Motion presence boundary for loading-to-application handoff. At widths below 768px the prefix sits above the changing word; larger viewports use one baseline. Brand strings live under `auth.loading*` in both core locale files, preserving the requested bilingual sequence in either UI language.
