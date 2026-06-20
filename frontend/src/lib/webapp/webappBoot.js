import { readMagicLoginToken } from "./authHelpers.js";
import { TELEGRAM_SDK_BOOT_TIMEOUT_MS } from "./constants.js";

/**
 * Initial auth / session bootstrap for the subscription webapp (non-preview).
 * Keeps side effects in App (mode, tg, token) via injected callbacks.
 */
export async function runWebappBoot({
  MOCK,
  setMode,
  hasTelegramLaunchParams,
  loadTelegramSdk,
  prepareTelegramMiniApp,
  loadData,
  showLogin,
  clearToken,
  clearManualLogoutFlag,
  isManuallyLoggedOut,
  hasEmailCodeLoginDeeplink,
  finalizeMagicLogin,
  finalizeTelegramAuth,
  getInitDataForBoot,
  getToken,
  getCsrfToken,
}) {
  setMode("loading");
  if (hasTelegramLaunchParams()) await loadTelegramSdk(TELEGRAM_SDK_BOOT_TIMEOUT_MS);
  prepareTelegramMiniApp();

  if (MOCK) {
    await loadData();
    return;
  }

  if (hasEmailCodeLoginDeeplink?.()) {
    clearManualLogoutFlag();
    clearToken();
    showLogin();
    return;
  }

  const magicToken = readMagicLoginToken();
  if (magicToken && (await finalizeMagicLogin(magicToken))) return;

  const initData = getInitDataForBoot();
  if (initData) {
    try {
      if (await finalizeTelegramAuth(initData, "init_data")) return;
    } catch (_error) {
      void _error;
    }
  }

  if (isManuallyLoggedOut()) {
    showLogin();
    return;
  }

  if (getToken() || getCsrfToken()) {
    try {
      await loadData();
      return;
    } catch {
      clearToken();
    }
  }

  showLogin();
}
