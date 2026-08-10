package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"strings"
)

// ValidateEvent checks stored provider, amount, identity, and trade ID before settlement.
func (s *Service) ValidateEvent(ctx context.Context, event ProviderEvent) (model.PaymentOrder, error) {
	order, err := s.repository.PaymentOrderByID(ctx, event.OrderID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if order.Provider != event.Provider {
		return model.PaymentOrder{}, database.ErrConflict
	}
	if event.Rail != "" && order.ProviderRail != "" && !strings.EqualFold(order.ProviderRail, event.Rail) {
		return model.PaymentOrder{}, database.ErrConflict
	}
	if event.Provider == "bepusdt" {
		var receivingAddress *string
		if order.ActualCryptoAmount != nil || order.ActualCryptoCurrency != nil {
			if order.ActualCryptoAmount == nil || order.ActualCryptoCurrency == nil ||
				!strings.EqualFold(*order.ActualCryptoCurrency, event.PayableCurrency) || !Equivalent(*order.ActualCryptoAmount, event.PayableAmount) {
				return model.PaymentOrder{}, database.ErrConflict
			}
			receivingAddress = order.ReceivingAddress
		} else {
			// Orders created before the separate crypto fields were introduced
			// stored the provider's USDT amount/currency in payable_* and the
			// receiving address in qr_payload. Preserve settlement for those
			// immutable pending snapshots; new-direction orders may not use this
			// compatibility path.
			if order.RateDirection != "" && order.RateDirection != "currency_per_txb" {
				return model.PaymentOrder{}, database.ErrConflict
			}
			if !strings.EqualFold(order.PayableCurrency, event.PayableCurrency) || !Equivalent(order.PayableAmount, event.PayableAmount) {
				return model.PaymentOrder{}, database.ErrConflict
			}
			receivingAddress = order.ReceivingAddress
			if receivingAddress == nil {
				receivingAddress = order.QRPayload
			}
		}
		rate, parseErr := ParseDecimal(order.RateSnapshot)
		var expectedFiat string
		var payableErr error
		if order.RateDirection == "txb_per_currency" {
			expectedFiat, payableErr = PayableFromTXBPerCurrency(order.TXBMinor, rate, 2)
		} else {
			expectedFiat, payableErr = Payable(order.TXBMinor, rate, 2)
		}
		if parseErr != nil || payableErr != nil || !strings.EqualFold(event.FiatCurrency, "USD") || !Equivalent(expectedFiat, event.FiatAmount) {
			return model.PaymentOrder{}, database.ErrConflict
		}
		// Reference versions overload callback token as either the receive
		// address or the literal currency code. Compare it to the exact checkout
		// address whenever the provider supplies address semantics.
		if event.Recipient != "" && !strings.EqualFold(event.Recipient, "USDT") &&
			(receivingAddress == nil || event.Recipient != *receivingAddress) {
			return model.PaymentOrder{}, database.ErrConflict
		}
	} else if !strings.EqualFold(order.PayableCurrency, event.PayableCurrency) || !Equivalent(order.PayableAmount, event.PayableAmount) {
		return model.PaymentOrder{}, database.ErrConflict
	}
	if order.ProviderTradeID != nil && event.TradeID != "" && *order.ProviderTradeID != event.TradeID {
		return model.PaymentOrder{}, database.ErrConflict
	}
	if event.TelegramID != nil {
		user, err := s.repository.UserByID(ctx, order.UserID)
		if err != nil || user.TelegramID != *event.TelegramID {
			return model.PaymentOrder{}, database.ErrConflict
		}
	}
	return order, nil
}

// AuthorizeEvent applies the stricter state and expiry checks required before a
// provider is allowed to collect funds. Authoritative paid events remain able to
// settle a just-expired order because the charge has already occurred.
func (s *Service) AuthorizeEvent(ctx context.Context, event ProviderEvent) (model.PaymentOrder, error) {
	order, err := s.ValidateEvent(ctx, event)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if order.Status != "pending" || !s.now().UTC().Before(order.ExpiresAt) {
		return model.PaymentOrder{}, database.ErrConflict
	}
	return order, nil
}

// Settle commits a previously cryptographically verified provider event exactly once.
func (s *Service) Settle(ctx context.Context, event ProviderEvent) (model.PaymentOrder, bool, error) {
	if _, err := s.ValidateEvent(ctx, event); err != nil {
		return model.PaymentOrder{}, false, err
	}
	if event.DedupeKey == "" {
		event.DedupeKey = event.TradeID
	}
	if event.DedupeKey == "" {
		return model.PaymentOrder{}, false, errors.New("provider event has no dedupe key")
	}
	if event.PayloadHash == "" {
		digest := sha256.Sum256([]byte(event.Provider + "\x00" + event.DedupeKey + "\x00" + event.OrderID + "\x00" + event.PayableAmount))
		event.PayloadHash = hex.EncodeToString(digest[:])
	}
	return s.repository.SettlePayment(ctx, event.Provider, event.DedupeKey, event.PayloadHash, event.OrderID, event.TradeID, event.ChargeID, s.now().UTC())
}

// OrderForUser returns durable status for payment-sheet polling.
func (s *Service) OrderForUser(ctx context.Context, orderID, userID string) (model.PaymentOrder, error) {
	return s.repository.PaymentOrderForUser(ctx, orderID, userID)
}

// Cancel stops client polling immediately and then performs provider cancellation
// on a best-effort basis. A later verified paid event remains authoritative.
func (s *Service) Cancel(ctx context.Context, orderID, userID string) (model.PaymentOrder, error) {
	repository, ok := s.repository.(cancellationRepository)
	if !ok {
		return model.PaymentOrder{}, errors.New("payment cancellation is unavailable")
	}
	order, changed, err := repository.CancelPaymentOrder(ctx, orderID, userID, "cancelled by user", s.now().UTC())
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if !changed || order.Provider != "bepusdt" || order.ProviderTradeID == nil {
		return order, nil
	}
	canceller, ok := s.gateway.(CancellationGateway)
	if !ok {
		_ = repository.SetPaymentProviderCancellation(ctx, order.ID, "unsupported", s.now().UTC())
		return s.repository.PaymentOrderForUser(ctx, order.ID, userID)
	}
	status := "cancelled"
	if err := canceller.Cancel(ctx, order); err != nil {
		status = "failed"
	}
	_ = repository.SetPaymentProviderCancellation(ctx, order.ID, status, s.now().UTC())
	return s.repository.PaymentOrderForUser(ctx, order.ID, userID)
}

func (s *Service) loadNewRate(ctx context.Context, provider string) (Decimal, error) {
	rateRaw, err := s.settings.Plaintext(ctx, "billing.rate.txb_per_"+currencyCode(provider))
	if err != nil {
		return Decimal{}, fmt.Errorf("%w: %v", errRateNotConfigured, err)
	}
	rate, err := ParseDecimal(rateRaw)
	if err != nil || !rate.Positive() {
		return Decimal{}, errRateNotConfigured
	}
	return rate, nil
}

func callbackCapability(secret, orderID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("bepusdt-callback\x00" + orderID))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyBEPusdtCallbackCapability authenticates unsigned v1.19-style callback
// URLs without exposing the configured API token.
func (s *Service) VerifyBEPusdtCallbackCapability(ctx context.Context, orderID, capability string) bool {
	secret, err := s.settings.Plaintext(ctx, "billing.bepusdt.api_token")
	if err != nil {
		return false
	}
	expected := callbackCapability(secret, orderID)
	return len(capability) == len(expected) && hmac.Equal([]byte(capability), []byte(expected))
}

func currencyCode(provider string) string {
	switch provider {
	case "ezpay":
		return "cny"
	case "bepusdt":
		return "usd"
	default:
		return "xtr"
	}
}

func (s *Service) absolute(path string) string {
	result := *s.publicURL
	result.Path = strings.TrimRight(result.Path, "/") + path
	result.RawQuery = ""
	if split := strings.IndexByte(path, '?'); split >= 0 {
		result.Path = strings.TrimRight(s.publicURL.Path, "/") + path[:split]
		result.RawQuery = path[split+1:]
	}
	return result.String()
}
