// Package httpapi contains HTTP route wiring.
package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/payments"
)

// BackendRouter builds the webhook/health HTTP router.
func BackendRouter(settings config.Settings, pool *pgxpool.Pool, redisClient *redis.Client, registry *payments.Registry) http.Handler {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	router.Use(requestBodyLimit(2 << 20))
	router.Get("/healthz", healthHandler(pool, redisClient))
	router.Get("/health", healthHandler(pool, redisClient))
	router.Post(settings.WebhookPath(), telegramWebhookHandler(settings))
	router.Post(settings.PanelWebhookPath, panelWebhookHandler(settings))
	for _, providerID := range registry.IDs() {
		providerID := providerID
		router.Post("/webhook/"+providerID, paymentWebhookHandler(registry, providerID))
	}
	return router
}

func healthHandler(pool *pgxpool.Pool, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{"status": "ok"}
		if pool != nil {
			stat := pool.Stat()
			payload["db_pool"] = map[string]any{
				"acquired":     stat.AcquiredConns(),
				"idle":         stat.IdleConns(),
				"total":        stat.TotalConns(),
				"max":          stat.MaxConns(),
				"constructing": stat.ConstructingConns(),
			}
		}
		if redisClient != nil {
			payload["redis"] = "configured"
		}
		writeJSON(w, http.StatusOK, payload)
	}
}

func telegramWebhookHandler(settings config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if settings.WebhookSecretToken != "" {
			got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if got != settings.WebhookSecretToken {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_secret"})
				return
			}
		}
		_, _ = io.Copy(io.Discard, r.Body)
		slog.Warn("telegram webhook accepted but bot update processing is not fully ported")
		writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "accepted"})
	}
}

func panelWebhookHandler(settings config.Settings) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if settings.PanelWebhookSecret != "" {
			got := r.Header.Get("X-Remnawave-Webhook-Secret")
			if got == "" {
				got = r.Header.Get("X-Webhook-Secret")
			}
			if got != settings.PanelWebhookSecret {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_secret"})
				return
			}
		}
		_, _ = io.Copy(io.Discard, r.Body)
		slog.Warn("panel webhook accepted but panel event processing is not fully ported")
		writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "accepted"})
	}
}

func paymentWebhookHandler(registry *payments.Registry, providerID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_body"})
			return
		}
		if err := registry.HandleWebhook(r.Context(), providerID, body); err != nil {
			slog.Warn("payment webhook not processed", "provider", providerID, "error", err)
			writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "accepted_not_processed"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Warn("failed to write json response", "error", err)
	}
}
