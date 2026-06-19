package httpapi

import (
	"context"
	"encoding/csv"
	"encoding/json"
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
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/tariffs"
	"remna-user-panel/internal/webassets"
)

func registerExtraAPIRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry) {
	_ = catalog
	router.Get("/api/tariffs/topup-options", webappPlansOptionsHandler(settings, pool, "topup"))
	router.Get("/api/devices/topup-options", webappPlansOptionsHandler(settings, pool, "devices"))
	router.Get("/api/tariffs/change-options", webappPlansOptionsHandler(settings, pool, "change"))
	router.Post("/api/tariffs/change", okSessionMutation(settings, pool))
	router.Post("/api/tariffs/change-payment", createPaymentHandler(settings, pool, registry))
	router.Post("/api/subscription/auto-renew", okSessionMutation(settings, pool))
	router.Get("/api/devices", devicesHandler(settings, pool))
	router.Post("/api/devices/disconnect", okSessionMutation(settings, pool))
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
	router.Post("/api/support/tickets/{ticket_id}/read", okSessionMutation(settings, pool))
	router.Get("/api/support/unread", supportUnreadHandler(settings, pool))

	router.Get("/api/admin/tariffs", adminTariffsHandler(settings, pool))
	router.Put("/api/admin/tariffs", adminTariffsHandler(settings, pool))
	router.Get("/api/admin/panel/internal-squads", adminSquadsHandler(settings, pool))
	router.Get("/api/admin/payments/export.csv", adminPaymentsExportHandler(settings, pool, registry))
	router.Get("/api/admin/health", adminHealthHandler(settings, pool))
	router.Get("/api/admin/stats", adminStatsHandler(settings, pool))
	router.Post("/api/admin/sync", okAdminMutation(settings, pool))
	router.Get("/api/admin/logs", adminLogsHandler(settings, pool))
	router.Get("/api/admin/users", adminUsersListHandler(settings, pool))
	router.Get("/api/admin/users/{user_id}", adminUserDetailHandler(settings, pool))
	router.Delete("/api/admin/users/{user_id}", adminUserDeleteHandler(settings, pool))
	router.Get("/api/admin/users/{user_id}/referrals", adminUserReferralsHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/ban", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/message", okAdminMutation(settings, pool))
	router.Post("/api/admin/users/{user_id}/message/preview", adminMessagePreviewHandler(settings, pool))
	router.Get("/api/admin/users/{user_id}/telegram-profile-link", adminTelegramProfileLinkHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/extend", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/tariff", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/reset-trial", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/premium-override", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/regular-traffic-override", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/hwid-device-limit", adminUserActionHandler(settings, pool))
	router.Post("/api/admin/users/{user_id}/traffic-grant", adminUserActionHandler(settings, pool))

	router.Get("/api/admin/promos", adminListSettingHandler(settings, pool, "ADMIN_PROMOS", "promos"))
	router.Post("/api/admin/promos", adminCreateSettingItemHandler(settings, pool, "ADMIN_PROMOS", "promo"))
	router.Patch("/api/admin/promos/{id}", adminPatchSettingItemHandler(settings, pool, "ADMIN_PROMOS", "promo"))
	router.Delete("/api/admin/promos/{id}", adminDeleteSettingItemHandler(settings, pool, "ADMIN_PROMOS"))
	router.Get("/api/admin/ads", adminAdsListHandler(settings, pool))
	router.Post("/api/admin/ads", adminCreateSettingItemHandler(settings, pool, "ADMIN_ADS", "campaign"))
	router.Post("/api/admin/ads/{id}/toggle", adminPatchSettingItemHandler(settings, pool, "ADMIN_ADS", "campaign"))
	router.Delete("/api/admin/ads/{id}", adminDeleteSettingItemHandler(settings, pool, "ADMIN_ADS"))
	router.Get("/api/admin/broadcast/audience-counts", adminBroadcastAudienceHandler(settings, pool))
	router.Post("/api/admin/broadcast", okAdminMutation(settings, pool))
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
	router.Post("/api/admin/support/tickets/{ticket_id}/read", okAdminMutation(settings, pool))
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

func devicesHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSession(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "devices": []any{}, "limit": 0, "used": 0})
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

func adminSquadsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "squads": []any{}})
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

func adminHealthHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		payload := map[string]any{"ok": true, "status": "ok", "checks": map[string]any{"database": pool != nil}}
		if pool != nil {
			payload["db_pool"] = pool.Stat()
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func adminStatsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		var users, payments int64
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&users)
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM payment_orders").Scan(&payments)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": users, "payments": payments, "revenue": 0, "series": []any{}})
	}
}

func adminLogsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "logs": []any{}, "total": 0})
	}
}

func adminUsersListHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize <= 0 || pageSize > 100 {
			pageSize = 25
		}
		rows, err := pool.Query(r.Context(), `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), is_banned, registration_date
FROM users ORDER BY registration_date DESC LIMIT $1 OFFSET $2`, pageSize, page*pageSize)
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
			if err := rows.Scan(&id, &telegramID, &username, &email, &firstName, &lastName, &language, &banned, &created); err != nil {
				continue
			}
			users = append(users, userAdminPayload(id, telegramID, username, email, firstName, lastName, language, banned, created))
		}
		var total int64
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&total)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "users": users, "total": total})
	}
}

func adminUserDetailHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		user, err := loadAdminUser(r.Context(), pool, chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "user_not_found"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user, "active_subscription": nil, "payments": []any{}, "logs": []any{}})
	}
}

func adminUserDeleteHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		userID, err := parsePositiveInt64(chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
			return
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

func adminUserActionHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		userID, err := parsePositiveInt64(chi.URLParam(r, "user_id"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_user_id"})
			return
		}
		if strings.HasSuffix(r.URL.Path, "/ban") {
			var payload struct {
				IsBanned bool `json:"is_banned"`
			}
			_ = decodeJSONBody(r, &payload)
			_, _ = pool.Exec(r.Context(), "UPDATE users SET is_banned=$2 WHERE user_id=$1", userID, payload.IsBanned)
		}
		user, _ := loadAdminUser(r.Context(), pool, strconv.FormatInt(userID, 10))
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "user": user, "active_subscription": nil})
	}
}

func adminMessagePreviewHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, true); !ok {
			return
		}
		var payload map[string]any
		_ = decodeJSONBody(r, &payload)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "preview": payload["message"], "html": payload["message"]})
	}
}

func adminTelegramProfileLinkHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
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
		var total int64
		_ = pool.QueryRow(r.Context(), "SELECT COUNT(*) FROM users").Scan(&total)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "total": total, "audiences": map[string]any{"all": total}})
	}
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
		if admin {
			if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
				return
			}
		} else if _, ok := requireSession(w, r, settings, pool, false); !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "tickets": tickets, "counts": supportCounts(tickets), "total": len(tickets)})
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
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		id := nextListID(tickets)
		now := time.Now().Format(time.RFC3339)
		ticket := map[string]any{"ticket_id": id, "id": id, "user_id": session.User.UserID, "status": "open", "subject": payload["subject"], "body": payload["body"], "created_at": now, "updated_at": now, "messages": []any{}}
		tickets = append(tickets, ticket)
		_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": ticket})
	}
}

func supportDetailHandler(settings config.Settings, pool *pgxpool.Pool, admin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if admin {
			if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
				return
			}
		} else if _, ok := requireSession(w, r, settings, pool, false); !ok {
			return
		}
		ticket, ok := findSettingItem(r.Context(), pool, "SUPPORT_TICKETS", chi.URLParam(r, "ticket_id"), "ticket_id")
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
			return
		}
		messages, _ := ticket["messages"].([]any)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": ticket, "messages": messages})
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
			Body string `json:"body"`
		}
		if err := decodeJSONBody(r, &payload); err != nil || strings.TrimSpace(payload.Body) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_message"})
			return
		}
		id := chi.URLParam(r, "ticket_id")
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		for index := range tickets {
			if fmt.Sprint(tickets[index]["ticket_id"]) == id {
				messages, _ := tickets[index]["messages"].([]any)
				message := map[string]any{"message_id": len(messages) + 1, "ticket_id": id, "body": payload.Body, "is_admin": admin, "user_id": session.User.UserID, "created_at": time.Now().Format(time.RFC3339)}
				tickets[index]["messages"] = append(messages, message)
				tickets[index]["updated_at"] = time.Now().Format(time.RFC3339)
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": tickets[index], "message": message})
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
					tickets[index][k] = v
				}
				_ = writeSettingList(r.Context(), pool, "SUPPORT_TICKETS", tickets)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true, "ticket": tickets[index]})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "ticket_not_found"})
	}
}

func supportUnreadHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireSession(w, r, settings, pool, false); !ok {
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "unread": 0})
	}
}

func adminSupportStatsHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, settings, pool, false); !ok {
			return
		}
		tickets := readSettingList(r.Context(), pool, "SUPPORT_TICKETS")
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "counts": supportCounts(tickets)})
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
				defer file.Close()
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
	counts := map[string]int{"active": 0, "closed": 0, "awaiting_admin": 0, "awaiting_user": 0, "open": 0, "total": len(tickets)}
	for _, ticket := range tickets {
		status := strings.ToLower(fmt.Sprint(ticket["status"]))
		if status == "" {
			status = "open"
		}
		counts[status]++
		if status != "closed" {
			counts["active"]++
		}
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
	err = pool.QueryRow(ctx, `
SELECT user_id, COALESCE(telegram_id,0), COALESCE(username,''), COALESCE(email,''), COALESCE(first_name,''), COALESCE(last_name,''),
	COALESCE(language_code,''), is_banned, registration_date
FROM users WHERE user_id=$1`, userID).Scan(&id, &telegramID, &username, &email, &firstName, &lastName, &language, &banned, &created)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return userAdminPayload(id, telegramID, username, email, firstName, lastName, language, banned, created), nil
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
