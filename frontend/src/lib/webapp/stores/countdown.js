export function createStoreCountdown({ update, field, intervalMs = 1000 }) {
  let timer = null;

  function clear() {
    if (timer !== null && typeof window !== "undefined") window.clearInterval(timer);
    timer = null;
  }

  function start(seconds = 60) {
    clear();
    const initial = Math.max(0, Number(seconds || 60));
    update((state) => ({ ...state, [field]: initial }));
    if (typeof window === "undefined" || initial <= 0) return;

    timer = window.setInterval(() => {
      update((state) => {
        const current = Math.max(0, Number(state[field] || 0));
        if (current <= 1) {
          clear();
          return { ...state, [field]: 0 };
        }
        return { ...state, [field]: current - 1 };
      });
    }, intervalMs);
  }

  return { clear, start };
}
