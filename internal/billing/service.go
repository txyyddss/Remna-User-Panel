package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// ErrProviderDisabled means the administrator has not enabled a payment method.
var ErrProviderDisabled = errors.New("payment provider is disabled")

// ErrInvalidOrder means user-controlled checkout selection failed validation.
var ErrInvalidOrder = errors.New("invalid payment order")

// ProviderCreateRequest is the server-owned checkout input passed to an adapter.
type ProviderCreateRequest struct {
	Provider        string
	MethodID        string
	Rail            string
	OrderID         string
	TelegramID      int64
	TXBMinor        int64
	PayableAmount   string
	PayableCurrency string
	NotifyURL       string
	ReturnURL       string
	RedirectURL     string
}

// ProviderCheckout is the exact display contract returned by a provider.
type ProviderCheckout struct {
	TradeID              *string
	PaymentURL           *string
	QRPayload            *string
	ReceivingAddress     *string
	ActualCryptoAmount   *string
	ActualCryptoCurrency *string
	PayableAmount        string
	PayableCurrency      string
	ProviderPayload      string
	ExpiresAt            time.Time
}

// ProviderEvent is a verified, normalized authoritative payment event.
type ProviderEvent struct {
	Provider        string
	Rail            string
	OrderID         string
	TradeID         string
	ChargeID        string
	PayableAmount   string
	PayableCurrency string
	FiatAmount      string
	FiatCurrency    string
	Recipient       string
	DedupeKey       string
	PayloadHash     string
	TelegramID      *int64
}

// Gateway creates provider-specific checkouts. It receives only server-owned values.
type Gateway interface {
	Create(context.Context, ProviderCreateRequest) (ProviderCheckout, error)
}

// CancellationGateway is implemented by providers that expose a signed order
// cancellation operation. Absence is a valid best-effort outcome.
type CancellationGateway interface {
	Cancel(context.Context, model.PaymentOrder) error
}

// Settings is the trusted runtime configuration surface.
type Settings interface {
	Plaintext(context.Context, string) (string, error)
	Optional(context.Context, string) (string, error)
}

// Repository is the transactional billing persistence contract.
type Repository interface {
	CreatePaymentOrder(context.Context, model.PaymentOrder) (model.PaymentOrder, error)
	UpdatePaymentCheckout(context.Context, string, *string, *string, *string, string, string, string, time.Time) (model.PaymentOrder, error)
	FailPaymentOrder(context.Context, string, string) error
	PaymentOrderByID(context.Context, string) (model.PaymentOrder, error)
	PaymentOrderForUser(context.Context, string, string) (model.PaymentOrder, error)
	SettlePayment(context.Context, string, string, string, string, string, string, time.Time) (model.PaymentOrder, bool, error)
	UserByID(context.Context, string) (model.User, error)
}

type checkoutDetailsRepository interface {
	UpdatePaymentCheckoutDetails(context.Context, string, *string, *string, *string, *string, *string, *string, string, string, string, time.Time) (model.PaymentOrder, error)
}

type cancellationRepository interface {
	CancelPaymentOrder(context.Context, string, string, string, time.Time) (model.PaymentOrder, bool, error)
	SetPaymentProviderCancellation(context.Context, string, string, time.Time) error
}

// Service calculates prices and commits verified provider events.
type Service struct {
	repository Repository
	settings   Settings
	gateway    Gateway
	publicURL  *url.URL
	now        func() time.Time
}

// NewService creates a payment service.
func NewService(repository Repository, settings Settings, gateway Gateway, publicURL *url.URL) *Service {
	return &Service{repository: repository, settings: settings, gateway: gateway, publicURL: publicURL, now: time.Now}
}

// Methods returns the configured ordered rail list. A method is selectable only
// when both its provider and new-direction rate are configured.
func (s *Service) Methods(ctx context.Context) ([]model.PaymentMethod, error) {
	result := make([]model.PaymentMethod, 0)
	for _, provider := range []string{"ezpay", "bepusdt", "stars"} {
		enabled, err := s.settings.Optional(ctx, "billing."+provider+".enabled")
		if err != nil {
			return nil, err
		}
		if enabled != "true" {
			continue
		}
		rate, rateErr := s.loadNewRate(ctx, provider)
		available := rateErr == nil && rate.Positive()
		note := ""
		if !available {
			note = "Administrator must enter the TXB rate"
		}
		if provider == "stars" {
			result = append(result, methodModel(provider, "", available, note))
			continue
		}
		raw, err := s.settings.Optional(ctx, "billing."+provider+".methods")
		if err != nil {
			return nil, err
		}
		allowed := ezpayRails
		if provider == "bepusdt" {
			allowed = bepusdtRails
		}
		rails, err := parseEnabledRails(raw, allowed)
		if err != nil {
			return nil, fmt.Errorf("load %s methods: %w", provider, err)
		}
		for _, rail := range rails {
			result = append(result, methodModel(provider, rail, available, note))
		}
	}
	return result, nil
}

// CreateOrder computes authoritative pricing, persists the attempt, then requests checkout data.
func (s *Service) CreateOrder(ctx context.Context, user model.User, methodID string, txbMinor int64) (model.PaymentOrder, error) {
	provider, rail, err := ParseMethodID(methodID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if txbMinor <= 0 || txbMinor > 100_000_000_00 {
		return model.PaymentOrder{}, fmt.Errorf("%w: TXB amount is out of range", ErrInvalidOrder)
	}
	enabled, err := s.settings.Optional(ctx, "billing."+provider+".enabled")
	if err != nil || enabled != "true" {
		return model.PaymentOrder{}, ErrProviderDisabled
	}
	if provider != "stars" {
		raw, settingsErr := s.settings.Optional(ctx, "billing."+provider+".methods")
		if settingsErr != nil {
			return model.PaymentOrder{}, settingsErr
		}
		allowed := ezpayRails
		if provider == "bepusdt" {
			allowed = bepusdtRails
		}
		enabledRails, parseErr := parseEnabledRails(raw, allowed)
		if parseErr != nil || !containsRail(enabledRails, rail) {
			return model.PaymentOrder{}, ErrProviderDisabled
		}
	}
	rate, err := s.loadNewRate(ctx, provider)
	if err != nil {
		return model.PaymentOrder{}, ErrProviderDisabled
	}
	precision := 2
	if provider == "stars" {
		precision = 0
	}
	payable, err := PayableFromTXBPerCurrency(txbMinor, rate, precision)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	now := s.now().UTC()
	canonicalMethodID := provider
	if rail != "" {
		canonicalMethodID += ":" + rail
	}
	order := model.PaymentOrder{
		UserID: user.ID, Provider: provider, MethodID: canonicalMethodID, ProviderRail: rail, Status: "creating", TXBMinor: txbMinor,
		PayableAmount: payable, PayableCurrency: strings.ToUpper(currencyCode(provider)), RateSnapshot: rate.Canonical(),
		RateDirection: "txb_per_currency", ProviderPayload: "{}", ExpiresAt: now.Add(30 * time.Minute),
	}
	order, err = s.repository.CreatePaymentOrder(ctx, order)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	request := ProviderCreateRequest{
		Provider: provider, MethodID: order.MethodID, Rail: rail, OrderID: order.ID, TelegramID: user.TelegramID, TXBMinor: txbMinor,
		PayableAmount: payable, PayableCurrency: order.PayableCurrency,
		NotifyURL:   s.absolute("/api/v1/webhooks/" + provider),
		ReturnURL:   s.absolute("/api/v1/payments/return/" + provider + "/" + url.PathEscape(order.ID)),
		RedirectURL: s.absolute("/api/v1/payments/return/" + provider + "/" + url.PathEscape(order.ID)),
	}
	if provider == "bepusdt" {
		secret, secretErr := s.settings.Plaintext(ctx, "billing.bepusdt.api_token")
		if secretErr != nil {
			_ = s.repository.FailPaymentOrder(ctx, order.ID, `{"error":"callback_capability_failed"}`)
			return model.PaymentOrder{}, fmt.Errorf("create callback capability: %w", secretErr)
		}
		request.NotifyURL = s.absolute("/api/v1/webhooks/bepusdt/" + callbackCapability(secret, order.ID))
	}
	checkout, err := s.gateway.Create(ctx, request)
	if err != nil && provider == "bepusdt" && ctx.Err() == nil {
		// BEPusdt duplicate order IDs are the safest way to resolve an ambiguous create timeout.
		checkout, err = s.gateway.Create(ctx, request)
	}
	if err != nil {
		_ = s.repository.FailPaymentOrder(ctx, order.ID, `{"error":"provider_create_failed"}`)
		return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w", err)
	}
	if checkout.PayableAmount == "" {
		checkout.PayableAmount = order.PayableAmount
	}
	if checkout.PayableCurrency == "" {
		checkout.PayableCurrency = order.PayableCurrency
	}
	// The provider response may add display and transaction metadata, but it
	// must never replace the server-priced fiat snapshot used to validate the
	// eventual callback.
	if !strings.EqualFold(checkout.PayableCurrency, order.PayableCurrency) ||
		!Equivalent(checkout.PayableAmount, order.PayableAmount) {
		_ = s.repository.FailPaymentOrder(ctx, order.ID, `{"error":"provider_checkout_mismatch"}`)
		return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: provider pricing conflicts with immutable order", ErrInvalidOrder)
	}
	checkout.PayableAmount = order.PayableAmount
	checkout.PayableCurrency = order.PayableCurrency
	if (checkout.ActualCryptoAmount == nil) != (checkout.ActualCryptoCurrency == nil) {
		_ = s.repository.FailPaymentOrder(ctx, order.ID, `{"error":"provider_checkout_mismatch"}`)
		return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: incomplete crypto pricing", ErrInvalidOrder)
	}
	if checkout.ActualCryptoAmount != nil {
		actual, actualErr := ParseDecimal(*checkout.ActualCryptoAmount)
		if provider != "bepusdt" || actualErr != nil || !actual.Positive() ||
			!strings.EqualFold(*checkout.ActualCryptoCurrency, "USDT") {
			_ = s.repository.FailPaymentOrder(ctx, order.ID, `{"error":"provider_checkout_mismatch"}`)
			return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: invalid crypto pricing", ErrInvalidOrder)
		}
	}
	if checkout.ExpiresAt.IsZero() {
		checkout.ExpiresAt = order.ExpiresAt
	}
	if checkout.ProviderPayload == "" {
		checkout.ProviderPayload = "{}"
	}
	if repository, ok := s.repository.(checkoutDetailsRepository); ok {
		return repository.UpdatePaymentCheckoutDetails(ctx, order.ID, checkout.TradeID, checkout.PaymentURL, checkout.QRPayload, checkout.ReceivingAddress, checkout.ActualCryptoAmount, checkout.ActualCryptoCurrency,
			checkout.PayableAmount, checkout.PayableCurrency, checkout.ProviderPayload, checkout.ExpiresAt)
	}
	return s.repository.UpdatePaymentCheckout(ctx, order.ID, checkout.TradeID, checkout.PaymentURL, checkout.QRPayload,
		checkout.PayableAmount, checkout.PayableCurrency, checkout.ProviderPayload, checkout.ExpiresAt)
}

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
