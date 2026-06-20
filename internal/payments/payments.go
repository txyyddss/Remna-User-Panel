// Package payments implements payment provider configuration, order storage, and webhooks.
package payments

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	appsettings "remna-user-panel/internal/settings"
)

const (
	// ProviderEZPay is the identifier for the EZPay payment provider.
	ProviderEZPay = "ezpay"
	// ProviderBEPUSDT is the identifier for the BEPUSDT payment provider.
	ProviderBEPUSDT       = "bepusdt"
	ProviderTelegramStars = "telegram_stars"
)

// Method is one checkout payment method shown in the Web App.
type Method struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	PaymentType string `json:"payment_type"`
	Icon        string `json:"icon"`
	LabelKey    string `json:"label_key"`
}

// Config is the effective payment configuration.
type Config struct {
	EZPay               config.EZPaySettings
	BEPUSDT             config.BEPUSDTSettings
	PaymentMethodsOrder []string
	WebhookBaseURL      string
	StarsEnabled        bool
	StarsUSDRate        float64
}

// CreateOrderRequest describes a server-trusted payment to create.
type CreateOrderRequest struct {
	UserID           int64
	MethodID         string
	Amount           float64
	Currency         string
	BaseAmount       float64
	BaseCurrency     string
	DisplayCNYAmount float64
	FXRate           float64
	FXSource         string
	FXUpdatedAt      time.Time
	PlanHash         string
	PlanSnapshot     json.RawMessage
	Description      string
	TariffKey        string
	SaleMode         string
	Months           int
	TrafficGB        float64
	DeviceCount      int
	ClientIP         string
	Language         string
}

// CreateOrderResponse is returned to the Web App.
type CreateOrderResponse struct {
	OK                bool    `json:"ok"`
	Action            string  `json:"action"`
	PaymentID         int64   `json:"payment_id"`
	OrderID           string  `json:"order_id"`
	Provider          string  `json:"provider"`
	Method            string  `json:"method"`
	PaymentType       string  `json:"payment_type"`
	Amount            float64 `json:"amount"`
	Currency          string  `json:"currency"`
	BaseAmount        float64 `json:"base_amount"`
	BaseCurrency      string  `json:"base_currency"`
	DisplayCNYAmount  float64 `json:"display_cny_amount,omitempty"`
	FXRate            float64 `json:"fx_rate,omitempty"`
	FXSource          string  `json:"fx_source,omitempty"`
	PlanHash          string  `json:"plan_hash,omitempty"`
	ProviderPaymentID string  `json:"provider_payment_id,omitempty"`
	PaymentURL        string  `json:"payment_url,omitempty"`
	QRContent         string  `json:"qr_content,omitempty"`
	DisplayAmount     string  `json:"display_amount,omitempty"`
	DisplayCurrency   string  `json:"display_currency,omitempty"`
	PaymentAddress    string  `json:"payment_address,omitempty"`
	Network           string  `json:"network,omitempty"`
	URLScheme         string  `json:"url_scheme,omitempty"`
	ExpiresInSeconds  int     `json:"expires_in_seconds,omitempty"`
	Status            string  `json:"status"`
}

// Order is one stored payment order.
type Order struct {
	PaymentID         int64           `json:"payment_id"`
	OrderID           string          `json:"order_id"`
	UserID            int64           `json:"user_id"`
	UserLabel         string          `json:"user_label,omitempty"`
	TelegramID        int64           `json:"telegram_id,omitempty"`
	Provider          string          `json:"provider"`
	Method            string          `json:"method"`
	PaymentType       string          `json:"payment_type"`
	Amount            float64         `json:"amount"`
	Currency          string          `json:"currency"`
	BaseAmount        float64         `json:"base_amount"`
	BaseCurrency      string          `json:"base_currency"`
	DisplayCNYAmount  float64         `json:"display_cny_amount,omitempty"`
	FXRate            float64         `json:"fx_rate,omitempty"`
	FXSource          string          `json:"fx_source,omitempty"`
	FXUpdatedAt       *time.Time      `json:"fx_updated_at,omitempty"`
	PlanHash          string          `json:"plan_hash,omitempty"`
	PlanSnapshot      json.RawMessage `json:"plan_snapshot,omitempty"`
	Status            string          `json:"status"`
	Description       string          `json:"description,omitempty"`
	TariffKey         string          `json:"tariff_key,omitempty"`
	SaleMode          string          `json:"sale_mode,omitempty"`
	Months            int             `json:"subscription_duration_months,omitempty"`
	TrafficGB         float64         `json:"traffic_regular_gb,omitempty"`
	ProviderPaymentID string          `json:"provider_payment_id,omitempty"`
	PaymentURL        string          `json:"payment_url,omitempty"`
	QRContent         string          `json:"qr_content,omitempty"`
	DisplayAmount     string          `json:"display_amount,omitempty"`
	DisplayCurrency   string          `json:"display_currency,omitempty"`
	PaymentAddress    string          `json:"payment_address,omitempty"`
	Network           string          `json:"network,omitempty"`
	URLScheme         string          `json:"url_scheme,omitempty"`
	RawWebhook        json.RawMessage `json:"raw_webhook,omitempty"`
	ProvisionedAt     *time.Time      `json:"provisioned_at,omitempty"`
	ProvisionError    string          `json:"provision_error,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	PaidAt            *time.Time      `json:"paid_at,omitempty"`
}

// Registry owns payment providers and order persistence.
type Registry struct {
	settings config.Settings
	pool     *pgxpool.Pool
	store    appsettings.Store
	client   *http.Client
}

// NewRegistry creates a payment registry.
func NewRegistry(settings config.Settings, pool *pgxpool.Pool) *Registry {
	return &Registry{
		settings: settings,
		pool:     pool,
		store:    appsettings.NewStore(pool),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// EffectiveConfig returns environment config with app_settings overrides applied.
func (r *Registry) EffectiveConfig(ctx context.Context) Config {
	ezpay := r.settings.EZPay
	ezpay.Enabled = r.store.Bool(ctx, "EZPAY_ENABLED", ezpay.Enabled)
	ezpay.BaseURL = strings.TrimRight(r.store.String(ctx, "EZPAY_BASE_URL", ezpay.BaseURL), "/")
	ezpay.PID = r.store.Int(ctx, "EZPAY_PID", ezpay.PID)
	ezpay.Key = r.store.String(ctx, "EZPAY_KEY", ezpay.Key)
	ezpay.ReturnURL = normalizedReturnURL(r.settings.SubscriptionMiniApp)

	bepusdt := r.settings.BEPUSDT
	bepusdt.Enabled = r.store.Bool(ctx, "BEPUSDT_ENABLED", bepusdt.Enabled)
	bepusdt.BaseURL = strings.TrimRight(r.store.String(ctx, "BEPUSDT_BASE_URL", bepusdt.BaseURL), "/")
	bepusdt.Token = r.store.String(ctx, "BEPUSDT_TOKEN", bepusdt.Token)
	bepusdt.ReturnURL = normalizedReturnURL(r.settings.SubscriptionMiniApp)

	orderRaw := r.store.String(ctx, "PAYMENT_METHODS_ORDER", strings.Join(r.settings.PaymentMethodsOrder, ","))
	return Config{
		EZPay:               ezpay,
		BEPUSDT:             bepusdt,
		PaymentMethodsOrder: splitList(orderRaw),
		WebhookBaseURL:      strings.TrimRight(r.store.String(ctx, "WEBHOOK_BASE_URL", r.settings.WebhookBaseURL), "/"),
		StarsEnabled:        r.store.Bool(ctx, "STARS_ENABLED", r.settings.StarsEnabled),
		StarsUSDRate:        r.store.Float(ctx, "STARS_USD_RATE", r.settings.StarsUSDRate),
	}
}

// IDs returns webhook provider ids.
func (r *Registry) IDs() []string {
	return []string{ProviderEZPay, ProviderBEPUSDT}
}

// Methods returns enabled and configured payment methods.
func (r *Registry) Methods(ctx context.Context, _ string, _ bool) []Method {
	cfg := r.EffectiveConfig(ctx)
	methods := []Method{}
	if cfg.EZPay.Enabled && cfg.EZPay.BaseURL != "" && cfg.EZPay.PID != 0 && cfg.EZPay.Key != "" {
		methods = append(methods,
			Method{ID: "ezpay:alipay", Name: "Alipay", LabelKey: "payment_method_ezpay_alipay", Provider: ProviderEZPay, PaymentType: "alipay", Icon: "CreditCard"},
			Method{ID: "ezpay:wxpay", Name: "WeChat Pay", LabelKey: "payment_method_ezpay_wxpay", Provider: ProviderEZPay, PaymentType: "wxpay", Icon: "WalletCards"},
			Method{ID: "ezpay:usdt", Name: "EZPay USDT", LabelKey: "payment_method_ezpay_usdt", Provider: ProviderEZPay, PaymentType: "usdt", Icon: "Bitcoin"},
		)
	}
	if cfg.BEPUSDT.Enabled && cfg.BEPUSDT.BaseURL != "" && cfg.BEPUSDT.Token != "" {
		methods = append(methods,
			Method{ID: "bepusdt:usdt.polygon", Name: "USDT Polygon", LabelKey: "payment_method_bepusdt_polygon", Provider: ProviderBEPUSDT, PaymentType: "usdt.polygon", Icon: "Bitcoin"},
			Method{ID: "bepusdt:usdt.arbitrum", Name: "USDT Arbitrum", LabelKey: "payment_method_bepusdt_arbitrum", Provider: ProviderBEPUSDT, PaymentType: "usdt.arbitrum", Icon: "Bitcoin"},
			Method{ID: "bepusdt:usdt.aptos", Name: "USDT Aptos", LabelKey: "payment_method_bepusdt_aptos", Provider: ProviderBEPUSDT, PaymentType: "usdt.aptos", Icon: "Bitcoin"},
		)
	}
	if cfg.StarsEnabled && strings.TrimSpace(r.settings.BotToken) != "" {
		methods = append(methods, Method{ID: "telegram_stars:xtr", Name: "Telegram Stars", LabelKey: "payment_method_telegram_stars", Provider: ProviderTelegramStars, PaymentType: "xtr", Icon: "Star"})
	}
	sortMethods(methods, cfg.PaymentMethodsOrder)
	return methods
}

// Create creates a payment order and calls the selected provider.
func (r *Registry) Create(ctx context.Context, req CreateOrderRequest) (CreateOrderResponse, error) {
	method, err := parseMethod(req.MethodID)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	cfg := r.EffectiveConfig(ctx)
	if err := validateProviderConfig(cfg, method.Provider); err != nil {
		return CreateOrderResponse{}, err
	}
	orderID, err := randomOrderID()
	if err != nil {
		return CreateOrderResponse{}, err
	}
	amount := math.Round(req.Amount*100) / 100
	if amount <= 0 {
		return CreateOrderResponse{}, fmt.Errorf("amount must be greater than zero")
	}
	baseAmount := math.Round(req.BaseAmount*100) / 100
	if baseAmount <= 0 {
		baseAmount = amount
	}
	baseCurrency := strings.ToUpper(strings.TrimSpace(req.BaseCurrency))
	if baseCurrency == "" {
		baseCurrency = strings.ToUpper(req.Currency)
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = baseCurrency
	}
	planSnapshot := req.PlanSnapshot
	if len(planSnapshot) == 0 || !json.Valid(planSnapshot) {
		planSnapshot = json.RawMessage(`{}`)
	}
	if r.pool == nil {
		return CreateOrderResponse{}, fmt.Errorf("database is not configured")
	}
	var paymentID int64
	err = r.pool.QueryRow(ctx, `
INSERT INTO payment_orders
	(order_id, user_id, provider, method, payment_type, amount, currency, base_amount, base_currency,
	 display_cny_amount, fx_rate, fx_source, fx_updated_at, plan_hash, plan_snapshot,
	 status, description, tariff_key, sale_mode, months, traffic_gb, device_count, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,'pending',$16,$17,$18,$19,$20,$21,NOW(),NOW())
RETURNING payment_id`,
		orderID, req.UserID, method.Provider, req.MethodID, method.PaymentType, amount, currency, baseAmount, baseCurrency,
		zeroFloatToNil(req.DisplayCNYAmount), zeroFloatToNil(req.FXRate), emptyToNil(req.FXSource), zeroTimeToNil(req.FXUpdatedAt),
		emptyToNil(req.PlanHash), planSnapshot,
		req.Description, emptyToNil(req.TariffKey), emptyToNil(req.SaleMode), zeroIntToNil(req.Months),
		zeroFloatToNil(req.TrafficGB), zeroIntToNil(req.DeviceCount),
	).Scan(&paymentID)
	if err != nil {
		return CreateOrderResponse{}, fmt.Errorf("insert payment order: %w", err)
	}

	providerResp, err := r.createProviderPayment(ctx, cfg, method, providerPaymentRequest{
		OrderID:     orderID,
		Amount:      amount,
		Currency:    currency,
		Description: req.Description,
		ClientIP:    req.ClientIP,
	})
	if err != nil {
		_ = r.markFailed(ctx, paymentID)
		return CreateOrderResponse{}, err
	}
	if err := r.updateProviderFields(ctx, paymentID, providerResp); err != nil {
		return CreateOrderResponse{}, err
	}
	if method.Provider == ProviderTelegramStars {
		_, _ = r.pool.Exec(ctx, "UPDATE payment_orders SET telegram_invoice_payload=$2 WHERE payment_id=$1", paymentID, orderID)
	}

	action := "show_checkout"
	if method.Provider == ProviderTelegramStars {
		action = "open_invoice"
	}
	return CreateOrderResponse{
		OK:                true,
		Action:            action,
		PaymentID:         paymentID,
		OrderID:           orderID,
		Provider:          method.Provider,
		Method:            req.MethodID,
		PaymentType:       method.PaymentType,
		Amount:            amount,
		Currency:          currency,
		BaseAmount:        baseAmount,
		BaseCurrency:      baseCurrency,
		DisplayCNYAmount:  req.DisplayCNYAmount,
		FXRate:            req.FXRate,
		FXSource:          req.FXSource,
		PlanHash:          req.PlanHash,
		ProviderPaymentID: providerResp.ProviderPaymentID,
		PaymentURL:        providerResp.PaymentURL,
		QRContent:         providerResp.QRContent,
		DisplayAmount:     providerResp.DisplayAmount,
		DisplayCurrency:   providerResp.DisplayCurrency,
		PaymentAddress:    providerResp.PaymentAddress,
		Network:           providerResp.Network,
		URLScheme:         providerResp.URLScheme,
		ExpiresInSeconds:  providerResp.ExpiresInSeconds,
		Status:            "pending",
	}, nil
}

// GetForUser returns one payment order visible to a user.
func (r *Registry) GetForUser(ctx context.Context, userID int64, paymentID int64) (Order, error) {
	return r.scanOrder(ctx, `
SELECT payment_id, order_id, user_id, COALESCE(provider,''), COALESCE(method,''), COALESCE(payment_type,''),
	amount::float8, COALESCE(currency,''), COALESCE(base_amount, amount)::float8, COALESCE(base_currency, currency),
	COALESCE(display_cny_amount,0)::float8, COALESCE(fx_rate,0)::float8, COALESCE(fx_source,''), fx_updated_at,
	COALESCE(plan_hash,''), COALESCE(plan_snapshot,'{}'::jsonb), COALESCE(status,''), COALESCE(description,''),
	COALESCE(tariff_key,''), COALESCE(sale_mode,''), COALESCE(months,0), COALESCE(traffic_gb,0)::float8,
	COALESCE(device_count,0), COALESCE(provider_payment_id,''), COALESCE(payment_url,''), COALESCE(qr_content,''),
	COALESCE(display_amount,''), COALESCE(display_currency,''), COALESCE(payment_address,''), COALESCE(network,''),
	COALESCE(url_scheme,''), COALESCE(raw_webhook,'{}'::jsonb), provisioned_at, COALESCE(provision_error,''), created_at, updated_at, paid_at
FROM payment_orders WHERE payment_id=$1 AND user_id=$2`, paymentID, userID)
}

// Get returns one payment order for admin views.
func (r *Registry) Get(ctx context.Context, paymentID int64) (Order, error) {
	return r.scanOrder(ctx, `
SELECT p.payment_id, p.order_id, p.user_id, COALESCE(p.provider,''), COALESCE(p.method,''), COALESCE(p.payment_type,''),
	p.amount::float8, COALESCE(p.currency,''), COALESCE(p.base_amount, p.amount)::float8, COALESCE(p.base_currency, p.currency),
	COALESCE(p.display_cny_amount,0)::float8, COALESCE(p.fx_rate,0)::float8, COALESCE(p.fx_source,''), p.fx_updated_at,
	COALESCE(p.plan_hash,''), COALESCE(p.plan_snapshot,'{}'::jsonb), COALESCE(p.status,''), COALESCE(p.description,''),
	COALESCE(p.tariff_key,''), COALESCE(p.sale_mode,''), COALESCE(p.months,0), COALESCE(p.traffic_gb,0)::float8,
	COALESCE(p.device_count,0), COALESCE(p.provider_payment_id,''), COALESCE(p.payment_url,''), COALESCE(p.qr_content,''),
	COALESCE(p.display_amount,''), COALESCE(p.display_currency,''), COALESCE(p.payment_address,''), COALESCE(p.network,''),
	COALESCE(p.url_scheme,''), COALESCE(p.raw_webhook,'{}'::jsonb), p.provisioned_at, COALESCE(p.provision_error,''), p.created_at, p.updated_at, p.paid_at
FROM payment_orders p WHERE p.payment_id=$1`, paymentID)
}

// List returns a paginated admin payment list.
func (r *Registry) List(ctx context.Context, page int, pageSize int) ([]Order, int64, error) {
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 25
	}
	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM payment_orders").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `
SELECT p.payment_id, p.order_id, p.user_id, COALESCE(u.username,''), COALESCE(u.telegram_id,0),
	COALESCE(p.provider,''), COALESCE(p.method,''), COALESCE(p.payment_type,''), p.amount::float8,
	COALESCE(p.currency,''), COALESCE(p.base_amount, p.amount)::float8, COALESCE(p.base_currency, p.currency),
	COALESCE(p.display_cny_amount,0)::float8, COALESCE(p.fx_rate,0)::float8, COALESCE(p.fx_source,''), p.fx_updated_at,
	COALESCE(p.plan_hash,''), COALESCE(p.plan_snapshot,'{}'::jsonb), COALESCE(p.status,''), COALESCE(p.description,''), COALESCE(p.tariff_key,''),
	COALESCE(p.sale_mode,''), COALESCE(p.months,0), COALESCE(p.traffic_gb,0)::float8, COALESCE(p.device_count,0),
	COALESCE(p.provider_payment_id,''), COALESCE(p.payment_url,''), COALESCE(p.qr_content,''), COALESCE(p.display_amount,''),
	COALESCE(p.display_currency,''), COALESCE(p.payment_address,''), COALESCE(p.network,''), COALESCE(p.url_scheme,''),
	COALESCE(p.raw_webhook,'{}'::jsonb), p.provisioned_at, COALESCE(p.provision_error,''), p.created_at, p.updated_at, p.paid_at
FROM payment_orders p
LEFT JOIN users u ON u.user_id = p.user_id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2`, pageSize, page*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	result := []Order{}
	for rows.Next() {
		var order Order
		var deviceCount int
		if err := rows.Scan(
			&order.PaymentID, &order.OrderID, &order.UserID, &order.UserLabel, &order.TelegramID,
			&order.Provider, &order.Method, &order.PaymentType, &order.Amount, &order.Currency,
			&order.BaseAmount, &order.BaseCurrency, &order.DisplayCNYAmount, &order.FXRate, &order.FXSource,
			&order.FXUpdatedAt, &order.PlanHash, &order.PlanSnapshot, &order.Status,
			&order.Description, &order.TariffKey, &order.SaleMode, &order.Months, &order.TrafficGB, &deviceCount,
			&order.ProviderPaymentID, &order.PaymentURL, &order.QRContent, &order.DisplayAmount, &order.DisplayCurrency,
			&order.PaymentAddress, &order.Network, &order.URLScheme, &order.RawWebhook, &order.ProvisionedAt, &order.ProvisionError,
			&order.CreatedAt, &order.UpdatedAt, &order.PaidAt,
		); err != nil {
			return nil, 0, err
		}
		result = append(result, order)
	}
	return result, total, rows.Err()
}

// HandleWebhook verifies and applies a provider webhook.
func (r *Registry) HandleWebhook(ctx context.Context, providerID string, body []byte) error {
	cfg := r.EffectiveConfig(ctx)
	var result webhookResult
	var err error
	switch providerID {
	case ProviderEZPay:
		result, err = handleEZPayWebhook(cfg.EZPay, body)
	case ProviderBEPUSDT:
		result, err = handleBEPUSDTWebhook(cfg.BEPUSDT, body)
	default:
		err = fmt.Errorf("provider not found")
	}
	if err != nil {
		return err
	}
	raw := json.RawMessage(body)
	if !json.Valid(raw) {
		formJSON, _ := json.Marshal(result.Raw)
		raw = formJSON
	}
	if result.Paid {
		_, err = r.pool.Exec(ctx, `
UPDATE payment_orders
SET status='paid', provider_payment_id=COALESCE(NULLIF($2,''), provider_payment_id), raw_webhook=$3,
	paid_at=COALESCE(paid_at, NOW()), updated_at=NOW()
WHERE order_id=$1 AND provider=$4`, result.OrderID, result.ProviderPaymentID, raw, providerID)
		return err
	}
	_, err = r.pool.Exec(ctx, `
UPDATE payment_orders
SET raw_webhook=$2, updated_at=NOW()
WHERE order_id=$1 AND provider=$3`, result.OrderID, raw, providerID)
	return err
}

func (r *Registry) createProviderPayment(ctx context.Context, cfg Config, method paymentMethod, req providerPaymentRequest) (providerPaymentResponse, error) {
	switch method.Provider {
	case ProviderEZPay:
		return createEZPayPayment(ctx, r.client, cfg.EZPay, cfg.WebhookBaseURL+"/webhook/ezpay", req, method.PaymentType)
	case ProviderBEPUSDT:
		return createBEPUSDTPayment(ctx, r.client, cfg.BEPUSDT, cfg.WebhookBaseURL+"/webhook/bepusdt", req, method.PaymentType)
	case ProviderTelegramStars:
		return createTelegramStarsInvoice(ctx, r.client, r.settings.BotToken, req)
	default:
		return providerPaymentResponse{}, fmt.Errorf("unsupported provider %s", method.Provider)
	}
}

func (r *Registry) updateProviderFields(ctx context.Context, paymentID int64, response providerPaymentResponse) error {
	var expiresAt any
	if response.ExpiresInSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(response.ExpiresInSeconds) * time.Second)
	}
	_, err := r.pool.Exec(ctx, `
UPDATE payment_orders
SET provider_payment_id=$2, payment_url=$3, qr_content=$4, display_amount=$5, display_currency=$6,
	payment_address=$7, network=$8, url_scheme=$9, expires_at=$10, updated_at=NOW()
WHERE payment_id=$1`,
		paymentID, emptyToNil(response.ProviderPaymentID), emptyToNil(response.PaymentURL), emptyToNil(response.QRContent),
		emptyToNil(response.DisplayAmount), emptyToNil(response.DisplayCurrency), emptyToNil(response.PaymentAddress),
		emptyToNil(response.Network), emptyToNil(response.URLScheme), expiresAt,
	)
	return err
}

func (r *Registry) markFailed(ctx context.Context, paymentID int64) error {
	_, err := r.pool.Exec(ctx, "UPDATE payment_orders SET status='failed', updated_at=NOW() WHERE payment_id=$1", paymentID)
	return err
}

func (r *Registry) scanOrder(ctx context.Context, query string, args ...any) (Order, error) {
	var order Order
	var deviceCount int
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&order.PaymentID, &order.OrderID, &order.UserID, &order.Provider, &order.Method, &order.PaymentType,
		&order.Amount, &order.Currency, &order.BaseAmount, &order.BaseCurrency, &order.DisplayCNYAmount,
		&order.FXRate, &order.FXSource, &order.FXUpdatedAt, &order.PlanHash, &order.PlanSnapshot,
		&order.Status, &order.Description, &order.TariffKey, &order.SaleMode,
		&order.Months, &order.TrafficGB, &deviceCount, &order.ProviderPaymentID, &order.PaymentURL, &order.QRContent,
		&order.DisplayAmount, &order.DisplayCurrency, &order.PaymentAddress, &order.Network, &order.URLScheme,
		&order.RawWebhook, &order.ProvisionedAt, &order.ProvisionError, &order.CreatedAt, &order.UpdatedAt, &order.PaidAt,
	)
	return order, err
}

type paymentMethod struct {
	Provider    string
	PaymentType string
}

func parseMethod(id string) (paymentMethod, error) {
	provider, paymentType, ok := strings.Cut(strings.TrimSpace(id), ":")
	if !ok || provider == "" || paymentType == "" {
		return paymentMethod{}, fmt.Errorf("invalid payment method")
	}
	switch provider {
	case ProviderEZPay:
		switch paymentType {
		case "alipay", "wxpay", "qqpay", "bank", "usdt":
			return paymentMethod{Provider: provider, PaymentType: paymentType}, nil
		}
	case ProviderBEPUSDT:
		switch paymentType {
		case "usdt.polygon", "usdt.arbitrum", "usdt.aptos":
			return paymentMethod{Provider: provider, PaymentType: paymentType}, nil
		}
	case ProviderTelegramStars:
		if paymentType == "xtr" {
			return paymentMethod{Provider: provider, PaymentType: paymentType}, nil
		}
	}
	return paymentMethod{}, fmt.Errorf("unsupported payment method")
}

func validateProviderConfig(cfg Config, provider string) error {
	if cfg.WebhookBaseURL == "" {
		return fmt.Errorf("WEBHOOK_BASE_URL is required for payment webhooks")
	}
	switch provider {
	case ProviderEZPay:
		if !cfg.EZPay.Enabled {
			return fmt.Errorf("EZPay is disabled")
		}
		if cfg.EZPay.BaseURL == "" || cfg.EZPay.PID == 0 || cfg.EZPay.Key == "" {
			return fmt.Errorf("EZPay is not configured")
		}
		if cfg.EZPay.ReturnURL == "" {
			return fmt.Errorf("SUBSCRIPTION_MINI_APP_URL is required")
		}
	case ProviderBEPUSDT:
		if !cfg.BEPUSDT.Enabled {
			return fmt.Errorf("BEPUSDT is disabled")
		}
		if cfg.BEPUSDT.BaseURL == "" || cfg.BEPUSDT.Token == "" {
			return fmt.Errorf("BEPUSDT is not configured")
		}
		if cfg.BEPUSDT.ReturnURL == "" {
			return fmt.Errorf("SUBSCRIPTION_MINI_APP_URL is required")
		}
	case ProviderTelegramStars:
		if !cfg.StarsEnabled {
			return fmt.Errorf("telegram Stars is disabled")
		}
	default:
		return fmt.Errorf("unsupported provider")
	}
	return nil
}

func normalizedReturnURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.TrimRight(value, "/") + "/"
}

func sortMethods(methods []Method, order []string) {
	rank := map[string]int{}
	for index, id := range order {
		rank[strings.TrimSpace(id)] = index
	}
	sort.SliceStable(methods, func(i, j int) bool {
		ri, iok := rank[methods[i].ID]
		rj, jok := rank[methods[j].ID]
		if iok && jok {
			return ri < rj
		}
		if iok {
			return true
		}
		if jok {
			return false
		}
		return methods[i].ID < methods[j].ID
	})
}

func splitList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(field); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func randomOrderID() (string, error) {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "po_" + base64.RawURLEncoding.EncodeToString(buf), nil
}

func emptyToNil(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func zeroIntToNil(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func zeroFloatToNil(value float64) any {
	if value == 0 {
		return nil
	}
	return value
}

func zeroTimeToNil(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func floatString(value float64) string {
	if value == math.Trunc(value) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}
