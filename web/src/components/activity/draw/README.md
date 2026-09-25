# Lucky-draw presentation

The draw coordinator sends one idempotent request before displaying a prize. Presenters receive the server result and can only animate toward it; they never choose a reward.

- `selection.ts` defines the seven display styles, random selection, bounded prize preview, exact grid/slot stopping steps, and reward celebration policy.
- `useDrawExperience.ts` owns the request-to-receipt state machine, retry, live reduced-motion path, and completion watchdog.
- `DrawExperienceModal.vue` hosts one accessible draw and receipt surface.
- `DrawCelebration.vue` renders a short, in-modal CSS burst for positive rewards so it stays above Telegram's modal surface without a canvas layer.
- `presenters/README.md` maps the individual visual modes.
- `selection.test.ts` checks Random coverage, bounded previews, selected-prize insertion, exact stopping for every preview size, and celebration eligibility.
- `useDrawExperience.test.ts` checks request, retry, reduced-motion, and stalled-animation flows.
