import { writable, get } from "svelte/store";

export function createDevicesStore({ api, t, showToast, panelUuid }) {
  const state = writable({
    ipsData: null,
    ipsLoaded: false,
    ipsBusy: false,
    ipsStatus: "",
    ipsIsError: false,
    ipsErrorCode: "",
    ipConfirmOpen: false,
    ipToDisconnect: null,
    ipDisconnectBusy: false,
  });

  async function loadDevices(devicesEnabled, force = false) {
    const s = get(state);
    if (!devicesEnabled || s.ipsBusy || (s.ipsLoaded && !force)) return;
    state.update((s) => ({
      ...s,
      ipsBusy: true,
      ipsStatus: "",
      ipsIsError: false,
      ipsErrorCode: "",
    }));
    try {
      const uuid = typeof panelUuid === "function" ? panelUuid() : panelUuid;
      // Use Remnawave IP control API to fetch active IPs
      const fetchRes = await api("/ip-control/fetch-ips/" + encodeURIComponent(uuid), { method: "POST" });
      if (!fetchRes?.ok) throw fetchRes;
      const jobId = fetchRes?.job_id;
      if (!jobId) throw { message: t("wa_ips_load_failed") };
      // Poll for results
      let result = null;
      for (let i = 0; i < 30; i++) {
        await new Promise((r) => setTimeout(r, 1000));
        const pollRes = await api("/ip-control/fetch-users-ips/result/" + encodeURIComponent(jobId));
        if (pollRes?.ok && pollRes?.ips) {
          result = pollRes;
          break;
        }
      }
      const ips = Array.isArray(result?.ips) ? result.ips : [];
      state.update((s) => ({
        ...s,
        ipsData: { ips, current_ips: ips.length },
        ipsLoaded: true,
        ipsErrorCode: "",
      }));
    } catch (error) {
      state.update((s) => ({
        ...s,
        ipsStatus: error?.message || t("wa_ips_load_failed"),
        ipsIsError: true,
        ipsErrorCode: String(error?.error || ""),
        ipsLoaded: true,
      }));
    } finally {
      state.update((s) => ({ ...s, ipsBusy: false }));
    }
  }

  function openDeviceDisconnectDialog(ipEntry) {
    state.update((s) => ({ ...s, ipToDisconnect: ipEntry, ipConfirmOpen: true }));
  }

  function closeDeviceDisconnectDialog() {
    const s = get(state);
    if (s.ipDisconnectBusy) return;
    state.update((s) => ({ ...s, ipConfirmOpen: false, ipToDisconnect: null }));
  }

  async function disconnectDevice(devicesEnabled) {
    const s = get(state);
    const ip = String(s.ipToDisconnect?.ip || "").trim();
    if (!ip || s.ipDisconnectBusy) return;
    state.update((s) => ({ ...s, ipDisconnectBusy: true }));
    try {
      const uuid = typeof panelUuid === "function" ? panelUuid() : panelUuid;
      const response = await api("/ip-control/drop-connections", {
        method: "POST",
        body: JSON.stringify({ uuid, ips: [ip] }),
      });
      if (!response?.ok) throw response;
      showToast(t("wa_ip_disconnected"));
      state.update((s) => ({
        ...s,
        ipConfirmOpen: false,
        ipToDisconnect: null,
        ipsLoaded: false,
      }));
      await loadDevices(devicesEnabled, true);
    } catch (error) {
      showToast(error?.message || t("wa_ip_disconnect_failed"));
    } finally {
      state.update((s) => ({ ...s, ipDisconnectBusy: false }));
    }
  }

  return {
    subscribe: state.subscribe,
    set: state.set,
    update: state.update,
    loadDevices,
    openDeviceDisconnectDialog,
    closeDeviceDisconnectDialog,
    disconnectDevice,
  };
}
