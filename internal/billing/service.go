package billing

import (
	"context"
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
	TradeID         *string
	PaymentURL      *string
	QRPayload       *string
	PayableAmount   string
	PayableCurrency string
	ProviderPayload string
	ExpiresAt       time.Time
}

// ProviderEvent is a verified, normalized authoritative payment event.
type ProviderEvent struct {
	Provider        string
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
func (s *Service) CreateOrder(ctx context.Context, user model.User, provider string, txbMinor int64) (model.PaymentOrder, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "ezpay" && provider != "bepusdt" && provider != "stars" {
		return model.PaymentOrder{}, fmt.Errorf("%w: unsupported provider", ErrInvalidOrder)
	}
	if txbMinor <= 0 || txbMinor > 100_000_000_00 {
		return model.PaymentOrder{}, fmt.Errorf("%w: TXB amount is out of range", ErrInvalidOrder)
	}
	enabled, err := s.settings.Optional(ctx, "billing."+provider+".enabled")
	if err != nil || enabled != "true" {
		return model.PaymentOrder{}, ErrProviderDisabled
	}
	rateRaw, err := s.settings.Plaintext(ctx, "billing.rate."+currencyCode(provider)+"_per_txb")
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("load payment rate: %w", err)
	}
	rate, err := ParseDecimal(rateRaw)
	if err != nil || !rate.Positive() {
		return model.PaymentOrder{}, errors.New("payment rate is invalid")
	}
	precision := 2
	if provider == "stars" {
		precision = 0
	}
	payable, err := Payable(txbMinor, rate, precision)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	now := s.now().UTC()
	order := model.PaymentOrder{
		UserID: user.ID, Provider: provider, Status: "creating", TXBMinor: txbMinor,
		PayableAmount: payable, PayableCurrency: strings.ToUpper(currencyCode(provider)), RateSnapshot: rate.Canonical(),
		ProviderPayload: "{}", ExpiresAt: now.Add(30 * time.Minute),
	}
	order, err = s.repository.CreatePaymentOrder(ctx, order)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	request := ProviderCreateRequest{
		Provider: provider, OrderID: order.ID, TelegramID: user.TelegramID, TXBMinor: txbMinor,
		PayableAmount: payable, PayableCurrency: order.PayableCurrency,
		NotifyURL:   s.absolute("/api/v1/webhooks/" + provider),
		ReturnURL:   s.absolute("/api/v1/payments/return/" + provider + "/" + url.PathEscape(order.ID)),
		RedirectURL: s.absolute("/api/v1/payments/return/" + provider + "/" + url.PathEscape(order.ID)),
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
	if checkout.ExpiresAt.IsZero() {
		checkout.ExpiresAt = order.ExpiresAt
	}
	if checkout.ProviderPayload == "" {
		checkout.ProviderPayload = "{}"
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
	if order.Provider != event.Provider || !strings.EqualFold(order.PayableCurrency, event.PayableCurrency) || !Equivalent(order.PayableAmount, event.PayableAmount) {
		return model.PaymentOrder{}, database.ErrConflict
	}
	if event.Provider == "bepusdt" {
		rate, parseErr := ParseDecimal(order.RateSnapshot)
		expectedFiat, payableErr := Payable(order.TXBMinor, rate, 2)
		if parseErr != nil || payableErr != nil || !strings.EqualFold(event.FiatCurrency, "USD") || !Equivalent(expectedFiat, event.FiatAmount) {
			return model.PaymentOrder{}, database.ErrConflict
		}
		// Reference versions overload callback token as either the receive
		// address or the literal currency code. Compare it to the exact checkout
		// address whenever the provider supplies address semantics.
		if event.Recipient != "" && !strings.EqualFold(event.Recipient, "USDT") &&
			(order.QRPayload == nil || event.Recipient != *order.QRPayload) {
			return model.PaymentOrder{}, database.ErrConflict
		}
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
