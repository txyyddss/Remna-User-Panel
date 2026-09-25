# Lucky-draw presentation

The draw coordinator sends one idempotent request before displaying a prize. Presenters receive the server result and can only animate toward it; they never choose a reward.

- `selection.ts` defines the seven display styles, random selection, bounded prize preview, and reward celebration policy.
- `useDrawExperience.ts` owns the request-to-receipt state machine, retry, reduced-motion path, and completion watchdog.
- `DrawExperienceModal.vue` hosts one accessible draw and receipt surface.
- `DrawCelebration.vue` owns positive-reward artwork and confetti.
- `presenters/README.md` maps the individual visual modes.
- `assets/README.md` maps bundled Lottie artwork.
- `selection.test.ts` checks Random coverage, bounded previews, selected-prize insertion, and celebration eligibility.
- `useDrawExperience.test.ts` checks request, retry, reduced-motion, and stalled-animation flows.
