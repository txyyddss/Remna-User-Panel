import { describe, expect, it, vi } from "vitest";

import { createLogsStore } from "../src/lib/admin/stores/logsStore.js";
import { createPaymentsStore } from "../src/lib/admin/stores/paymentsStore.js";
import { createAdminSupportStore } from "../src/lib/admin/stores/supportStore.js";
import { createUsersStore } from "../src/lib/admin/stores/usersStore.js";
import { deferred, snapshot } from "./helpers.js";

const at = (key, _params, fallback) => fallback || key;

describe("admin request ordering", () => {
  it("keeps the most recently opened payment detail", async () => {
    const first = deferred();
    const second = deferred();
    const api = vi.fn((path) => (path.endsWith("/1") ? first.promise : second.promise));
    const store = createPaymentsStore({ api, at });

    const firstRequest = store.openPayment(1);
    const secondRequest = store.openPayment(2);
    second.resolve({ ok: true, payment: { payment_id: 2, status: "pending" } });
    await secondRequest;
    first.resolve({ ok: true, payment: { payment_id: 1, status: "succeeded" } });
    await firstRequest;

    expect(snapshot(store)).toMatchObject({
      openedPaymentId: 2,
      openedPayment: { payment_id: 2 },
      paymentDetailLoading: false,
    });
  });

  it("keeps the latest payment page", async () => {
    const first = deferred();
    const second = deferred();
    const api = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const store = createPaymentsStore({ api, at });

    const firstRequest = store.loadPayments();
    store.update((state) => ({ ...state, paymentsPage: 1 }));
    const secondRequest = store.loadPayments();
    second.resolve({ ok: true, payments: [{ payment_id: 2 }], total: 1 });
    await secondRequest;
    first.resolve({ ok: true, payments: [{ payment_id: 1 }], total: 1 });
    await firstRequest;

    expect(snapshot(store).payments).toEqual([{ payment_id: 2 }]);
  });

  it("keeps the latest users and logs pages", async () => {
    const userFirst = deferred();
    const userSecond = deferred();
    const userApi = vi
      .fn()
      .mockReturnValueOnce(userFirst.promise)
      .mockReturnValueOnce(userSecond.promise);
    const users = createUsersStore({ api: userApi, onToast: vi.fn(), at });
    const firstUsersRequest = users.loadUsers();
    users.update((state) => ({ ...state, usersPage: 1 }));
    const secondUsersRequest = users.loadUsers();
    userSecond.resolve({ ok: true, users: [{ user_id: 2 }], total: 1 });
    await secondUsersRequest;
    userFirst.resolve({ ok: true, users: [{ user_id: 1 }], total: 1 });
    await firstUsersRequest;
    expect(snapshot(users).users).toEqual([{ user_id: 2 }]);

    const logFirst = deferred();
    const logSecond = deferred();
    const logApi = vi
      .fn()
      .mockReturnValueOnce(logFirst.promise)
      .mockReturnValueOnce(logSecond.promise);
    const logs = createLogsStore({ api: logApi });
    const firstLogsRequest = logs.loadLogs();
    logs.update((state) => ({ ...state, logsPage: 1 }));
    const secondLogsRequest = logs.loadLogs();
    logSecond.resolve({ ok: true, logs: [{ id: 2 }], total: 1 });
    await secondLogsRequest;
    logFirst.resolve({ ok: true, logs: [{ id: 1 }], total: 1 });
    await firstLogsRequest;
    expect(snapshot(logs).logs).toEqual([{ id: 2 }]);
  });

  it("keeps the latest support filter result", async () => {
    const first = deferred();
    const second = deferred();
    const api = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const store = createAdminSupportStore({ api, onToast: vi.fn(), at });

    const firstRequest = store.loadList();
    store.setFilter("priority", "urgent");
    const secondRequest = store.loadList();
    second.resolve({ ok: true, tickets: [{ ticket_id: 2 }] });
    await secondRequest;
    first.resolve({ ok: true, tickets: [{ ticket_id: 1 }] });
    await firstRequest;

    expect(snapshot(store).tickets).toEqual([{ ticket_id: 2 }]);
    store.destroy();
  });
});
