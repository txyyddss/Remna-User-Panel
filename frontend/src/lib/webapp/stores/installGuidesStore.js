import { get, writable } from "svelte/store";
import { createRequestTracker } from "$lib/shared/requestTracker.js";

export function createInstallGuidesStore({ api, t, showToast }) {
  let inFlight = null;
  const requestTracker = createRequestTracker();
  const state = writable({
    enabled: false,
    config: null,
    source: null,
    subscription: null,
    error: "",
    loading: false,
    loaded: false,
  });

  async function fetchGuides(path, force = false) {
    if (inFlight?.path === path) return inFlight.promise;
    const snapshot = get(state);
    if (!force && snapshot?.loaded) return snapshot;
    const generation = requestTracker.next();
    const promise = (async () => {
      state.update((s) => ({
        ...s,
        loading: true,
        loaded: force ? false : s.loaded,
        error: "",
      }));
      try {
        const response = await api(path);
        if (requestTracker.isStale(generation)) return get(state);
        const next = {
          enabled: Boolean(response?.enabled),
          config: response?.config || null,
          source: response?.source || null,
          subscription: response?.subscription || null,
          error: response?.error || "",
          loading: false,
          loaded: true,
        };
        state.set(next);
        return next;
      } catch (error) {
        if (requestTracker.isStale(generation)) return get(state);
        const message =
          error?.message || t("wa_install_unavailable", {}, "Instructions unavailable");
        if (typeof showToast === "function") showToast(message);
        const next = {
          enabled: false,
          config: null,
          source: null,
          subscription: null,
          error: message,
          loading: false,
          loaded: true,
        };
        state.set(next);
        return next;
      } finally {
        if (inFlight?.generation === generation) inFlight = null;
      }
    })();
    inFlight = { path, promise, generation };
    return promise;
  }

  async function load(force = false) {
    return fetchGuides("/subscription-guides", force);
  }

  async function loadPublic(shareToken, force = false) {
    const encoded = encodeURIComponent(String(shareToken || ""));
    return fetchGuides(`/subscription-guides/public/${encoded}`, force);
  }

  function reset() {
    requestTracker.next(); // invalidate pending requests
    inFlight = null;
    state.set({
      enabled: false,
      config: null,
      source: null,
      subscription: null,
      error: "",
      loading: false,
      loaded: false,
    });
  }

  return {
    subscribe: state.subscribe,
    set: state.set,
    update: state.update,
    load,
    loadPublic,
    reset,
  };
}
