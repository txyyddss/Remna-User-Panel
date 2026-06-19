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
	"remna-user-panel/internal/remnawave"
)

// BackendRouter builds the webhook/health HTTP router.
func BackendRouter(settings config.Settings, pool *pgxpool.Pool, redisClient *redis.Client, registry *payments.Registry, panel *remnawave.Client) http.Handler {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	router.Use(requestBodyLimit(2 << 20))
	RegisterBackendRoutes(router, settings, pool, redisClient, registry, panel)
	return router
}

// RegisterBackendRoutes adds webhook and health routes (without middleware) to an existing router.
func RegisterBackendRoutes(router chi.Router, settings config.Settings, pool *pgxpool.Pool, redisClient *redis.Client, registry *payments.Registry, panel *remnawave.Client) {
	router.Get("/healthz", healthHandler(pool, redisClient))
	router.Get("/health", healthHandler(pool, redisClient))
	router.Post(settings.WebhookPath(), telegramWebhookHandler(settings, pool))
	router.Post(settings.PanelWebhookPath, panelWebhookHandler(settings, pool))
	for _, providerID := range registry.IDs() {
		providerID := providerID
		router.Post("/webhook/"+providerID, paymentWebhookHandler(settings, pool, registry, panel, providerID))
	}
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

func telegramWebhookHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if settings.WebhookSecretToken != "" {
			got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if got != settings.WebhookSecretToken {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"ok": false, "error": "invalid_secret"})
				return
			}
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_body"})
			return
		}
		if pool != nil {
			payload := json.RawMessage(body)
			if !json.Valid(payload) {
				payload, _ = json.Marshal(map[string]any{"raw": string(body)})
			}
			if _, err := pool.Exec(r.Context(), "INSERT INTO webhook_events (provider, payload, status) VALUES ($1,$2,'queued')", "telegram", payload); err != nil {
				slog.Warn("telegram webhook accepted but queue insert failed", "error", err)
				recordMessageLog(r.Context(), pool, messageLogEntry{EventType: "telegram_webhook_queue_failed", Content: err.Error(), RawUpdatePreview: string(body)})
				writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "accepted_not_queued"})
				return
			}
			recordMessageLog(r.Context(), pool, messageLogEntry{EventType: "telegram_webhook_queued", Content: "queued", RawUpdatePreview: string(body)})
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "queued"})
	}
}

func panelWebhookHandler(settings config.Settings, pool *pgxpool.Pool) http.HandlerFunc {
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
		body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid_body"})
			return
		}
		if pool != nil {
			payload := json.RawMessage(body)
			if !json.Valid(payload) {
				payload, _ = json.Marshal(map[string]any{"raw": string(body)})
			}
			if _, err := pool.Exec(r.Context(), "INSERT INTO webhook_events (provider, payload, status) VALUES ($1,$2,'queued')", "remnawave", payload); err != nil {
				slog.Warn("panel webhook accepted but queue insert failed", "error", err)
				recordMessageLog(r.Context(), pool, messageLogEntry{EventType: "remnawave_webhook_queue_failed", Content: err.Error(), Payload: map[string]any{"provider": "remnawave"}})
				writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "accepted_not_queued"})
				return
			}
			recordMessageLog(r.Context(), pool, messageLogEntry{EventType: "remnawave_webhook_queued", Content: "queued", Payload: map[string]any{"provider": "remnawave"}})
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "queued"})
	}
}

func paymentWebhookHandler(settings config.Settings, pool *pgxpool.Pool, registry *payments.Registry, panel *remnawave.Client, providerID string) http.HandlerFunc {
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
		provision, err := ProvisionPendingPaidOrders(r.Context(), settings, pool, panel, 10)
		if err != nil {
			slog.Warn("payment webhook accepted but provisioning is pending", "provider", providerID, "error", err, "scanned", provision.Scanned, "provisioned", provision.Provisioned, "failed", provision.Failed)
		}
		recordMessageLog(r.Context(), pool, messageLogEntry{
			EventType: "payment_webhook",
			Content:   providerID,
			Payload: map[string]any{
				"provider":     providerID,
				"provisioning": provision,
				"error":        errorString(err),
			},
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":           true,
			"provisioning": provision,
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Warn("failed to write json response", "error", err)
	}
}
