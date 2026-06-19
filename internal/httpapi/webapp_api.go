package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/auth"
	"remna-user-panel/internal/config"
	"remna-user-panel/internal/fx"
	"remna-user-panel/internal/payments"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/tariffs"
)

const (
	sessionCookieName = "rw_webapp_session"
	csrfCookieName    = "rw_webapp_csrf"
)

type sessionContext struct {
	Claims auth.SessionClaims
	User   webappUser
}

type webappUser struct {
	UserID       int64  `json:"user_id"`
	TelegramID   int64  `json:"telegram_id,omitempty"`
	Username     string `json:"username,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	LanguageCode string `json:"language_code"`
	PhotoURL     string `json:"telegram_photo_url,omitempty"`
	IsAdmin      bool   `json:"is_admin"`
}

func authTokenHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if pool == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "database_not_configured"})
			return
		}
		var payload struct {
			InitData string `json:"init_data"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		tgUser, err := auth.ValidateTelegramInitData(payload.InitData, settings.BotToken, 24*time.Hour)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_telegram_init_data"})
			return
		}
		language := normalizeWebLanguage(tgUser.LanguageCode, settings.DefaultLanguage)
		if err := upsertTelegramUser(r.Context(), pool, tgUser, language); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "user_upsert_failed"})
			return
		}
		manager := auth.NewManager(settings.WebAppSessionSecret, "")
		token, csrf, err := manager.Sign(tgUser.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "session_sign_failed"})
			return
		}
		setSessionCookies(w, r, token, csrf)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":         true,
			"csrf_token": csrf,
			"user": webappUser{
				UserID:       tgUser.ID,
				TelegramID:   tgUser.ID,
				Username:     tgUser.Username,
				FirstName:    tgUser.FirstName,
				LastName:     tgUser.LastName,
				LanguageCode: language,
				PhotoURL:     tgUser.PhotoURL,
				IsAdmin:      isAdminID(settings.AdminIDs, tgUser.ID),
			},
		})
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearCookie(w, r, sessionCookieName)
		clearCookie(w, r, csrfCookieName)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func meHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		catalog, err := tariffs.Load("data/tariffs.json")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_load_failed"})
			return
		}
		rate := fx.NewService(appsettings.NewStore(pool)).USDCNY(r.Context())
		plans := tariffs.WithCNYDisplay(
			catalog.Plans(session.User.LanguageCode, effectiveDefaultCurrency(r.Context(), settings, pool)),
			rate.Rate,
			rate.Source,
			rate.UpdatedAt,
		)
		methods := []payments.Method{}
		if registry != nil {
			methods = registry.Methods(r.Context(), session.User.LanguageCode, session.User.IsAdmin)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":              true,
			"user":            session.User,
			"subscription":    map[string]any{"active": false},
			"settings":        webappFeatureSettings(),
			"plans":           plans,
			"payment_methods": methods,
		})
	}
}

func createPaymentHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		if registry == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "payments_not_configured"})
			return
		}
		var payload struct {
			Method           string `json:"method"`
			PlanHash         string `json:"plan_hash"`
			RenewHWIDDevices bool   `json:"renew_hwid_devices"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if strings.TrimSpace(payload.PlanHash) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "plan_hash_required"})
			return
		}
		catalog, err := tariffs.Load("data/tariffs.json")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_load_failed"})
			return
		}
		defaultCurrency := effectiveDefaultCurrency(r.Context(), settings, pool)
		if len(catalog.Plans(session.User.LanguageCode, defaultCurrency)) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "tariffs_not_configured"})
			return
		}
		plan, found := catalog.FindPlanByHash(payload.PlanHash, session.User.LanguageCode, defaultCurrency)
		if !found {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "plan_not_found"})
			return
		}
		rate := fx.NewService(appsettings.NewStore(pool)).USDCNY(r.Context())
		plan = tariffs.WithCNYDisplay([]tariffs.Plan{plan}, rate.Rate, rate.Source, rate.UpdatedAt)[0]
		providerAmount, providerCurrency, err := providerCheckoutAmount(payload.Method, plan, rate)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "payment_method_not_supported"})
			return
		}
		planSnapshot, err := json.Marshal(plan)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "plan_snapshot_failed"})
			return
		}
		description := plan.Title
		if plan.Months > 0 {
			description = fmt.Sprintf("%s %d month", plan.Title, plan.Months)
		} else if plan.TrafficGB > 0 {
			description = fmt.Sprintf("%s %s GB", plan.Title, compactFloat(plan.TrafficGB))
		}
		response, err := registry.Create(r.Context(), payments.CreateOrderRequest{
			UserID:           session.User.UserID,
			MethodID:         payload.Method,
			Amount:           providerAmount,
			Currency:         providerCurrency,
			BaseAmount:       plan.BaseAmount,
			BaseCurrency:     plan.BaseCurrency,
			DisplayCNYAmount: plan.DisplayCNYAmount,
			FXRate:           plan.FXRate,
			FXSource:         plan.FXSource,
			FXUpdatedAt:      rate.UpdatedAt,
			PlanHash:         plan.PlanHash,
			PlanSnapshot:     planSnapshot,
			Description:      description,
			TariffKey:        plan.TariffKey,
			SaleMode:         plan.SaleMode,
			Months:           plan.Months,
			TrafficGB:        plan.TrafficGB,
			DeviceCount:      boolToInt(payload.RenewHWIDDevices),
			ClientIP:         clientIP(r),
			Language:         session.User.LanguageCode,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": paymentErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func paymentStatusHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		paymentID, err := parsePositiveInt64(chi.URLParam(r, "payment_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_payment_id"})
			return
		}
		order, err := registry.GetForUser(r.Context(), session.User.UserID, paymentID)
		if err != nil {
			status := http.StatusInternalServerError
			code := "payment_load_failed"
			if errors.Is(err, pgx.ErrNoRows) {
				status = http.StatusNotFound
				code = "payment_not_found"
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": code})
			return
		}
		writeJSON(w, http.StatusOK, paymentStatusPayload(order))
	}
}

func adminSettingsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	store := appsettings.NewStore(pool)
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet); !ok {
			return
		}
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":       true,
				"features": []string{"payments"},
				"sections": []map[string]any{{
					"id":     "payments",
					"fields": paymentSettingsFields(r.Context(), settings, store),
				}},
			})
			return
		}
		var payload struct {
			Updates map[string]any `json:"updates"`
			Deletes []string       `json:"deletes"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		allowed := allowedPaymentSettingKeys()
		errorsByKey := map[string]string{}
		for key, value := range payload.Updates {
			if !allowed[key] {
				errorsByKey[key] = "unsupported_setting"
				continue
			}
			normalized, err := normalizeSettingValue(key, value)
			if err != nil {
				errorsByKey[key] = err.Error()
				continue
			}
			if err := store.Upsert(r.Context(), key, normalized); err != nil {
				errorsByKey[key] = "save_failed"
			}
		}
		for _, key := range payload.Deletes {
			if !allowed[key] {
				errorsByKey[key] = "unsupported_setting"
				continue
			}
			if err := store.Delete(r.Context(), key); err != nil {
				errorsByKey[key] = "delete_failed"
			}
		}
		if len(errorsByKey) > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "errors": errorsByKey})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func adminPaymentsListHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		orders, total, err := registry.List(r.Context(), page, pageSize)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "payments_load_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "payments": orders, "total": total})
	}
}

func adminPaymentDetailHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		paymentID, err := parsePositiveInt64(chi.URLParam(r, "payment_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_payment_id"})
			return
		}
		order, err := registry.Get(r.Context(), paymentID)
		if err != nil {
			status := http.StatusInternalServerError
			code := "payment_load_failed"
			if errors.Is(err, pgx.ErrNoRows) {
				status = http.StatusNotFound
				code = "payment_not_found"
			}
			writeJSON(w, status, map[string]any{"ok": false, "error": code})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "payment": order})
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request, settings config.Settings, pool *pgxpool.Pool, csrf bool) (sessionContext, bool) {
	session, ok := requireSession(w, r, settings, pool, csrf)
	if !ok {
		return sessionContext{}, false
	}
	if !session.User.IsAdmin {
		writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "forbidden"})
		return sessionContext{}, false
	}
	return session, true
}

func requireSession(w http.ResponseWriter, r *http.Request, settings config.Settings, pool *pgxpool.Pool, csrf bool) (sessionContext, bool) {
	if pool == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "database_not_configured"})
		return sessionContext{}, false
	}
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "auth_required"})
		return sessionContext{}, false
	}
	manager := auth.NewManager(settings.WebAppSessionSecret, "")
	claims, err := manager.Verify(cookie.Value)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "auth_required"})
		return sessionContext{}, false
	}
	if csrf {
		got := r.Header.Get("X-CSRF-Token")
		csrfCookie, _ := r.Cookie(csrfCookieName)
		if got == "" || got != claims.CSRF || csrfCookie == nil || csrfCookie.Value != claims.CSRF {
			writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "invalid_csrf"})
			return sessionContext{}, false
		}
	}
	user, err := loadWebappUser(r.Context(), pool, claims.UserID, settings)
	if err != nil {
		status := http.StatusInternalServerError
		code := "user_load_failed"
		if errors.Is(err, pgx.ErrNoRows) {
			status = http.StatusUnauthorized
			code = "auth_required"
		}
		writeJSON(w, status, map[string]any{"ok": false, "error": code})
		return sessionContext{}, false
	}
	return sessionContext{Claims: claims, User: user}, true
}

func upsertTelegramUser(ctx context.Context, pool *pgxpool.Pool, user auth.TelegramUser, language string) error {
	_, err := pool.Exec(ctx, `
INSERT INTO users (user_id, telegram_id, username, first_name, last_name, language_code, telegram_photo_url, registration_date)
VALUES ($1,$1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (user_id) DO UPDATE SET
	telegram_id=EXCLUDED.telegram_id,
	username=EXCLUDED.username,
	first_name=EXCLUDED.first_name,
	last_name=EXCLUDED.last_name,
	language_code=EXCLUDED.language_code,
	telegram_photo_url=EXCLUDED.telegram_photo_url`,
		user.ID, emptyStringToNil(user.Username), emptyStringToNil(user.FirstName), emptyStringToNil(user.LastName),
		language, emptyStringToNil(user.PhotoURL))
	return err
}

func loadWebappUser(ctx context.Context, pool *pgxpool.Pool, userID int64, settings config.Settings) (webappUser, error) {
	var user webappUser
	var username, firstName, lastName, language, photo string
	err := pool.QueryRow(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), COALESCE(telegram_photo_url,'')
FROM users WHERE user_id=$1`, userID).Scan(
		&user.UserID, &user.TelegramID, &username, &firstName, &lastName, &language, &photo,
	)
	if err != nil {
		return webappUser{}, err
	}
	user.Username = username
	user.FirstName = firstName
	user.LastName = lastName
	user.LanguageCode = normalizeWebLanguage(language, settings.DefaultLanguage)
	user.PhotoURL = photo
	user.IsAdmin = isAdminID(settings.AdminIDs, user.UserID) || isAdminID(settings.AdminIDs, user.TelegramID)
	return user, nil
}

func setSessionCookies(w http.ResponseWriter, r *http.Request, token string, csrf string) {
	secure := isSecureRequest(r)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrf,
		Path:     "/",
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: name == sessionCookieName,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func webappFeatureSettings() map[string]any {
	return map[string]any{
		"my_devices_enabled":              false,
		"support_tickets_enabled":         false,
		"subscription_guides_enabled":     false,
		"trial_enabled":                   false,
		"tariff_change_enabled":           false,
		"subscription_auto_renew_enabled": false,
	}
}

func paymentStatusPayload(order payments.Order) map[string]any {
	paid := strings.EqualFold(order.Status, "paid") || strings.EqualFold(order.Status, "succeeded")
	return map[string]any{
		"ok":                  true,
		"payment_id":          order.PaymentID,
		"order_id":            order.OrderID,
		"status":              order.Status,
		"paid":                paid,
		"provider":            order.Provider,
		"method":              order.Method,
		"payment_type":        order.PaymentType,
		"payment_url":         order.PaymentURL,
		"qr_content":          order.QRContent,
		"display_amount":      order.DisplayAmount,
		"display_currency":    order.DisplayCurrency,
		"payment_address":     order.PaymentAddress,
		"network":             order.Network,
		"provider_payment_id": order.ProviderPaymentID,
	}
}

func paymentSettingsFields(ctx context.Context, settings config.Settings, store appsettings.Store) []map[string]any {
	effectiveWebhookBaseURL := store.String(ctx, "WEBHOOK_BASE_URL", settings.WebhookBaseURL)
	webhookConfigured := strings.TrimSpace(effectiveWebhookBaseURL) != ""
	fields := []paymentSettingField{
		{Key: "WEBHOOK_BASE_URL", Type: "string", Label: "Webhook base URL", Description: "Public HTTPS backend URL used by payment callbacks.", Subsection: "General", Fallback: settings.WebhookBaseURL},
		{Key: "DEFAULT_CURRENCY_SYMBOL", Type: "string", Label: "Default catalog currency", Description: "Primary catalog currency. New deployments should use USD.", Subsection: "Currency", Fallback: effectiveDefaultCurrency(ctx, settings, nil), Choices: []settingChoice{{Value: "USD", Label: "USD"}, {Value: "CNY", Label: "CNY"}}},
		{Key: "FX_PROVIDER", Type: "string", Label: "USD/CNY rate provider", Description: "Use frankfurter by default; custom uses the fixed rate below.", Subsection: "Currency", Fallback: "frankfurter", Choices: []settingChoice{{Value: "frankfurter", Label: "Frankfurter"}, {Value: "exchange_rate_api", Label: "ExchangeRate-API"}, {Value: "custom", Label: "Custom"}}},
		{Key: "FX_CUSTOM_USD_CNY", Type: "float", Label: "Custom USD/CNY rate", Description: "Used only when the provider is custom.", Subsection: "Currency", Fallback: ""},
		{Key: "FX_CACHE_TTL_SECONDS", Type: "int", Label: "FX cache TTL seconds", Description: "How long a successful rate response is reused.", Subsection: "Currency", Fallback: 3600},
		{Key: "PAYMENT_METHODS_ORDER", Type: "text", Label: "Payment method order", Description: "Comma-separated method ids, e.g. ezpay:alipay,bepusdt:usdt.polygon.", Fallback: strings.Join(settings.PaymentMethodsOrder, ",")},
		{Key: "EZPAY_ENABLED", Type: "bool", Label: "EZPay enabled", Description: "Enable EZPay payment methods.", Subsection: "EZPay", Fallback: settings.EZPay.Enabled, WebhookPath: "/webhook/ezpay", ProviderID: "ezpay", WebhookConfigured: webhookConfigured},
		{Key: "EZPAY_BASE_URL", Type: "string", Label: "EZPay base URL", Description: "Merchant API base URL.", Subsection: "EZPay", Fallback: settings.EZPay.BaseURL},
		{Key: "EZPAY_PID", Type: "int", Label: "EZPay PID", Description: "Merchant PID.", Subsection: "EZPay", Fallback: settings.EZPay.PID},
		{Key: "EZPAY_KEY", Type: "string", Label: "EZPay key", Description: "Merchant signing key.", Subsection: "EZPay", Fallback: settings.EZPay.Key, Secret: true},
		{Key: "EZPAY_RETURN_URL", Type: "string", Label: "EZPay return URL", Description: "Return URL after checkout.", Subsection: "EZPay", Fallback: settings.EZPay.ReturnURL},
		{Key: "PAYMENT_EZPAY_ALIPAY_LABEL_ZH", Type: "string", Label: "Alipay label ZH", Subsection: "EZPay", Fallback: "支付宝"},
		{Key: "PAYMENT_EZPAY_ALIPAY_LABEL_EN", Type: "string", Label: "Alipay label EN", Subsection: "EZPay", Fallback: "Alipay"},
		{Key: "PAYMENT_EZPAY_WXPAY_LABEL_ZH", Type: "string", Label: "WeChat label ZH", Subsection: "EZPay", Fallback: "微信支付"},
		{Key: "PAYMENT_EZPAY_WXPAY_LABEL_EN", Type: "string", Label: "WeChat label EN", Subsection: "EZPay", Fallback: "WeChat Pay"},
		{Key: "PAYMENT_EZPAY_USDT_LABEL_ZH", Type: "string", Label: "EZPay USDT label ZH", Subsection: "EZPay", Fallback: "EZPay USDT"},
		{Key: "PAYMENT_EZPAY_USDT_LABEL_EN", Type: "string", Label: "EZPay USDT label EN", Subsection: "EZPay", Fallback: "EZPay USDT"},
		{Key: "BEPUSDT_ENABLED", Type: "bool", Label: "BEPUSDT enabled", Description: "Enable BEPUSDT USDT methods.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.Enabled, WebhookPath: "/webhook/bepusdt", ProviderID: "bepusdt", WebhookConfigured: webhookConfigured},
		{Key: "BEPUSDT_BASE_URL", Type: "string", Label: "BEPUSDT base URL", Description: "BEPUSDT API base URL.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.BaseURL},
		{Key: "BEPUSDT_TOKEN", Type: "string", Label: "BEPUSDT token", Description: "BEPUSDT signing token.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.Token, Secret: true},
		{Key: "BEPUSDT_RETURN_URL", Type: "string", Label: "BEPUSDT return URL", Description: "Return URL after checkout.", Subsection: "BEPUSDT", Fallback: settings.BEPUSDT.ReturnURL},
		{Key: "PAYMENT_BEPUSDT_POLYGON_LABEL_ZH", Type: "string", Label: "Polygon label ZH", Subsection: "BEPUSDT", Fallback: "USDT Polygon"},
		{Key: "PAYMENT_BEPUSDT_POLYGON_LABEL_EN", Type: "string", Label: "Polygon label EN", Subsection: "BEPUSDT", Fallback: "USDT Polygon"},
		{Key: "PAYMENT_BEPUSDT_ARBITRUM_LABEL_ZH", Type: "string", Label: "Arbitrum label ZH", Subsection: "BEPUSDT", Fallback: "USDT Arbitrum"},
		{Key: "PAYMENT_BEPUSDT_ARBITRUM_LABEL_EN", Type: "string", Label: "Arbitrum label EN", Subsection: "BEPUSDT", Fallback: "USDT Arbitrum"},
		{Key: "PAYMENT_BEPUSDT_APTOS_LABEL_ZH", Type: "string", Label: "Aptos label ZH", Subsection: "BEPUSDT", Fallback: "USDT Aptos"},
		{Key: "PAYMENT_BEPUSDT_APTOS_LABEL_EN", Type: "string", Label: "Aptos label EN", Subsection: "BEPUSDT", Fallback: "USDT Aptos"},
	}
	result := make([]map[string]any, 0, len(fields))
	for _, field := range fields {
		result = append(result, field.toMap(ctx, store))
	}
	return result
}

type paymentSettingField struct {
	Key               string
	Type              string
	Label             string
	Description       string
	Subsection        string
	Fallback          any
	Secret            bool
	WebhookPath       string
	ProviderID        string
	WebhookConfigured bool
	Choices           []settingChoice
}

type settingChoice struct {
	Value string
	Label string
}

func (f paymentSettingField) toMap(ctx context.Context, store appsettings.Store) map[string]any {
	raw, overridden, _ := store.Get(ctx, f.Key)
	value := f.Fallback
	hasValue := false
	if overridden {
		hasValue = true
		switch f.Type {
		case "bool":
			var parsed bool
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
			}
		case "int":
			var parsed float64
			if json.Unmarshal(raw, &parsed) == nil {
				value = int(parsed)
			}
		default:
			var parsed string
			if json.Unmarshal(raw, &parsed) == nil {
				value = parsed
				hasValue = parsed != ""
			}
		}
	}
	mappedType := f.Type
	if mappedType == "string" {
		mappedType = "input"
	}
	item := map[string]any{
		"key":         f.Key,
		"type":        mappedType,
		"label":       f.Label,
		"description": f.Description,
		"value":       value,
		"overridden":  overridden,
	}
	if f.Subsection != "" {
		item["subsection"] = f.Subsection
	}
	if f.Secret {
		item["secret"] = true
		item["has_value"] = hasValue || fmt.Sprint(value) != ""
	}
	if f.WebhookPath != "" {
		item["webhook_path"] = f.WebhookPath
		item["provider_id"] = f.ProviderID
		item["webhook_requires_base_url"] = true
		item["webhook_base_url_configured"] = f.WebhookConfigured
	}
	if len(f.Choices) > 0 {
		choices := make([]map[string]string, 0, len(f.Choices))
		for _, choice := range f.Choices {
			choices = append(choices, map[string]string{"value": choice.Value, "label": choice.Label})
		}
		item["choices"] = choices
	}
	return item
}

func allowedPaymentSettingKeys() map[string]bool {
	result := map[string]bool{}
	for _, key := range []string{
		"WEBHOOK_BASE_URL", "DEFAULT_CURRENCY_SYMBOL", "FX_PROVIDER", "FX_CUSTOM_USD_CNY", "FX_CACHE_TTL_SECONDS",
		"PAYMENT_METHODS_ORDER",
		"EZPAY_ENABLED", "EZPAY_BASE_URL", "EZPAY_PID", "EZPAY_KEY", "EZPAY_RETURN_URL",
		"BEPUSDT_ENABLED", "BEPUSDT_BASE_URL", "BEPUSDT_TOKEN", "BEPUSDT_RETURN_URL",
		"PAYMENT_EZPAY_ALIPAY_LABEL_ZH", "PAYMENT_EZPAY_ALIPAY_LABEL_EN",
		"PAYMENT_EZPAY_WXPAY_LABEL_ZH", "PAYMENT_EZPAY_WXPAY_LABEL_EN",
		"PAYMENT_EZPAY_USDT_LABEL_ZH", "PAYMENT_EZPAY_USDT_LABEL_EN",
		"PAYMENT_BEPUSDT_POLYGON_LABEL_ZH", "PAYMENT_BEPUSDT_POLYGON_LABEL_EN",
		"PAYMENT_BEPUSDT_ARBITRUM_LABEL_ZH", "PAYMENT_BEPUSDT_ARBITRUM_LABEL_EN",
		"PAYMENT_BEPUSDT_APTOS_LABEL_ZH", "PAYMENT_BEPUSDT_APTOS_LABEL_EN",
	} {
		result[key] = true
	}
	return result
}

func normalizeSettingValue(key string, value any) (any, error) {
	switch key {
	case "EZPAY_ENABLED", "BEPUSDT_ENABLED":
		if typed, ok := value.(bool); ok {
			return typed, nil
		}
		switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
		case "true", "1", "yes", "on":
			return true, nil
		case "false", "0", "no", "off":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid_bool")
		}
	case "EZPAY_PID", "FX_CACHE_TTL_SECONDS":
		switch typed := value.(type) {
		case float64:
			return int(typed), nil
		case int:
			return typed, nil
		default:
			parsed, err := strconv.Atoi(strings.TrimSpace(fmt.Sprint(value)))
			if err != nil {
				return nil, fmt.Errorf("invalid_int")
			}
			return parsed, nil
		}
	case "DEFAULT_CURRENCY_SYMBOL":
		value := strings.ToUpper(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "USD", "CNY":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_currency")
		}
	case "FX_PROVIDER":
		value := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
		switch value {
		case "frankfurter", "exchange_rate_api", "custom":
			return value, nil
		default:
			return nil, fmt.Errorf("unsupported_fx_provider")
		}
	case "FX_CUSTOM_USD_CNY":
		if strings.TrimSpace(fmt.Sprint(value)) == "" {
			return "", nil
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(value)), 64)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid_float")
		}
		return strconv.FormatFloat(parsed, 'f', -1, 64), nil
	default:
		return strings.TrimSpace(fmt.Sprint(value)), nil
	}
}

func paymentErrorCode(err error) string {
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "disabled"):
		return "provider_disabled"
	case strings.Contains(text, "not configured"):
		return "provider_not_configured"
	case strings.Contains(text, "webhook"):
		return "webhook_base_url_required"
	case strings.Contains(text, "unsupported"):
		return "payment_method_not_supported"
	default:
		return "payment_create_failed"
	}
}

func normalizeWebLanguage(raw string, fallback string) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(raw), "_", "-"))
	if strings.HasPrefix(value, "zh") {
		return "zh"
	}
	if strings.HasPrefix(value, "en") {
		return "en"
	}
	fallback = strings.ToLower(strings.TrimSpace(fallback))
	if fallback == "en" {
		return "en"
	}
	return "zh"
}

func isAdminID(adminIDs []int64, id int64) bool {
	if id == 0 {
		return false
	}
	for _, adminID := range adminIDs {
		if adminID == id {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if first := strings.TrimSpace(strings.Split(value, ",")[0]); first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func emptyStringToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func parsePositiveInt64(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid_id")
	}
	return id, nil
}

func compactFloat(value float64) string {
	if value == float64(int64(value)) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func effectiveDefaultCurrency(ctx context.Context, settings config.Settings, pool *pgxpool.Pool) string {
	value := appsettings.NewStore(pool).String(ctx, "DEFAULT_CURRENCY_SYMBOL", settings.DefaultCurrency)
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || value == "RUB" {
		return "USD"
	}
	return value
}

func providerCheckoutAmount(methodID string, plan tariffs.Plan, rate fx.Rate) (float64, string, error) {
	provider, _, ok := strings.Cut(strings.TrimSpace(methodID), ":")
	if !ok || provider == "" {
		return 0, "", fmt.Errorf("invalid payment method")
	}
	baseAmount := plan.BaseAmount
	if baseAmount <= 0 {
		baseAmount = plan.Price
	}
	baseCurrency := strings.ToUpper(strings.TrimSpace(plan.BaseCurrency))
	if baseCurrency == "" {
		baseCurrency = strings.ToUpper(plan.Currency)
	}
	switch provider {
	case payments.ProviderEZPay:
		switch baseCurrency {
		case "CNY", "RMB":
			return roundMoney(baseAmount), "CNY", nil
		case "USD":
			return roundMoney(baseAmount * rate.Rate), "CNY", nil
		default:
			return roundMoney(plan.Price), strings.ToUpper(plan.Currency), nil
		}
	case payments.ProviderBEPUSDT:
		switch baseCurrency {
		case "USD":
			return roundMoney(baseAmount), "USD", nil
		case "CNY", "RMB":
			if rate.Rate <= 0 {
				return 0, "", fmt.Errorf("missing fx rate")
			}
			return roundMoney(baseAmount / rate.Rate), "USD", nil
		default:
			if strings.ToUpper(plan.Currency) == "USD" {
				return roundMoney(plan.Price), "USD", nil
			}
			return 0, "", fmt.Errorf("unsupported currency")
		}
	default:
		return 0, "", fmt.Errorf("unsupported provider")
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
