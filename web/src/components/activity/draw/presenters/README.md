# Presentation modes

Every presenter accepts the same prize-preview and authoritative-result props and emits finished only after its visual sequence. Unmounting cancels timers and GSAP work.

- `SimplePresenter.vue` provides an immediate receipt path.
- `WheelPresenter.vue` waits for a center press, spins custom SVG sectors, then slows toward the selected prize. A second press requests an earlier slowdown.
- `GridPresenter.vue` waits for a center press before stepping through eight SVG perimeter positions.
- `CardsPresenter.vue` lets the member flip one of three cards, then reveals the recorded prize; a short idle fallback finishes the sequence.
- `GiftPresenter.vue` opens a CSS gift box by tap and raises the recorded prize from behind the box; a short idle fallback finishes the sequence.
- `SlotPresenter.vue` scrolls a continuous reel at its base speed until the member pulls the attached lever. The lever accelerates a pending reel or settles it on the recorded prize.
- `ScratchPresenter.vue` runs a canvas scrape when the card is tapped or the modal requests Reveal result.
- `usePreview.ts` rotates a bounded eight-prize preview while waiting for the server.
- `useSelectionTicker.ts` drives the grid highlight without deciding the prize.
