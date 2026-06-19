package httpapi

import (
	"encoding/json"
	"html"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/i18n"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/webassets"
)

// WebAppRouter builds the Mini App and admin HTTP router.
func WebAppRouter(settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry) http.Handler {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	router.Use(requestBodyLimit(8 << 20))
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})
	router.Get("/api/bootstrap", bootstrapHandler(settings, catalog))
	router.Get("/api/i18n", i18nHandler(settings, catalog))
	router.Get("/api/me", meHandler(settings, pool, registry))
	router.Post("/api/auth/token", authTokenHandler(settings, pool))
	router.Post("/api/auth/logout", logoutHandler())
	router.Post("/api/payments", createPaymentHandler(settings, pool, registry))
	router.Get("/api/payments/{payment_id}", paymentStatusHandler(settings, pool, registry))
	router.Get("/api/admin/settings", adminSettingsHandler(settings, pool))
	router.Patch("/api/admin/settings", adminSettingsHandler(settings, pool))
	router.Get("/api/admin/payments", adminPaymentsListHandler(settings, pool, registry))
	registerExtraAPIRoutes(router, settings, pool, catalog, assets, registry)
	router.Get("/api/admin/payments/{payment_id}", adminPaymentDetailHandler(settings, pool, registry))
	router.Get("/api/*", notImplementedAPI("unknown_api"))
	router.Post("/api/*", notImplementedAPI("unknown_api"))
	router.Put("/api/*", notImplementedAPI("unknown_api"))
	router.Patch("/api/*", notImplementedAPI("unknown_api"))
	router.Delete("/api/*", notImplementedAPI("unknown_api"))
	registerAssetRoutes(router, assets)
	registerIndexRoutes(router, settings, catalog, assets)
	return router
}

func bootstrapHandler(settings config.Settings, catalog *i18n.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := r.URL.Query().Get("i18n_scope")
		_ = scope
		writeJSON(w, http.StatusOK, map[string]any{
			"config": map[string]any{
				"title":        "Subscription",
				"primaryColor": "#00fe7a",
				"apiBase":      "/api",
				"language":     settings.DefaultLanguage,
				"languages":    languageOptions(catalog),
			},
			"i18n": localePayload(catalog),
		})
	}
}

func i18nHandler(settings config.Settings, catalog *i18n.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"default_language": settings.DefaultLanguage,
			"languages":        languageOptions(catalog),
			"messages":         localePayload(catalog),
		})
	}
}

func okAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func notImplementedAPI(code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusNotImplemented
		if code == "auth_required" {
			status = http.StatusUnauthorized
		}
		writeJSON(w, status, map[string]any{"ok": false, "error": code})
	}
}

func registerAssetRoutes(router chi.Router, assets webassets.Paths) {
	fileServer := http.FileServer(http.Dir(assets.TemplatesDir))
	for _, pattern := range []string{
		"/subscription_webapp.css",
		"/subscription_webapp.js",
		"/subscription_webapp_admin.css",
		"/subscription_webapp_admin.js",
		"/subscription_webapp.{asset_hash}.css",
		"/subscription_webapp.min.{asset_hash}.js",
		"/subscription_webapp_admin.{asset_hash}.css",
		"/subscription_webapp_admin.min.{asset_hash}.js",
		"/favicon.ico",
		"/apple-touch-icon.png",
		"/apple-touch-icon-precomposed.png",
		"/icon-192.png",
		"/icon-512.png",
	} {
		router.Get(pattern, func(w http.ResponseWriter, r *http.Request) {
			name := filepath.Base(r.URL.Path)
			if strings.HasPrefix(name, "apple-touch") || strings.HasPrefix(name, "icon-") || name == "favicon.ico" {
				http.ServeFile(w, r, filepath.Join(assets.TemplatesDir, "default-brand", "favicons", "19b2a242e5b7bc2d", name))
				return
			}
			fileServer.ServeHTTP(w, r)
		})
	}
	router.Get("/webapp-theme-css/*", func(w http.ResponseWriter, r *http.Request) {
		serveThemeFile(w, r, assets.ThemesDir, "/webapp-theme-css/")
	})
	router.Get("/webapp-theme-assets/*", func(w http.ResponseWriter, r *http.Request) {
		serveThemeFile(w, r, assets.ThemesDir, "/webapp-theme-assets/")
	})
}

func serveThemeFile(w http.ResponseWriter, r *http.Request, root string, prefix string) {
	rel := strings.TrimPrefix(r.URL.Path, prefix)
	rel = strings.TrimPrefix(filepath.Clean("/"+rel), string(filepath.Separator))
	if rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
		http.NotFound(w, r)
		return
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fullAbs, err := filepath.Abs(filepath.Join(rootAbs, rel))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if fullAbs != rootAbs && !strings.HasPrefix(fullAbs, rootAbs+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, fullAbs)
}

func registerIndexRoutes(router chi.Router, settings config.Settings, catalog *i18n.Catalog, assets webassets.Paths) {
	index := indexHandler(settings, catalog, assets)
	for _, route := range []string{
		"/", "/login/password", "/home", "/install", "/trial", "/invite", "/devices",
		"/settings", "/support", "/admin", "/open-app",
	} {
		router.Get(route, index)
	}
	router.Get("/s/{share_token}", index)
	router.Get("/support/{ticket_id}", index)
	router.Get("/admin/*", index)
}

func indexHandler(settings config.Settings, catalog *i18n.Catalog, assets webassets.Paths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := os.ReadFile(filepath.Join(assets.TemplatesDir, "subscription_webapp.html"))
		if err != nil {
			slog.Error("failed to read webapp template", "error", err)
			http.Error(w, "template not found", http.StatusInternalServerError)
			return
		}
		configScript := scriptJSON("webapp-config", map[string]any{
			"title":        "Subscription",
			"primaryColor": "#00fe7a",
			"apiBase":      "/api",
			"language":     settings.DefaultLanguage,
			"languages":    languageOptions(catalog),
		})
		i18nScript := scriptJSON("i18n", localePayload(catalog))
		htmlBody := string(body)
		htmlBody = strings.ReplaceAll(htmlBody, "<!-- WEBAPP_CONFIG_SCRIPT -->", configScript)
		htmlBody = strings.ReplaceAll(htmlBody, "<!-- WEBAPP_I18N_SCRIPT -->", i18nScript)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(htmlBody))
	}
}

func scriptJSON(id string, payload any) string {
	body, err := json.Marshal(payload)
	if err != nil {
		body = []byte("{}")
	}
	return `<script id="` + html.EscapeString(id) + `" type="application/json">` + string(body) + `</script>`
}

func localePayload(catalog *i18n.Catalog) map[string]map[string]string {
	payload := map[string]map[string]string{}
	for _, lang := range catalog.Languages() {
		payload[lang] = map[string]string{}
		for _, key := range []string{
			"wa_language_default", "wa_auth_telegram_cancelled", "wa_auth_telegram_not_confirmed",
			"wa_back", "wa_close", "wa_apply", "wa_copied",
		} {
			payload[lang][key] = catalog.Translate(lang, key)
		}
	}
	return payload
}

func languageOptions(catalog *i18n.Catalog) []map[string]string {
	labels := map[string]string{"zh": "中文", "en": "English"}
	flags := map[string]string{"zh": "🇨🇳", "en": "🇬🇧"}
	ordered := []string{"zh", "en"}
	seen := map[string]bool{}
	options := make([]map[string]string, 0, len(ordered))
	for _, lang := range ordered {
		seen[lang] = true
		options = append(options, map[string]string{"code": lang, "label": labels[lang], "flag": flags[lang]})
	}
	for _, lang := range catalog.Languages() {
		if seen[lang] {
			continue
		}
		if lang == "zh" || lang == "en" {
			options = append(options, map[string]string{"code": lang, "label": labels[lang], "flag": flags[lang]})
		}
	}
	return options
}
