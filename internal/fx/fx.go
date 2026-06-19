// Package fx resolves and caches fiat exchange rates used for display and checkout.
package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	appsettings "remna-user-panel/internal/settings"
)

const (
	defaultProvider = "frankfurter"
	defaultRate     = 7.20
)

// Rate is a fiat conversion snapshot.
type Rate struct {
	Base      string    `json:"base"`
	Quote     string    `json:"quote"`
	Rate      float64   `json:"rate"`
	Source    string    `json:"source"`
	UpdatedAt time.Time `json:"updated_at"`
	Stale     bool      `json:"stale,omitempty"`
}

// Service reads exchange-rate settings and caches successful USD/CNY responses.
type Service struct {
	store  appsettings.Store
	client *http.Client
	now    func() time.Time
}

// NewService creates an exchange-rate service.
func NewService(store appsettings.Store) *Service {
	return &Service{
		store:  store,
		client: &http.Client{Timeout: 8 * time.Second},
		now:    time.Now,
	}
}

// USDCNY returns the current USD to CNY conversion rate.
func (s *Service) USDCNY(ctx context.Context) Rate {
	provider := strings.ToLower(strings.TrimSpace(s.store.String(ctx, "FX_PROVIDER", defaultProvider)))
	if provider == "" {
		provider = defaultProvider
	}
	if provider == "custom" {
		rate := parsePositiveFloat(s.store.String(ctx, "FX_CUSTOM_USD_CNY", ""), 0)
		if rate > 0 {
			return Rate{Base: "USD", Quote: "CNY", Rate: rate, Source: "custom", UpdatedAt: s.now()}
		}
	}
	ttl := time.Duration(s.store.Int(ctx, "FX_CACHE_TTL_SECONDS", 3600)) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	if cached, ok := s.cached(ctx, ttl); ok {
		return cached
	}
	if provider != "exchange_rate_api" {
		provider = defaultProvider
	}
	rate, err := s.fetch(ctx, provider)
	if err == nil && rate.Rate > 0 {
		_ = s.store.Upsert(ctx, "FX_USD_CNY_RATE", rate.Rate)
		_ = s.store.Upsert(ctx, "FX_USD_CNY_SOURCE", rate.Source)
		_ = s.store.Upsert(ctx, "FX_USD_CNY_UPDATED_AT", rate.UpdatedAt.Format(time.RFC3339))
		return rate
	}
	if cached, ok := s.cached(ctx, 0); ok {
		cached.Stale = true
		return cached
	}
	return Rate{Base: "USD", Quote: "CNY", Rate: defaultRate, Source: "fallback", UpdatedAt: s.now(), Stale: true}
}

func (s *Service) cached(ctx context.Context, ttl time.Duration) (Rate, bool) {
	rate := parsePositiveFloat(s.store.String(ctx, "FX_USD_CNY_RATE", ""), 0)
	if rate <= 0 {
		return Rate{}, false
	}
	source := s.store.String(ctx, "FX_USD_CNY_SOURCE", "cache")
	updatedAt := parseTime(s.store.String(ctx, "FX_USD_CNY_UPDATED_AT", ""))
	if updatedAt.IsZero() {
		updatedAt = s.now()
	}
	if ttl > 0 && s.now().Sub(updatedAt) > ttl {
		return Rate{}, false
	}
	return Rate{Base: "USD", Quote: "CNY", Rate: rate, Source: source, UpdatedAt: updatedAt}, true
}

func (s *Service) fetch(ctx context.Context, provider string) (Rate, error) {
	switch provider {
	case "exchange_rate_api":
		return s.fetchExchangeRateAPI(ctx)
	default:
		return s.fetchFrankfurter(ctx)
	}
}

func (s *Service) fetchFrankfurter(ctx context.Context) (Rate, error) {
	var payload struct {
		Rates map[string]float64 `json:"rates"`
		Date  string             `json:"date"`
	}
	if err := s.fetchJSON(ctx, "https://api.frankfurter.app/latest?from=USD&to=CNY", &payload); err != nil {
		return Rate{}, err
	}
	rate := payload.Rates["CNY"]
	if rate <= 0 {
		return Rate{}, fmt.Errorf("frankfurter response missing CNY rate")
	}
	return Rate{Base: "USD", Quote: "CNY", Rate: rate, Source: "frankfurter", UpdatedAt: s.now()}, nil
}

func (s *Service) fetchExchangeRateAPI(ctx context.Context) (Rate, error) {
	var payload struct {
		Result string             `json:"result"`
		Rates  map[string]float64 `json:"rates"`
	}
	if err := s.fetchJSON(ctx, "https://open.er-api.com/v6/latest/USD", &payload); err != nil {
		return Rate{}, err
	}
	if payload.Result != "" && !strings.EqualFold(payload.Result, "success") {
		return Rate{}, fmt.Errorf("exchange rate api result %q", payload.Result)
	}
	rate := payload.Rates["CNY"]
	if rate <= 0 {
		return Rate{}, fmt.Errorf("exchange rate api response missing CNY rate")
	}
	return Rate{Base: "USD", Quote: "CNY", Rate: rate, Source: "exchange_rate_api", UpdatedAt: s.now()}, nil
}

func (s *Service) fetchJSON(ctx context.Context, url string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("fx provider returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func parsePositiveFloat(raw string, fallback float64) float64 {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseTime(raw string) time.Time {
	value, err := time.Parse(time.RFC3339, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}
	}
	return value
}
