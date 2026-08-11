package httpapi

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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
	if provider != "ezpay" && provider != "bepusdt" {
		s.writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Payment return route not found.")
		return
	}
	orderID := r.URL.Query().Get("order")
	if pathOrderID := chiURLParam(r, "orderID"); pathOrderID != "" {
		orderID = pathOrderID
	}
	target := *s.deps.PublicURL
	target.Path = strings.TrimRight(target.Path, "/") + "/payment-result"
	target.RawQuery = url.Values{"paymentOrder": {orderID}, "provider": {provider}}.Encode()
	http.Redirect(w, r, target.String(), http.StatusSeeOther)
}
