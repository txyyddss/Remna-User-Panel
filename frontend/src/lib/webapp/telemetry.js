const FONT_CANDIDATES = [
  "Arial",
  "Calibri",
  "Cambria",
  "Courier New",
  "Georgia",
  "Helvetica Neue",
  "Noto Sans",
  "PingFang SC",
  "Roboto",
  "Segoe UI",
  "SimSun",
  "Times New Roman",
];

async function digest(value) {
  const bytes = new TextEncoder().encode(String(value || ""));
  const hash = await crypto.subtle.digest("SHA-256", bytes);
  return Array.from(new Uint8Array(hash), (byte) => byte.toString(16).padStart(2, "0")).join("");
}

function canvasSignal() {
  try {
    const canvas = document.createElement("canvas");
    canvas.width = 280;
    canvas.height = 60;
    const ctx = canvas.getContext("2d");
    ctx.textBaseline = "alphabetic";
    ctx.fillStyle = "#e35b35";
    ctx.fillRect(8, 8, 52, 31);
    ctx.fillStyle = "#17324d";
    ctx.font = "17px Georgia";
    ctx.fillText("Remna 指纹 ◇ 123", 12, 32);
    ctx.globalCompositeOperation = "multiply";
    ctx.fillStyle = "rgba(23, 214, 122, .72)";
    ctx.beginPath();
    ctx.arc(64, 29, 18, 0, Math.PI * 2);
    ctx.fill();
    return canvas.toDataURL();
  } catch {
    return "unavailable";
  }
}

function webglSignal() {
  try {
    const gl = document.createElement("canvas").getContext("webgl");
    if (!gl) return "unavailable";
    const extension = gl.getExtension("WEBGL_debug_renderer_info");
    return JSON.stringify({
      vendor: extension
        ? gl.getParameter(extension.UNMASKED_VENDOR_WEBGL)
        : gl.getParameter(gl.VENDOR),
      renderer: extension
        ? gl.getParameter(extension.UNMASKED_RENDERER_WEBGL)
        : gl.getParameter(gl.RENDERER),
      version: gl.getParameter(gl.VERSION),
      shading: gl.getParameter(gl.SHADING_LANGUAGE_VERSION),
      maxTexture: gl.getParameter(gl.MAX_TEXTURE_SIZE),
    });
  } catch {
    return "unavailable";
  }
}

function fontSignal() {
  if (!document.fonts?.check) return "unavailable";
  return FONT_CANDIDATES.filter((font) => document.fonts.check(`14px "${font}"`)).join("|");
}

async function audioSignal() {
  try {
    const OfflineAudio = window.OfflineAudioContext || window.webkitOfflineAudioContext;
    if (!OfflineAudio) return "unavailable";
    const context = new OfflineAudio(1, 4410, 44100);
    const oscillator = context.createOscillator();
    const compressor = context.createDynamicsCompressor();
    oscillator.type = "triangle";
    oscillator.frequency.value = 10000;
    oscillator.connect(compressor);
    compressor.connect(context.destination);
    oscillator.start(0);
    const rendered = await context.startRendering();
    const data = rendered.getChannelData(0);
    let sum = 0;
    for (let index = 500; index < 1500; index += 1) sum += Math.abs(data[index]);
    return sum.toFixed(12);
  } catch {
    return "unavailable";
  }
}

function bucket(value, step) {
  const number = Number(value || 0);
  return number ? String(Math.ceil(number / step) * step) : "0";
}

export async function collectBrowserFingerprint() {
  if (!window.crypto?.subtle) throw new Error("fingerprint_crypto_unavailable");
  const hints = navigator.userAgentData
    ? await navigator.userAgentData.getHighEntropyValues([
        "architecture",
        "bitness",
        "model",
        "platformVersion",
        "uaFullVersion",
      ])
    : {};
  const raw = {
    canvas: canvasSignal(),
    webgl: webglSignal(),
    fonts: fontSignal(),
    audio: await audioSignal(),
    browser: JSON.stringify({ ua: navigator.userAgent, hints }),
    platform: JSON.stringify({ platform: navigator.platform, vendor: navigator.vendor }),
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "unknown",
    screen: `${bucket(screen.width, 100)}x${bucket(screen.height, 100)}@${bucket(devicePixelRatio, 0.25)}`,
    hardware: `${bucket(navigator.hardwareConcurrency, 2)}:${bucket(navigator.deviceMemory, 2)}:${navigator.maxTouchPoints || 0}`,
    language: `${navigator.language}|${(navigator.languages || []).join(",")}`,
  };
  const fingerprint = {};
  await Promise.all(
    Object.entries(raw).map(async ([key, value]) => {
      fingerprint[key] = await digest(value);
    })
  );
  fingerprint.full = await digest(
    Object.keys(fingerprint)
      .sort()
      .map((key) => fingerprint[key])
      .join("|")
  );
  return fingerprint;
}

export async function sendTelemetryHeartbeat() {
  try {
    const fingerprint = await collectBrowserFingerprint();
    const params = new URLSearchParams(window.location.search);
    const inviteCode =
      params.get("ref") ||
      params.get("start") ||
      params.get("start_param") ||
      params.get("startapp") ||
      "";
    await fetch("/api/telemetry/heartbeat", {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "application/json" },
      body: JSON.stringify({ fingerprint, invite_code: inviteCode }),
    });
  } catch (_error) {
    void _error;
  }
}
