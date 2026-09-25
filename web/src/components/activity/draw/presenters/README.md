# Presentation modes

Every presenter accepts the same prize-preview and authoritative-result props and emits finished only after its visual sequence. Unmounting cancels timers and GSAP work.

- `SimplePresenter.vue` provides an immediate receipt path.
- `WheelPresenter.vue` spins custom SVG sectors toward the selected prize.
- `GridPresenter.vue` steps through eight SVG perimeter positions.
- `CardsPresenter.vue` flips a prize card with Motion.
- `GiftPresenter.vue` plays the locally bundled gift animation and waits for its completion event.
- `SlotPresenter.vue` decelerates an SVG reel to the selected prize.
- `ScratchPresenter.vue` reveals the server prize through pointer erasure.
- `usePreview.ts` rotates a bounded eight-prize preview while waiting for the server.
- `useSelectionTicker.ts` drives grid and slot highlight without deciding the prize.
