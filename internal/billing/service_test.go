package billing

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func TestServiceCreateOrderByProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		rate     string
		txbMinor int64
		amount   string
		currency string
	}{
		{provider: "ezpay", rate: "6.50", txbMinor: 123, amount: "8.00", currency: "CNY"},
		{provider: "bepusdt", rate: "0.125", txbMinor: 101, amount: "0.13", currency: "USD"},
		{provider: "stars", rate: "3", txbMinor: 250, amount: "8", currency: "XTR"},
	}

	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			t.Parallel()
			repository := newBillingRepository()
			settings := &billingSettings{values: map[string]string{
				"billing." + test.provider + ".enabled":                       "true",
				"billing.rate." + strings.ToLower(test.currency) + "_per_txb": test.rate,
			}}
			gateway := &billingGateway{}
			service := newBillingServiceForTest(repository, settings, gateway)
			user := model.User{ID: "user-1", TelegramID: 9876}

			order, err := service.CreateOrder(context.Background(), user, " "+strings.ToUpper(test.provider)+" ", test.txbMinor)
			if err != nil {
				t.Fatalf("CreateOrder(): %v", err)
			}
			if order.Provider != test.provider || order.PayableAmount != test.amount || order.PayableCurrency != test.currency {
				t.Fatalf("order = (%q, %q, %q), want (%q, %q, %q)", order.Provider, order.PayableAmount, order.PayableCurrency, test.provider, test.amount, test.currency)
			}
			if repository.updatedOrderID != "order-1" {
				t.Fatalf("updated order ID = %q, want order-1", repository.updatedOrderID)
			}
			if gateway.request.OrderID != "order-1" || gateway.request.TelegramID != user.TelegramID || gateway.request.TXBMinor != test.txbMinor {
				t.Fatalf("gateway request identity = %+v", gateway.request)
			}
			if gateway.request.NotifyURL != "https://example.test/base/api/v1/webhooks/"+test.provider {
				t.Fatalf("notify URL = %q", gateway.request.NotifyURL)
			}
			wantReturn := "https://example.test/base/api/v1/payments/return/" + test.provider + "/order-1"
			if gateway.request.ReturnURL != wantReturn || gateway.request.RedirectURL != wantReturn {
				t.Fatalf("return URLs = %q, %q, want %q", gateway.request.ReturnURL, gateway.request.RedirectURL, wantReturn)
			}
			if order.ExpiresAt != service.now().UTC().Add(30*time.Minute) {
				t.Fatalf("expiry = %s", order.ExpiresAt)
			}
		})
	}
}

func TestServiceCreateOrderPreservesProviderCheckout(t *testing.T) {
	t.Parallel()

	repository := newBillingRepository()
	settings := &billingSettings{values: map[string]string{
		"billing.ezpay.enabled":    "true",
		"billing.rate.cny_per_txb": "1",
	}}
	tradeID, paymentURL, qr := "trade", "https://pay.test/order", "qr-payload"
	expires := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	gateway := &billingGateway{checkout: ProviderCheckout{
		TradeID: &tradeID, PaymentURL: &paymentURL, QRPayload: &qr,
		PayableAmount: "1.234", PayableCurrency: "ALT", ProviderPayload: `{"ok":true}`, ExpiresAt: expires,
	}}
	service := newBillingServiceForTest(repository, settings, gateway)

	order, err := service.CreateOrder(context.Background(), model.User{ID: "user-1"}, "ezpay", 100)
	if err != nil {
		t.Fatalf("CreateOrder(): %v", err)
	}
	if order.ProviderTradeID == nil || *order.ProviderTradeID != tradeID || order.PaymentURL == nil || *order.PaymentURL != paymentURL || order.QRPayload == nil || *order.QRPayload != qr {
		t.Fatalf("provider checkout not preserved: %+v", order)
	}
	if order.PayableAmount != "1.234" || order.PayableCurrency != "ALT" || order.ProviderPayload != `{"ok":true}` || !order.ExpiresAt.Equal(expires) {
		t.Fatalf("provider checkout values not preserved: %+v", order)
	}
}

func TestServiceCreateOrderFailures(t *testing.T) {
	t.Parallel()

	testError := errors.New("test failure")
	tests := []struct {
		name       string
		provider   string
		txbMinor   int64
		configure  func(*billingRepository, *billingSettings, *billingGateway)
		want       error
		wantFailed bool
	}{
		{name: "unsupported provider", provider: "cash", txbMinor: 100, want: ErrInvalidOrder},
		{name: "zero amount", provider: "ezpay", txbMinor: 0, want: ErrInvalidOrder},
		{name: "too large", provider: "ezpay", txbMinor: 100_000_000_01, want: ErrInvalidOrder},
		{name: "disabled", provider: "ezpay", txbMinor: 100, want: ErrProviderDisabled},
		{name: "enabled lookup error", provider: "ezpay", txbMinor: 100, configure: func(_ *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.errs["billing.ezpay.enabled"] = testError
		}, want: ErrProviderDisabled},
		{name: "rate lookup error", provider: "ezpay", txbMinor: 100, configure: func(_ *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.errs["billing.rate.cny_per_txb"] = testError
		}},
		{name: "invalid rate", provider: "ezpay", txbMinor: 100, configure: func(_ *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.values["billing.rate.cny_per_txb"] = "zero"
		}},
		{name: "zero rate", provider: "ezpay", txbMinor: 100, configure: func(_ *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.values["billing.rate.cny_per_txb"] = "0"
		}},
		{name: "repository create", provider: "ezpay", txbMinor: 100, configure: func(repository *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.values["billing.rate.cny_per_txb"] = "1"
			repository.createErr = testError
		}},
		{name: "gateway create", provider: "ezpay", txbMinor: 100, configure: func(_ *billingRepository, settings *billingSettings, gateway *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.values["billing.rate.cny_per_txb"] = "1"
			gateway.err = testError
		}, wantFailed: true},
		{name: "repository update", provider: "ezpay", txbMinor: 100, configure: func(repository *billingRepository, settings *billingSettings, _ *billingGateway) {
			settings.values["billing.ezpay.enabled"] = "true"
			settings.values["billing.rate.cny_per_txb"] = "1"
			repository.updateErr = testError
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := newBillingRepository()
			settings := &billingSettings{values: map[string]string{}, errs: map[string]error{}}
			gateway := &billingGateway{}
			if test.configure != nil {
				test.configure(repository, settings, gateway)
			}
			service := newBillingServiceForTest(repository, settings, gateway)
			_, err := service.CreateOrder(context.Background(), model.User{ID: "user-1"}, test.provider, test.txbMinor)
			if err == nil {
				t.Fatal("CreateOrder() unexpectedly succeeded")
			}
			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("CreateOrder() error = %v, want %v", err, test.want)
			}
			if repository.failed != test.wantFailed {
				t.Fatalf("FailPaymentOrder called = %t, want %t", repository.failed, test.wantFailed)
			}
		})
	}
}

func TestServiceRetriesAmbiguousBEPusdtCreateWithSameOrder(t *testing.T) {
	t.Parallel()

	repository := newBillingRepository()
	settings := &billingSettings{values: map[string]string{
		"billing.bepusdt.enabled":  "true",
		"billing.rate.usd_per_txb": "1",
	}}
	gateway := &bepusdtRetryGateway{firstErr: errors.New("ambiguous timeout")}
	service := newBillingServiceForTest(repository, settings, gateway)
	order, err := service.CreateOrder(context.Background(), model.User{ID: "user-1"}, "bepusdt", 100)
	if err != nil {
		t.Fatalf("CreateOrder(): %v", err)
	}
	if gateway.calls != 2 || gateway.orderIDs[0] != "order-1" || gateway.orderIDs[1] != "order-1" || order.Status != "pending" {
		t.Fatalf("retry calls/order IDs/status = %d/%v/%q", gateway.calls, gateway.orderIDs, order.Status)
	}
}

func TestServiceValidateEvent(t *testing.T) {
	t.Parallel()

	tradeID := "trade-1"
	baseOrder := model.PaymentOrder{ID: "order-1", UserID: "user-1", Provider: "ezpay", PayableAmount: "1.00", PayableCurrency: "CNY", ProviderTradeID: &tradeID}
	telegramID := int64(42)
	tests := []struct {
		name      string
		mutate    func(*billingRepository, *ProviderEvent)
		wantError error
	}{
		{name: "valid", mutate: func(_ *billingRepository, event *ProviderEvent) { event.TelegramID = &telegramID }},
		{name: "order lookup", mutate: func(repository *billingRepository, _ *ProviderEvent) { repository.lookupErr = errors.New("lookup") }},
		{name: "provider mismatch", mutate: func(_ *billingRepository, event *ProviderEvent) { event.Provider = "stars" }, wantError: database.ErrConflict},
		{name: "currency mismatch", mutate: func(_ *billingRepository, event *ProviderEvent) { event.PayableCurrency = "USD" }, wantError: database.ErrConflict},
		{name: "amount mismatch", mutate: func(_ *billingRepository, event *ProviderEvent) { event.PayableAmount = "1.01" }, wantError: database.ErrConflict},
		{name: "trade mismatch", mutate: func(_ *billingRepository, event *ProviderEvent) { event.TradeID = "other" }, wantError: database.ErrConflict},
		{name: "telegram mismatch", mutate: func(_ *billingRepository, event *ProviderEvent) { wrong := int64(99); event.TelegramID = &wrong }, wantError: database.ErrConflict},
		{name: "user lookup error", mutate: func(repository *billingRepository, event *ProviderEvent) {
			event.TelegramID = &telegramID
			repository.userErr = errors.New("user lookup")
		}, wantError: database.ErrConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := newBillingRepository()
			repository.orders[baseOrder.ID] = baseOrder
			repository.user = model.User{ID: "user-1", TelegramID: telegramID}
			event := ProviderEvent{Provider: "ezpay", OrderID: baseOrder.ID, TradeID: tradeID, PayableAmount: "1.0", PayableCurrency: "cny"}
			if test.mutate != nil {
				test.mutate(repository, &event)
			}
			_, err := newBillingServiceForTest(repository, &billingSettings{}, &billingGateway{}).ValidateEvent(context.Background(), event)
			if test.wantError != nil && !errors.Is(err, test.wantError) {
				t.Fatalf("ValidateEvent() error = %v, want %v", err, test.wantError)
			}
			if test.wantError == nil && test.name != "order lookup" && err != nil {
				t.Fatalf("ValidateEvent(): %v", err)
			}
			if test.name == "order lookup" && err == nil {
				t.Fatal("ValidateEvent() unexpectedly succeeded")
			}
		})
	}
}

func TestServiceSettleAndOrderForwarding(t *testing.T) {
	t.Parallel()

	repository := newBillingRepository()
	repository.orders["order-1"] = model.PaymentOrder{ID: "order-1", UserID: "user-1", Provider: "stars", PayableAmount: "5", PayableCurrency: "XTR"}
	service := newBillingServiceForTest(repository, &billingSettings{}, &billingGateway{})

	order, applied, err := service.Settle(context.Background(), ProviderEvent{
		Provider: "stars", OrderID: "order-1", TradeID: "trade-1", ChargeID: "charge-1", PayableAmount: "5.0", PayableCurrency: "xtr",
	})
	if err != nil || !applied || order.ID != "order-1" {
		t.Fatalf("Settle() = (%+v, %t, %v)", order, applied, err)
	}
	if repository.settleDedupe != "trade-1" || len(repository.settleHash) != 64 {
		t.Fatalf("settlement dedupe/hash = %q/%q", repository.settleDedupe, repository.settleHash)
	}

	repository.paymentForUser = model.PaymentOrder{ID: "poll-order"}
	poll, err := service.OrderForUser(context.Background(), "poll-order", "user-1")
	if err != nil || poll.ID != "poll-order" || repository.pollUserID != "user-1" {
		t.Fatalf("OrderForUser() = (%+v, %v), user %q", poll, err, repository.pollUserID)
	}

	_, _, err = service.Settle(context.Background(), ProviderEvent{Provider: "stars", OrderID: "order-1", PayableAmount: "5", PayableCurrency: "XTR"})
	if err == nil {
		t.Fatal("Settle() without dedupe unexpectedly succeeded")
	}
}

func TestServiceAuthorizeEventRequiresLivePendingOrder(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		status    string
		expiresAt time.Time
		wantError bool
	}{
		{name: "pending", status: "pending", expiresAt: now.Add(time.Minute)},
		{name: "creating", status: "creating", expiresAt: now.Add(time.Minute), wantError: true},
		{name: "paid", status: "paid", expiresAt: now.Add(time.Minute), wantError: true},
		{name: "refunded", status: "refunded", expiresAt: now.Add(time.Minute), wantError: true},
		{name: "failed", status: "failed", expiresAt: now.Add(time.Minute), wantError: true},
		{name: "expired by status", status: "expired", expiresAt: now.Add(time.Minute), wantError: true},
		{name: "expired by time", status: "pending", expiresAt: now, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := newBillingRepository()
			repository.orders["order-1"] = model.PaymentOrder{ID: "order-1", UserID: "user-1", Provider: "stars", Status: test.status, PayableAmount: "5", PayableCurrency: "XTR", ExpiresAt: test.expiresAt}
			_, err := newBillingServiceForTest(repository, &billingSettings{}, &billingGateway{}).AuthorizeEvent(context.Background(), ProviderEvent{
				Provider: "stars", OrderID: "order-1", PayableAmount: "5", PayableCurrency: "XTR",
			})
			if test.wantError && !errors.Is(err, database.ErrConflict) {
				t.Fatalf("AuthorizeEvent() error = %v, want ErrConflict", err)
			}
			if !test.wantError && err != nil {
				t.Fatalf("AuthorizeEvent(): %v", err)
			}
		})
	}
}

func TestServiceValidateBEPusdtFiatAndCryptoAmounts(t *testing.T) {
	t.Parallel()

	recipient := "TXExactReceiveAddress"
	repository := newBillingRepository()
	repository.orders["order-1"] = model.PaymentOrder{
		ID: "order-1", UserID: "user-1", Provider: "bepusdt", TXBMinor: 1_000,
		PayableAmount: "0.950001", PayableCurrency: "USDT", RateSnapshot: "0.1", QRPayload: &recipient,
	}
	service := newBillingServiceForTest(repository, &billingSettings{}, &billingGateway{})
	event := ProviderEvent{
		Provider: "bepusdt", OrderID: "order-1", PayableAmount: "0.950001", PayableCurrency: "USDT",
		FiatAmount: "1.00", FiatCurrency: "USD", Recipient: recipient,
	}
	if _, err := service.ValidateEvent(context.Background(), event); err != nil {
		t.Fatalf("ValidateEvent(valid): %v", err)
	}
	event.FiatAmount = "1.01"
	if _, err := service.ValidateEvent(context.Background(), event); !errors.Is(err, database.ErrConflict) {
		t.Fatalf("ValidateEvent(fiat mismatch) error = %v, want ErrConflict", err)
	}
	event.FiatAmount = "1.00"
	event.PayableAmount = "0.950002"
	if _, err := service.ValidateEvent(context.Background(), event); !errors.Is(err, database.ErrConflict) {
		t.Fatalf("ValidateEvent(crypto mismatch) error = %v, want ErrConflict", err)
	}
	event.PayableAmount = "0.950001"
	event.Recipient = "different-address"
	if _, err := service.ValidateEvent(context.Background(), event); !errors.Is(err, database.ErrConflict) {
		t.Fatalf("ValidateEvent(recipient mismatch) error = %v, want ErrConflict", err)
	}
	event.Recipient = "USDT"
	if _, err := service.ValidateEvent(context.Background(), event); err != nil {
		t.Fatalf("ValidateEvent(overloaded token currency): %v", err)
	}
}

type billingRepository struct {
	orders         map[string]model.PaymentOrder
	user           model.User
	createErr      error
	updateErr      error
	lookupErr      error
	userErr        error
	failed         bool
	updatedOrderID string
	settleDedupe   string
	settleHash     string
	paymentForUser model.PaymentOrder
	pollUserID     string
}

func newBillingRepository() *billingRepository {
	return &billingRepository{orders: make(map[string]model.PaymentOrder), user: model.User{ID: "user-1"}}
}

func (r *billingRepository) CreatePaymentOrder(_ context.Context, order model.PaymentOrder) (model.PaymentOrder, error) {
	if r.createErr != nil {
		return model.PaymentOrder{}, r.createErr
	}
	order.ID = "order-1"
	r.orders[order.ID] = order
	return order, nil
}

func (r *billingRepository) UpdatePaymentCheckout(_ context.Context, orderID string, tradeID, paymentURL, qrPayload *string, amount, currency, payload string, expires time.Time) (model.PaymentOrder, error) {
	if r.updateErr != nil {
		return model.PaymentOrder{}, r.updateErr
	}
	r.updatedOrderID = orderID
	order := r.orders[orderID]
	order.Status = "pending"
	order.ProviderTradeID = tradeID
	order.PaymentURL = paymentURL
	order.QRPayload = qrPayload
	order.PayableAmount = amount
	order.PayableCurrency = currency
	order.ProviderPayload = payload
	order.ExpiresAt = expires
	r.orders[orderID] = order
	return order, nil
}

func (r *billingRepository) FailPaymentOrder(context.Context, string, string) error {
	r.failed = true
	return nil
}

func (r *billingRepository) PaymentOrderByID(_ context.Context, id string) (model.PaymentOrder, error) {
	if r.lookupErr != nil {
		return model.PaymentOrder{}, r.lookupErr
	}
	return r.orders[id], nil
}

func (r *billingRepository) PaymentOrderForUser(_ context.Context, _, userID string) (model.PaymentOrder, error) {
	r.pollUserID = userID
	return r.paymentForUser, nil
}

func (r *billingRepository) SettlePayment(_ context.Context, _, dedupeKey, payloadHash, orderID, _, _ string, _ time.Time) (model.PaymentOrder, bool, error) {
	r.settleDedupe = dedupeKey
	r.settleHash = payloadHash
	return r.orders[orderID], true, nil
}

func (r *billingRepository) UserByID(context.Context, string) (model.User, error) {
	return r.user, r.userErr
}

type billingSettings struct {
	values map[string]string
	errs   map[string]error
}

func (s *billingSettings) Plaintext(_ context.Context, key string) (string, error) {
	if s.errs != nil && s.errs[key] != nil {
		return "", s.errs[key]
	}
	return s.values[key], nil
}

func (s *billingSettings) Optional(ctx context.Context, key string) (string, error) {
	return s.Plaintext(ctx, key)
}

type billingGateway struct {
	request  ProviderCreateRequest
	checkout ProviderCheckout
	err      error
}

type bepusdtRetryGateway struct {
	calls    int
	orderIDs []string
	firstErr error
}

func (g *bepusdtRetryGateway) Create(_ context.Context, request ProviderCreateRequest) (ProviderCheckout, error) {
	g.calls++
	g.orderIDs = append(g.orderIDs, request.OrderID)
	if g.calls == 1 {
		return ProviderCheckout{}, g.firstErr
	}
	return ProviderCheckout{}, nil
}

func (g *billingGateway) Create(_ context.Context, request ProviderCreateRequest) (ProviderCheckout, error) {
	g.request = request
	return g.checkout, g.err
}

func newBillingServiceForTest(repository Repository, settings Settings, gateway Gateway) *Service {
	publicURL, _ := url.Parse("https://example.test/base")
	service := NewService(repository, settings, gateway, publicURL)
	service.now = func() time.Time { return time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC) }
	return service
}
