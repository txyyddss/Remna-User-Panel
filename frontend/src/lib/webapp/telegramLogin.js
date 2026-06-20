const TELEGRAM_LOGIN_LIBRARY_URL = "https://oauth.telegram.org/js/telegram-login.js";
export const TELEGRAM_LOGIN_TIMEOUT_MS = 10_000;

let telegramLoginLibraryPromise = null;

function abortError(message = "telegram_login_timeout") {
  const error = new Error(message);
  error.name = "AbortError";
  return error;
}

export async function loadTelegramLoginLibrary(options = {}) {
  if (window.Telegram?.Login?.auth) return window.Telegram.Login;
  if (options.signal?.aborted) throw abortError("telegram_login_cancelled");
  if (telegramLoginLibraryPromise) return telegramLoginLibraryPromise;

  telegramLoginLibraryPromise = new Promise((resolve, reject) => {
    let script = document.querySelector(`script[src="${TELEGRAM_LOGIN_LIBRARY_URL}"]`);
    if (script?.dataset.telegramLoginState === "loaded") {
      script.remove();
      script = null;
    }
    const created = !script;
    script ||= document.createElement("script");
    const timeoutMs = Math.max(1, Number(options.timeoutMs || TELEGRAM_LOGIN_TIMEOUT_MS));
    let settled = false;

    const cleanup = () => {
      window.clearTimeout(timeoutId);
      script.removeEventListener("load", onLoad);
      script.removeEventListener("error", onError);
      options.signal?.removeEventListener("abort", onAbort);
    };
    const finish = (callback, value) => {
      if (settled) return;
      settled = true;
      cleanup();
      callback(value);
    };
    const fail = (error) => {
      script.dataset.telegramLoginState = "failed";
      script.remove();
      finish(reject, error);
    };
    const onLoad = () => {
      script.dataset.telegramLoginState = "loaded";
      const login = window.Telegram?.Login;
      if (login?.auth) finish(resolve, login);
      else fail(new Error("telegram_login_library_unavailable"));
    };
    const onError = () => fail(new Error("telegram_login_library_unavailable"));
    const onAbort = () => fail(abortError("telegram_login_cancelled"));
    const timeoutId = window.setTimeout(
      () => fail(abortError("telegram_login_timeout")),
      timeoutMs
    );

    script.addEventListener("load", onLoad, { once: true });
    script.addEventListener("error", onError, { once: true });
    options.signal?.addEventListener("abort", onAbort, { once: true });
    if (created) {
      script.src = TELEGRAM_LOGIN_LIBRARY_URL;
      script.async = true;
      script.dataset.telegramLoginState = "loading";
      document.head.appendChild(script);
    }
  }).finally(() => {
    telegramLoginLibraryPromise = null;
  });

  return telegramLoginLibraryPromise;
}

export async function browserTelegramLogin(clientId, language = () => "en", options = {}) {
  if (options.signal?.aborted) throw abortError("telegram_login_cancelled");
  const timeoutMs = Math.max(1, Number(options.timeoutMs || TELEGRAM_LOGIN_TIMEOUT_MS));
  const controller = new AbortController();
  let timedOut = false;
  const onAbort = () => controller.abort();
  const timeoutId = window.setTimeout(() => {
    timedOut = true;
    controller.abort();
  }, timeoutMs);
  options.signal?.addEventListener("abort", onAbort, { once: true });

  try {
    const nonceResponse = await fetch("/api/auth/telegram/nonce", {
      credentials: "include",
      headers: { Accept: "application/json" },
      signal: controller.signal,
    });
    const nonceData = await nonceResponse.json();
    if (!nonceResponse.ok || !nonceData?.nonce) throw nonceData;
    const login = await loadTelegramLoginLibrary({ ...options, signal: controller.signal });
    return await new Promise((resolve, reject) => {
      let settled = false;
      const cleanup = () => controller.signal.removeEventListener("abort", rejectOnAbort);
      const finish = (callback, value) => {
        if (settled) return;
        settled = true;
        cleanup();
        callback(value);
      };
      const rejectOnAbort = () =>
        finish(
          reject,
          abortError(timedOut ? "telegram_login_timeout" : "telegram_login_cancelled")
        );
      controller.signal.addEventListener("abort", rejectOnAbort, { once: true });
      try {
        login.auth(
          { client_id: Number(clientId), nonce: nonceData.nonce, lang: language?.() || "en" },
          (result) => {
            if (result?.id_token) {
              finish(resolve, { id_token: result.id_token, nonce: nonceData.nonce });
            } else {
              finish(reject, new Error(result?.error || "telegram_login_cancelled"));
            }
          }
        );
      } catch (error) {
        finish(reject, error);
      }
    });
  } catch (error) {
    if (controller.signal.aborted) {
      throw abortError(timedOut ? "telegram_login_timeout" : "telegram_login_cancelled");
    }
    throw error;
  } finally {
    window.clearTimeout(timeoutId);
    options.signal?.removeEventListener("abort", onAbort);
  }
}
