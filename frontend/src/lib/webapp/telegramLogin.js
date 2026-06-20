const TELEGRAM_LOGIN_LIBRARY_URL = "https://oauth.telegram.org/js/telegram-login.js";

async function loadTelegramLoginLibrary() {
  if (window.Telegram?.Login?.auth) return window.Telegram.Login;
  await new Promise((resolve, reject) => {
    const existing = document.querySelector(`script[src="${TELEGRAM_LOGIN_LIBRARY_URL}"]`);
    if (existing) {
      existing.addEventListener("load", resolve, { once: true });
      existing.addEventListener("error", reject, { once: true });
      return;
    }
    const script = document.createElement("script");
    script.src = TELEGRAM_LOGIN_LIBRARY_URL;
    script.async = true;
    script.onload = resolve;
    script.onerror = reject;
    document.head.appendChild(script);
  });
  if (!window.Telegram?.Login?.auth) throw new Error("telegram_login_library_unavailable");
  return window.Telegram.Login;
}

export async function browserTelegramLogin(clientId, language = () => "en") {
  const nonceResponse = await fetch("/api/auth/telegram/nonce", {
    credentials: "include",
    headers: { Accept: "application/json" },
  });
  const nonceData = await nonceResponse.json();
  if (!nonceResponse.ok || !nonceData?.nonce) throw nonceData;
  const login = await loadTelegramLoginLibrary();
  return new Promise((resolve, reject) => {
    login.auth(
      { client_id: Number(clientId), nonce: nonceData.nonce, lang: language?.() || "en" },
      (result) => {
        if (result?.id_token) resolve({ id_token: result.id_token, nonce: nonceData.nonce });
        else reject(new Error(result?.error || "telegram_login_cancelled"));
      }
    );
  });
}
