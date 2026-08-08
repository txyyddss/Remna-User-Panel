package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
)

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 2*time.Second)
	defer cancel()
	if err := s.deps.Store.DB().PingContext(ctx); err != nil {
		s.writeError(w, r, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "Database is unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := contextWithTimeout(r, 3*time.Second)
	defer cancel()
	if err := s.deps.Store.DB().PingContext(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "missing": []string{"database"}})
		return
	}
	combos, err := s.deps.Store.ListCombos(ctx, true)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "missing": []string{"catalog"}})
		return
	}
	issues := s.deps.Settings.Readiness(ctx, len(combos))
	if len(issues) > 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "setup_required", "missing": issues})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) telegramWebhook(w http.ResponseWriter, r *http.Request) {
	secret, err := s.deps.Settings.Plaintext(r.Context(), "telegram.webhook_secret")
	if err != nil {
		s.writeError(w, r, http.StatusServiceUnavailable, "WEBHOOK_NOT_CONFIGURED", "Webhook setup is incomplete.")
		return
	}
	if !telegram.VerifyWebhookSecret(r.Header.Get(telegram.WebhookSecretHeader), secret) {
		s.writeError(w, r, http.StatusUnauthorized, "INVALID_WEBHOOK_SECRET", "Webhook authentication failed.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_TELEGRAM_UPDATE", "Telegram update is invalid.")
		return
	}
	var update telegram.Update
	if err := json.Unmarshal(body, &update); err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_TELEGRAM_UPDATE", "Telegram update is invalid.")
		return
	}
	if update.ChatJoinRequest != nil && update.ChatJoinRequest.InviteLink != nil {
		join := update.ChatJoinRequest
		if err := s.deps.Accounts.HandleJoinRequest(r.Context(), join.From.ID, join.Chat.ID, join.InviteLink.InviteLink); err != nil {
			s.deps.Logger.Warn("Telegram join request was not approved", "request_id", middlewareRequestID(r), "chat_id", join.Chat.ID, "telegram_id", join.From.ID, "error", err)
		}
	}
	if update.ChatMember != nil {
		member := update.ChatMember.NewChatMember.User
		if _, err := s.deps.Accounts.RefreshMembershipByTelegramID(r.Context(), member.ID); err != nil {
			s.deps.Logger.Warn("Telegram membership refresh failed", "request_id", middlewareRequestID(r), "chat_id", update.ChatMember.Chat.ID, "telegram_id", member.ID, "error", err)
		}
	}
	if update.PreCheckoutQuery != nil {
		query := update.PreCheckoutQuery
		telegramID := query.From.ID
		event := billing.ProviderEvent{Provider: "stars", OrderID: query.InvoicePayload, PayableAmount: strconv.FormatInt(query.TotalAmount, 10),
			PayableCurrency: query.Currency, TelegramID: &telegramID}
		approved := true
		errorMessage := ""
		if _, err := s.deps.Billing.AuthorizeEvent(r.Context(), event); err != nil {
			approved = false
			errorMessage = "This payment order is no longer valid."
		}
		answerCtx, cancel := contextWithTimeout(r, 8*time.Second)
		err := s.deps.Telegram.AnswerPreCheckoutQuery(answerCtx, query.ID, approved, errorMessage)
		cancel()
		if err != nil {
			s.writeError(w, r, http.StatusBadGateway, "PRECHECKOUT_REPLY_FAILED", "Telegram pre-checkout could not be answered.")
			return
		}
	}
	if update.Message != nil && update.Message.SuccessfulPayment != nil {
		message := update.Message
		if message.From == nil {
			s.writeError(w, r, http.StatusBadRequest, "PAYMENT_USER_MISSING", "Telegram payment has no user identity.")
			return
		}
		payment := message.SuccessfulPayment
		telegramID := message.From.ID
		event := billing.ProviderEvent{Provider: "stars", OrderID: payment.InvoicePayload, TradeID: payment.TelegramPaymentChargeID,
			ChargeID: payment.ProviderPaymentChargeID, PayableAmount: strconv.FormatInt(payment.TotalAmount, 10), PayableCurrency: payment.Currency,
			DedupeKey: payment.TelegramPaymentChargeID, TelegramID: &telegramID}
		if _, _, err := s.deps.Billing.Settle(r.Context(), event); err != nil {
			s.writeError(w, r, http.StatusConflict, "PAYMENT_SETTLEMENT_FAILED", "Telegram payment did not match the stored order.")
			return
		}
	}
	if update.Message != nil && update.Message.RefundedPayment != nil {
		message := update.Message
		payment := message.RefundedPayment
		event := billing.ProviderEvent{Provider: "stars", OrderID: payment.InvoicePayload, TradeID: payment.TelegramPaymentChargeID,
			PayableAmount: strconv.FormatInt(payment.TotalAmount, 10), PayableCurrency: payment.Currency}
		if _, err := s.deps.Billing.ValidateEvent(r.Context(), event); err != nil {
			s.writeError(w, r, http.StatusConflict, "REFUND_MISMATCH", "Telegram refund did not match the stored order.")
			return
		}
		if _, err := s.deps.Store.RefundPayment(r.Context(), nil, payment.InvoicePayload, "Telegram Stars refund", time.Now().UTC()); err != nil {
			s.writeError(w, r, http.StatusInternalServerError, "REFUND_FAILED", "Telegram refund could not be applied.")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) ezpayWebhook(w http.ResponseWriter, r *http.Request) {
	event, paid, err := s.deps.Webhooks.VerifyEZPay(r.Context(), r.URL.Query())
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("fail"))
		return
	}
	if paid {
		if _, _, err := s.deps.Billing.Settle(r.Context(), event); err != nil {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte("fail"))
			return
		}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("success"))
}

func (s *Server) bepusdtWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_WEBHOOK", "Webhook body is invalid.")
		return
	}
	event, status, err := s.deps.Webhooks.VerifyBEPusdt(r.Context(), body)
	if err != nil {
		capability := chiURLParam(r, "capability")
		unsigned, ok := s.deps.Webhooks.(bepusdtUnsignedVerifier)
		if !ok || capability == "" {
			s.writeError(w, r, http.StatusUnauthorized, "INVALID_SIGNATURE", "Webhook authentication failed.")
			return
		}
		event, status, err = unsigned.VerifyBEPusdtUnsigned(r.Context(), body)
		if err != nil || !s.deps.Billing.VerifyBEPusdtCallbackCapability(r.Context(), event.OrderID, capability) {
			s.writeError(w, r, http.StatusUnauthorized, "INVALID_CAPABILITY", "Webhook authentication failed.")
			return
		}
	}
	if status < 1 || status > 3 {
		s.writeError(w, r, http.StatusBadRequest, "INVALID_PAYMENT_STATUS", "Payment status is invalid.")
		return
	}
	switch status {
	case 1:
		if _, err := s.deps.Billing.ValidateEvent(r.Context(), event); err != nil {
			s.writeError(w, r, http.StatusConflict, "PAYMENT_STATUS_FAILED", "Pending payment status did not match the stored order.")
			return
		}
	case 2:
		if _, _, err := s.deps.Billing.Settle(r.Context(), event); err != nil {
			s.writeError(w, r, http.StatusConflict, "PAYMENT_SETTLEMENT_FAILED", "Payment did not match the stored order.")
			return
		}
	case 3:
		if _, err := s.deps.Billing.ValidateEvent(r.Context(), event); err != nil {
			s.writeError(w, r, http.StatusConflict, "PAYMENT_EXPIRY_FAILED", "Payment timeout did not match the stored order.")
			return
		}
		if err := s.deps.Store.ExpirePaymentOrder(r.Context(), event.OrderID, "bepusdt", time.Now().UTC()); err != nil {
			s.writeError(w, r, http.StatusConflict, "PAYMENT_EXPIRY_FAILED", "Payment timeout did not match the stored order.")
			return
		}
	}
	ack, _ := s.deps.Settings.Optional(r.Context(), "billing.bepusdt.ack")
	if ack == "" {
		ack = "ok"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(ack))
}

// paymentReturn is deliberately navigation-only and never changes payment or balance state.
func (s *Server) paymentReturn(w http.ResponseWriter, r *http.Request) {
	provider := chiURLParam(r, "provider")
	orderID := r.URL.Query().Get("order")
	if pathOrderID := chiURLParam(r, "orderID"); pathOrderID != "" {
		orderID = pathOrderID
	}
	target := *s.deps.PublicURL
	target.Path = strings.TrimRight(target.Path, "/") + "/balance"
	target.RawQuery = url.Values{"paymentOrder": {orderID}, "provider": {provider}}.Encode()
	http.Redirect(w, r, target.String(), http.StatusSeeOther)
}

func contextWithTimeout(r *http.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), timeout)
}

func middlewareRequestID(r *http.Request) string {
	return middleware.GetReqID(r.Context())
}
