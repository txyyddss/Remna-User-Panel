# Presentation modes

Every presenter accepts the same prize-preview and authoritative-result props and emits finished only after its visual sequence. Unmounting cancels timers and GSAP work.

- `SimplePresenter.vue` provides an immediate receipt path.
- `WheelPresenter.vue` spins custom SVG sectors toward the selected prize.
- `GridPresenter.vue` steps through eight SVG perimeter positions.
- `CardsPresenter.vue` lets the member flip one of three cards, then reveals the recorded prize; a short idle fallback finishes the sequence.
- `GiftPresenter.vue` opens a CSS gift box by tap and raises the recorded prize; a short idle fallback finishes the sequence.
- `SlotPresenter.vue` scrolls a continuous reel and lets the member pull a lever to accelerate or stop it on the recorded prize.
- `ScratchPresenter.vue` reveals the server prize through pointer erasure.
- `usePreview.ts` rotates a bounded eight-prize preview while waiting for the server.
- `useSelectionTicker.ts` drives the grid highlight without deciding the prize.
