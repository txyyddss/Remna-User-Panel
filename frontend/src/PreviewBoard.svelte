<script>
  import {
    ArrowRight,
    CheckCircle2,
    Circle,
    Copy,
    Crown,
    Database,
    Download,
    Gift,
    Globe2,
    LockKeyhole,
    Mail,
    RefreshCw,
    Repeat2,
    Send,
    Ticket,
    UserRound,
    Zap,
  } from "$components/ui/icons.js";

  import Button from "$components/ui/button.svelte";
  import Card from "$components/ui/card.svelte";
  import { LinearProgress } from "$components/patterns/webapp/index.js";
  import BackTitle from "./preview/BackTitle.svelte";
  import PhoneFrame from "./preview/PhoneFrame.svelte";
  import PreviewMethods from "./preview/PreviewMethods.svelte";
  import PreviewNav from "./preview/PreviewNav.svelte";

  export let config = {};
  export let mockData = {};
  export let t = (key, params, fallback) => fallback || key;

  const title = config.title || "Subscription";
  const logoUrl = config.logoUrl || "/webapp-default-logo.webp";
  const plans = mockData.plans || [];
  const sub = mockData.subscription || {};
  const methods = mockData.payment_methods || [];
  const user = mockData.user || {};
  const tariffs = [
    [
      "subscription",
      t("pb_subscription", {}, "Подписка"),
      t("pb_unlimited_traffic", {}, "Безлимитный трафик"),
      t("pb_ideal_for_regular", {}, "Идеально для постоянного использования"),
      Zap,
    ],
    ["traffic", t("pb_traffic", {}, "Трафик"), t("pb_traffic_packages", {}, "Пакеты гигабайт"), t("pb_pay_for_what_you_need", {}, "Платите только за нужный объем"), Database],
    ["premium", t("pb_premium", {}, "Премиум"), t("pb_max_speed", {}, "Максимальная скорость"), t("pb_priority_servers", {}, "Приоритетные серверы и поддержка"), Crown],
  ];
  const traffic = [
    [20, 290],
    [50, 590],
    [100, 990],
    [300, 2190],
  ];
  const settingsRows = [
    [Globe2, t("pb_language", {}, "Язык интерфейса"), "中文"],
    [
      Send,
      t("pb_telegram_link", {}, "Привязка Telegram"),
      user.telegram_linked ? `@${user.username || "username"}` : t("pb_not_linked", {}, "Не привязан"),
    ],
    [Mail, t("pb_email_link", {}, "Привязка почты"), user.email || t("pb_not_linked", {}, "Не привязана")],
    [UserRound, t("pb_logout", {}, "Выйти"), t("pb_end_session", {}, "Завершить сессию")],
  ];
  const previewTelegramName =
    user.first_name || (user.username ? `@${user.username}` : t("pb_tg_not_linked", {}, "Telegram не привязан"));
  const previewEmail = user.email || t("pb_email_not_linked", {}, "Почта не привязана");
  const previewTelegramId = user.telegram_id ? `TG ID ${user.telegram_id}` : t("pb_tg_id_not_linked", {}, "TG ID не привязан");
  const previewAvatar = user.telegram_photo_url || "";

  function money(value) {
    return `${value} ₽`;
  }
</script>

<div class="preview-board" style={`--accent: ${config.primaryColor || "#00fe7a"};`}>
  <PhoneFrame number="1" label={t("pb_home_screen", {}, "Главный экран")}>
    <main class="home-layout">
      <div class="login-brand home-brand">
        <div class="brand-mark brand-mark-xl"><img class="loaded" src={logoUrl} alt="" /></div>
        <h1>{title}</h1>
      </div>
      <div class="home-bottom">
        <Card class="status-card">
          <div class="sub-status">
            <CheckCircle2 size={23} />
            <div>
              <h2>{t("pb_subscription_active", {}, "Подписка активна")}</h2>
              <p>до {sub.end_date_text}</p>
            </div>
          </div>
        </Card>
        <Card>
          <div class="traffic-top">
            <span>{t("pb_traffic_used", {}, "Использовано трафика")}</span><strong
              >{sub.traffic_used} из {sub.traffic_limit}</strong
            >
          </div>
          <LinearProgress value={18} />
          <div class="traffic-percent">18%</div>
        </Card>
        <div class="action-stack">
          <Button class="wide"><Download size={17} />{t("pb_install_and_configure", {}, "Установить и настроить")}</Button>
          <Button variant="secondary" class="wide"><RefreshCw size={17} />{t("pb_renew", {}, "Продлить")}</Button>
          <Button variant="secondary" class="wide"><Repeat2 size={17} />{t("pb_change_tariff", {}, "Сменить тариф")}</Button>
        </div>
      </div>
      <PreviewNav active="home" />
    </main>
  </PhoneFrame>

  <PhoneFrame number="2" label={t("pb_select_tariff", {}, "Выбор тарифа")}>
    <div class="preview-header">
      <div class="brand-row">
        <div class="brand-mark"><img class="loaded" src={logoUrl} alt="" /></div>
        <strong>{title}</strong>
      </div>
    </div>
    <div class="tariff-list">
      {#each tariffs as tariff, index}
        <div class:active={index === 0} class="select-card">
          <span class="select-icon"><svelte:component this={tariff[4]} size={24} /></span>
          <span><strong>{tariff[1]}</strong><small>{tariff[2]}</small><em>{tariff[3]}</em></span>
          {#if index === 0}<CheckCircle2 size={21} />{:else}<Circle size={21} />{/if}
        </div>
      {/each}
    </div>
    <Button class="wide bottom-action">Далее <ArrowRight size={17} /></Button>
  </PhoneFrame>

  <PhoneFrame number="3" label={t("pb_payment_subscription", {}, "Оплата тарифа — подписка")} wide>
    <BackTitle title={t("pb_subscription", {}, "Подписка")} subtitle={t("pb_choose_period", {}, "Выберите срок подписки")} />
    <div class="period-grid">
      {#each plans as plan, index}
        <div class:active={index === 1} class="period-card">
          <strong>{plan.title}</strong><span>{money(plan.price)}</span><small
            >{money(Math.round(plan.price / plan.months))}{t("pb_per_month_short", {}, "/мес")}</small
          >
          {#if index === 1}<CheckCircle2 size={18} />{/if}
        </div>
      {/each}
    </div>
    <Card class="total-card"
      ><span>{t("pb_total", {}, "Итого")}<br /><small>{t("pb_to_pay", {}, "К оплате")}</small></span><strong>790 ₽</strong></Card
    >
    <PreviewMethods {methods} />
    <Button class="wide bottom-action">{t("pb_pay_amount", { amount: "790 ₽" }, "Оплатить 790 ₽")} <LockKeyhole size={16} /></Button>
  </PhoneFrame>

  <PhoneFrame number="4" label={t("pb_payment_traffic", {}, "Оплата тарифа — трафик")} wide>
    <BackTitle title={t("pb_traffic", {}, "Трафик")} subtitle={t("pb_choose_traffic_package", {}, "Выберите пакет трафика")} />
    <div class="period-grid">
      {#each traffic as pack, index}
        <div class:active={index === 2} class="period-card">
          <strong>{pack[0]} ГБ</strong><span>{money(pack[1])}</span><small
            >{money(Math.round(pack[1] / pack[0]))}{t("pb_per_gb_short", {}, "/ГБ")}</small
          >
          {#if index === 2}<CheckCircle2 size={18} />{/if}
        </div>
      {/each}
    </div>
    <Card class="total-card"
      ><span>{t("pb_total", {}, "Итого")}<br /><small>{t("pb_to_pay", {}, "К оплате")}</small></span><strong>990 ₽</strong></Card
    >
    <PreviewMethods {methods} />
    <Button class="wide bottom-action">{t("pb_pay_amount", { amount: "990 ₽" }, "Оплатить 990 ₽")} <LockKeyhole size={16} /></Button>
  </PhoneFrame>

  <PhoneFrame number="5" label={t("pb_change_tariff_title", {}, "Смена тарифа")}>
    <BackTitle title={t("pb_change_tariff_title", {}, "Смена тарифа")} subtitle={t("pb_remaining_days_recalc", { days: 12 }, "Остаток 12 дней будет пересчитан")} />
    <div class="tariff-list compact">
      <div class="select-card">
        <span><strong>{t("pb_subscription", {}, "Подписка")}</strong><small>{t("pb_unlimited_traffic", {}, "Безлимитный трафик")}</small></span><em>{t("pb_surcharge", { amount: "190 ₽" }, "Доплата 190 ₽")}</em
        ><Circle size={20} />
      </div>
      <div class="select-card active">
        <span><strong>{t("pb_traffic", {}, "Трафик")}</strong><small>{t("pb_traffic_packages", {}, "Пакеты гигабайт")}</small></span><em
          >{t("pb_no_surcharge", {}, "Доплата не требуется")}</em
        ><CheckCircle2 size={20} />
      </div>
      <div class="select-card">
        <span><strong>{t("pb_premium", {}, "Премиум")}</strong><small>{t("pb_max_speed", {}, "Максимальная скорость")}</small></span><em
          >{t("pb_surcharge", { amount: "390 ₽" }, "Доплата 390 ₽")}</em
        ><Circle size={20} />
      </div>
    </div>
    <Button class="wide bottom-action">{t("pb_next", {}, "Далее")} <ArrowRight size={17} /></Button>
    <div class="preview-modal">
      <Repeat2 size={30} />
      <strong>{t("pb_change_without_surcharge", {}, "Сменить тариф без доплаты?")}</strong>
      <p>{t("pb_remaining_recalc", {}, "Остаток 12 дней будет пересчитан по новому тарифу.")}</p>
      <Button>{t("pb_yes_change", {}, "Да, сменить")}</Button>
      <Button variant="secondary">{t("pb_cancel", {}, "Отмена")}</Button>
    </div>
  </PhoneFrame>

  <PhoneFrame number="6" label={t("pb_invite_friend", {}, "Пригласить друга")}>
    <div class="preview-header">
      <div class="brand-row">
        <div class="brand-mark"><img class="loaded" src={logoUrl} alt="" /></div>
        <strong>{title}</strong>
      </div>
    </div>
    <Card>
      <div class="card-label">{t("pb_your_referral_link", {}, "Ваша реферальная ссылка")}</div>
      <div class="copy-row">
        <code>https://minishop.app/ref/ABCD1234</code><Button>{t("pb_copy", {}, "Копировать")} <Copy size={16} /></Button>
      </div>
    </Card>
    <Card class="bonus-card">
      <Gift size={42} />
      <div>
        <span>{t("pb_your_bonus", {}, "Ваш бонус")}</span><strong>{t("pb_days_per_friend", {}, "+7 дней за каждого друга")}</strong>
        <p>{t("pb_friend_gets_days", {}, "Друг получит +3 дня к подписке.")}</p>
      </div>
    </Card>
    <Button variant="outline" class="wide"><Ticket size={18} />{t("pb_activate_promo", {}, "Активировать промокод")}</Button>
  </PhoneFrame>

  <PhoneFrame number="7" label={t("pb_settings", {}, "Настройки")}>
    <div class="preview-header">
      <div class="brand-row">
        <div class="brand-mark"><img class="loaded" src={logoUrl} alt="" /></div>
        <strong>{title}</strong>
      </div>
    </div>
    <Card class="settings-profile">
      <div class="settings-avatar">
        {#if previewAvatar}
          <img src={previewAvatar} alt={t("pb_user_avatar", {}, "Аватар пользователя")} />
        {:else}
          <UserRound size={27} />
        {/if}
      </div>
      <div class="settings-profile-meta">
        <strong>{previewTelegramName}</strong>
        <small>{previewEmail}</small>
        <small>{previewTelegramId}</small>
      </div>
    </Card>
    <div class="settings-list">
      {#each settingsRows as row}
        <div class="settings-row">
          <svelte:component this={row[0]} size={20} />
          <span><strong>{row[1]}</strong><small>{row[2]}</small></span>
          <ArrowRight size={16} />
        </div>
      {/each}
    </div>
  </PhoneFrame>

  <PhoneFrame number="8" label={t("pb_login", {}, "Логин")} wide>
    <div class="login-brand small">
      <div class="brand-mark brand-mark-xl"><img class="loaded" src={logoUrl} alt="" /></div>
      <h1>{title}</h1>
      <p>{t("pb_login_description", {}, "Войдите в свой аккаунт")}</p>
    </div>
    <Card class="auth-card">
      <div class="field-label">{t("pb_email_login", {}, "Вход по email")}</div>
      <div class="auth-email-stack">
        <div class="input muted">Email</div>
        <Button class="wide"><Mail size={17} />{t("pb_login_by_email", {}, "Войти по почте")}</Button>
      </div>
      <div class="or-line"><span></span>{t("pb_or", {}, "или")}<span></span></div>
      <Button variant="telegram" class="wide telegram-login-button">
        <span class="telegram-login-text"><Send size={17} />{t("pb_login_via_telegram", {}, "Войти через телеграм")}</span>
      </Button>
    </Card>
  </PhoneFrame>

  <PhoneFrame number="9" label={t("pb_code_verification", {}, "Подтверждение по коду")} wide>
    <BackTitle title={t("pb_code_verification", {}, "Подтверждение по email")} subtitle={t("pb_code_sent_to", { email: "user@example.com" }, "Мы отправили код на user@example.com")} />
    <div class="otp-slots static">
      {#each [1, 2, 3, 4, 5, 6] as digit}<span>{digit}</span>{/each}
    </div>
    <Button class="wide bottom-action">{t("pb_confirm", {}, "Подтвердить")}</Button>
    <button class="link-button"><RefreshCw size={15} />{t("pb_resend_code", { time: "00:45" }, "Отправить код повторно (00:45)")}</button>
  </PhoneFrame>
</div>
