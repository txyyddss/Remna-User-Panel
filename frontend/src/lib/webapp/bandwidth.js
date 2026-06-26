/**
 * Formats bandwidth data from a subscription payload for chart display.
 * Handles multiple payload shapes returned by the backend.
 * @param {object} sub - Subscription object with optional .bandwidth field.
 * @returns {Array<{bytes: number, label: string, value: string}>}
 */
export function formatBandwidthData(sub) {
  const payload = sub?.bandwidth;
  const raw =
    payload?.bandwidthLastSevenDays ||
    payload?.bandwidthLast30Days ||
    payload?.stats ||
    payload?.items ||
    payload?.data ||
    payload;
  const formatBytes = (value) => {
    const bytes = Number(value || 0);
    if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
    const units = ["B", "KB", "MB", "GB", "TB"];
    const unit = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
    const amount = bytes / 1024 ** unit;
    return `${amount >= 10 || unit === 0 ? amount.toFixed(0) : amount.toFixed(1)} ${units[unit]}`;
  };
  const entries = Array.isArray(raw)
    ? raw
    : raw && typeof raw === "object"
      ? Object.entries(raw)
          .filter(([, value]) => Number.isFinite(Number(value)))
          .map(([label, bytes]) => ({ label, bytes }))
      : [];
  return entries
    .map((entry) => {
      const bytes = Number(
        entry?.bytes ?? entry?.total ?? entry?.totalBytes ?? entry?.usedTrafficBytes ?? 0
      );
      return {
        bytes,
        label: entry?.label || entry?.date || entry?.day || entry?.timestamp || "",
        value: entry?.value || entry?.display || entry?.formatted || formatBytes(bytes),
      };
    })
    .filter((entry) => Number.isFinite(entry.bytes));
}
