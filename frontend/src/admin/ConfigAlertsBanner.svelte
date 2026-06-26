<script>
  import { getContext } from "svelte";
  import { RefreshCw, TriangleAlert } from "$components/ui/icons.js";
  import { AdminButton } from "$components/patterns/admin/index.js";

  export let at = (key, _params = {}, fallback = "") => fallback || key;
  export let section = "stats";
  export let onNavigate = () => {};

  const healthStore = getContext("healthStore");

  const MESSAGE_FALLBACKS = {
    data_dir_missing:
      "Data directory ({path}) not found. Verify the data volume is mounted.",
    data_dir_not_writable: "No write permission to {path} — backups, tariffs and logos won't be saved.",
    backups_dir_not_writable: "Backups directory {path} is not writable.",
    tariffs_config_invalid: "Tariffs file {path} cannot be read: {error}",
    subscription_page_config_invalid: "Subscription guides config cannot be read: {error}",
    provider_not_configured:
      "Provider {provider} is enabled but not configured — payments won't work.",
    provider_webhook_needs_base_url:
      "Provider {provider} needs WEBHOOK_BASE_URL for webhooks, but it is not set.",
    no_payment_methods: "No payment methods are enabled.",
    mini_app_url_missing:
      "SUBSCRIPTION_MINI_APP_URL is not set — the Mini App button won't appear in the bot.",
    mini_app_url_not_https:
      "SUBSCRIPTION_MINI_APP_URL must start with https:// (currently {url}).",
    redis_not_configured:
      "REDIS_URL is not set — bot dialog state and cache won't persist across restarts.",
    smtp_incomplete: "SMTP is not fully configured — email login won't work.",
    proxy_not_trusted:
      "Запросы приходят через прокси {remote}, которого нет в TRUSTED_PROXIES — вебхуки платёжных провайдеров могут отклоняться по IP.",
    bot_token_invalid: "Telegram отверг BOT_TOKEN — бот не работает.",
    telegram_api_error: "Не удалось обратиться к Telegram API: {error}",
    telegram_webhook_missing: "Вебхук Telegram не установлен — бот не получает обновления.",
    telegram_webhook_mismatch: "Вебхук Telegram указывает на {actual}, ожидается {expected}.",
    telegram_webhook_error: "Telegram сообщает об ошибке доставки вебхука: {error}",
    telegram_webhook_pending: "В очереди Telegram скопилось {count} необработанных обновлений.",
    panel_api_not_configured:
      "PANEL_API_URL и PANEL_API_KEY не заданы — синхронизация и выдача подписок не работают.",
    panel_api_unreachable: "Панель Remnawave недоступна по адресу {url}.",
  };

  const SECTION_FALLBACK_LABELS = {
    settings: "Настройки",
    payments: "Платежи",
    backups: "Бэкапы",
    tariffs: "Тарифы",
    appearance: "Внешний вид",
    users: "Пользователи",
  };

  function interpolate(template, params = {}) {
    return String(template || "").replace(/\{(\w+)\}/g, (match, key) =>
      params[key] !== undefined && params[key] !== null ? String(params[key]) : match
    );
  }

  function alertText(alert) {
    const fallback = interpolate(
      MESSAGE_FALLBACKS[alert.message_key] || alert.message_key,
      alert.params
    );
    return at(`health_${alert.message_key}`, alert.params || {}, fallback);
  }

  function sectionLabel(id) {
    return at(`nav_${id}`, {}, SECTION_FALLBACK_LABELS[id] || id);
  }

  $: alerts = $healthStore?.alerts || [];
  $: healthLoading = $healthStore?.healthLoading;
  $: isDashboard = section === "stats";
  $: visibleAlerts = isDashboard
    ? alerts
    : alerts.filter((alert) => (alert.sections || []).includes(section));
  $: errorCount = visibleAlerts.filter((alert) => alert.severity === "error").length;
</script>

{#if visibleAlerts.length}
  <div
    class="admin-config-alerts"
    class:admin-config-alerts-error={errorCount > 0}
    role="alert"
    aria-live="polite"
  >
    <div class="admin-config-alerts-head">
      <span class="admin-config-alerts-title">
        <TriangleAlert size={15} />
        {at("health_title", {}, "Проблемы конфигурации")}
      </span>
      {#if isDashboard}
        <AdminButton
          onclick={() => healthStore.loadHealth({ refresh: true })}
          disabled={healthLoading}
        >
          <RefreshCw size={13} />
          {at("health_refresh", {}, "Проверить снова")}
        </AdminButton>
      {/if}
    </div>
    <ul class="admin-config-alerts-list">
      {#each visibleAlerts as alert (alert.id)}
        <li class="admin-config-alert admin-config-alert-{alert.severity}">
          <span class="admin-config-alert-dot" aria-hidden="true"></span>
          <span class="admin-config-alert-text">{alertText(alert)}</span>
          {#if isDashboard && (alert.sections || []).length}
            <span class="admin-config-alert-links">
              {#each alert.sections as sectionId (sectionId)}
                <button
                  type="button"
                  class="admin-config-alert-link"
                  on:click={() => onNavigate(sectionId)}
                >
                  {sectionLabel(sectionId)}
                </button>
              {/each}
            </span>
          {/if}
        </li>
      {/each}
    </ul>
  </div>
{/if}
