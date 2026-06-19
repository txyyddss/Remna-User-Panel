package httpapi

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/fx"
	"remna-user-panel/internal/i18n"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/tariffs"
	"remna-user-panel/internal/webassets"
)

func registerExtraAPIRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry, panel *remnawave.Client) {
	_ = catalog
	router.Get("/api/tariffs/topup-options", webappPlansOptionsHandler(settings, pool, "topup"))
	router.Get("/api/devices/topup-options", webappPlansOptionsHandler(settings, pool, "devices"))
	router.Get("/api/tariffs/change-options", webappPlansOptionsHandler(settings, pool, "change"))
	router.Post("/api/tariffs/change", userTariffChangeHandler(settings, pool, panel))
	router.Post("/api/tariffs/change-payment", createPaymentHandler(settings, pool, registry))
	router.Post("/api/subscription/auto-renew", autoRenewHandler(settings, pool))
	router.Post("/api/promo/apply", promoApplyHandler(settings, pool, panel))
	router.Post("/api/referral/welcome-bonus/claim", referralWelcomeBonusHandler(settings, pool, panel))
	router.Post("/api/trial/activate", trialActivateHandler(settings, pool, panel))
	router.Get("/api/devices", devicesHandler(settings, pool, panel))
	router.Post("/api/devices/disconnect", disconnectDeviceHandler(settings, pool, panel))
	router.Post("/api/account/email/request", unavailableSessionMutation(settings, pool, "email_delivery_not_configured"))
	router.Post("/api/account/email/verify", unavailableSessionMutation(settings, pool, "email_delivery_not_configured"))
	router.Post("/api/account/password/request", unavailableSessionMutation(settings, pool, "email_delivery_not_configured"))
	router.Post("/api/account/password/confirm", unavailableSessionMutation(settings, pool, "email_delivery_not_configured"))
	router.Post("/api/account/telegram/link", unavailableSessionMutation(settings, pool, "telegram_link_requires_mini_app_login"))
	router.Post("/api/account/language", accountLanguageHandler(settings, pool))
	router.Post("/api/auth/email/request", unavailablePublicMutation("email_delivery_not_configured"))
	router.Post("/api/auth/email/verify", unavailablePublicMutation("email_delivery_not_configured"))
	router.Post("/api/auth/email/password", unavailablePublicMutation("email_password_login_not_configured"))
	router.Post("/api/auth/email/magic", unavailablePublicMutation("email_magic_login_not_configured"))
	router.Get("/api/subscription-guides", subscriptionGuidesHandler(settings, pool))
	router.Get("/api/subscription-guides/public/{share_token}", subscriptionGuidesHandler(settings, pool))

	router.Get("/api/support/tickets", supportListHandler(settings, pool, false))
	router.Post("/api/support/tickets", supportCreateHandler(settings, pool, false))
	router.Get("/api/support/tickets/{ticket_id}", supportDetailHandler(settings, pool, false))
	router.Post("/api/support/tickets/{ticket_id}/messages", supportMessageHandler(settings, pool, false))
	router.Post("/api/support/tickets/{ticket_id}/read", supportReadHandler(settings, pool, false))
	router.Get("/api/support/unread", supportUnreadHandler(settings, pool))

	router.Get("/api/admin/tariffs", adminTariffsHandler(settings, pool))
	router.Put("/api/admin/tariffs", adminTariffsHandler(settings, pool))
	router.Get("/api/admin/panel/internal-squads", adminSquadsHandler(settings, pool, panel))
	router.Get("/api/admin/payments/export.csv", adminPaymentsExportHandler(settings, pool, registry))
	router.Get("/api/admin/health", adminHealthHandler(settings, pool, panel))
	router.Get("/api/admin/stats", adminStatsHandler(settings, pool, panel))
	router.Post("/api/admin/sync", adminSyncHandler(settings, pool, panel))
	router.Get("/api/admin/logs", adminLogsHandler(settings, pool))
	router.Get("/api/admin/users", adminUsersListHandler(settings, pool, panel))
	router.Get("/api/admin/users/{user_id}", adminUserDetailHandler(settings, pool, panel))
	router.Delete("/api/admin/users/{user_id}", adminUserDeleteHandler(settings, pool, panel))
	router.Get("/api/admin/users/{user_id}/referrals", adminUserReferralsHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/ban", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/message", adminUserMessageHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/message/preview", adminMessagePreviewHandler(settings, pool))
	router.Get("/api/admin/users/{user_id}/telegram-profile-link", adminTelegramProfileLinkHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/telegram-profile-link", adminTelegramProfileLinkHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/extend", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/tariff", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/reset-trial", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/premium-override", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/regular-traffic-override", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/hwid-device-limit", adminUserActionHandler(settings, pool, panel))
	router.Post("/api/admin/users/{user_id}/traffic-grant", adminUserActionHandler(settings, pool, panel))

	router.Get("/api/admin/promos", adminListSettingHandler(settings, pool, "ADMIN_PROMOS", "promos"))
	router.Post("/api/admin/promos", adminCreateSettingItemHandler(settings, pool, "ADMIN_PROMOS", "promo"))
	router.Patch("/api/admin/promos/{id}", adminPatchSettingItemHandler(settings, pool, "ADMIN_PROMOS", "promo"))
	router.Delete("/api/admin/promos/{id}", adminDeleteSettingItemHandler(settings, pool, "ADMIN_PROMOS"))
	router.Get("/api/admin/ads", adminAdsListHandler(settings, pool))
	router.Post("/api/admin/ads", adminCreateSettingItemHandler(settings, pool, "ADMIN_ADS", "campaign"))
	router.Post("/api/admin/ads/{id}/toggle", adminPatchSettingItemHandler(settings, pool, "ADMIN_ADS", "campaign"))
	router.Delete("/api/admin/ads/{id}", adminDeleteSettingItemHandler(settings, pool, "ADMIN_ADS"))
	router.Get("/api/admin/broadcast/audience-counts", adminBroadcastAudienceHandler(settings, pool))
	router.Post("/api/admin/broadcast", adminBroadcastHandler(settings, pool))
	router.Get("/api/admin/backups", adminBackupsHandler(settings, pool))
	router.Post("/api/admin/backups/create", adminBackupCreateHandler(settings, pool))
	router.Post("/api/admin/backups/upload", adminUploadPlaceholderHandler(settings, pool, "archive"))
	router.Post("/api/admin/backups/restore", okAdminMutation(settings, pool))
	router.Get("/api/admin/themes", adminThemesHandler(settings, pool, assets))
	router.Put("/api/admin/themes", adminThemesHandler(settings, pool, assets))
	router.Post("/api/admin/appearance/logo", adminUploadPlaceholderHandler(settings, pool, "logo"))
	router.Delete("/api/admin/appearance/logo", okAdminMutation(settings, pool))
	router.Post("/api/admin/appearance/favicon", adminUploadPlaceholderHandler(settings, pool, "favicon"))
	router.Delete("/api/admin/appearance/favicon", okAdminMutation(settings, pool))
	router.Get("/api/admin/translations", adminTranslationsHandler(settings, pool))
	router.Put("/api/admin/translations", adminTranslationsHandler(settings, pool))
	router.Get("/api/admin/support/stats", adminSupportStatsHandler(settings, pool))
	router.Get("/api/admin/support/tickets", supportListHandler(settings, pool, true))
	router.Get("/api/admin/support/tickets/{ticket_id}", supportDetailHandler(settings, pool, true))
	router.Post("/api/admin/support/tickets/{ticket_id}/messages", supportMessageHandler(settings, pool, true))
	router.Post("/api/admin/support/tickets/{ticket_id}/read", supportReadHandler(settings, pool, true))
	router.Patch("/api/admin/support/tickets/{ticket_id}", supportPatchHandler(settings, pool))
}

func webappPlansOptionsHandler(settings config.Settings, pool *pgxpool.Pool, kind string) http.HandlerFunc {
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
		plans := tariffs.WithCNYDisplay(catalog.Plans(session.User.LanguageCode, effectiveDefaultCurrency(r.Context(), settings, pool)), rate.Rate, rate.Source, rate.UpdatedAt)
		response := map[string]any{"ok": true, "plans": plans, "fx": rate}
		switch kind {
		case "devices":
			response["tariff_name"] = ""
			response["plans"] = devicePlansFrom(plans)
		case "change":
			response["current"] = nil
			response["targets"] = changeTargetsFrom(plans)
		default:
			response["topup_kind"] = r.URL.Query().Get("kind")
			response["tariff_name"] = ""
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func userTariffChangeHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		if err := adminChangePanelTariff(r.Context(), settings, pool, panel, session.User.UserID, r); err != nil {
			writeAdminActionError(w, err)
			return
		}
		webUser, _ := loadWebappUser(r.Context(), pool, session.User.UserID, settings)
		var subscription any
		if panel != nil && panel.Configured(r.Context()) {
			if panelUser, found, _ := panelUserForWebUser(r.Context(), pool, panel, webUser); found {
				subscription = subscriptionFromPanelUser(r.Context(), pool, webUser, panelUser)
			}
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			TargetUserID: session.User.UserID,
			EventType:    "user_tariff_change",
			Content:      "tariff changed",
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "subscription": subscription, "active_subscription": subscription})
	}
}

func autoRenewHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Enabled bool `json:"enabled"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if err := saveUserAutoRenew(r.Context(), pool, session.User.UserID, payload.Enabled); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			TargetUserID: session.User.UserID,
			EventType:    "user_auto_renew",
			Content:      strconv.FormatBool(payload.Enabled),
			Payload:      map[string]any{"enabled": payload.Enabled},
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": payload.Enabled, "auto_renew_enabled": payload.Enabled})
	}
}

func devicesHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		if panel == nil || !panel.Configured(r.Context()) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "panel_not_configured"})
			return
		}
		panelUser, found, err := panelUserForWebUser(r.Context(), pool, panel, session.User)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		if !found {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "subscription_not_active"})
			return
		}
		devices, err := panel.GetUserDevices(r.Context(), stringValue(panelUser, "uuid"))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, mapDevicePayload(devices, panelUser))
	}
}

func disconnectDeviceHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		if panel == nil || !panel.Configured(r.Context()) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "panel_not_configured"})
			return
		}
		var payload struct {
			Token string `json:"token"`
			HWID  string `json:"hwid"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		hwid := strings.TrimSpace(firstNonEmpty(payload.HWID, payload.Token))
		if hwid == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "hwid_required"})
			return
		}
		panelUser, found, err := panelUserForWebUser(r.Context(), pool, panel, session.User)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		if !found {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "subscription_not_active"})
			return
		}
		if err := panel.DisconnectDevice(r.Context(), stringValue(panelUser, "uuid"), hwid); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func accountLanguageHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Language string `json:"language"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		language := normalizeWebLanguage(payload.Language, settings.DefaultLanguage)
		if _, err := pool.Exec(r.Context(), "UPDATE users SET language_code=$2 WHERE user_id=$1", session.User.UserID, language); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "language_update_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "language": language})
	}
}

func promoApplyHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Code string `json:"code"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		code := strings.TrimSpace(payload.Code)
		if code == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "promo_code_required"})
			return
		}
		promos := readSettingList(r.Context(), pool, "ADMIN_PROMOS")
		for index := range promos {
			if !strings.EqualFold(strings.TrimSpace(fmt.Sprint(promos[index]["code"])), code) {
				continue
			}
			if !settingItemActive(promos[index]) || promoExpired(promos[index]) {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_inactive"})
				return
			}
			if promoActivatedByUser(promos[index], session.User.UserID) {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_already_used"})
				return
			}
			maxActivations := int(int64Value(promos[index], "max_activations"))
			currentActivations := promoActivationCount(promos[index])
			if maxActivations > 0 && currentActivations >= maxActivations {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_exhausted"})
				return
			}
			days := int(int64Value(promos[index], "bonus_days"))
			if days <= 0 {
				days = int(int64Value(promos[index], "days"))
			}
			if days <= 0 {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_invalid_reward"})
				return
			}
			panelUser, err := grantPanelAccessDays(r.Context(), pool, panel, session.User, days, accessGrantOptions{Source: "promo:" + strings.ToUpper(code)})
			if err != nil {
				writePanelActionError(w, err)
				return
			}
			activations := anyList(promos[index]["activations"])
			activations = append(activations, map[string]any{"user_id": session.User.UserID, "activated_at": time.Now().UTC().Format(time.RFC3339)})
			promos[index]["activations"] = activations
			promos[index]["current_activations"] = currentActivations + 1
			_ = writeSettingList(r.Context(), pool, "ADMIN_PROMOS", promos)
			response := grantResponse(r.Context(), pool, session.User, panelUser, days)
			response["code"] = strings.TrimSpace(fmt.Sprint(promos[index]["code"]))
			writeJSON(w, http.StatusOK, response)
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "promo_not_found"})
	}
}

func trialActivateHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		store := appsettings.NewStore(pool)
		if !store.Bool(r.Context(), "TRIAL_ENABLED", false) {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "trial_not_enabled"})
			return
		}
		days := store.Int(r.Context(), "TRIAL_DURATION_DAYS", 0)
		if days <= 0 {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "trial_not_configured"})
			return
		}
		if !trialAvailableForUser(r.Context(), pool, session.User.UserID) {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "trial_already_used"})
			return
		}
		panelUser, err := grantPanelAccessDays(r.Context(), pool, panel, session.User, days, accessGrantOptions{
			Source:          "trial",
			TrafficLimitGB:  store.Float(r.Context(), "TRIAL_TRAFFIC_LIMIT_GB", 0),
			TrafficStrategy: store.String(r.Context(), "TRIAL_TRAFFIC_STRATEGY", settings.UserTrafficStrategy),
			SquadUUIDs:      splitRuntimeList(store.String(r.Context(), "TRIAL_SQUAD_UUIDS", "")),
			SetTrafficLimit: true,
		})
		if err != nil {
			writePanelActionError(w, err)
			return
		}
		recordTrialActivation(r.Context(), pool, session.User.UserID)
		writeJSON(w, http.StatusOK, grantResponse(r.Context(), pool, session.User, panelUser, days))
	}
}

func referralWelcomeBonusHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		store := appsettings.NewStore(pool)
		days := store.Int(r.Context(), "REFERRAL_WELCOME_BONUS_DAYS", 0)
		if days <= 0 {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "referral_welcome_not_enabled"})
			return
		}
		var referredBy int64
		var claimedAt sql.NullTime
		err := pool.QueryRow(r.Context(), "SELECT COALESCE(referred_by_id,0), referral_welcome_bonus_claimed_at FROM users WHERE user_id=$1", session.User.UserID).Scan(&referredBy, &claimedAt)
		if err != nil || referredBy == 0 {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "referral_welcome_not_available"})
			return
		}
		if claimedAt.Valid {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "referral_welcome_already_claimed"})
			return
		}
		panelUser, err := grantPanelAccessDays(r.Context(), pool, panel, session.User, days, accessGrantOptions{Source: "referral_welcome"})
		if err != nil {
			writePanelActionError(w, err)
			return
		}
		_, _ = pool.Exec(r.Context(), "UPDATE users SET referral_welcome_bonus_claimed_at=NOW() WHERE user_id=$1", session.User.UserID)
		writeJSON(w, http.StatusOK, grantResponse(r.Context(), pool, session.User, panelUser, days))
	}
}

func adminTariffsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet); !ok {
			return
		}
		path := "data/tariffs.json"
		if r.Method == http.MethodGet {
			catalog, err := loadTariffCatalogForAdmin(path)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_load_failed"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "catalog": catalog, "path": path, "provider_currency_support": providerCurrencySupport()})
			return
		}
		var payload struct {
			Catalog json.RawMessage `json:"catalog"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || !json.Valid(payload.Catalog) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_save_failed"})
			return
		}
		if err := os.WriteFile(path, append(payload.Catalog, '\n'), 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "tariffs_save_failed"})
			return
		}
		var saved any
		_ = json.Unmarshal(payload.Catalog, &saved)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "catalog": saved, "path": path, "provider_currency_support": providerCurrencySupport()})
	}
}

func adminSquadsHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		if panel == nil || !panel.Configured(r.Context()) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "squads": []any{}, "configured": false})
			return
		}
		squads, err := panel.GetInternalSquads(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "squads": squads, "configured": true})
	}
}

func adminPaymentsExportHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="payments.csv"`)
		writer := csv.NewWriter(w)
		_ = writer.Write([]string{"payment_id", "order_id", "user_id", "provider", "method", "amount", "currency", "base_amount", "base_currency", "status", "created_at", "paid_at"})
		if registry != nil {
			orders, _, err := registry.List(r.Context(), 0, 100)
			if err == nil {
				for _, order := range orders {
					paidAt := ""
					if order.PaidAt != nil {
						paidAt = order.PaidAt.Format(time.RFC3339)
					}
					_ = writer.Write([]string{
						strconv.FormatInt(order.PaymentID, 10), order.OrderID, strconv.FormatInt(order.UserID, 10),
						order.Provider, order.Method, fmt.Sprintf("%.2f", order.Amount), order.Currency,
						fmt.Sprintf("%.2f", order.BaseAmount), order.BaseCurrency, order.Status, order.CreatedAt.Format(time.RFC3339), paidAt,
					})
				}
			}
		}
		writer.Flush()
	}
}

func adminHealthHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		alerts := []map[string]any{}
		checks := map[string]any{"database": pool != nil}
		panelConfigured := panel != nil && panel.Configured(r.Context())
		checks["remnawave_configured"] = panelConfigured
		if !panelConfigured {
			alerts = append(alerts, map[string]any{"level": "warning", "code": "panel_not_configured", "message": "PANEL_API_URL and PANEL_API_KEY are required for Remnawave integration."})
		} else if _, err := panel.GetSystemStats(r.Context()); err != nil {
			checks["remnawave_api"] = false
			alerts = append(alerts, map[string]any{"level": "error", "code": panelErrorCode(err), "message": err.Error()})
		} else {
			checks["remnawave_api"] = true
		}
		payload := map[string]any{"ok": true, "status": "ok", "checks": checks, "alerts": alerts, "checked_at": time.Now().UTC()}
		if pool != nil {
			payload["db_pool"] = pool.Stat()
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func adminStatsHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		var users, payments int64
		var revenue float64
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&users)
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM payment_orders").Scan(&payments)
		_ = pool.QueryRow(r.Context(), "SELECT COALESCE(SUM(base_amount),0)::float8 FROM payment_orders WHERE status IN ('paid','succeeded')").Scan(&revenue)
		payload := map[string]any{
			"ok":         true,
			"users":      users,
			"payments":   payments,
			"revenue":    revenue,
			"series":     []any{},
			"panel_sync": LastPanelSyncStatus(r.Context(), pool),
		}
		if panel != nil && panel.Configured(r.Context()) {
			panelStats := map[string]any{}
			if stats, err := panel.GetSystemStats(r.Context()); err == nil {
				panelStats["system"] = stats
			}
			if bandwidth, err := panel.GetBandwidthStats(r.Context()); err == nil {
				panelStats["bandwidth"] = bandwidth
			}
			if nodes, err := panel.GetNodesStats(r.Context()); err == nil {
				panelStats["nodes"] = nodes
			}
			payload["panel"] = panelStats
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func adminSyncHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, true)
		if !ok {
			return
		}
		result, err := RunPanelSync(r.Context(), settings, pool, panel, 500)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			EventType:    "admin_panel_sync",
			Content:      result.Status,
			IsAdminEvent: true,
			Payload:      result,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error(), "panel_sync": result})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "panel_sync": result, "sync": result})
	}
}

func adminLogsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		var userID int64
		if raw := strings.TrimSpace(r.URL.Query().Get("user_id")); raw != "" {
			parsed, err := parsePositiveInt64(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
				return
			}
			userID = parsed
		}
		logs, total, err := adminMessageLogs(r.Context(), pool, page, pageSize, userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "logs_load_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "logs": logs, "total": total, "page": page, "page_size": pageSize})
	}
}

func adminUsersListHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 0 {
			page = 0
		}
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 25
		}
		where, args := adminUsersWhereClause(r)
		orderBy := adminUsersOrderBy(r.URL.Query().Get("sort"))
		needsMemoryFilter := needsAdminUsersMemoryFilter(r)
		limit := pageSize
		offset := page * pageSize
		if needsMemoryFilter {
			limit = 1000
			offset = 0
		}
		query := `
SELECT u.user_id, COALESCE(u.telegram_id,0), COALESCE(u.username,''), COALESCE(u.email,''), COALESCE(u.first_name,''), COALESCE(u.last_name,''),
	COALESCE(u.language_code,''), u.is_banned, u.registration_date,
	COALESCE((SELECT SUM(p.base_amount)::float8 FROM payment_orders p WHERE p.user_id=u.user_id AND p.status IN ('paid','succeeded')),0),
	COALESCE((SELECT COUNT(*) FROM payment_orders p WHERE p.user_id=u.user_id),0),
	COALESCE((SELECT COUNT(*) FROM users invitee WHERE invitee.referred_by_id=u.user_id),0)
FROM users u ` + where + ` ORDER BY ` + orderBy + ` LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
		args = append(args, limit, offset)
		rows, err := pool.Query(r.Context(), query, args...)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "users_load_failed"})
			return
		}
		defer rows.Close()
		users := []map[string]any{}
		for rows.Next() {
			var id, telegramID int64
			var username, email, firstName, lastName, language string
			var banned bool
			var created time.Time
			var paymentsTotal float64
			var paymentsCount, invitedCount int64
			if err := rows.Scan(&id, &telegramID, &username, &email, &firstName, &lastName, &language, &banned, &created, &paymentsTotal, &paymentsCount, &invitedCount); err != nil {
				continue
			}
			user := userAdminPayload(id, telegramID, username, email, firstName, lastName, language, banned, created)
			user["payments_total_amount"] = paymentsTotal
			user["payments_count"] = paymentsCount
			user["payments_currency"] = settings.DefaultCurrency
			user["invited_users_count"] = invitedCount
			users = append(users, panelAwareAdminUser(r.Context(), pool, panel, user))
		}
		if needsMemoryFilter {
			users = filterAdminUsersInMemory(users, r)
			sortAdminUsersInMemory(users, r.URL.Query().Get("sort"))
		}
		var total int64
		countArgs := args[:len(args)-2]
		countQuery := "SELECT COUNT(*) FROM users u " + where
		_ = pool.QueryRow(r.Context(), countQuery, countArgs...).Scan(&total)
		if needsMemoryFilter {
			total = int64(len(users))
			start := page * pageSize
			end := start + pageSize
			if start > len(users) {
				start = len(users)
			}
			if end > len(users) {
				end = len(users)
			}
			users = users[start:end]
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": users, "total": total})
	}
}

func adminUsersWhereClause(r *http.Request) (string, []any) {
	query := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	filter := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("filter")))
	clauses := []string{}
	args := []any{}
	if query != "" {
		args = append(args, "%"+query+"%")
		idx := len(args)
		clauses = append(clauses, fmt.Sprintf(`(
			LOWER(COALESCE(u.username,'')) LIKE $%[1]d OR
			LOWER(COALESCE(u.email,'')) LIKE $%[1]d OR
			LOWER(COALESCE(u.first_name,'')) LIKE $%[1]d OR
			LOWER(COALESCE(u.last_name,'')) LIKE $%[1]d OR
			u.user_id::text LIKE $%[1]d OR
			COALESCE(u.telegram_id,0)::text LIKE $%[1]d
		)`, idx))
	}
	switch filter {
	case "active":
		clauses = append(clauses, "u.is_banned = FALSE")
	case "banned":
		clauses = append(clauses, "u.is_banned = TRUE")
	case "tg_linked":
		clauses = append(clauses, "COALESCE(u.telegram_id,0) <> 0")
	case "no_tg":
		clauses = append(clauses, "COALESCE(u.telegram_id,0) = 0")
	case "email_linked":
		clauses = append(clauses, "COALESCE(u.email,'') <> ''")
	case "no_email":
		clauses = append(clauses, "COALESCE(u.email,'') = ''")
	case "panel_linked":
		clauses = append(clauses, "COALESCE(u.panel_user_uuid,'') <> ''")
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func adminUsersOrderBy(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "registered_asc":
		return "u.registration_date ASC, u.user_id ASC"
	case "name_asc":
		return "LOWER(COALESCE(NULLIF(u.username,''), NULLIF(u.first_name,''), NULLIF(u.email,''), u.user_id::text)) ASC, u.user_id ASC"
	case "name_desc":
		return "LOWER(COALESCE(NULLIF(u.username,''), NULLIF(u.first_name,''), NULLIF(u.email,''), u.user_id::text)) DESC, u.user_id DESC"
	case "payments_total_asc":
		return "COALESCE((SELECT SUM(p.base_amount)::float8 FROM payment_orders p WHERE p.user_id=u.user_id AND p.status IN ('paid','succeeded')),0) ASC, u.registration_date DESC"
	case "payments_total_desc":
		return "COALESCE((SELECT SUM(p.base_amount)::float8 FROM payment_orders p WHERE p.user_id=u.user_id AND p.status IN ('paid','succeeded')),0) DESC, u.registration_date DESC"
	case "payments_count_asc":
		return "COALESCE((SELECT COUNT(*) FROM payment_orders p WHERE p.user_id=u.user_id),0) ASC, u.registration_date DESC"
	case "payments_count_desc":
		return "COALESCE((SELECT COUNT(*) FROM payment_orders p WHERE p.user_id=u.user_id),0) DESC, u.registration_date DESC"
	case "invited_users_count_asc":
		return "COALESCE((SELECT COUNT(*) FROM users invitee WHERE invitee.referred_by_id=u.user_id),0) ASC, u.registration_date DESC"
	case "invited_users_count_desc":
		return "COALESCE((SELECT COUNT(*) FROM users invitee WHERE invitee.referred_by_id=u.user_id),0) DESC, u.registration_date DESC"
	default:
		return "u.registration_date DESC, u.user_id DESC"
	}
}

func needsAdminUsersMemoryFilter(r *http.Request) bool {
	panelStatus := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("panel_status")))
	premiumTraffic := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("premium_traffic")))
	sortKey := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("sort")))
	if panelStatus != "" && panelStatus != "all" {
		return true
	}
	if premiumTraffic != "" && premiumTraffic != "all" {
		return true
	}
	return strings.HasPrefix(sortKey, "premium_ratio_") || strings.HasPrefix(sortKey, "subscription_expires_at_")
}

func filterAdminUsersInMemory(users []map[string]any, r *http.Request) []map[string]any {
	panelStatus := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("panel_status")))
	premiumTraffic := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("premium_traffic")))
	filtered := make([]map[string]any, 0, len(users))
	for _, user := range users {
		if panelStatus != "" && panelStatus != "all" && strings.ToLower(stringValue(user, "panel_status")) != panelStatus {
			continue
		}
		if premiumTraffic != "" && premiumTraffic != "all" {
			premium := mapValue(user, "premium_traffic")
			if strings.ToLower(stringValue(premium, "state")) != premiumTraffic {
				continue
			}
		}
		filtered = append(filtered, user)
	}
	return filtered
}

func sortAdminUsersInMemory(users []map[string]any, raw string) {
	sortKey := strings.TrimSpace(strings.ToLower(raw))
	switch sortKey {
	case "subscription_expires_at_asc", "subscription_expires_at_desc":
		desc := strings.HasSuffix(sortKey, "_desc")
		sort.SliceStable(users, func(i, j int) bool {
			left := parsePanelTime(firstNonEmpty(stringValue(users[i], "subscription_expires_at"), stringValue(users[i], "panel_status_expired_at")))
			right := parsePanelTime(firstNonEmpty(stringValue(users[j], "subscription_expires_at"), stringValue(users[j], "panel_status_expired_at")))
			if left.Equal(right) {
				return int64Value(users[i], "user_id") < int64Value(users[j], "user_id")
			}
			if left.IsZero() {
				return false
			}
			if right.IsZero() {
				return true
			}
			if desc {
				return left.After(right)
			}
			return left.Before(right)
		})
	case "premium_ratio_asc", "premium_ratio_desc":
		desc := strings.HasSuffix(sortKey, "_desc")
		sort.SliceStable(users, func(i, j int) bool {
			left := premiumTrafficRank(mapValue(users[i], "premium_traffic"))
			right := premiumTrafficRank(mapValue(users[j], "premium_traffic"))
			if left == right {
				return int64Value(users[i], "user_id") < int64Value(users[j], "user_id")
			}
			if desc {
				return left > right
			}
			return left < right
		})
	}
}

func premiumTrafficRank(premium map[string]any) int {
	switch strings.ToLower(stringValue(premium, "state")) {
	case "critical":
		return 4
	case "warn":
		return 3
	case "good":
		return 2
	case "unlimited":
		return 1
	default:
		return 0
	}
}

func adminUserDetailHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		rawUserID := chi.URLParam(r, "user_id")
		user, err := loadAdminUser(r.Context(), pool, rawUserID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		user["payments_currency"] = settings.DefaultCurrency
		user = panelAwareAdminUser(r.Context(), pool, panel, user)
		userID, _ := parsePositiveInt64(rawUserID)
		webUser, _ := loadWebappUser(r.Context(), pool, userID, settings)
		var activeSubscription any
		subscriptionURL := ""
		lastConnectedAt := ""
		vpnStatus := "unknown"
		if panel != nil && panel.Configured(r.Context()) {
			panelUser, found, err := panelUserForWebUser(r.Context(), pool, panel, webUser)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
				return
			}
			if found {
				activeSubscription = subscriptionFromPanelUser(r.Context(), pool, webUser, panelUser)
				subscriptionURL = stringValue(panelUser, "subscriptionUrl")
				traffic := mapValue(panelUser, "userTraffic")
				if onlineAt := parsePanelTime(traffic["onlineAt"]); !onlineAt.IsZero() {
					lastConnectedAt = timeString(onlineAt)
					if time.Since(onlineAt) <= 5*time.Minute {
						vpnStatus = "online"
					} else {
						vpnStatus = "offline"
					}
				}
			}
		}
		recentPayments := loadRecentPaymentsForUser(r.Context(), pool, userID)
		userLogs, logCount, _ := adminMessageLogs(r.Context(), pool, 0, 20, userID)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":                    true,
			"user":                  user,
			"active_subscription":   activeSubscription,
			"subscriptions":         []any{},
			"payments":              recentPayments,
			"recent_payments":       recentPayments,
			"logs":                  userLogs,
			"log_count":             logCount,
			"total_paid":            userTotalPaid(r.Context(), pool, userID),
			"subscription_url":      subscriptionURL,
			"last_vpn_connected_at": lastConnectedAt,
			"vpn_connection_status": vpnStatus,
			"referral":              map[string]any{"code": user["referral_code"], "bot_link": nil, "webapp_link": nil, "inviter": nil, "invitees_total": 0},
		})
	}
}

func adminUserDeleteHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		userID, err := parsePositiveInt64(chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
			return
		}
		if panel != nil && panel.Configured(r.Context()) {
			if panelUUID := loadPanelUUIDForUser(r.Context(), pool, userID); panelUUID != "" {
				if err := panel.DeleteUser(r.Context(), panelUUID); err != nil {
					writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
					return
				}
			}
		}
		_, _ = pool.Exec(r.Context(), "DELETE FROM users WHERE user_id=$1", userID)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func adminUserReferralsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "referrals": []any{}, "total": 0})
	}
}

func telegramChatIDForAdminUser(user map[string]any) int64 {
	if telegramID := int64Value(user, "telegram_id"); telegramID != 0 {
		return telegramID
	}
	return int64Value(user, "user_id")
}

func telegramChatIDForWebUser(user webappUser) int64 {
	if user.TelegramID != 0 {
		return user.TelegramID
	}
	return user.UserID
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func adminUserMessageHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, true)
		if !ok {
			return
		}
		userID, err := parsePositiveInt64(chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
			return
		}
		var payload struct {
			Text string `json:"text"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		user, err := loadAdminUser(r.Context(), pool, strconv.FormatInt(userID, 10))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		chatID := telegramChatIDForAdminUser(user)
		err = sendTelegramText(r.Context(), settings, chatID, payload.Text)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			TargetUserID: userID,
			EventType:    "admin_user_message",
			Content:      payload.Text,
			IsAdminEvent: true,
			Payload:      map[string]any{"error": errorString(err)},
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func adminUserActionHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, true)
		if !ok {
			return
		}
		userID, err := parsePositiveInt64(chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
			return
		}
		path := r.URL.Path
		if strings.HasSuffix(r.URL.Path, "/ban") {
			var payload struct {
				IsBanned bool  `json:"is_banned"`
				Banned   *bool `json:"banned"`
			}
			_ = decodeJSONBody(r, &payload)
			banned := payload.IsBanned
			if payload.Banned != nil {
				banned = *payload.Banned
			}
			_, _ = pool.Exec(r.Context(), "UPDATE users SET is_banned=$2 WHERE user_id=$1", userID, banned)
			if panel != nil && panel.Configured(r.Context()) {
				if panelUUID := loadPanelUUIDForUser(r.Context(), pool, userID); panelUUID != "" {
					if err := panel.SetUserEnabled(r.Context(), panelUUID, !banned); err != nil {
						writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
						return
					}
				}
			}
		} else if strings.HasSuffix(path, "/extend") {
			if err := adminExtendPanelUser(r.Context(), settings, pool, panel, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/tariff") {
			if err := adminChangePanelTariff(r.Context(), settings, pool, panel, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/premium-override") {
			if err := adminSetPremiumTrafficOverride(r.Context(), pool, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/regular-traffic-override") {
			if err := adminSetRegularTrafficOverride(r.Context(), settings, pool, panel, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/hwid-device-limit") {
			if err := adminSetPanelHWIDLimit(r.Context(), settings, pool, panel, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/traffic-grant") {
			if err := adminGrantPanelTraffic(r.Context(), settings, pool, panel, userID, r); err != nil {
				writeAdminActionError(w, err)
				return
			}
		} else if strings.HasSuffix(path, "/reset-trial") {
			_, _ = pool.Exec(r.Context(), "UPDATE users SET trial_eligibility_reset_at=NOW() WHERE user_id=$1", userID)
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			TargetUserID: userID,
			EventType:    "admin_user_action",
			Content:      strings.TrimPrefix(path, "/api/admin/users/"+strconv.FormatInt(userID, 10)+"/"),
			IsAdminEvent: true,
			Payload: map[string]any{
				"path": path,
			},
		})
		user, _ := loadAdminUser(r.Context(), pool, strconv.FormatInt(userID, 10))
		if user == nil {
			user = map[string]any{"user_id": userID}
		}
		user["payments_currency"] = settings.DefaultCurrency
		user = panelAwareAdminUser(r.Context(), pool, panel, user)
		webUser, _ := loadWebappUser(r.Context(), pool, userID, settings)
		var subscription any
		if panel != nil && panel.Configured(r.Context()) {
			if panelUser, found, _ := panelUserForWebUser(r.Context(), pool, panel, webUser); found {
				subscription = subscriptionFromPanelUser(r.Context(), pool, webUser, panelUser)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user, "active_subscription": subscription, "subscription": subscription})
	}
}

func adminMessagePreviewHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload map[string]any
		_ = decodeJSONBody(r, &payload)
		text := strings.TrimSpace(firstNonEmpty(stringFromAny(payload["text"]), stringFromAny(payload["message"])))
		if text == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		err := sendTelegramText(r.Context(), settings, telegramChatIDForWebUser(session.User), text)
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			TargetUserID: session.User.UserID,
			EventType:    "admin_user_message_preview",
			Content:      text,
			IsAdminEvent: true,
			Payload:      map[string]any{"error": errorString(err)},
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "preview": text, "html": text})
	}
}

func adminTelegramProfileLinkHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet)
		if !ok {
			return
		}
		user, err := loadAdminUser(r.Context(), pool, chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		username, _ := user["username"].(string)
		url := ""
		if username != "" {
			url = "https://t.me/" + strings.TrimPrefix(username, "@")
		}
		if url == "" && int64Value(user, "telegram_id") != 0 {
			url = "tg://user?id=" + strconv.FormatInt(int64Value(user, "telegram_id"), 10)
		}
		if r.Method == http.MethodPost {
			if url == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "telegram_profile_not_available"})
				return
			}
			err := sendTelegramText(r.Context(), settings, telegramChatIDForWebUser(session.User), url)
			recordMessageLog(r.Context(), pool, messageLogEntry{
				UserID:       session.User.UserID,
				TargetUserID: int64Value(user, "user_id"),
				EventType:    "admin_telegram_profile_link",
				Content:      url,
				IsAdminEvent: true,
				Payload:      map[string]any{"error": errorString(err)},
			})
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error(), "url": url})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "url": url})
	}
}

func adminAdsListHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		items := readSettingList(r.Context(), pool, "ADMIN_ADS")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "campaigns": items, "totals": map[string]any{"cost": 0, "revenue": 0}})
	}
}

func adminListSettingHandler(settings config.Settings, pool *pgxpool.Pool, key string, responseKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		items := readSettingList(r.Context(), pool, key)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, responseKey: items, "total": len(items)})
	}
}

func adminCreateSettingItemHandler(settings config.Settings, pool *pgxpool.Pool, key string, responseKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		items := readSettingList(r.Context(), pool, key)
		payload["id"] = nextListID(items)
		if _, ok := payload["is_active"]; !ok {
			payload["is_active"] = true
		}
		payload["created_at"] = time.Now().Format(time.RFC3339)
		items = append(items, payload)
		if err := writeSettingList(r.Context(), pool, key, items); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, responseKey: payload})
	}
}

func adminPatchSettingItemHandler(settings config.Settings, pool *pgxpool.Pool, key string, responseKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		id := chi.URLParam(r, "id")
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		items := readSettingList(r.Context(), pool, key)
		for index := range items {
			if fmt.Sprint(items[index]["id"]) == id {
				for k, v := range payload {
					items[index][k] = v
				}
				_ = writeSettingList(r.Context(), pool, key, items)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, responseKey: items[index]})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not_found"})
	}
}

func adminDeleteSettingItemHandler(settings config.Settings, pool *pgxpool.Pool, key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		id := chi.URLParam(r, "id")
		items := readSettingList(r.Context(), pool, key)
		next := items[:0]
		for _, item := range items {
			if fmt.Sprint(item["id"]) != id {
				next = append(next, item)
			}
		}
		_ = writeSettingList(r.Context(), pool, key, next)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func adminBroadcastAudienceHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		counts := map[string]int64{"all": 0, "active": 0, "inactive": 0, "expired": 0, "never": 0}
		var allCount, activeCount, inactiveCount int64
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users WHERE is_banned=FALSE").Scan(&allCount)
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users WHERE is_banned=FALSE AND COALESCE(panel_user_uuid,'')<>''").Scan(&activeCount)
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users WHERE is_banned=FALSE AND COALESCE(panel_user_uuid,'')=''").Scan(&inactiveCount)
		counts["all"] = allCount
		counts["active"] = activeCount
		counts["inactive"] = inactiveCount
		counts["never"] = counts["inactive"]
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "total": counts["all"], "counts": counts, "audiences": counts})
	}
}

func adminBroadcastHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload struct {
			Target string `json:"target"`
			Text   string `json:"text"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		recipients, err := broadcastRecipients(r.Context(), pool, payload.Target, 5000)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "broadcast_recipients_failed"})
			return
		}
		queued := 0
		failed := 0
		for _, recipient := range recipients {
			if err := sendTelegramText(r.Context(), settings, recipient.ChatID, payload.Text); err != nil {
				failed++
				recordMessageLog(r.Context(), pool, messageLogEntry{
					UserID:       session.User.UserID,
					TargetUserID: recipient.UserID,
					EventType:    "admin_broadcast_failed",
					Content:      payload.Text,
					IsAdminEvent: true,
					Payload:      map[string]any{"target": payload.Target, "error": err.Error()},
				})
				continue
			}
			queued++
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			EventType:    "admin_broadcast",
			Content:      payload.Text,
			IsAdminEvent: true,
			Payload:      map[string]any{"target": payload.Target, "queued": queued, "failed": failed},
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "queued": queued, "failed": failed})
	}
}

type broadcastRecipient struct {
	UserID int64
	ChatID int64
}

func broadcastRecipients(ctx context.Context, pool *pgxpool.Pool, target string, limit int) ([]broadcastRecipient, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		target = "all"
	}
	where := "is_banned=FALSE"
	switch target {
	case "active":
		where += " AND COALESCE(panel_user_uuid,'')<>''"
	case "inactive", "never":
		where += " AND COALESCE(panel_user_uuid,'')=''"
	case "expired":
		return []broadcastRecipient{}, nil
	}
	if limit <= 0 || limit > 10000 {
		limit = 5000
	}
	rows, err := pool.Query(ctx, `
SELECT user_id, COALESCE(telegram_id,0)
FROM users
WHERE `+where+`
ORDER BY registration_date DESC
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	recipients := []broadcastRecipient{}
	for rows.Next() {
		var userID, telegramID int64
		if err := rows.Scan(&userID, &telegramID); err != nil {
			return nil, err
		}
		chatID := telegramID
		if chatID == 0 {
			chatID = userID
		}
		if chatID != 0 {
			recipients = append(recipients, broadcastRecipient{UserID: userID, ChatID: chatID})
		}
	}
	return recipients, rows.Err()
}

func adminBackupsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "archives": []any{}, "backup_dir": "data/backups"})
	}
}

func adminBackupCreateHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		result := map[string]any{"created_at": time.Now().Format(time.RFC3339), "note": "database backup requires deployment storage configuration"}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result, "archive": nil})
	}
}

func adminThemesHandler(settings config.Settings, pool *pgxpool.Pool, assets webassets.Paths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet); !ok {
			return
		}
		store := appsettings.NewStore(pool)
		if r.Method == http.MethodGet {
			catalog := readThemeCatalog(r.Context(), store, assets.ThemesDir)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "catalog": catalog, "themes_dir": assets.ThemesDir})
			return
		}
		var payload struct {
			Catalog map[string]any `json:"catalog"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if err := store.Upsert(r.Context(), "THEME_CATALOG", payload.Catalog); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "themes_save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "catalog": payload.Catalog, "themes_dir": assets.ThemesDir})
	}
}

func adminTranslationsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, r.Method != http.MethodGet); !ok {
			return
		}
		store := appsettings.NewStore(pool)
		if r.Method == http.MethodGet {
			raw, ok, _ := store.Get(r.Context(), "TRANSLATION_OVERRIDES")
			overrides := map[string]any{}
			if ok {
				_ = json.Unmarshal(raw, &overrides)
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "translations": overrides, "overrides": overrides})
			return
		}
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		value := payload["translations"]
		if value == nil {
			value = payload["overrides"]
		}
		if value == nil {
			value = map[string]any{}
		}
		if err := store.Upsert(r.Context(), "TRANSLATION_OVERRIDES", value); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "translations_save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "translations": value, "overrides": value})
	}
}

func supportListHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, false)
			if !ok {
				return
			}
		} else {
			session, ok = requireSession(w, r, settings, pool, false)
			if !ok {
				return
			}
		}
		allTickets := supportVisibleTickets(r.Context(), pool, readSettingList(r.Context(), pool, "SUPPORT_TICKETS"), session.User.UserID, admin)
		counts := supportCounts(allTickets)
		filtered := filterSupportTickets(allTickets, r.URL.Query())
		total := len(filtered)
		filtered = paginateSupportTickets(filtered, r.URL.Query())
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "tickets": filtered, "counts": counts, "total": total})
	}
}

func supportCreateHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		subject := strings.TrimSpace(fmt.Sprint(payload["subject"]))
		body := strings.TrimSpace(fmt.Sprint(payload["body"]))
		if subject == "" || body == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_ticket"})
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		id := nextListID(tickets)
		now := time.Now().Format(time.RFC3339)
		message := map[string]any{
			"message_id":       1,
			"ticket_id":        id,
			"body":             body,
			"is_admin":         false,
			"is_internal_note": false,
			"author_role":      "user",
			"user_id":          session.User.UserID,
			"created_at":       now,
		}
		ticket := map[string]any{
			"ticket_id":            id,
			"id":                   id,
			"user_id":              session.User.UserID,
			"status":               "awaiting_admin",
			"priority":             defaultString(payload["priority"], "normal"),
			"category":             defaultString(payload["category"], "general"),
			"subject":              subject,
			"body":                 body,
			"created_at":           now,
			"updated_at":           now,
			"last_message_at":      now,
			"last_message_preview": supportPreview(body),
			"message_count":        1,
			"unread_admin_count":   1,
			"unread_user_count":    0,
			"messages":             []any{message},
		}
		tickets = append(tickets, ticket)
		_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, ticket, false)})
	}
}

func supportDetailHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, false)
			if !ok {
				return
			}
		} else {
			session, ok = requireSession(w, r, settings, pool, false)
			if !ok {
				return
			}
		}
		ticket, ok := findSettingItem(r.Context(), pool, "SUPPORT_TICKETS", chi.URLParam(r, "ticket_id"), "ticket_id")
		if !ok || (!admin && !supportTicketBelongsToUser(ticket, session.User.UserID)) {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
			return
		}
		response := map[string]any{
			"ok":       true,
			"ticket":   supportTicketResponse(r.Context(), pool, ticket, admin),
			"messages": supportVisibleMessages(ticket, admin),
		}
		if admin {
			response["user_snapshot"] = supportUserSnapshot(r.Context(), pool, int64Value(ticket, "user_id"))
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func supportMessageHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, true)
		} else {
			session, ok = requireSession(w, r, settings, pool, true)
		}
		if !ok {
			return
		}
		var payload struct {
			Body           string `json:"body"`
			IsInternalNote bool   `json:"is_internal_note"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Body) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		id := chi.URLParam(r, "ticket_id")
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == id {
				if !admin && !supportTicketBelongsToUser(tickets[index], session.User.UserID) {
					writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
					return
				}
				messages := supportAllMessages(tickets[index])
				now := time.Now().Format(time.RFC3339)
				internalNote := admin && payload.IsInternalNote
				role := "user"
				if admin {
					role = "admin"
				}
				if internalNote {
					role = "internal"
				}
				message := map[string]any{
					"message_id":       len(messages) + 1,
					"ticket_id":        id,
					"body":             strings.TrimSpace(payload.Body),
					"is_admin":         admin,
					"is_internal_note": internalNote,
					"author_role":      role,
					"user_id":          session.User.UserID,
					"created_at":       now,
				}
				tickets[index]["messages"] = append(messages, message)
				tickets[index]["updated_at"] = now
				tickets[index]["message_count"] = len(messages) + 1
				if !internalNote {
					tickets[index]["last_message_at"] = now
					tickets[index]["last_message_preview"] = supportPreview(payload.Body)
				}
				switch {
				case internalNote:
				case admin:
					tickets[index]["status"] = "awaiting_user"
					tickets[index]["unread_user_count"] = int64Value(tickets[index], "unread_user_count") + 1
				default:
					tickets[index]["status"] = "awaiting_admin"
					tickets[index]["unread_admin_count"] = int64Value(tickets[index], "unread_admin_count") + 1
				}
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], admin), "message": message})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportPatchHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		id := chi.URLParam(r, "ticket_id")
		var payload map[string]any
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == id || fmt.Sprint(tickets[index]["id"]) == id {
				for k, v := range payload {
					switch k {
					case "status":
						tickets[index][k] = normalizeSupportStatus(v)
					case "priority":
						tickets[index][k] = normalizeSupportPriority(v)
					case "category":
						tickets[index][k] = defaultString(v, "general")
					case "subject":
						if subject := strings.TrimSpace(fmt.Sprint(v)); subject != "" {
							tickets[index][k] = subject
						}
					default:
						tickets[index][k] = v
					}
				}
				tickets[index]["updated_at"] = time.Now().Format(time.RFC3339)
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], true)})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportReadHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessionContext
		var ok bool
		if admin {
			session, ok = requireAdmin(w, r, settings, pool, true)
		} else {
			session, ok = requireSession(w, r, settings, pool, true)
		}
		if !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == chi.URLParam(r, "ticket_id") || fmt.Sprint(tickets[index]["id"]) == chi.URLParam(r, "ticket_id") {
				if !admin && !supportTicketBelongsToUser(tickets[index], session.User.UserID) {
					writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
					return
				}
				if admin {
					tickets[index]["unread_admin_count"] = 0
				} else {
					tickets[index]["unread_user_count"] = 0
				}
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": supportTicketResponse(r.Context(), pool, tickets[index], admin)})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportUnreadHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, false)
		if !ok {
			return
		}
		unread := 0
		for _, ticket := range readSettingList(r.Context(), pool, "SUPPORT_TICKETS") {
			if supportTicketBelongsToUser(ticket, session.User.UserID) {
				unread += int(int64Value(ticket, "unread_user_count"))
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "unread": unread})
	}
}

func adminSupportStatsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "stats": supportCounts(tickets), "counts": supportCounts(tickets)})
	}
}

func adminUploadPlaceholderHandler(settings config.Settings, pool *pgxpool.Pool, kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "multipart/form-data") {
			if err := r.ParseMultipartForm(8 << 20); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_upload"})
				return
			}
			file, header, err := r.FormFile("file")
			if err == nil {
				defer func() { _ = file.Close() }()
				if header.Size > 8<<20 {
					writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_too_large"})
					return
				}
				buf := make([]byte, 512)
				n, _ := file.Read(buf)
				mime := http.DetectContentType(buf[:n])
				if kind != "archive" && !strings.HasPrefix(mime, "image/") {
					writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "unsupported_mime"})
					return
				}
			}
		} else if r.Body != nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, 1<<20))
		}
		response := map[string]any{"ok": true}
		switch kind {
		case "logo":
			response["logo_url"] = "/default-brand/default-logo.webp"
			response["favicon_url"] = "/favicon.ico"
		case "favicon":
			response["favicon_url"] = "/favicon.ico"
			response["variants"] = map[string]string{}
		case "archive":
			response["archive"] = nil
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func okSessionMutation(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSession(w, r, settings, pool, true); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func okAdminMutation(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func unavailableSessionMutation(settings config.Settings, pool *pgxpool.Pool, code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSession(w, r, settings, pool, true); !ok {
			return
		}
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": code})
	}
}

func unavailablePublicMutation(code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, maxJSONBodyBytes))
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": code})
	}
}

func subscriptionGuidesHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/public/") {
			if _, ok := requireSession(w, r, settings, pool, false); !ok {
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": false, "config": nil, "source": nil, "subscription": nil})
	}
}

func loadTariffCatalogForAdmin(path string) (any, error) {
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		body, err = os.ReadFile("data/tariffs.example.json")
	}
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{"default_tariff": "", "default_currency": "usd", "tariffs": []any{}}, nil
		}
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func providerCurrencySupport() []map[string]any {
	return []map[string]any{
		{"provider": payments.ProviderEZPay, "currencies": []string{"CNY"}, "note": "EZPay checkout amount is converted from USD to CNY."},
		{"provider": payments.ProviderBEPUSDT, "currencies": []string{"USD"}, "note": "BEPUSDT creates USDT invoices from USD fiat amount."},
	}
}

func devicePlansFrom(plans []tariffs.Plan) []map[string]any {
	result := []map[string]any{}
	for _, plan := range plans {
		if plan.SaleMode != "subscription" {
			continue
		}
		result = append(result, map[string]any{
			"plan_hash":          plan.PlanHash,
			"title":              plan.Title,
			"months":             plan.Months,
			"device_count":       plan.Months,
			"price":              plan.Price,
			"currency":           plan.Currency,
			"base_amount":        plan.BaseAmount,
			"base_currency":      plan.BaseCurrency,
			"display_cny_amount": plan.DisplayCNYAmount,
			"fx_rate":            plan.FXRate,
			"fx_source":          plan.FXSource,
			"sale_mode":          "hwid_devices",
			"tariff_key":         plan.TariffKey,
		})
	}
	return result
}

func changeTargetsFrom(plans []tariffs.Plan) []map[string]any {
	grouped := map[string][]tariffs.Plan{}
	for _, plan := range plans {
		grouped[plan.TariffKey] = append(grouped[plan.TariffKey], plan)
	}
	keys := make([]string, 0, len(grouped))
	for key := range grouped {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := []map[string]any{}
	for _, key := range keys {
		group := grouped[key]
		if len(group) == 0 {
			continue
		}
		actions := []map[string]any{}
		for _, plan := range group {
			action := map[string]any{
				"kind":               "payment",
				"mode":               "buy_period",
				"plan_hash":          plan.PlanHash,
				"tariff_key":         plan.TariffKey,
				"months":             plan.Months,
				"title":              plan.Title,
				"price":              plan.Price,
				"currency":           plan.Currency,
				"base_amount":        plan.BaseAmount,
				"base_currency":      plan.BaseCurrency,
				"display_cny_amount": plan.DisplayCNYAmount,
				"fx_rate":            plan.FXRate,
				"fx_source":          plan.FXSource,
			}
			if plan.SaleMode == "traffic_package" {
				action["mode"] = "buy_package"
				action["traffic_gb"] = plan.TrafficGB
			}
			actions = append(actions, action)
		}
		result = append(result, map[string]any{"tariff_key": key, "title": group[0].Title, "description": group[0].Description, "billing_model": group[0].BillingModel, "actions": actions})
	}
	return result
}

func readSettingList(ctx context.Context, pool *pgxpool.Pool, key string) []map[string]any {
	store := appsettings.NewStore(pool)
	raw, ok, err := store.Get(ctx, key)
	if err != nil || !ok {
		return []map[string]any{}
	}
	var items []map[string]any
	if json.Unmarshal(raw, &items) != nil {
		return []map[string]any{}
	}
	return items
}

func writeSettingList(ctx context.Context, pool *pgxpool.Pool, key string, items []map[string]any) error {
	return appsettings.NewStore(pool).Upsert(ctx, key, items)
}

func nextListID(items []map[string]any) int {
	next := 1
	for _, item := range items {
		id, _ := strconv.Atoi(fmt.Sprint(item["id"]))
		if id >= next {
			next = id + 1
		}
	}
	return next
}

func findSettingItem(ctx context.Context, pool *pgxpool.Pool, key string, id string, idField string) (map[string]any, bool) {
	for _, item := range readSettingList(ctx, pool, key) {
		if fmt.Sprint(item[idField]) == id || fmt.Sprint(item["id"]) == id {
			return item, true
		}
	}
	return nil, false
}

func settingItemActive(item map[string]any) bool {
	if _, ok := item["is_active"]; !ok {
		if _, enabled := item["enabled"]; !enabled {
			return true
		}
	}
	if value, ok := item["is_active"]; ok {
		return boolish(value, true)
	}
	return boolish(item["enabled"], true)
}

func promoExpired(item map[string]any) bool {
	now := time.Now().UTC()
	for _, key := range []string{"valid_until", "expires_at", "expired_at"} {
		if value := parsePanelTime(item[key]); !value.IsZero() && now.After(value) {
			return true
		}
	}
	validDays := int(int64Value(item, "valid_days"))
	if validDays <= 0 {
		return false
	}
	createdAt := parsePanelTime(item["created_at"])
	return !createdAt.IsZero() && now.After(createdAt.AddDate(0, 0, validDays))
}

func promoActivatedByUser(item map[string]any, userID int64) bool {
	for _, raw := range anyList(item["activations"]) {
		activation, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if int64Value(activation, "user_id") == userID {
			return true
		}
	}
	return false
}

func promoActivationCount(item map[string]any) int {
	current := int(int64Value(item, "current_activations"))
	if activations := len(anyList(item["activations"])); activations > current {
		return activations
	}
	return current
}

func trialAvailableForUser(ctx context.Context, pool *pgxpool.Pool, userID int64) bool {
	if pool == nil || userID == 0 {
		return false
	}
	var resetAt sql.NullTime
	_ = pool.QueryRow(ctx, "SELECT trial_eligibility_reset_at FROM users WHERE user_id=$1", userID).Scan(&resetAt)
	for _, raw := range readSettingList(ctx, pool, "TRIAL_ACTIVATIONS") {
		if int64Value(raw, "user_id") != userID {
			continue
		}
		activatedAt := parsePanelTime(raw["activated_at"])
		if !resetAt.Valid || activatedAt.IsZero() || activatedAt.After(resetAt.Time) {
			return false
		}
	}
	return true
}

func recordTrialActivation(ctx context.Context, pool *pgxpool.Pool, userID int64) {
	activations := readSettingList(ctx, pool, "TRIAL_ACTIVATIONS")
	activations = append(activations, map[string]any{"user_id": userID, "activated_at": time.Now().UTC().Format(time.RFC3339)})
	_ = writeSettingList(ctx, pool, "TRIAL_ACTIVATIONS", activations)
}

func referralPayload(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, user webappUser) map[string]any {
	code := ensureReferralCode(ctx, pool, user.UserID)
	store := appsettings.NewStore(pool)
	welcomeDays := store.Int(ctx, "REFERRAL_WELCOME_BONUS_DAYS", 0)
	var invitees int64
	var referredBy int64
	var claimedAt sql.NullTime
	if pool != nil {
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE referred_by_id=$1", user.UserID).Scan(&invitees)
		_ = pool.QueryRow(ctx, "SELECT COALESCE(referred_by_id,0), referral_welcome_bonus_claimed_at FROM users WHERE user_id=$1", user.UserID).Scan(&referredBy, &claimedAt)
	}
	return map[string]any{
		"code":                         code,
		"bot_link":                     nil,
		"webapp_link":                  referralWebAppLink(settings, code),
		"invited_count":                invitees,
		"invitees_total":               invitees,
		"purchased_count":              0,
		"bonus_details":                []any{},
		"one_bonus_per_referee":        true,
		"welcome_bonus_days":           welcomeDays,
		"welcome_bonus_available":      welcomeDays > 0 && referredBy != 0 && !claimedAt.Valid,
		"welcome_bonus_claimed":        claimedAt.Valid,
		"welcome_bonus_claimed_at":     nullableTimeString(claimedAt),
		"referral_welcome_bonus_days":  welcomeDays,
		"referral_welcome_available":   welcomeDays > 0 && referredBy != 0 && !claimedAt.Valid,
		"referral_welcome_claimed":     claimedAt.Valid,
		"referral_welcome_claimed_at":  nullableTimeString(claimedAt),
		"referral_welcome_requires_tg": false,
	}
}

func ensureReferralCode(ctx context.Context, pool *pgxpool.Pool, userID int64) string {
	if pool == nil || userID == 0 {
		return ""
	}
	var code string
	err := pool.QueryRow(ctx, "SELECT COALESCE(referral_code,'') FROM users WHERE user_id=$1", userID).Scan(&code)
	if err != nil {
		return ""
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code != "" {
		return code
	}
	code = "R" + strings.ToUpper(strconv.FormatInt(userID, 36))
	_, _ = pool.Exec(ctx, "UPDATE users SET referral_code=$2 WHERE user_id=$1 AND (referral_code IS NULL OR referral_code='')", userID, code)
	return code
}

func bindReferralCode(ctx context.Context, pool *pgxpool.Pool, userID int64, rawCode string) {
	if pool == nil || userID == 0 {
		return
	}
	code := strings.ToUpper(strings.TrimSpace(rawCode))
	code = strings.TrimPrefix(code, "REF_")
	code = strings.TrimPrefix(code, "REF-")
	if code == "" {
		return
	}
	var referrerID int64
	err := pool.QueryRow(ctx, "SELECT user_id FROM users WHERE UPPER(referral_code)=UPPER($1) LIMIT 1", code).Scan(&referrerID)
	if err != nil || referrerID == 0 || referrerID == userID {
		return
	}
	_, _ = pool.Exec(ctx, "UPDATE users SET referred_by_id=$2 WHERE user_id=$1 AND referred_by_id IS NULL", userID, referrerID)
}

func referralWebAppLink(settings config.Settings, code string) string {
	base := strings.TrimSpace(settings.SubscriptionMiniApp)
	if base == "" || code == "" {
		return ""
	}
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	return base + separator + "startapp=ref_" + code
}

func grantResponse(ctx context.Context, pool *pgxpool.Pool, user webappUser, panelUser map[string]any, days int) map[string]any {
	subscription := subscriptionFromPanelUser(ctx, pool, user, panelUser)
	response := map[string]any{
		"ok":           true,
		"bonus_days":   days,
		"subscription": subscription,
	}
	for _, key := range []string{"end_date", "end_date_text", "config_link", "connect_url", "days_left"} {
		if value, ok := subscription[key]; ok {
			response[key] = value
		}
	}
	return response
}

func writePanelActionError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	if errors.Is(err, remnawave.ErrNotConfigured) {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{"ok": false, "error": panelErrorCode(err), "message": err.Error()})
}

func writeAdminActionError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	code := strings.TrimSpace(err.Error())
	if errors.Is(err, remnawave.ErrNotConfigured) {
		status = http.StatusServiceUnavailable
		code = panelErrorCode(err)
	} else {
		var apiErr remnawave.APIError
		if errors.As(err, &apiErr) {
			status = http.StatusBadGateway
			code = panelErrorCode(err)
		}
	}
	if code == "" {
		code = "action_failed"
	}
	writeJSON(w, status, map[string]any{"ok": false, "error": code, "message": err.Error()})
}

func anyList(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []map[string]any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result
	default:
		return []any{}
	}
}

func splitRuntimeList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(field); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func boolish(value any, fallback bool) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	case float64:
		return typed != 0
	case int:
		return typed != 0
	case int64:
		return typed != 0
	}
	return fallback
}

func nullableTimeString(value sql.NullTime) any {
	if !value.Valid {
		return nil
	}
	return value.Time.UTC().Format(time.RFC3339)
}

func supportVisibleTickets(ctx context.Context, pool *pgxpool.Pool, tickets []map[string]any, userID int64, admin bool) []map[string]any {
	result := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		if !admin && !supportTicketBelongsToUser(ticket, userID) {
			continue
		}
		result = append(result, supportTicketResponse(ctx, pool, ticket, admin))
	}
	return result
}

func supportTicketResponse(ctx context.Context, pool *pgxpool.Pool, ticket map[string]any, admin bool) map[string]any {
	item := make(map[string]any, len(ticket)+2)
	for key, value := range ticket {
		if key == "messages" {
			continue
		}
		item[key] = value
	}
	item["status"] = normalizeSupportStatus(item["status"])
	item["priority"] = normalizeSupportPriority(item["priority"])
	item["category"] = defaultString(item["category"], "general")
	item["message_count"] = len(supportVisibleMessages(ticket, admin))
	if item["last_message_at"] == nil || fmt.Sprint(item["last_message_at"]) == "" {
		item["last_message_at"] = item["updated_at"]
	}
	if admin {
		item["user"] = supportTicketUser(ctx, pool, int64Value(ticket, "user_id"))
	} else {
		delete(item, "unread_admin_count")
	}
	return item
}

func supportTicketUser(ctx context.Context, pool *pgxpool.Pool, userID int64) map[string]any {
	user := map[string]any{"user_id": userID}
	if pool == nil || userID == 0 {
		return user
	}
	loaded, err := loadAdminUser(ctx, pool, strconv.FormatInt(userID, 10))
	if err != nil {
		return user
	}
	loaded["name"] = strings.TrimSpace(strings.Join([]string{stringValue(loaded, "first_name"), stringValue(loaded, "last_name")}, " "))
	if loaded["name"] == "" {
		loaded["name"] = firstNonEmpty(stringValue(loaded, "username"), stringValue(loaded, "email"), strconv.FormatInt(userID, 10))
	}
	return loaded
}

func supportUserSnapshot(ctx context.Context, pool *pgxpool.Pool, userID int64) map[string]any {
	user := supportTicketUser(ctx, pool, userID)
	name := firstNonEmpty(stringValue(user, "name"), stringValue(user, "username"), stringValue(user, "email"), strconv.FormatInt(userID, 10))
	return map[string]any{
		"name":         name,
		"tariff":       "-",
		"panel_status": "-",
		"remaining":    "-",
	}
}

func supportTicketBelongsToUser(ticket map[string]any, userID int64) bool {
	return userID != 0 && int64Value(ticket, "user_id") == userID
}

func supportAllMessages(ticket map[string]any) []any {
	switch messages := ticket["messages"].(type) {
	case []any:
		return messages
	case []map[string]any:
		result := make([]any, 0, len(messages))
		for _, message := range messages {
			result = append(result, message)
		}
		return result
	default:
		return []any{}
	}
}

func supportVisibleMessages(ticket map[string]any, admin bool) []any {
	messages := supportAllMessages(ticket)
	if admin {
		return messages
	}
	result := make([]any, 0, len(messages))
	for _, message := range messages {
		if mapped, ok := message.(map[string]any); ok && supportBoolValue(mapped, "is_internal_note") {
			continue
		}
		result = append(result, message)
	}
	return result
}

func filterSupportTickets(tickets []map[string]any, query map[string][]string) []map[string]any {
	status := strings.ToLower(strings.TrimSpace(firstQuery(query, "status")))
	priority := strings.ToLower(strings.TrimSpace(firstQuery(query, "priority")))
	category := strings.ToLower(strings.TrimSpace(firstQuery(query, "category")))
	search := strings.ToLower(strings.TrimSpace(firstQuery(query, "search")))
	result := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		ticketStatus := normalizeSupportStatus(ticket["status"])
		if status != "" && status != "all" {
			active := ticketStatus != "closed" && ticketStatus != "resolved"
			if status == "active" && !active {
				continue
			}
			if status != "active" && ticketStatus != status {
				continue
			}
		}
		if priority != "" && strings.ToLower(fmt.Sprint(ticket["priority"])) != priority {
			continue
		}
		if category != "" && strings.ToLower(fmt.Sprint(ticket["category"])) != category {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(fmt.Sprint(ticket)), search) {
			continue
		}
		result = append(result, ticket)
	}
	sortSupportTickets(result, firstQuery(query, "sort"))
	return result
}

func paginateSupportTickets(tickets []map[string]any, query map[string][]string) []map[string]any {
	offset, _ := strconv.Atoi(firstQuery(query, "offset"))
	limit, _ := strconv.Atoi(firstQuery(query, "limit"))
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset >= len(tickets) {
		return []map[string]any{}
	}
	end := offset + limit
	if end > len(tickets) {
		end = len(tickets)
	}
	return tickets[offset:end]
}

func sortSupportTickets(tickets []map[string]any, rawSort string) {
	sortKey := strings.ToLower(strings.TrimSpace(rawSort))
	sort.SliceStable(tickets, func(i, j int) bool {
		left := tickets[i]
		right := tickets[j]
		switch sortKey {
		case "created_asc":
			return supportTimeValue(left, "created_at").Before(supportTimeValue(right, "created_at"))
		case "created_desc":
			return supportTimeValue(left, "created_at").After(supportTimeValue(right, "created_at"))
		case "updated_asc":
			return supportTimeValue(left, "updated_at").Before(supportTimeValue(right, "updated_at"))
		default:
			leftUnread := int64Value(left, "unread_admin_count")
			rightUnread := int64Value(right, "unread_admin_count")
			if leftUnread != rightUnread {
				return leftUnread > rightUnread
			}
			leftPriority := supportPriorityRank(left["priority"])
			rightPriority := supportPriorityRank(right["priority"])
			if leftPriority != rightPriority {
				return leftPriority > rightPriority
			}
			return supportTimeValue(left, "updated_at").After(supportTimeValue(right, "updated_at"))
		}
	})
}

func supportTimeValue(ticket map[string]any, key string) time.Time {
	value := stringValue(ticket, key)
	if value == "" && key == "updated_at" {
		value = firstNonEmpty(stringValue(ticket, "last_message_at"), stringValue(ticket, "created_at"))
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	return time.Time{}
}

func supportPriorityRank(value any) int {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "urgent":
		return 4
	case "high":
		return 3
	case "normal":
		return 2
	case "low":
		return 1
	default:
		return 2
	}
}

func normalizeSupportStatus(value any) string {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "closed", "resolved", "awaiting_admin", "awaiting_user", "open":
		return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	default:
		return "open"
	}
}

func normalizeSupportPriority(value any) string {
	switch strings.ToLower(strings.TrimSpace(fmt.Sprint(value))) {
	case "low", "normal", "high", "urgent":
		return strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	default:
		return "normal"
	}
}

func supportBoolValue(m map[string]any, key string) bool {
	switch value := m[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
	default:
		return false
	}
}

func supportPreview(value string) string {
	clean := strings.Join(strings.Fields(value), " ")
	runes := []rune(clean)
	if len(runes) <= 160 {
		return clean
	}
	return string(runes[:160])
}

func defaultString(value any, fallback string) string {
	clean := strings.TrimSpace(fmt.Sprint(value))
	if clean == "" || clean == "<nil>" {
		return fallback
	}
	return clean
}

func firstQuery(query map[string][]string, key string) string {
	values := query[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func readThemeCatalog(ctx context.Context, store appsettings.Store, themesDir string) map[string]any {
	raw, ok, _ := store.Get(ctx, "THEME_CATALOG")
	if ok {
		var saved map[string]any
		if json.Unmarshal(raw, &saved) == nil {
			return saved
		}
	}
	themes := []any{}
	entries, _ := os.ReadDir(themesDir)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		body, err := os.ReadFile(filepath.Join(themesDir, entry.Name(), "theme.json"))
		if err != nil {
			continue
		}
		var theme map[string]any
		if json.Unmarshal(body, &theme) == nil {
			themes = append(themes, theme)
		}
	}
	return map[string]any{"default_theme": "dark", "themes": themes}
}

func supportCounts(tickets []map[string]any) map[string]int {
	counts := map[string]int{"active": 0, "closed": 0, "resolved": 0, "awaiting_admin": 0, "awaiting_user": 0, "open": 0, "total": len(tickets), "total_unread_admin": 0, "total_unread_user": 0}
	for _, ticket := range tickets {
		status := normalizeSupportStatus(ticket["status"])
		counts[status]++
		if status != "closed" && status != "resolved" {
			counts["active"]++
		}
		counts["total_unread_admin"] += int(int64Value(ticket, "unread_admin_count"))
		counts["total_unread_user"] += int(int64Value(ticket, "unread_user_count"))
	}
	return counts
}

func loadAdminUser(ctx context.Context, pool *pgxpool.Pool, rawID string) (map[string]any, error) {
	userID, err := parsePositiveInt64(rawID)
	if err != nil {
		return nil, err
	}
	var id, telegramID int64
	var username, email, firstName, lastName, language string
	var banned bool
	var created time.Time
	var paymentsTotal float64
	var paymentsCount, invitedCount int64
	err = pool.QueryRow(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), is_banned, registration_date,
	COALESCE((SELECT SUM(p.base_amount)::float8 FROM payment_orders p WHERE p.user_id=users.user_id AND p.status IN ('paid','succeeded')),0),
	COALESCE((SELECT COUNT(*) FROM payment_orders p WHERE p.user_id=users.user_id),0),
	COALESCE((SELECT COUNT(*) FROM users invitee WHERE invitee.referred_by_id=users.user_id),0)
FROM users WHERE user_id=$1`, userID).Scan(&id, &telegramID, &username, &email, &firstName, &lastName, &language, &banned, &created, &paymentsTotal, &paymentsCount, &invitedCount)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	user := userAdminPayload(id, telegramID, username, email, firstName, lastName, language, banned, created)
	user["payments_total_amount"] = paymentsTotal
	user["payments_count"] = paymentsCount
	user["invited_users_count"] = invitedCount
	return user, nil
}

func userAdminPayload(id int64, telegramID int64, username string, email string, firstName string, lastName string, language string, banned bool, created time.Time) map[string]any {
	return map[string]any{
		"user_id":           id,
		"telegram_id":       telegramID,
		"username":          username,
		"email":             email,
		"first_name":        firstName,
		"last_name":         lastName,
		"language_code":     language,
		"is_banned":         banned,
		"registration_date": created,
		"created_at":        created,
	}
}
