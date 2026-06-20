<script>
  import { CircleX, Globe, RefreshCw } from "$components/ui/icons.js";

  import Button from "$components/ui/button.svelte";
  import Card from "$components/ui/card.svelte";
  import { EmptyCard, StatusMessage } from "$components/patterns/webapp/index.js";

  export let devicesBusy = false;
  export let devicesData = {};
  export let devicesErrorCode = "";
  export let devicesIsError = false;
  export let devicesLoaded = false;
  export let devicesStatus = "";
  export const subscription = {};

  export let loadDevices = () => {};
  export let openDeviceDisconnectDialog = () => {};
  export let t = (key) => key;

  $: ipList = Array.isArray(devicesData?.ips) ? devicesData.ips : [];
  $: hasIPs = ipList.length > 0;
  $: subscriptionNotActiveError =
    devicesErrorCode === "subscription_not_active" ||
    devicesStatus === "Subscription is not active";
  $: showInactiveNotice =
    !hasIPs &&
    !(devicesBusy && !devicesLoaded) &&
    (!devicesStatus || subscriptionNotActiveError);
</script>

<main class="content with-nav">
  <Card class="devices-summary-card">
    <div class="devices-summary-head">
      <Globe size={28} />
      <span>
        <strong>{t("wa_ips_title")}</strong>
        <small>{t("wa_ips_count", { current: ipList.length })}</small>
      </span>
      <Button
        variant="icon"
        size="icon"
        onclick={() => loadDevices(true)}
        disabled={devicesBusy}
        aria-label={t("wa_ips_refresh")}
      >
        <RefreshCw size={18} />
      </Button>
    </div>
  </Card>

  {#if devicesBusy && !devicesLoaded}
    <EmptyCard>{t("wa_ips_loading")}</EmptyCard>
  {:else if showInactiveNotice}
    <EmptyCard class="devices-empty-card devices-inactive-card">
      <CircleX size={28} />
      <span>{t("wa_home_subscription_inactive")}</span>
    </EmptyCard>
  {:else if devicesStatus}
    <EmptyCard>
      <StatusMessage error={devicesIsError}>{devicesStatus}</StatusMessage>
    </EmptyCard>
  {:else if !hasIPs}
    <EmptyCard class="devices-empty-card">
      <Globe size={28} />
      <span>{t("wa_ips_empty")}</span>
      <small>{t("wa_ips_empty_hint")}</small>
    </EmptyCard>
  {:else}
    <div class="devices-list">
      {#each ipList as ipEntry (ipEntry.ip)}
        <Card class="device-card">
          <div class="device-card-head">
            <div class="device-icon"><Globe size={20} /></div>
            <span>
              <strong>{ipEntry.ip}</strong>
              {#if ipEntry.location}
                <small>{ipEntry.location}</small>
              {/if}
            </span>
          </div>
          <div class="device-meta">
            {#if ipEntry.last_seen_text}
              <div>
                <span>{t("wa_ips_last_seen")}</span>
                <strong>{ipEntry.last_seen_text}</strong>
              </div>
            {/if}
            {#if ipEntry.user_agent}
              <div class="device-user-agent">
                <span>{t("wa_ips_user_agent")}</span>
                <small>{ipEntry.user_agent}</small>
              </div>
            {/if}
          </div>
          <Button
            variant="outline"
            class="wide device-disconnect-button"
            onclick={() => openDeviceDisconnectDialog(ipEntry)}
          >
            <CircleX size={17} />
            {t("wa_ips_disconnect")}
          </Button>
        </Card>
      {/each}
    </div>
  {/if}
</main>
