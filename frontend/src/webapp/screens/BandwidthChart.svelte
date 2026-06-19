<script>
  import Card from "$components/ui/card.svelte";

  export let bandwidthData = [];
  export let t = (key) => key;

  const BAR_MAX_HEIGHT = 120;
  $: maxBytes = Math.max(1, ...bandwidthData.map((d) => Number(d?.bytes || 0)));
  $: bars = bandwidthData.map((d) => ({
    ...d,
    height: Math.max(2, (Number(d?.bytes || 0) / maxBytes) * BAR_MAX_HEIGHT),
    label: d?.label || "",
    value: d?.value || "0",
  }));
</script>

<Card class="bandwidth-chart-card">
  <h3 class="bandwidth-chart-title">{t("wa_bandwidth_usage_title")}</h3>
  <div
    class="bandwidth-bars"
    role="img"
    aria-label={t("wa_bandwidth_usage_chart_aria")}
  >
    {#each bars as bar, i (i)}
      <div class="bandwidth-bar-col" style="animation-delay: {i * 0.08}s">
        <div class="bandwidth-bar-value">{bar.value}</div>
        <div
          class="bandwidth-bar"
          style="height: {bar.height}px"
          aria-hidden="true"
        ></div>
        <div class="bandwidth-bar-label">{bar.label}</div>
      </div>
    {/each}
  </div>
</Card>

<style>
  .bandwidth-chart-card {
    margin-top: 1rem;
    padding: 1rem;
    text-align: center;
  }

  .bandwidth-chart-title {
    font-size: 0.875rem;
    font-weight: 600;
    margin: 0 0 1rem;
    opacity: 0.8;
  }

  .bandwidth-bars {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    gap: 0.5rem;
    height: 160px;
    padding: 0 0.25rem;
  }

  .bandwidth-bar-col {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    flex: 1;
    max-width: 48px;
    animation: barFadeIn 0.5s ease both;
  }

  .bandwidth-bar-value {
    font-size: 0.65rem;
    font-weight: 600;
    opacity: 0.7;
    white-space: nowrap;
  }

  .bandwidth-bar {
    width: 100%;
    max-width: 36px;
    border-radius: 6px 6px 2px 2px;
    background: linear-gradient(180deg, var(--color-accent, #00fe7a) 0%, var(--color-accent-muted, #00c060) 100%);
    transition: height 0.4s ease;
    min-height: 2px;
  }

  .bandwidth-bar-label {
    font-size: 0.6rem;
    opacity: 0.5;
    white-space: nowrap;
  }

  @keyframes barFadeIn {
    from {
      opacity: 0;
      transform: translateY(12px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
