import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { browserTelegramLogin, loadTelegramLoginLibrary } from "../src/lib/webapp/telegramLogin.js";

let appendedScripts = [];

beforeEach(() => {
  appendedScripts = [];
  vi.spyOn(document.head, "appendChild").mockImplementation((node) => {
    appendedScripts.push(node);
    return node;
  });
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  document.head
    .querySelectorAll('script[src*="telegram-login.js"]')
    .forEach((node) => node.remove());
  delete window.Telegram;
});

describe("Telegram browser login", () => {
  it("resolves after the library loads and OAuth returns an ID token", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({ ok: true, json: async () => ({ nonce: "nonce-1" }) }))
    );
    const loginPromise = browserTelegramLogin(10, () => "en", { timeoutMs: 1000 });
    await vi.waitFor(() => {
      expect(appendedScripts).toHaveLength(1);
    });
    const script = appendedScripts[0];
    window.Telegram = {
      Login: {
        auth: (_options, callback) => callback({ id_token: "token-1" }),
      },
    };
    script.dispatchEvent(new Event("load"));

    await expect(loginPromise).resolves.toEqual({ id_token: "token-1", nonce: "nonce-1" });
  });

  it("times out a library that never finishes loading", async () => {
    vi.useFakeTimers();
    const promise = loadTelegramLoginLibrary({ timeoutMs: 25 });
    const assertion = expect(promise).rejects.toMatchObject({ name: "AbortError" });
    await vi.advanceTimersByTimeAsync(25);
    await assertion;
  });

  it("times out when the OAuth callback never fires", async () => {
    vi.useFakeTimers();
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({ ok: true, json: async () => ({ nonce: "nonce-2" }) }))
    );
    window.Telegram = { Login: { auth: vi.fn() } };
    const promise = browserTelegramLogin(10, () => "en", { timeoutMs: 25 });
    const assertion = expect(promise).rejects.toMatchObject({ name: "AbortError" });
    await vi.advanceTimersByTimeAsync(25);
    await assertion;
  });
});
