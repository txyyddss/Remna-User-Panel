import { describe, expect, it, vi } from "vitest";

import { createInstallGuidesStore } from "../src/lib/webapp/stores/installGuidesStore.js";
import { deferred, snapshot } from "./helpers.js";

describe("install guide request ordering", () => {
  it("ignores older share-token responses", async () => {
    const first = deferred();
    const second = deferred();
    const api = vi.fn().mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const store = createInstallGuidesStore({ api, t: (key) => key, showToast: vi.fn() });

    const firstRequest = store.loadPublic("first", true);
    const secondRequest = store.loadPublic("second", true);
    second.resolve({ enabled: true, subscription: { user: "second" } });
    await secondRequest;
    first.resolve({ enabled: true, subscription: { user: "first" } });
    await firstRequest;

    expect(snapshot(store).subscription).toEqual({ user: "second" });
  });

  it("invalidates in-flight responses when reset", async () => {
    const request = deferred();
    const store = createInstallGuidesStore({
      api: vi.fn(() => request.promise),
      t: (key) => key,
      showToast: vi.fn(),
    });

    const pending = store.load(true);
    store.reset();
    request.resolve({ enabled: true, subscription: { user: "stale" } });
    await pending;

    expect(snapshot(store)).toMatchObject({ loaded: false, subscription: null });
  });
});
