import { tick } from "svelte";

export function createHeightStageAnimator({
  durationMs = 360,
  getElement = () => null,
  settleDelayMs = 80,
  setState = () => {},
} = {}) {
  let timer = null;
  let frame = null;
  let animationId = 0;

  function clearPending() {
    if (typeof window === "undefined") return;
    if (timer) {
      window.clearTimeout(timer);
      timer = null;
    }
    if (frame) {
      window.cancelAnimationFrame(frame);
      frame = null;
    }
  }

  async function animate(applyChange) {
    const element = getElement();
    if (typeof window === "undefined" || !element) {
      applyChange();
      return;
    }

    const currentAnimationId = ++animationId;
    clearPending();

    const startHeight = Math.ceil(element.getBoundingClientRect().height);
    if (!startHeight) {
      applyChange();
      return;
    }

    setState({ instant: true, locked: true, style: `height:${startHeight}px;` });
    await tick();
    if (currentAnimationId !== animationId) return;

    void element.offsetHeight;
    applyChange();
    await tick();
    if (currentAnimationId !== animationId) return;

    const endHeight = Math.max(1, Math.ceil(element.scrollHeight));
    frame = window.requestAnimationFrame(() => {
      frame = null;
      if (currentAnimationId !== animationId) return;

      setState({ instant: false, locked: true, style: `height:${endHeight}px;` });
      timer = window.setTimeout(() => {
        timer = null;
        if (currentAnimationId !== animationId) return;
        setState({ instant: false, locked: false, style: "" });
      }, durationMs + settleDelayMs);
    });
  }

  function destroy() {
    animationId += 1;
    clearPending();
    setState({ instant: false, locked: false, style: "" });
  }

  return {
    animate,
    destroy,
  };
}
