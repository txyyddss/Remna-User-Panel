package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/telegram"
)

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
		expiresAt := time.Unix(join.InviteLink.ExpireDate, 0).UTC()
		if err := s.deps.Accounts.HandleSignedJoinRequest(r.Context(), join.From.ID, join.Chat.ID, join.InviteLink.InviteLink, join.InviteLink.Name, expiresAt); err != nil {
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
	if update.Message != nil {
		s.processTelegramGroupMessage(r.Context(), update.Message)
	}
	w.WriteHeader(http.StatusNoContent)
}
