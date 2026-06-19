export function activationPaymentFailed(status) {
  const normalized = String(status?.status || "").toLowerCase();
  return (
    normalized === "failed" ||
    normalized === "canceled" ||
    normalized === "cancelled" ||
    normalized.startsWith("failed_")
  );
}

export function createActivationHandoff({ storageKey, ttlMs, now = () => Date.now() } = {}) {
  let fallbackState = null;

  function normalizeState(value) {
    return value && typeof value === "object"
      ? {
          pending: value.pending && typeof value.pending === "object" ? value.pending : null,
          acknowledged:
            value.acknowledged && typeof value.acknowledged === "object"
              ? value.acknowledged
              : null,
        }
      : { pending: null, acknowledged: null };
  }

  function isPendingFresh(pending) {
    const startedAt = Number(pending?.startedAt || 0);
    return Boolean(startedAt && now() - startedAt <= ttlMs);
  }

  function write(state) {
    const normalized = normalizeState(state);
    fallbackState = normalized;
    if (!storageKey) return;
    try {
      localStorage.setItem(storageKey, JSON.stringify(normalized));
    } catch (_error) {
      void _error;
    }
  }

  function read() {
    let state = fallbackState || { pending: null, acknowledged: null };
    if (storageKey) {
      try {
        const raw = localStorage.getItem(storageKey);
        if (raw) state = JSON.parse(raw);
      } catch (_error) {
        void _error;
      }
    }
    state = normalizeState(state);
    if (state.pending && !isPendingFresh(state.pending)) {
      state = { ...state, pending: null };
      write(state);
    }
    return state;
  }

  function userKey(payload = {}) {
    const payloadUser = payload?.user || {};
    return String(payloadUser.user_id ?? payloadUser.id ?? payloadUser.telegram_id ?? "").trim();
  }

  function subscriptionKey(payload = {}) {
    const payloadSubscription = payload?.subscription || {};
    if (!payloadSubscription?.active) return "";
    return [
      userKey(payload) || "anonymous",
      payloadSubscription.panel_short_uuid ||
        payloadSubscription.panel_uuid ||
        payloadSubscription.uuid ||
        payloadSubscription.subscription_id ||
        payloadSubscription.config_link ||
        payloadSubscription.connect_url ||
        "active",
      payloadSubscription.end_date || payloadSubscription.end_date_text || "",
      payloadSubscription.tariff_key || payloadSubscription.tariff_name || "",
      payloadSubscription.status || "",
    ]
      .map((part) => String(part || "").trim())
      .join("|");
  }

  function pendingMatchesUser(pending, payload = {}) {
    if (!pending) return false;
    const pendingUserKey = String(pending.userKey || "").trim();
    const currentUserKey = userKey(payload);
    return !pendingUserKey || !currentUserKey || pendingUserKey === currentUserKey;
  }

  function hasPending(payload = {}) {
    const pending = read().pending;
    return Boolean(pending && pendingMatchesUser(pending, payload));
  }

  function rememberPending(context = {}, payload = {}) {
    if (context.initialSubscriptionPayment === false) return;
    const state = read();
    write({
      ...state,
      pending: {
        kind: "initial_subscription",
        source: String(context.source || "payment"),
        paymentId: String(context.paymentId || ""),
        userKey: userKey(payload),
        startedAt: now(),
      },
    });
  }

  function clearPending() {
    const state = read();
    if (!state.pending) return;
    write({ ...state, pending: null });
  }

  function isAcknowledged(nextSubscriptionKey, state = read()) {
    return Boolean(
      nextSubscriptionKey && state.acknowledged?.subscriptionKey === nextSubscriptionKey
    );
  }

  function acknowledge(nextSubscriptionKey, context = {}, payload = {}, state = read()) {
    const pending = state.pending || {};
    write({
      ...state,
      pending: null,
      acknowledged: {
        subscriptionKey: nextSubscriptionKey,
        source: String(context.source || pending.source || "payment"),
        paymentId: String(context.paymentId || pending.paymentId || ""),
        userKey: userKey(payload),
        acknowledgedAt: now(),
      },
    });
  }

  return {
    acknowledge,
    clearPending,
    hasPending,
    isAcknowledged,
    isPendingFresh,
    pendingMatchesUser,
    read,
    rememberPending,
    subscriptionKey,
    write,
  };
}
