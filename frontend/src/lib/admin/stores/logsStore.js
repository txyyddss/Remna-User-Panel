import { get, writable } from "svelte/store";
import { createRequestTracker } from "$lib/shared/requestTracker.js";

export function createLogsStore({ api }) {
  const state = writable({
    logs: [],
    logsTotal: 0,
    logsPage: 0,
    logsUserFilter: "",
    logsLoading: false,
  });

  const LOGS_PAGE_SIZE = 50;
  const logsTracker = createRequestTracker();

  async function loadLogs() {
    const requestId = logsTracker.next();
    const snapshot = get(state);
    const currentPage = snapshot.logsPage;
    const filter = snapshot.logsUserFilter;
    state.update((s) => ({ ...s, logsLoading: true }));

    try {
      let q = `/admin/logs?page=${currentPage}&page_size=${LOGS_PAGE_SIZE}`;
      if (filter.trim()) {
        q += `&user_id=${encodeURIComponent(filter.trim())}`;
      }
      const data = await api(q);
      if (!logsTracker.isStale(requestId) && data?.ok) {
        state.update((s) => ({
          ...s,
          logs: data.logs || [],
          logsTotal: data.total || 0,
        }));
      }
    } finally {
      if (!logsTracker.isStale(requestId)) {
        state.update((s) => ({ ...s, logsLoading: false }));
      }
    }
  }

  function setPage(page) {
    state.update((s) => ({ ...s, logsPage: page }));
    loadLogs();
  }

  function setFilter(filter) {
    state.update((s) => ({ ...s, logsUserFilter: filter }));
  }

  return {
    subscribe: state.subscribe,
    set: state.set,
    update: state.update,
    loadLogs,
    setPage,
    setFilter,
  };
}
