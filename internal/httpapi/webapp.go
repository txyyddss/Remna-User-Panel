package httpapi

import (
	"context"
	"encoding/json"
	"html"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/i18n"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
	"remna-user-panel/internal/webassets"
)

// WebAppRouter builds the Mini App and admin HTTP router.
func WebAppRouter(settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry, panel *remnawave.Client) http.Handler {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	router.Use(requestBodyLimit(8 << 20))
	RegisterWebAppRoutes(router, settings, pool, catalog, assets, registry, panel)
	return router
}

// RegisterWebAppRoutes adds Mini App, admin, and asset routes (without middleware) to an existing router.
// Note: /health is registered by RegisterBackendRoutes; this function does not
// register a duplicate /health endpoint to avoid route conflicts when both
// route sets are combined in CombinedRouter.
func RegisterWebAppRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry, panel *remnawave.Client) {
	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})
	router.Get("/api/bootstrap", bootstrapHandler(settings, pool, catalog, assets))
	router.Get("/api/i18n", i18nHandler(settings, pool, catalog))
	router.Get("/api/me", meHandler(settings, pool, registry, panel))
	router.Post("/api/auth/token", authTokenHandler(settings, pool))
	router.Get("/api/auth/telegram/nonce", telegramLoginNonceHandler(settings, pool))
	router.Post("/api/auth/logout", logoutHandler())
	router.Post("/api/payments", createPaymentHandler(settings, pool, registry))
	router.Get("/api/payments/{payment_id}", paymentStatusHandler(settings, pool, registry, panel))
	router.Get("/api/admin/settings", adminSettingsHandler(settings, pool))
	router.Patch("/api/admin/settings", adminSettingsHandler(settings, pool))
	router.Get("/api/admin/payments", adminPaymentsListHandler(settings, pool, registry))
	registerExtraAPIRoutes(router, settings, pool, catalog, assets, registry, panel)
	router.Get("/api/admin/payments/{payment_id}", adminPaymentDetailHandler(settings, pool, registry))
	router.Get("/webapp-uploaded-logo/{filename}", serveUploadedLogo)
	router.Get("/webapp-logo", serveWebAppLogo)
	router.Get("/api/*", unknownAPIHandler())
	router.Post("/api/*", unknownAPIHandler())
	router.Put("/api/*", unknownAPIHandler())
	router.Patch("/api/*", unknownAPIHandler())
	router.Delete("/api/*", unknownAPIHandler())
	registerAssetRoutes(router, assets)
	registerIndexRoutes(router, settings, pool, catalog, assets)
}

// CombinedRouter builds a single router that serves both backend webhooks and
// the Mini App on one port. Webhook routes use a 2 MiB body limit while Mini
// App routes allow up to 8 MiB.
func CombinedRouter(settings config.Settings, pool *pgxpool.Pool, redisClient *redis.Client, catalog *i18n.Catalog, assets webassets.Paths, registry *payments.Registry, panel *remnawave.Client) http.Handler {
	router := chi.NewRouter()
	router.Use(securityHeaders)

	// Webhook routes with smaller body limit.
	router.Group(func(r chi.Router) {
		r.Use(requestBodyLimit(2 << 20))
		RegisterBackendRoutes(r, settings, pool, redisClient, registry, panel)
	})

	// Mini App and admin routes with larger body limit.
	router.Group(func(r chi.Router) {
		r.Use(requestBodyLimit(8 << 20))
		RegisterWebAppRoutes(r, settings, pool, catalog, assets, registry, panel)
	})

	return router
}

func bootstrapHandler(settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := r.URL.Query().Get("i18n_scope")
		_ = scope
		i18nPayload := localePayload(r.Context(), pool, catalog)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"config":   webappRuntimeConfig(r.Context(), settings, pool, catalog, assets),
			"i18n":     i18nPayload,
			"messages": i18nPayload,
		})
	}
}

func webappRuntimeConfig(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths) map[string]any {
	store := appsettings.NewStore(pool)
	logoURL := store.String(ctx, "WEBAPP_LOGO_URL", "")
	faviconURL := store.String(ctx, "WEBAPP_FAVICON_URL", "")
	useCustomFavicon := store.Bool(ctx, "WEBAPP_FAVICON_USE_CUSTOM", faviconURL != "")
	if !useCustomFavicon {
		faviconURL = logoURL
	}
	emailAuthEnabled := settings.AdminEmail != "" && settings.AdminPassword != ""
	if store.Bool(ctx, "SMTP_ENABLED", false) && mailerConfigFromSettings(ctx, store).IsConfigured() {
		emailAuthEnabled = true
	}
	return map[string]any{
		"title": store.String(ctx, "WEBAPP_TITLE", "Subscription"), "primaryColor": store.String(ctx, "WEBAPP_PRIMARY_COLOR", "#00fe7a"),
		"apiBase": "/api", "language": effectiveDefaultLanguage(ctx, pool, settings), "languages": languageOptions(catalog),
		"emailAuthEnabled": emailAuthEnabled, "logoUrl": logoURL,
		"telegramLoginClientId": store.String(ctx, "TELEGRAM_LOGIN_CLIENT_ID", os.Getenv("TELEGRAM_LOGIN_CLIENT_ID")),
		"faviconUrl": faviconURL, "faviconUseCustom": useCustomFavicon,
		"supportUrl": store.String(ctx, "SUPPORT_LINK", ""), "serverStatusUrl": store.String(ctx, "SERVER_STATUS_URL", ""),
		"privacyPolicyUrl": store.String(ctx, "PRIVACY_POLICY_URL", ""), "userAgreementUrl": store.String(ctx, "USER_AGREEMENT_URL", ""),
		"themesCatalog": readThemeCatalog(ctx, store, assets.ThemesDir),
	}
}

func i18nHandler(settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		i18nPayload := localePayload(r.Context(), pool, catalog)
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":               true,
			"default_language": settings.DefaultLanguage,
			"languages":        languageOptions(catalog),
			"i18n":             i18nPayload,
			"messages":         i18nPayload,
		})
	}
}

func unknownAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "unknown_api"})
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

func serveUploadedLogo(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	if filename == "." || filename == "" || strings.Contains(filename, "..") {
		http.NotFound(w, r)
		return
	}
	// Check both logos and favicons directories
	for _, subdir := range []string{"logos", "favicons"} {
		dir := filepath.Join("data", "uploads", subdir)
		filePath := filepath.Join(dir, filename)
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
	}
	http.NotFound(w, r)
}

func serveWebAppLogo(w http.ResponseWriter, r *http.Request) {
	// Serve the most recent uploaded logo
	defaultPath := filepath.Join("data", "uploads", "logos")
	if entries, err := os.ReadDir(defaultPath); err == nil && len(entries) > 0 {
		var latest string
		var latestTime time.Time
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latest = entry.Name()
			}
		}
		if latest != "" {
			http.ServeFile(w, r, filepath.Join(defaultPath, latest))
			return
		}
	}
	http.NotFound(w, r)
}

func registerIndexRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths) {
	index := indexHandler(settings, pool, catalog, assets)
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

func indexHandler(settings config.Settings, pool *pgxpool.Pool, catalog *i18n.Catalog, assets webassets.Paths) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := os.ReadFile(filepath.Join(assets.TemplatesDir, "subscription_webapp.html"))
		if err != nil {
			slog.Error("failed to read webapp template", "error", err)
			http.Error(w, "template not found", http.StatusInternalServerError)
			return
		}
		configScript := scriptJSON("webapp-config", webappRuntimeConfig(r.Context(), settings, pool, catalog, assets))
		i18nScript := scriptJSON("i18n", localePayload(r.Context(), pool, catalog))
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

func localePayload(ctx context.Context, pool *pgxpool.Pool, catalog *i18n.Catalog) map[string]map[string]string {
	_ = ctx
	_ = pool
	return catalog.Messages()
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
