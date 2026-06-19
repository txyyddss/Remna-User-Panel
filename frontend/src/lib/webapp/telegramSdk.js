export function readTelegramMiniAppInitDataFromLocation() {
  if (typeof window === "undefined") return "";
  const queryText = window.location.search.replace(/^\?/, "");
  const hashText = window.location.hash.replace(/^#/, "");
  for (const text of [queryText, hashText]) {
    if (!text) continue;
    const params = new URLSearchParams(text);
    const initData = params.get("tgWebAppData");
    if (initData) return initData;
  }
  return "";
}

export function createTelegramSdk({
  scriptUrl,
  bootTimeoutMs,
  actionTimeoutMs,
  miniAppAuthTimeoutMs,
  onStatusChange = () => {},
  onInitDataChange = () => {},
} = {}) {
  let tg = resolve();
  let sdkPromise = null;
  let launchParamsDetected = false;
  let initData = tg?.initData || readTelegramMiniAppInitDataFromLocation();
  if (initData) launchParamsDetected = true;

  function resolve() {
    return window.Telegram?.WebApp || null;
  }

  function setStatus(status) {
    onStatusChange(status);
  }

  function refresh() {
    tg = resolve();
    if (tg) setStatus("ready");
    initData = tg?.initData || readTelegramMiniAppInitDataFromLocation();
    onInitDataChange(initData);
    if (initData) launchParamsDetected = true;
    return tg;
  }

  function hasLaunchParams() {
    refresh();
    if (launchParamsDetected || initData) {
      launchParamsDetected = true;
      return true;
    }
    const queryText = window.location.search.replace(/^\?/, "");
    const hashText = window.location.hash.replace(/^#/, "");
    const detected = [queryText, hashText].some((text) => {
      if (!text) return false;
      const params = new URLSearchParams(text);
      return ["tgWebAppData", "tgWebAppVersion", "tgWebAppPlatform", "tgWebAppThemeParams"].some(
        (key) => params.has(key)
      );
    });
    if (detected) launchParamsDetected = true;
    return detected;
  }

  function load(timeoutMs = bootTimeoutMs) {
    if (refresh()) return Promise.resolve(tg);
    if (sdkPromise) return sdkPromise;
    if (typeof document === "undefined") return Promise.resolve(null);

    setStatus("loading");
    sdkPromise = new Promise((resolvePromise) => {
      const existingScript = document.querySelector("script[data-rw-telegram-web-app-sdk]");
      const script = existingScript || document.createElement("script");
      let resolved = false;
      let timeoutId = null;

      const resolveOnce = (value) => {
        if (resolved) return;
        resolved = true;
        if (timeoutId) window.clearTimeout(timeoutId);
        resolvePromise(value);
      };

      const refreshFromScript = () => {
        tg = resolve();
        setStatus(tg ? "ready" : "unavailable");
        return tg;
      };

      script.addEventListener("load", () => resolveOnce(refreshFromScript()), { once: true });
      script.addEventListener(
        "error",
        () => {
          setStatus("unavailable");
          resolveOnce(null);
        },
        { once: true }
      );

      if (!existingScript) {
        script.src = scriptUrl;
        script.async = true;
        script.defer = true;
        script.dataset.rwTelegramWebAppSdk = "1";
        document.head.appendChild(script);
      }

      timeoutId = window.setTimeout(() => {
        if (!tg) setStatus("unavailable");
        resolveOnce(tg);
      }, timeoutMs);
    }).finally(() => {
      sdkPromise = null;
    });
    return sdkPromise;
  }

  async function ensureForAction() {
    if (refresh()) return tg;
    return await load(actionTimeoutMs);
  }

  function createMiniAppAuthTimeout() {
    const controller = typeof AbortController === "undefined" ? null : new AbortController();
    let timedOut = false;
    let timeoutId = null;
    let timeoutPromise = new Promise(() => {});

    if (typeof window !== "undefined") {
      timeoutPromise = new Promise((_, reject) => {
        timeoutId = window.setTimeout(() => {
          timedOut = true;
          controller?.abort();
          const error = new Error("telegram_mini_app_auth_timeout");
          error.name = "AbortError";
          reject(error);
        }, miniAppAuthTimeoutMs);
      });
    }

    return {
      promise: timeoutPromise,
      get signal() {
        return controller?.signal;
      },
      get timedOut() {
        return timedOut;
      },
      clear() {
        if (timeoutId) window.clearTimeout(timeoutId);
        timeoutId = null;
      },
    };
  }

  return {
    get tg() {
      return tg;
    },
    get initData() {
      return initData;
    },
    refresh,
    hasLaunchParams,
    load,
    ensureForAction,
    createMiniAppAuthTimeout,
    readInitDataFromLocation: readTelegramMiniAppInitDataFromLocation,
  };
}
