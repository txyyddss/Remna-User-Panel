package billing

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// ErrProviderDisabled means the administrator has not enabled a payment method.
var ErrProviderDisabled = errors.New("payment provider is disabled")

// ErrInvalidOrder means user-controlled checkout selection failed validation.
var ErrInvalidOrder = errors.New("invalid payment order")

// Settings is the trusted runtime configuration surface.
type Settings interface {
	Plaintext(context.Context, string) (string, error)
	Optional(context.Context, string) (string, error)
}

type paymentProfileReader interface {
	PaymentProfiles(context.Context) ([]model.PaymentProfile, error)
}

type paymentProfileRuntimeReader interface {
	PaymentProfile(context.Context, string, string) (model.PaymentProfileRuntime, error)
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
	providerEnabled, settingsErr := s.providerMethodEnabled(ctx, provider, rail)
	if settingsErr != nil {
		return model.PaymentOrder{}, settingsErr
	}
	if !providerEnabled {
		return model.PaymentOrder{}, ErrProviderDisabled
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
		secret, secretErr := s.providerCredential(ctx, "bepusdt")
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
