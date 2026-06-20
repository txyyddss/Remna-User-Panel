package httpapi

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
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

func registerExtraAPIRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, _ *i18n.Catalog, assets webassets.Paths, registry *payments.Registry, panel *remnawave.Client) {
	router.Get("/api/tariffs/topup-options", webappPlansOptionsHandler(settings, pool, "topup"))
	router.Get("/api/tariffs/change-options", webappPlansOptionsHandler(settings, pool, "change"))
	router.Post("/api/tariffs/change", userTariffChangeHandler(settings, pool, panel))
	router.Post("/api/tariffs/change-payment", createPaymentHandler(settings, pool, registry))
	router.Post("/api/subscription/auto-renew", autoRenewHandler(settings, pool))
	router.Post("/api/promo/apply", promoApplyHandler(settings, pool, panel))
	router.Post("/api/referral/welcome-bonus/claim", referralWelcomeBonusHandler(settings, pool, panel))
	router.Post("/api/telemetry/heartbeat", telemetryHeartbeatHandler(settings, pool))
	router.Post("/api/trial/activate", trialActivateHandler(settings, pool, panel))
	router.Post("/api/devices/ips", devicesIPsHandler(settings, pool, panel))
	router.Post("/api/devices/ips/disconnect", devicesIPsDisconnectHandler(settings, pool, panel))
	router.Post("/api/account/email/request", emailRequestHandler(settings, pool))
	router.Post("/api/account/email/verify", emailVerifyHandler(settings, pool))
	router.Post("/api/account/password/request", passwordRequestHandler(settings, pool))
	router.Post("/api/account/password/confirm", passwordConfirmHandler(settings, pool))
	router.Post("/api/account/telegram/link", accountTelegramLinkHandler(settings, pool))
	router.Post("/api/account/language", accountLanguageHandler(settings, pool))
	router.Get("/api/account/notifications", userNotificationPrefsHandler(settings, pool))
	router.Post("/api/account/notifications", userNotificationPrefsHandler(settings, pool))
	router.Post("/api/auth/email/request", emailLoginRequestHandler(settings, pool))
	router.Post("/api/auth/email/verify", emailLoginVerifyHandler(settings, pool))
	router.Post("/api/auth/email/password", passwordLoginHandler(settings, pool))
	router.Post("/api/auth/email/magic", emailMagicLinkHandler(settings, pool))
	router.Get("/api/subscription-guides", subscriptionGuidesHandler(settings, pool, panel))
	router.Get("/api/subscription-guides/public/{share_token}", subscriptionGuidesHandler(settings, pool, panel))

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
	router.Post("/api/admin/backups/upload", adminBackupUploadHandler(settings, pool))
	router.Post("/api/admin/backups/restore", adminBackupRestoreHandler(settings, pool))
	router.Get("/api/admin/backups/{name}/download", adminBackupDownloadHandler(settings, pool))
	router.Get("/api/admin/themes", adminThemesHandler(settings, pool, assets))
	router.Put("/api/admin/themes", adminThemesHandler(settings, pool, assets))
	router.Post("/api/admin/appearance/logo", adminAppearanceLogoHandler(settings, pool))
	router.Post("/api/admin/appearance/favicon", adminAppearanceFaviconHandler(settings, pool))
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
		store := appsettings.NewStore(pool)
		plans := tariffs.WithCNYDisplay(catalog.Plans(session.User.LanguageCode, effectiveDefaultCurrency(r.Context(), settings, pool)), rate.Rate, rate.Source, rate.UpdatedAt)
		plans = tariffs.WithStarsPrice(plans, store.Float(r.Context(), "STARS_USD_RATE", settings.StarsUSDRate))
		response := map[string]any{"ok": true, "plans": plans, "fx": rate}
		switch kind {
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

func devicesIPsHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
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
		uuid := stringValue(panelUser, "uuid")
		if uuid == "" {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "subscription_not_active"})
			return
		}

		// Trigger IP fetch job
		fetchRes, err := panel.FetchUserIPs(r.Context(), uuid)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		jobID := stringValue(fetchRes, "job_id")
		if jobID == "" {
			// Try alternative field names
			jobID = stringValue(fetchRes, "jobId")
		}
		if jobID == "" {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": "ips_fetch_failed"})
			return
		}

		// Poll for results (up to 30 attempts, 1 second apart)
		for i := 0; i < 30; i++ {
			select {
			case <-r.Context().Done():
				writeJSON(w, http.StatusGatewayTimeout, map[string]any{"ok": false, "error": "ips_poll_timeout"})
				return
			case <-time.After(time.Second):
			}
			result, err := panel.GetFetchUserIPsResult(r.Context(), jobID)
			if err != nil {
				continue
			}
			// Check if results are ready
			if ips, ok := result["ips"]; ok && ips != nil {
				ipsList := anySlice(ips)
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":          true,
					"ips":         ipsList,
					"current_ips": len(ipsList),
				})
				return
			}
			// Some panels may return the result directly
			if _, hasData := result["ips"]; hasData {
				ipsList := anySlice(result["ips"])
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":          true,
					"ips":         ipsList,
					"current_ips": len(ipsList),
				})
				return
			}
		}
		writeJSON(w, http.StatusGatewayTimeout, map[string]any{"ok": false, "error": "ips_poll_timeout"})
	}
}

func devicesIPsDisconnectHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		if panel == nil || !panel.Configured(r.Context()) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "panel_not_configured"})
			return
		}
		var payload struct {
			IP string `json:"ip"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		ip := strings.TrimSpace(payload.IP)
		if ip == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "ip_required"})
			return
		}
		// Use the Remnawave IP drop endpoint
		if err := panel.DropIPConnections(r.Context(), []string{ip}); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": panelErrorCode(err)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func anySlice(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	case []map[string]any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result
	}
	return nil
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
		language := normalizeWebLanguage(payload.Language, effectiveDefaultLanguage(r.Context(), pool, settings))
		if _, err := pool.Exec(r.Context(), "UPDATE users SET language_code=$2 WHERE user_id=$1", session.User.UserID, language); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "language_update_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "language": language})
	}
}

func userNotificationPrefsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, r.Method != http.MethodGet)
		if !ok {
			return
		}
		if r.Method == http.MethodGet {
			prefs := loadUserNotificationPrefs(r.Context(), pool, session.User.UserID)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "notification_prefs": prefs})
			return
		}
		var payload struct {
			ExpiryEnabled       *bool `json:"expiry_enabled"`
			ExpiryDaysBefore    *int  `json:"expiry_days_before"`
			TrafficEnabled      *bool `json:"traffic_enabled"`
			TrafficThresholdPct *int  `json:"traffic_threshold_pct"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		prefs := loadUserNotificationPrefs(r.Context(), pool, session.User.UserID)
		if payload.ExpiryEnabled != nil {
			prefs.ExpiryEnabled = *payload.ExpiryEnabled
		}
		if payload.ExpiryDaysBefore != nil {
			if *payload.ExpiryDaysBefore < 1 || *payload.ExpiryDaysBefore > 30 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "expiry_days_before_must_be_1_to_30"})
				return
			}
			prefs.ExpiryDaysBefore = *payload.ExpiryDaysBefore
		}
		if payload.TrafficEnabled != nil {
			prefs.TrafficEnabled = *payload.TrafficEnabled
		}
		if payload.TrafficThresholdPct != nil {
			if *payload.TrafficThresholdPct < 50 || *payload.TrafficThresholdPct > 100 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "traffic_threshold_pct_must_be_50_to_100"})
				return
			}
			prefs.TrafficThresholdPct = *payload.TrafficThresholdPct
		}
		if err := saveUserNotificationPrefs(r.Context(), pool, session.User.UserID, prefs); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "notification_prefs_save_failed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "notification_prefs": prefs})
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
			promoID := strings.TrimSpace(fmt.Sprint(promos[index]["id"]))
			if promoID == "" {
				promoID = strings.ToUpper(code)
			}
			if promoActivatedByUser(promos[index], session.User.UserID) {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_already_used"})
				return
			}
			maxActivations := int(int64Value(promos[index], "max_activations"))
			currentActivations := promoActivationCount(promos[index])
			_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM promo_activations WHERE promo_id=$1 AND status='applied'", promoID).Scan(&currentActivations)
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
			var activationStatus string
			err := pool.QueryRow(r.Context(), `
INSERT INTO promo_activations(promo_id,promo_code,user_id,status,bonus_days)
VALUES($1,$2,$3,'processing',$4)
ON CONFLICT(promo_id,user_id) DO UPDATE SET status=CASE WHEN promo_activations.status='failed' THEN 'processing' ELSE promo_activations.status END,
 error_code=NULL, updated_at=NOW()
RETURNING status`, promoID, strings.ToUpper(code), session.User.UserID, days).Scan(&activationStatus)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "promo_activation_reserve_failed"})
				return
			}
			if activationStatus != "processing" {
				writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "promo_already_used"})
				return
			}
			panelUser, err := grantPanelAccessDays(r.Context(), pool, panel, session.User, days, accessGrantOptions{Source: "promo:" + strings.ToUpper(code)})
			if err != nil {
				_, _ = pool.Exec(r.Context(), "UPDATE promo_activations SET status='failed', error_code=$3, updated_at=NOW() WHERE promo_id=$1 AND user_id=$2", promoID, session.User.UserID, "grant_failed")
				writePanelActionError(w, err)
				return
			}
			activations := anyList(promos[index]["activations"])
			activations = append(activations, map[string]any{"user_id": session.User.UserID, "activated_at": time.Now().UTC().Format(time.RFC3339)})
			promos[index]["activations"] = activations
			promos[index]["current_activations"] = currentActivations + 1
			if _, err := pool.Exec(r.Context(), "UPDATE promo_activations SET status='applied', applied_at=NOW(), updated_at=NOW() WHERE promo_id=$1 AND user_id=$2 AND status='processing'", promoID, session.User.UserID); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "promo_activation_finalize_failed"})
				return
			}
			if err := writeSettingList(r.Context(), pool, "ADMIN_PROMOS", promos); err != nil {
				slog.Warn("promo legacy counter update failed", "error", err, "promo_id", promoID)
			}
			recordMessageLog(r.Context(), pool, messageLogEntry{UserID: session.User.UserID, EventType: "promo_activation", Content: strings.ToUpper(code), Payload: map[string]any{"promo_id": promoID, "bonus_days": days}})
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
		// Apply risk control to trial activation (same as referral welcome bonus)
		if telemetryEnabled(r.Context(), pool) {
			_, riskScore, riskRule := evaluateWelcomeRisk(r.Context(), settings, pool, r, session.User.UserID, 0)
			if riskRule != "" {
				recordMessageLog(r.Context(), pool, messageLogEntry{UserID: session.User.UserID, EventType: "trial_risk_rejected", Content: "automatic trial risk rejection", Payload: map[string]any{"risk_score": riskScore, "rule_code": riskRule}})
				writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "trial_risk_rejected"})
				return
			}
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
		if referredBy == session.User.UserID {
			writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "referral_welcome_risk_rejected"})
			return
		}
		var fingerprint storedFingerprint
		var riskScore int
		var riskRule string
		if telemetryEnabled(r.Context(), pool) {
			fingerprint, riskScore, riskRule = evaluateWelcomeRisk(r.Context(), settings, pool, r, session.User.UserID, referredBy)
		}
		status := "processing"
		if riskRule != "" {
			status = "rejected"
		}
		var reservedStatus string
		err = pool.QueryRow(r.Context(), `INSERT INTO referral_welcome_claims
(user_id,referrer_id,visitor_hash,fingerprint_hash,status,risk_score,rule_code,bonus_days)
VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8)
ON CONFLICT(user_id) DO UPDATE SET status=CASE WHEN referral_welcome_claims.status='failed' THEN EXCLUDED.status ELSE referral_welcome_claims.status END,
 visitor_hash=CASE WHEN referral_welcome_claims.status='failed' THEN EXCLUDED.visitor_hash ELSE referral_welcome_claims.visitor_hash END,
 fingerprint_hash=CASE WHEN referral_welcome_claims.status='failed' THEN EXCLUDED.fingerprint_hash ELSE referral_welcome_claims.fingerprint_hash END,
 risk_score=CASE WHEN referral_welcome_claims.status='failed' THEN EXCLUDED.risk_score ELSE referral_welcome_claims.risk_score END,
 rule_code=CASE WHEN referral_welcome_claims.status='failed' THEN EXCLUDED.rule_code ELSE referral_welcome_claims.rule_code END, updated_at=NOW()
RETURNING status`, session.User.UserID, referredBy, fingerprint.Visitor, fingerprint.Full, status, riskScore, riskRule, days).Scan(&reservedStatus)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "referral_welcome_reserve_failed"})
			return
		}
		if reservedStatus != "processing" || riskRule != "" {
			recordMessageLog(r.Context(), pool, messageLogEntry{UserID: session.User.UserID, TargetUserID: referredBy, EventType: "referral_welcome_risk_rejected", Content: "automatic risk rejection", Payload: map[string]any{"risk_score": riskScore, "rule_code": riskRule}})
			writeJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "referral_welcome_risk_rejected"})
			return
		}
		panelUser, err := grantPanelAccessDays(r.Context(), pool, panel, session.User, days, accessGrantOptions{Source: "referral_welcome"})
		if err != nil {
			_, _ = pool.Exec(r.Context(), "UPDATE referral_welcome_claims SET status='failed', rule_code='grant_failed', updated_at=NOW() WHERE user_id=$1", session.User.UserID)
			writePanelActionError(w, err)
			return
		}
		tx, err := pool.Begin(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "referral_welcome_finalize_failed"})
			return
		}
		defer func() { _ = tx.Rollback(r.Context()) }()
		if _, err = tx.Exec(r.Context(), "UPDATE users SET referral_welcome_bonus_claimed_at=NOW() WHERE user_id=$1 AND referral_welcome_bonus_claimed_at IS NULL", session.User.UserID); err == nil {
			_, err = tx.Exec(r.Context(), "UPDATE referral_welcome_claims SET status='applied', applied_at=NOW(), updated_at=NOW() WHERE user_id=$1 AND status='processing'", session.User.UserID)
		}
		if err != nil || tx.Commit(r.Context()) != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "referral_welcome_finalize_failed"})
			return
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{UserID: session.User.UserID, TargetUserID: referredBy, EventType: "referral_welcome_applied", Content: "welcome bonus applied", Payload: map[string]any{"bonus_days": days}})
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
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
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
			for page := 0; ; page++ {
				orders, total, err := registry.List(r.Context(), page, 100)
				if err != nil {
					break
				}
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
				if len(orders) == 0 || int64((page+1)*100) >= total {
					break
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
		ctx := r.Context()
		store := appsettings.NewStore(pool)
		alerts := []map[string]any{}
		addAlert := func(severity, key string, params map[string]any, sections ...string) {
			alerts = append(alerts, map[string]any{"id": fmt.Sprintf("%s:%d", key, len(alerts)), "severity": severity, "message_key": key, "params": params, "sections": sections})
		}
		if info, err := os.Stat("data"); err != nil || !info.IsDir() {
			addAlert("error", "data_dir_missing", map[string]any{"path": "data"}, "settings", "backups", "tariffs")
		} else if info.Mode().Perm()&0o200 == 0 {
			addAlert("error", "data_dir_not_writable", map[string]any{"path": "data"}, "backups", "tariffs", "appearance")
		}
		if _, err := tariffs.Load("data/tariffs.json"); err != nil {
			if _, fallbackErr := tariffs.Load("data/tariffs.example.json"); fallbackErr != nil {
				addAlert("error", "tariffs_config_invalid", map[string]any{"path": "data/tariffs.json", "error": err.Error()}, "tariffs")
			}
		}
		guidesEnabled := store.Bool(ctx, "SUBSCRIPTION_GUIDES_ENABLED", false)
		if guidesEnabled {
			configured := false
			if raw, ok, _ := store.Get(ctx, "SUBSCRIPTION_GUIDES_CONFIG"); ok {
				var object map[string]any
				configured = json.Unmarshal(raw, &object) == nil && validSubscriptionGuidesConfig(object)
			}
			if !configured {
				for _, path := range []string{"data/subscription-guides.json", "data/subscription-guides.example.json"} {
					if body, err := os.ReadFile(path); err == nil {
						var object map[string]any
						if json.Unmarshal(body, &object) == nil && validSubscriptionGuidesConfig(object) {
							configured = true
							break
						}
					}
				}
			}
			if !configured {
				addAlert("error", "subscription_page_config_invalid", map[string]any{"error": "configuration is empty"}, "settings")
			}
		}
		webhookBase := store.String(ctx, "WEBHOOK_BASE_URL", settings.WebhookBaseURL)
		if store.Bool(ctx, "EZPAY_ENABLED", settings.EZPay.Enabled) && (store.String(ctx, "EZPAY_BASE_URL", settings.EZPay.BaseURL) == "" || store.String(ctx, "EZPAY_KEY", settings.EZPay.Key) == "") {
			addAlert("error", "provider_not_configured", map[string]any{"provider": "EZPay"}, "settings", "payments")
		}
		if store.Bool(ctx, "BEPUSDT_ENABLED", settings.BEPUSDT.Enabled) && (store.String(ctx, "BEPUSDT_BASE_URL", settings.BEPUSDT.BaseURL) == "" || store.String(ctx, "BEPUSDT_TOKEN", settings.BEPUSDT.Token) == "") {
			addAlert("error", "provider_not_configured", map[string]any{"provider": "BEPUSDT"}, "settings", "payments")
		}
		if (store.Bool(ctx, "EZPAY_ENABLED", settings.EZPay.Enabled) || store.Bool(ctx, "BEPUSDT_ENABLED", settings.BEPUSDT.Enabled)) && strings.TrimSpace(webhookBase) == "" {
			addAlert("warning", "provider_webhook_needs_base_url", map[string]any{"provider": "payment"}, "settings", "payments")
		}
		if strings.TrimSpace(settings.SubscriptionMiniApp) == "" {
			addAlert("warning", "mini_app_url_missing", map[string]any{}, "settings")
		}
		if store.Bool(ctx, "SMTP_ENABLED", false) && !mailerConfigFromSettings(ctx, store).IsConfigured() {
			addAlert("error", "smtp_incomplete", map[string]any{}, "settings")
		}
		if strings.TrimSpace(settings.BotToken) == "" {
			addAlert("error", "bot_token_invalid", map[string]any{}, "settings")
		}
		checks := map[string]any{"database": pool != nil, "remnawave_configured": panel != nil && panel.Configured(ctx)}
		if panel == nil || !panel.Configured(ctx) {
			addAlert("warning", "panel_api_not_configured", map[string]any{}, "settings", "users", "tariffs")
		} else if _, err := panel.GetSystemStats(ctx); err != nil {
			checks["remnawave_api"] = false
			addAlert("error", "panel_api_unreachable", map[string]any{"url": store.String(ctx, "PANEL_API_URL", settings.PanelAPIURL)}, "settings", "users")
		} else {
			checks["remnawave_api"] = true
		}
		status := "ok"
		for _, alert := range alerts {
			if alert["severity"] == "error" {
				status = "error"
				break
			}
			status = "warning"
		}
		payload := map[string]any{"ok": true, "status": status, "checks": checks, "alerts": alerts, "checked_at": time.Now().UTC()}
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
		ctx := r.Context()
		users := map[string]any{}
		var totalUsers, activeToday, bannedUsers, referralUsers, activeSubscriptions, paidSubscriptions, trialUsers, expiredSubscriptions int64
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE registration_date>=CURRENT_DATE").Scan(&activeToday)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_banned=TRUE").Scan(&bannedUsers)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE referred_by_id IS NOT NULL").Scan(&referralUsers)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW())").Scan(&activeSubscriptions)
		_ = pool.QueryRow(ctx, "SELECT COUNT(DISTINCT user_id) FROM payment_orders WHERE status IN ('paid','succeeded')").Scan(&paidSubscriptions)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE UPPER(COALESCE(panel_status,''))='EXPIRED' OR panel_expire_at<=NOW()").Scan(&expiredSubscriptions)
		trialUserIDs := map[int64]struct{}{}
		for _, activation := range readSettingList(ctx, pool, "TRIAL_ACTIVATIONS") {
			if userID := int64Value(activation, "user_id"); userID > 0 {
				trialUserIDs[userID] = struct{}{}
			}
		}
		trialUsers = int64(len(trialUserIDs))
		users["total_users"] = totalUsers
		users["active_today"] = activeToday
		users["banned_users"] = bannedUsers
		users["referral_users"] = referralUsers
		users["active_subscriptions"] = activeSubscriptions
		users["paid_subscriptions"] = paidSubscriptions
		users["free_subscription_users"] = maxInt64(0, activeSubscriptions-paidSubscriptions)
		users["trial_users"] = trialUsers
		users["inactive_users"] = maxInt64(0, totalUsers-activeSubscriptions)
		users["expired_subscription_users"] = expiredSubscriptions

		financial := map[string]any{}
		var todayRevenue, weekRevenue, monthRevenue, allRevenue float64
		var todayPayments int64
		_ = pool.QueryRow(ctx, "SELECT COALESCE(SUM(COALESCE(base_amount,amount)),0)::float8,COUNT(*) FROM payment_orders WHERE status IN ('paid','succeeded') AND COALESCE(paid_at,updated_at)>=CURRENT_DATE").Scan(&todayRevenue, &todayPayments)
		_ = pool.QueryRow(ctx, "SELECT COALESCE(SUM(COALESCE(base_amount,amount)),0)::float8 FROM payment_orders WHERE status IN ('paid','succeeded') AND COALESCE(paid_at,updated_at)>=NOW()-INTERVAL '7 days'").Scan(&weekRevenue)
		_ = pool.QueryRow(ctx, "SELECT COALESCE(SUM(COALESCE(base_amount,amount)),0)::float8 FROM payment_orders WHERE status IN ('paid','succeeded') AND COALESCE(paid_at,updated_at)>=NOW()-INTERVAL '30 days'").Scan(&monthRevenue)
		_ = pool.QueryRow(ctx, "SELECT COALESCE(SUM(COALESCE(base_amount,amount)),0)::float8 FROM payment_orders WHERE status IN ('paid','succeeded')").Scan(&allRevenue)
		financial["today_revenue"] = todayRevenue
		financial["today_payments_count"] = todayPayments
		financial["week_revenue"] = weekRevenue
		financial["month_revenue"] = monthRevenue
		financial["all_time_revenue"] = allRevenue
		dailySeries := []map[string]any{}
		rows, err := pool.Query(ctx, `SELECT day::date::text,COALESCE(SUM(COALESCE(p.base_amount,p.amount)),0)::float8
FROM generate_series(CURRENT_DATE-INTERVAL '364 days',CURRENT_DATE,INTERVAL '1 day') day
LEFT JOIN payment_orders p ON p.status IN ('paid','succeeded') AND COALESCE(p.paid_at,p.updated_at)>=day AND COALESCE(p.paid_at,p.updated_at)<day+INTERVAL '1 day'
GROUP BY day ORDER BY day`)
		if err == nil {
			for rows.Next() {
				var date string
				var amount float64
				if rows.Scan(&date, &amount) == nil {
					dailySeries = append(dailySeries, map[string]any{"date": date, "amount": amount})
				}
			}
			rows.Close()
		}
		financial["daily_series"] = dailySeries

		recentPayments := []map[string]any{}
		paymentRows, err := pool.Query(ctx, `SELECT p.payment_id,p.user_id,COALESCE(NULLIF(u.username,''),NULLIF(u.email,''),p.user_id::text),p.amount::float8,p.currency,p.provider,p.status,p.created_at
FROM payment_orders p LEFT JOIN users u ON u.user_id=p.user_id ORDER BY p.created_at DESC LIMIT 10`)
		if err == nil {
			for paymentRows.Next() {
				var paymentID, userID int64
				var userLabel, currency, provider, status string
				var amount float64
				var createdAt time.Time
				if paymentRows.Scan(&paymentID, &userID, &userLabel, &amount, &currency, &provider, &status, &createdAt) == nil {
					recentPayments = append(recentPayments, map[string]any{"payment_id": paymentID, "user_id": userID, "user_label": userLabel, "amount": amount, "currency": currency, "provider": provider, "status": status, "created_at": createdAt})
				}
			}
			paymentRows.Close()
		}

		var anonymousVisitors, inviteVisits, rejectedRewards int64
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM visitor_telemetry").Scan(&anonymousVisitors)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM invite_visits").Scan(&inviteVisits)
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM referral_welcome_claims WHERE status='rejected'").Scan(&rejectedRewards)
		var heartbeatDate, version, provenance, osName, locale, userRange string
		_ = pool.QueryRow(ctx, `SELECT heartbeat_date::text,version,provenance,os,locale,user_count_range FROM installation_heartbeats ORDER BY heartbeat_date DESC LIMIT 1`).Scan(&heartbeatDate, &version, &provenance, &osName, &locale, &userRange)
		payload := map[string]any{
			"ok":              true,
			"currency_symbol": effectiveDefaultCurrency(ctx, settings, pool),
			"users":           users,
			"financial":       financial,
			"recent_payments": recentPayments,
			"panel_sync":      LastPanelSyncStatus(ctx, pool),
			"local_analytics": map[string]any{"anonymous_visitors": anonymousVisitors, "invite_visits": inviteVisits, "rejected_welcome_rewards": rejectedRewards, "heartbeat": map[string]any{"date": heartbeatDate, "version": version, "provenance": provenance, "os": osName, "locale": locale, "user_count_range": userRange}},
		}
		var queuedMessages, failedMessages int64
		_ = pool.QueryRow(ctx, "SELECT COUNT(*) FILTER (WHERE status='queued'),COUNT(*) FILTER (WHERE status='failed') FROM telegram_outbox").Scan(&queuedMessages, &failedMessages)
		payload["queue"] = map[string]any{"user_queue_size": queuedMessages, "failed_messages": failedMessages}
		if panel != nil && panel.Configured(ctx) {
			panelStats := map[string]any{}
			if stats, err := panel.GetSystemStats(ctx); err == nil {
				panelStats["system"] = stats
			}
			if bandwidth, err := panel.GetBandwidthStats(ctx); err == nil {
				panelStats["bandwidth"] = bandwidth
			}
			if nodes, err := panel.GetNodesStats(ctx); err == nil {
				panelStats["nodes"] = nodes
			}
			payload["panel"] = panelStats
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
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
		for _, item := range items {
			code := strings.TrimSpace(fmt.Sprint(item["start_param"]))
			var visits, registrations, conversions int64
			_ = pool.QueryRow(r.Context(), `SELECT COUNT(*),COUNT(registered_user_id),COUNT(converted_at) FROM invite_visits WHERE kind='campaign' AND UPPER(code)=UPPER($1)`, code).Scan(&visits, &registrations, &conversions)
			item["stats"] = map[string]any{"visits": visits, "registrations": registrations, "conversions": conversions}
			item["invite_link"] = referralWebAppLink(settings, code)
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "campaigns": items, "totals": map[string]any{"cost": 0, "revenue": 0}})
	}
}

func adminListSettingHandler(settings config.Settings, pool *pgxpool.Pool, key string, responseKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		items := readSettingList(r.Context(), pool, key)
		if key == "ADMIN_PROMOS" {
			for _, item := range items {
				promoID := strings.TrimSpace(fmt.Sprint(item["id"]))
				if promoID == "" {
					promoID = strings.ToUpper(strings.TrimSpace(fmt.Sprint(item["code"])))
				}
				var count int
				_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM promo_activations WHERE promo_id=$1 AND status='applied'", promoID).Scan(&count)
				item["current_activations"] = count
			}
		}
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
		counts := map[string]int64{}
		queries := map[string]string{
			"all":                    "is_banned=FALSE",
			"active":                 "is_banned=FALSE AND UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW())",
			"inactive":               "is_banned=FALSE AND NOT (UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW()))",
			"expired":                "is_banned=FALSE AND (UPPER(COALESCE(panel_status,''))='EXPIRED' OR panel_expire_at<=NOW())",
			"active_never_connected": "is_banned=FALSE AND UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW()) AND COALESCE(lifetime_used_traffic_bytes,0)=0",
			"never":                  "is_banned=FALSE AND COALESCE(panel_user_uuid,'')=''",
		}
		for key, where := range queries {
			var count int64
			_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users WHERE "+where).Scan(&count)
			counts[key] = count
		}
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
		recipients, err := broadcastRecipients(r.Context(), pool, payload.Target)
		if err != nil {
			if err.Error() == "invalid_broadcast_target" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_broadcast_target"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "broadcast_recipients_failed"})
			return
		}
		batchID := fmt.Sprintf("broadcast-%d-%d", session.User.UserID, time.Now().UnixNano())
		tx, err := pool.Begin(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "broadcast_queue_failed"})
			return
		}
		queued := 0
		for _, recipient := range recipients {
			if _, err := tx.Exec(r.Context(), `INSERT INTO telegram_outbox(batch_id,user_id,chat_id,text,parse_mode) VALUES($1,$2,$3,$4,'HTML')`, batchID, recipient.UserID, recipient.ChatID, strings.TrimSpace(payload.Text)); err != nil {
				_ = tx.Rollback(r.Context())
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "broadcast_queue_failed"})
				return
			}
			queued++
		}
		if err := tx.Commit(r.Context()); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "broadcast_queue_failed"})
			return
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			UserID:       session.User.UserID,
			EventType:    "admin_broadcast",
			Content:      payload.Text,
			IsAdminEvent: true,
			Payload:      map[string]any{"target": payload.Target, "queued": queued, "batch_id": batchID},
		})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "queued": queued, "failed": 0, "batch_id": batchID})
	}
}

type broadcastRecipient struct {
	UserID int64
	ChatID int64
}

func broadcastRecipients(ctx context.Context, pool *pgxpool.Pool, target string) ([]broadcastRecipient, error) {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" {
		target = "all"
	}
	where := "is_banned=FALSE"
	switch target {
	case "active":
		where += " AND UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW())"
	case "inactive":
		where += " AND NOT (UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW()))"
	case "never":
		where += " AND COALESCE(panel_user_uuid,'')=''"
	case "expired":
		where += " AND (UPPER(COALESCE(panel_status,''))='EXPIRED' OR panel_expire_at<=NOW())"
	case "active_never_connected":
		where += " AND UPPER(COALESCE(panel_status,'')) IN ('ACTIVE','LIMITED') AND (panel_expire_at IS NULL OR panel_expire_at>NOW()) AND COALESCE(lifetime_used_traffic_bytes,0)=0"
	default:
		if target != "all" {
			return nil, fmt.Errorf("invalid_broadcast_target")
		}
	}
	rows, err := pool.Query(ctx, `
SELECT user_id, COALESCE(telegram_id,0)
FROM users
WHERE `+where+`
ORDER BY registration_date DESC`)
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
		backupDir := filepath.Join("data", "backups")
		_ = os.MkdirAll(backupDir, 0o750)
		entries, err := os.ReadDir(backupDir)
		archives := []backupArchiveInfo{}
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				archives = append(archives, inspectBackupArchive(filepath.Join(backupDir, entry.Name())))
			}
			// Sort by creation time descending
			sort.Slice(archives, func(i, j int) bool {
				return archives[i].CreatedAt > archives[j].CreatedAt
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "archives": archives, "backup_dir": backupDir})
	}
}

func adminBackupCreateHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		backupDir := filepath.Join("data", "backups")
		_ = os.MkdirAll(backupDir, 0o750)
		filename := "backup-" + time.Now().Format("20060102-150405") + ".zip"
		savePath := filepath.Join(backupDir, filename)
		result, err := createBackupArchive(r.Context(), settings, savePath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "backup_create_failed", "note": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result, "archive": result})
	}
}

func adminBackupDownloadHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		name := filepath.Base(chi.URLParam(r, "name"))
		if name == "." || name == "" || !strings.HasSuffix(strings.ToLower(name), ".zip") {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join("data", "backups", name)
		if _, err := os.Stat(path); err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		w.Header().Set("Content-Type", "application/zip")
		http.ServeFile(w, r, path)
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

func adminAppearanceLogoHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}

		contentType := r.Header.Get("Content-Type")

		// Handle JSON URL upload
		if strings.Contains(contentType, "application/json") {
			var payload struct {
				URL string `json:"url"`
			}
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if json.Unmarshal(body, &payload) == nil && safeAppearanceURL(payload.URL) {
				logoURL := strings.TrimSpace(payload.URL)
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":          true,
					"logo_url":    logoURL,
					"favicon_url": "",
				})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}

		// Handle multipart file upload
		if err := r.ParseMultipartForm(5 << 20); err != nil { //nolint:gosec // G120: bounded to 5 MB.
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_upload"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_required"})
			return
		}
		defer func() { _ = file.Close() }()
		if header.Size > 5<<20 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_too_large"})
			return
		}
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mime := http.DetectContentType(buf[:n])
		ext, ok := safeImageExtension(mime)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "unsupported_mime"})
			return
		}
		// Read full file
		_, _ = file.Seek(0, io.SeekStart)
		full, err := io.ReadAll(io.LimitReader(file, 5<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "read_failed"})
			return
		}

		// Save to uploads directory
		uploadDir := filepath.Join("data", "uploads", "logos")
		_ = os.MkdirAll(uploadDir, 0o750)
		hash := sha256.Sum256(full)
		hashStr := hex.EncodeToString(hash[:])[:16]
		filename := "logo-" + hashStr + "-" + time.Now().Format("20060102150405") + ext
		savePath := filepath.Join(uploadDir, filename)
		if err := os.WriteFile(savePath, full, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}

		logoURL := "/webapp-uploaded-logo/" + filename
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"logo_url":    logoURL,
			"favicon_url": "",
		})
	}
}

func adminAppearanceFaviconHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}

		contentType := r.Header.Get("Content-Type")

		// Handle JSON URL upload
		if strings.Contains(contentType, "application/json") {
			var payload struct {
				URL string `json:"url"`
			}
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if json.Unmarshal(body, &payload) == nil && safeAppearanceURL(payload.URL) {
				faviconURL := strings.TrimSpace(payload.URL)
				writeJSON(w, http.StatusOK, map[string]any{
					"ok":          true,
					"favicon_url": faviconURL,
					"variants":    map[string]string{},
				})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}

		// Handle multipart file upload
		if err := r.ParseMultipartForm(5 << 20); err != nil { //nolint:gosec // G120: bounded to 5 MB.
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_upload"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_required"})
			return
		}
		defer func() { _ = file.Close() }()
		if header.Size > 5<<20 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_too_large"})
			return
		}
		buf := make([]byte, 512)
		n, _ := file.Read(buf)
		mime := http.DetectContentType(buf[:n])
		ext, ok := safeImageExtension(mime)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "unsupported_mime"})
			return
		}
		_, _ = file.Seek(0, io.SeekStart)
		full, err := io.ReadAll(io.LimitReader(file, 5<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "read_failed"})
			return
		}

		uploadDir := filepath.Join("data", "uploads", "favicons")
		_ = os.MkdirAll(uploadDir, 0o750)
		hash := sha256.Sum256(full)
		hashStr := hex.EncodeToString(hash[:])[:16]
		filename := "favicon-" + hashStr + "-" + time.Now().Format("20060102150405") + ext
		savePath := filepath.Join(uploadDir, filename)
		if err := os.WriteFile(savePath, full, 0o600); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"favicon_url": "/webapp-uploaded-logo/" + filename,
			"variants":    map[string]string{},
		})
	}
}

func safeAppearanceURL(raw string) bool {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, "\r\n\x00") {
		return false
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		return true
	}
	lower := strings.ToLower(value)
	return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://")
}

func safeImageExtension(mime string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0])) {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico", true
	default:
		return "", false
	}
}

func adminBackupUploadHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		if err := r.ParseMultipartForm(256 << 20); err != nil { //nolint:gosec // G120: bounded to 256 MB for backup uploads.
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_upload"})
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_required"})
			return
		}
		defer func() { _ = file.Close() }()
		if header.Size > 256<<20 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "file_too_large"})
			return
		}
		backupDir := filepath.Join("data", "backups")
		_ = os.MkdirAll(backupDir, 0o750)
		if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "zip_required"})
			return
		}
		filename := "uploaded-" + time.Now().Format("20060102-150405") + ".zip"
		savePath := filepath.Join(backupDir, filename)
		f, err := os.Create(savePath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}
		defer func() { _ = f.Close() }()
		if _, err := io.Copy(f, file); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save_failed"})
			return
		}
		info := inspectBackupArchive(savePath)
		if slices.Contains(info.Warnings, "invalid_zip") {
			_ = os.Remove(savePath)
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_zip"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "archive": info})
	}
}

func adminBackupRestoreHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		var payload struct {
			ArchiveName     string `json:"archive_name"`
			RestoreDatabase bool   `json:"restore_database"`
			RestoreCompose  bool   `json:"restore_compose"`
			Confirm         any    `json:"confirm"`
		}
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		if strings.TrimSpace(payload.ArchiveName) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "archive_name_required"})
			return
		}
		confirmed := boolish(payload.Confirm, false) || strings.EqualFold(strings.TrimSpace(fmt.Sprint(payload.Confirm)), "restore")
		if !confirmed {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "confirm_restore_required"})
			return
		}

		backupDir := filepath.Join("data", "backups")
		archivePath := filepath.Join(backupDir, filepath.Base(payload.ArchiveName))
		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "archive_not_found"})
			return
		}

		result := map[string]any{"restored_database": false, "restored_compose": false, "errors": []string{}}
		if payload.RestoreCompose {
			snapshot, snapshotErr := snapshotComposeBeforeRestore(backupDir)
			if snapshotErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "pre_restore_snapshot_failed"})
				return
			}
			result["pre_restore_snapshot"] = snapshot
		}

		tempDir, err := os.MkdirTemp("", "remna-restore-")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "restore_prepare_failed"})
			return
		}
		defer func() { _ = os.RemoveAll(tempDir) }()
		databasePath := archivePath
		if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
			var composeRestored bool
			databasePath, composeRestored, err = extractBackupArchive(archivePath, tempDir, payload.RestoreCompose)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_backup_archive", "message": err.Error()})
				return
			}
			result["restored_compose"] = composeRestored
		}

		if payload.RestoreDatabase {
			// Try to restore database using pg_restore if postgres tools are available
			if settings.DatabaseURL != "" && databasePath != "" {
				cmd := safePgCommand(r.Context(), settings.DatabaseURL, "pg_restore",
					"--clean",
					"--if-exists",
					"--no-owner",
					"--no-privileges",
					databasePath,
				)
				output, err := cmd.CombinedOutput()
				if err != nil {
					result["errors"] = append(result["errors"].([]string), "database_restore_failed: "+string(output))
				} else {
					result["restored_database"] = true
				}
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"result": result,
		})
	}
}

func accountTelegramLinkHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireSession(w, r, settings, pool, true)
		if !ok {
			return
		}
		var payload telegramAuthPayload
		if err := decodeJSONBody(r, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_json"})
			return
		}
		tgUser, err := validateTelegramAuthPayload(r.Context(), r, settings, pool, payload)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_telegram_auth"})
			return
		}
		var ownerID int64
		err = pool.QueryRow(r.Context(), "SELECT user_id FROM users WHERE telegram_id=$1", tgUser.ID).Scan(&ownerID)
		if err == nil && ownerID != session.User.UserID {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "telegram_already_linked"})
			return
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "telegram_link_failed"})
			return
		}
		_, err = pool.Exec(r.Context(), `UPDATE users SET telegram_id=$2,username=$3,first_name=COALESCE(NULLIF($4,''),first_name),last_name=COALESCE(NULLIF($5,''),last_name),telegram_photo_url=COALESCE(NULLIF($6,''),telegram_photo_url) WHERE user_id=$1`,
			session.User.UserID, tgUser.ID, emptyStringToNil(tgUser.Username), tgUser.FirstName, tgUser.LastName, tgUser.PhotoURL)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "telegram_link_failed"})
			return
		}
		http.SetCookie(w, &http.Cookie{Name: telegramNonceCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: requestIsHTTPS(r), SameSite: http.SameSiteLaxMode}) //nolint:gosec // G124: attributes set dynamically.
		user, _ := loadWebappUser(r.Context(), pool, session.User.UserID, settings)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user, "csrf_token": session.Claims.CSRF})
	}
}

func subscriptionGuidesHandler(settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isPublic := strings.Contains(r.URL.Path, "/public/")
		store := appsettings.NewStore(pool)
		var user webappUser
		shareToken := ""
		if isPublic {
			shareToken = strings.ToLower(strings.TrimSpace(chi.URLParam(r, "share_token")))
			if len(shareToken) != 32 || strings.Trim(shareToken, "0123456789abcdef") != "" {
				writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "share_not_found"})
				return
			}
			var userID int64
			if err := pool.QueryRow(r.Context(), "SELECT user_id FROM subscription_share_tokens WHERE token=$1", shareToken).Scan(&userID); err != nil {
				writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "share_not_found"})
				return
			}
			var err error
			user, err = loadWebappUser(r.Context(), pool, userID, settings)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "share_not_found"})
				return
			}
			_, _ = pool.Exec(r.Context(), "UPDATE subscription_share_tokens SET last_used_at=NOW() WHERE token=$1", shareToken)
		} else {
			session, ok := requireSession(w, r, settings, pool, false)
			if !ok {
				return
			}
			user = session.User
		}

		enabled := store.Bool(r.Context(), "SUBSCRIPTION_GUIDES_ENABLED", false)
		if !enabled {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "enabled": false, "config": nil, "source": nil, "subscription": nil})
			return
		}

		var config map[string]any
		source := "runtime"
		if raw, ok, _ := store.Get(r.Context(), "SUBSCRIPTION_GUIDES_CONFIG"); ok {
			_ = json.Unmarshal(raw, &config)
			if !validSubscriptionGuidesConfig(config) {
				config = nil
			}
		}
		if config == nil {
			source = "file"
			for _, configPath := range []string{filepath.Join("data", "subscription-guides.json"), filepath.Join("data", "subscription-guides.example.json")} {
				if body, err := os.ReadFile(configPath); err == nil && json.Unmarshal(body, &config) == nil && validSubscriptionGuidesConfig(config) {
					break
				}
			}
		}
		if !validSubscriptionGuidesConfig(config) {
			writeJSON(w, http.StatusConflict, map[string]any{"ok": false, "error": "subscription_guides_not_configured"})
			return
		}
		subscription := map[string]any{"active": false}
		if panel != nil && panel.Configured(r.Context()) {
			if panelUser, found, _ := panelUserForWebUser(r.Context(), pool, panel, user); found {
				subscription = subscriptionFromPanelUser(r.Context(), pool, user, panelUser)
			}
		}
		if shareToken != "" {
			subscription["install_share_token"] = shareToken
			subscription["share_url"] = strings.TrimRight(webappPublicBaseURL(r, settings), "/") + "/s/" + shareToken
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":           true,
			"enabled":      enabled,
			"config":       config,
			"source":       source,
			"subscription": subscription,
		})
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
		for _, campaign := range readSettingList(ctx, pool, "ADMIN_ADS") {
			if strings.EqualFold(strings.TrimSpace(fmt.Sprint(campaign["start_param"])), code) {
				_, _ = pool.Exec(ctx, `UPDATE invite_visits SET registered_user_id=$2 WHERE visit_id=(SELECT visit_id FROM invite_visits WHERE kind='campaign' AND UPPER(code)=UPPER($1) ORDER BY last_seen_at DESC LIMIT 1)`, code, userID)
				return
			}
		}
		return
	}
	result, _ := pool.Exec(ctx, "UPDATE users SET referred_by_id=$2 WHERE user_id=$1 AND referred_by_id IS NULL", userID, referrerID)
	if result.RowsAffected() > 0 {
		_, _ = pool.Exec(ctx, `UPDATE invite_visits SET registered_user_id=$2 WHERE visit_id=(SELECT visit_id FROM invite_visits WHERE kind='referral' AND UPPER(code)=UPPER($1) ORDER BY last_seen_at DESC LIMIT 1)`, code, userID)
	}
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
