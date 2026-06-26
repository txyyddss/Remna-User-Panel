import { readCookie } from "./session.js";

export function createApiClient({
  apiBase = "",
  csrfCookieName = "rw_webapp_csrf",
  getCsrfToken = () => "",
  onUnauthorized = () => {},
} = {}) {
  const isFormDataBody = (body) => typeof FormData !== "undefined" && body instanceof FormData;

  async function api(path, options = {}) {
    const method = String(options.method || "GET").toUpperCase();
    const headers = { ...(options.headers || {}) };

    const csrf = getCsrfToken() || readCookie(csrfCookieName) || "";
    if (csrf && ["POST", "PUT", "PATCH", "DELETE"].includes(method)) {
      headers["X-CSRF-Token"] = csrf;
    }
    if (options.body && !headers["Content-Type"] && !isFormDataBody(options.body)) {
      headers["Content-Type"] = "application/json";
    }

    const response = await fetch(`${apiBase}${path}`, {
      ...options,
      headers,
      credentials: "same-origin",
    });
    const payload = await response.json().catch(() => ({}));
    if (response.status === 401) onUnauthorized();
    return payload;
  }

  async function publicApi(path, payload = {}, options = {}) {
    const fetchOptions = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      signal: options.signal,
      credentials: "same-origin",
    };
    if (options.keepalive) {
      fetchOptions.keepalive = true;
    }
    const response = await fetch(`${apiBase}${path}`, fetchOptions);
    return response.json();
  }

  return { api, publicApi };
}
