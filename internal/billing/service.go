package billing

import (
	"context"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"net/url"
	"strings"
	"time"
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

type paymentProfileReader interface {
	PaymentProfiles(context.Context) ([]model.PaymentProfile, error)
}

// Repository is the transactional billing persistence contract.
type Repository interface {
	CreatePaymentOrder(context.Context, model.PaymentOrder) (model.PaymentOrder, error)
	UpdatePaymentCheckout(context.Context, string, *string, *string, *string, string, string, time.Time) (model.PaymentOrder, error)
	FailPaymentOrder(context.Context, string) error
	PaymentOrderByID(context.Context, string) (model.PaymentOrder, error)
	PaymentOrderForUser(context.Context, string, string) (model.PaymentOrder, error)
	SettlePayment(context.Context, string, string, string, string, string, string, time.Time) (model.PaymentOrder, bool, error)
	UserByID(context.Context, string) (model.User, error)
}

type checkoutDetailsRepository interface {
	UpdatePaymentCheckoutDetails(context.Context, string, *string, *string, *string, *string, *string, *string, string, string, time.Time) (model.PaymentOrder, error)
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

// CreateOrder computes authoritative pricing, persists the attempt, then requests checkout data.
func (s *Service) CreateOrder(ctx context.Context, user model.User, methodID string, txbMinor int64) (model.PaymentOrder, error) {
	provider, rail, err := ParseMethodID(methodID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if provider == "coupon" {
		return model.PaymentOrder{}, fmt.Errorf("%w: coupon funding is not a provider order", ErrInvalidOrder)
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
		RateDirection: "txb_per_currency", ExpiresAt: now.Add(30 * time.Minute),
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
			_ = s.repository.FailPaymentOrder(ctx, order.ID)
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
		_ = s.repository.FailPaymentOrder(ctx, order.ID)
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
		_ = s.repository.FailPaymentOrder(ctx, order.ID)
		return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: provider pricing conflicts with immutable order", ErrInvalidOrder)
	}
	checkout.PayableAmount = order.PayableAmount
	checkout.PayableCurrency = order.PayableCurrency
	if (checkout.ActualCryptoAmount == nil) != (checkout.ActualCryptoCurrency == nil) {
		_ = s.repository.FailPaymentOrder(ctx, order.ID)
		return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: incomplete crypto pricing", ErrInvalidOrder)
	}
	if checkout.ActualCryptoAmount != nil {
		actual, actualErr := ParseDecimal(*checkout.ActualCryptoAmount)
		if provider != "bepusdt" || actualErr != nil || !actual.Positive() ||
			!strings.EqualFold(*checkout.ActualCryptoCurrency, "USDT") {
			_ = s.repository.FailPaymentOrder(ctx, order.ID)
			return model.PaymentOrder{}, fmt.Errorf("create provider checkout: %w: invalid crypto pricing", ErrInvalidOrder)
		}
	}
	if checkout.ExpiresAt.IsZero() {
		checkout.ExpiresAt = order.ExpiresAt
	}
	if repository, ok := s.repository.(checkoutDetailsRepository); ok {
		return repository.UpdatePaymentCheckoutDetails(ctx, order.ID, checkout.TradeID, checkout.PaymentURL, checkout.QRPayload, checkout.ReceivingAddress, checkout.ActualCryptoAmount, checkout.ActualCryptoCurrency,
			checkout.PayableAmount, checkout.PayableCurrency, checkout.ExpiresAt)
	}
	return s.repository.UpdatePaymentCheckout(ctx, order.ID, checkout.TradeID, checkout.PaymentURL, checkout.QRPayload,
		checkout.PayableAmount, checkout.PayableCurrency, checkout.ExpiresAt)
}
