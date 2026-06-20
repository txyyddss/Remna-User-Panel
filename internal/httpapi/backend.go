// Package httpapi contains HTTP route wiring.
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/payments"
	"remna-user-panel/internal/remnawave"
	appsettings "remna-user-panel/internal/settings"
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
		if handled, err := handleTelegramPaymentUpdate(r.Context(), settings, pool, body); handled {
			if err != nil {
				slog.Warn("telegram payment update rejected", "error", err)
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
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

func handleTelegramPaymentUpdate(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, body []byte) (bool, error) {
	if pool == nil {
		return false, nil
	}
	var update struct {
		PreCheckout *struct {
			ID   string `json:"id"`
			From struct {
				ID int64 `json:"id"`
			} `json:"from"`
			Currency       string `json:"currency"`
			TotalAmount    int    `json:"total_amount"`
			InvoicePayload string `json:"invoice_payload"`
		} `json:"pre_checkout_query"`
		Message *struct {
			From struct {
				ID int64 `json:"id"`
			} `json:"from"`
			Chat struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			Text              string `json:"text"`
			SuccessfulPayment *struct {
				Currency         string `json:"currency"`
				TotalAmount      int    `json:"total_amount"`
				InvoicePayload   string `json:"invoice_payload"`
				TelegramChargeID string `json:"telegram_payment_charge_id"`
			} `json:"successful_payment"`
		} `json:"message"`
	}
	if json.Unmarshal(body, &update) != nil {
		return false, nil
	}
	if update.PreCheckout != nil {
		query := update.PreCheckout
		valid, reason := validateStarsOrder(ctx, pool, query.InvoicePayload, query.From.ID, query.Currency, query.TotalAmount)
		payload := map[string]any{"pre_checkout_query_id": query.ID, "ok": valid}
		if !valid {
			payload["error_message"] = "This order is unavailable or no longer valid."
		}
		err := callTelegramBotAPI(ctx, settings, "answerPreCheckoutQuery", payload)
		if err != nil {
			return true, err
		}
		if !valid {
			return true, fmt.Errorf("precheckout_rejected: %s", reason)
		}
		return true, nil
	}
	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		payment := update.Message.SuccessfulPayment
		valid, reason := validateStarsOrder(ctx, pool, payment.InvoicePayload, update.Message.From.ID, payment.Currency, payment.TotalAmount)
		if !valid {
			return true, fmt.Errorf("successful_payment_rejected: %s", reason)
		}
		command, err := pool.Exec(ctx, `UPDATE payment_orders SET status='paid',paid_at=COALESCE(paid_at,NOW()),updated_at=NOW(),telegram_payment_charge_id=$2,
 raw_webhook=$3 WHERE order_id=$1 AND status='pending' AND (telegram_payment_charge_id IS NULL OR telegram_payment_charge_id=$2)`, payment.InvoicePayload, payment.TelegramChargeID, body)
		if err != nil {
			return true, err
		}
		if command.RowsAffected() > 0 {
			if _, err := pool.Exec(ctx, `UPDATE invite_visits SET converted_at=NOW() WHERE registered_user_id=(SELECT user_id FROM users WHERE telegram_id=$1) AND converted_at IS NULL`, update.Message.From.ID); err != nil {
				slog.Error("failed to update invite_visits", "error", err)
			}
			recordMessageLog(ctx, pool, messageLogEntry{UserID: update.Message.From.ID, EventType: "telegram_stars_paid", Content: payment.InvoicePayload, Payload: map[string]any{"amount": payment.TotalAmount, "currency": payment.Currency}})
		}
		return true, nil
	}
	if update.Message != nil && strings.HasPrefix(strings.TrimSpace(update.Message.Text), "/paysupport") {
		text := "For payment support, contact the service administrator."
		link := appsettings.NewStore(pool).String(ctx, "SUPPORT_LINK", "")
		if link != "" {
			text += "\n" + link
		}
		err := sendTelegramText(ctx, settings, update.Message.Chat.ID, text)
		return true, err
	}
	return false, nil
}

func validateStarsOrder(ctx context.Context, pool *pgxpool.Pool, orderID string, telegramID int64, currency string, amount int) (bool, string) {
	var expected float64
	var status string
	var storedTelegram int64
	err := pool.QueryRow(ctx, `SELECT p.amount::float8,p.status,COALESCE(u.telegram_id,0) FROM payment_orders p JOIN users u ON u.user_id=p.user_id
 WHERE p.order_id=$1 AND p.provider='telegram_stars'`, orderID).Scan(&expected, &status, &storedTelegram)
	if err != nil {
		return false, "order_not_found"
	}
	if status != "pending" {
		return false, "order_not_pending"
	}
	if storedTelegram != telegramID {
		return false, "user_mismatch"
	}
	if currency != "XTR" || int(expected) != amount {
		return false, "amount_mismatch"
	}
	return true, ""
}

func callTelegramBotAPI(ctx context.Context, settings config.Settings, method string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+strings.TrimSpace(settings.BotToken)+"/"+method, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := telegramHTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode/100 != 2 {
		return fmt.Errorf("telegram_status_%d", response.StatusCode)
	}
	return nil
}

// RegisterTelegramWebhook sets the Telegram Bot webhook URL so that payment
// callbacks (Stars, etc.) are delivered to this backend. It is safe to call
// repeatedly — Telegram deduplicates identical URLs.
func RegisterTelegramWebhook(ctx context.Context, settings config.Settings) error {
	webhookURL := settings.WebhookURL()
	if webhookURL == "" || strings.TrimSpace(settings.BotToken) == "" {
		return nil
	}
	payload := map[string]any{"url": webhookURL}
	if settings.WebhookSecretToken != "" {
		payload["secret_token"] = settings.WebhookSecretToken
	}
	if err := callTelegramBotAPI(ctx, settings, "setWebhook", payload); err != nil {
		return fmt.Errorf("setWebhook %s: %w", webhookURL, err)
	}
	slog.Info("telegram webhook registered", "url", webhookURL)
	return nil
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
		if _, err := pool.Exec(r.Context(), `UPDATE invite_visits v SET converted_at=NOW() FROM payment_orders p WHERE v.registered_user_id=p.user_id AND p.status IN ('paid','succeeded') AND v.converted_at IS NULL`); err != nil {
			slog.Error("failed to batch update invite_visits", "error", err)
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
