export function createBillingActions({ api }) {
  async function fetchTopupOptions(kind) {
    return api(`/tariffs/topup-options?kind=${encodeURIComponent(kind)}`);
  }

  async function fetchDeviceTopupOptions() {
    return api("/devices/topup-options");
  }

  async function fetchTariffChangeOptions() {
    return api("/tariffs/change-options");
  }

  async function postPayment(body) {
    return api("/payments", { method: "POST", body: JSON.stringify(body) });
  }

  async function fetchPaymentStatus(paymentId) {
    return api(`/payments/${encodeURIComponent(paymentId)}`);
  }

  async function postTariffChange(body) {
    return api("/tariffs/change", { method: "POST", body: JSON.stringify(body) });
  }

  async function postTariffChangePayment(body) {
    return api("/tariffs/change-payment", { method: "POST", body: JSON.stringify(body) });
  }

  async function postAutoRenew(enabled) {
    return api("/subscription/auto-renew", {
      method: "POST",
      body: JSON.stringify({ enabled: Boolean(enabled) }),
    });
  }

  function planPaymentBody(plan, method, options = {}) {
    return {
      plan_hash: plan.plan_hash,
      renew_hwid_devices: Boolean(options.renewHwidDevices),
      method,
    };
  }

  function topupPaymentBody(plan, method, fallbackTariffKey) {
    void fallbackTariffKey;
    return {
      plan_hash: plan.plan_hash,
      method,
    };
  }

  function deviceTopupPaymentBody(plan, method, fallbackTariffKey) {
    void fallbackTariffKey;
    return {
      plan_hash: plan.plan_hash,
      method,
    };
  }

  function changePaymentBody(action, target, method) {
    void target;
    return { plan_hash: action.plan_hash, method };
  }

  return {
    fetchTopupOptions,
    fetchDeviceTopupOptions,
    fetchTariffChangeOptions,
    postPayment,
    fetchPaymentStatus,
    postTariffChange,
    postTariffChangePayment,
    postAutoRenew,
    planPaymentBody,
    topupPaymentBody,
    deviceTopupPaymentBody,
    changePaymentBody,
  };
}
