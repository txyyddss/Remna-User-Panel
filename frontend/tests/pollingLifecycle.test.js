import { afterEach, describe, expect, it, vi } from "vitest";

import { createAdminSupportStore } from "../src/lib/admin/stores/supportStore.js";
import { createBillingStore } from "../src/lib/webapp/stores/billingStore.js";

afterEach(() => {
  vi.useRealTimers();
});

describe("polling lifecycle", () => {
  it("stops admin support polling after destroy", async () => {
    vi.useFakeTimers();
    Object.defineProperty(document, "visibilityState", { configurable: true, value: "visible" });
    const api = vi.fn(async () => ({ ok: true, stats: {} }));
    const store = createAdminSupportStore({ api, onToast: vi.fn(), at: (key) => key });

    store.startStatsPolling();
    await vi.advanceTimersByTimeAsync(0);
    expect(api).toHaveBeenCalledTimes(1);
    store.destroy();
    await vi.advanceTimersByTimeAsync(120_000);
    expect(api).toHaveBeenCalledTimes(1);
  });

  it("cancels payment status polling after destroy", async () => {
    vi.useFakeTimers();
    const fetchPaymentStatus = vi.fn(async () => ({ ok: true, status: "pending" }));
    const store = createBillingStore({
      billing: {
        fetchPaymentStatus,
        planPaymentBody: () => ({}),
        postPayment: async () => ({ ok: true, action: "invoice_sent", payment_id: "payment-1" }),
      },
      loadData: vi.fn(),
      t: (key) => key,
      showToast: vi.fn(),
      openExternalLink: vi.fn(),
    });
    store.update((state) => ({
      ...state,
      selectedMethod: "provider",
      selectedPlan: { plan_hash: "plan" },
    }));

    await store.createPayment();
    store.destroy();
    await vi.advanceTimersByTimeAsync(10_000);
    expect(fetchPaymentStatus).not.toHaveBeenCalled();
  });
});
