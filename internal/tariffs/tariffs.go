// Package tariffs loads the JSON tariff catalog used by the Web App.
package tariffs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Catalog is the on-disk tariff catalog.
type Catalog struct {
	DefaultTariff   string   `json:"default_tariff"`
	DefaultCurrency string   `json:"default_currency"`
	Tariffs         []Tariff `json:"tariffs"`
}

// Tariff is one product family in the catalog.
type Tariff struct {
	Key          string            `json:"key"`
	Names        map[string]string `json:"names"`
	Descriptions map[string]string `json:"descriptions"`
	BillingModel string            `json:"billing_model"`
	MonthlyGB    float64           `json:"monthly_gb"`
	Enabled      bool              `json:"enabled"`
	raw          map[string]json.RawMessage
}

// Plan is the Web App purchase option shape.
type Plan struct {
	ID           string  `json:"id"`
	TariffKey    string  `json:"tariff_key"`
	TariffName   string  `json:"tariff_name"`
	Title        string  `json:"title"`
	Description  string  `json:"description,omitempty"`
	BillingModel string  `json:"billing_model"`
	SaleMode     string  `json:"sale_mode"`
	Months       int     `json:"months,omitempty"`
	TrafficGB    float64 `json:"traffic_gb,omitempty"`
	Price        float64 `json:"price"`
	Currency     string  `json:"currency"`
	MonthlyGB    float64 `json:"monthly_gb,omitempty"`
	IsDefault    bool    `json:"is_default_tariff,omitempty"`
}

// PaymentSelection identifies the plan a user chose.
type PaymentSelection struct {
	TariffKey string
	SaleMode  string
	Months    int
	TrafficGB float64
}

// UnmarshalJSON preserves unknown price/package fields for dynamic currency lookup.
func (t *Tariff) UnmarshalJSON(data []byte) error {
	type alias Tariff
	var dst alias
	if err := json.Unmarshal(data, &dst); err != nil {
		return err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*t = Tariff(dst)
	t.raw = raw
	return nil
}

// Load reads a catalog. A missing file is treated as an empty catalog.
func Load(path string) (Catalog, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Catalog{}, nil
	}
	if err != nil {
		return Catalog{}, fmt.Errorf("read tariffs: %w", err)
	}
	var catalog Catalog
	if err := json.Unmarshal(body, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("parse tariffs: %w", err)
	}
	return catalog, nil
}

// Plans returns enabled purchase plans for a display language and currency.
func (c Catalog) Plans(language string, fallbackCurrency string) []Plan {
	currency := normalizedCurrency(c.DefaultCurrency)
	if currency == "" {
		currency = normalizedCurrency(fallbackCurrency)
	}
	if currency == "" {
		currency = "rub"
	}
	result := []Plan{}
	for _, tariff := range c.Tariffs {
		if !tariff.Enabled || tariff.Key == "" {
			continue
		}
		name := localized(tariff.Names, language, tariff.Key)
		description := localized(tariff.Descriptions, language, "")
		model := strings.ToLower(strings.TrimSpace(tariff.BillingModel))
		if model == "" {
			model = "period"
		}
		if model == "traffic" {
			for _, pkg := range packagePrices(tariff.raw["traffic_packages"], currency) {
				result = append(result, Plan{
					ID:           fmt.Sprintf("%s:traffic:%s", tariff.Key, compactNumber(pkg.Amount)),
					TariffKey:    tariff.Key,
					TariffName:   name,
					Title:        name,
					Description:  description,
					BillingModel: "traffic",
					SaleMode:     "traffic_package",
					Months:       int(pkg.Amount),
					TrafficGB:    pkg.Amount,
					Price:        pkg.Price,
					Currency:     strings.ToUpper(currency),
					MonthlyGB:    tariff.MonthlyGB,
					IsDefault:    tariff.Key == c.DefaultTariff,
				})
			}
			continue
		}
		for _, period := range periodPrices(tariff.raw["prices_"+currency]) {
			result = append(result, Plan{
				ID:           fmt.Sprintf("%s:subscription:%d", tariff.Key, period.Months),
				TariffKey:    tariff.Key,
				TariffName:   name,
				Title:        name,
				Description:  description,
				BillingModel: "period",
				SaleMode:     "subscription",
				Months:       period.Months,
				Price:        period.Price,
				Currency:     strings.ToUpper(currency),
				MonthlyGB:    tariff.MonthlyGB,
				IsDefault:    tariff.Key == c.DefaultTariff,
			})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].TariffKey != result[j].TariffKey {
			return result[i].TariffKey < result[j].TariffKey
		}
		if result[i].SaleMode != result[j].SaleMode {
			return result[i].SaleMode < result[j].SaleMode
		}
		if result[i].Months != result[j].Months {
			return result[i].Months < result[j].Months
		}
		return result[i].TrafficGB < result[j].TrafficGB
	})
	return result
}

// FindPlan returns the server-trusted plan matching a user selection.
func (c Catalog) FindPlan(selection PaymentSelection, language string, fallbackCurrency string) (Plan, bool) {
	saleMode := strings.ToLower(strings.TrimSpace(selection.SaleMode))
	if saleMode == "" {
		saleMode = "subscription"
	}
	if saleMode == "topup" || saleMode == "premium_topup" || saleMode == "traffic" {
		saleMode = "traffic_package"
	}
	for _, plan := range c.Plans(language, fallbackCurrency) {
		if selection.TariffKey != "" && plan.TariffKey != selection.TariffKey {
			continue
		}
		if plan.SaleMode != saleMode {
			continue
		}
		if plan.SaleMode == "traffic_package" {
			if almostEqual(plan.TrafficGB, selection.TrafficGB) || almostEqual(plan.TrafficGB, float64(selection.Months)) {
				return plan, true
			}
			continue
		}
		if plan.Months == selection.Months {
			return plan, true
		}
	}
	return Plan{}, false
}

type periodPrice struct {
	Months int
	Price  float64
}

type packagePrice struct {
	Amount float64
	Price  float64
}

func periodPrices(raw json.RawMessage) []periodPrice {
	values := map[string]float64{}
	if len(raw) == 0 || json.Unmarshal(raw, &values) != nil {
		return nil
	}
	result := make([]periodPrice, 0, len(values))
	for key, price := range values {
		months, err := strconv.Atoi(key)
		if err == nil && months > 0 && price > 0 {
			result = append(result, periodPrice{Months: months, Price: price})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Months < result[j].Months })
	return result
}

func packagePrices(raw json.RawMessage, currency string) []packagePrice {
	values := map[string][]struct {
		GB    float64 `json:"gb"`
		Count float64 `json:"count"`
		Price float64 `json:"price"`
	}{}
	if len(raw) == 0 || json.Unmarshal(raw, &values) != nil {
		return nil
	}
	result := []packagePrice{}
	for _, item := range values[currency] {
		amount := item.GB
		if amount == 0 {
			amount = item.Count
		}
		if amount > 0 && item.Price > 0 {
			result = append(result, packagePrice{Amount: amount, Price: item.Price})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Amount < result[j].Amount })
	return result
}

func localized(values map[string]string, language string, fallback string) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	for _, candidate := range []string{lang, strings.Split(lang, "-")[0], "zh", "en"} {
		if value := strings.TrimSpace(values[candidate]); value != "" {
			return value
		}
	}
	return fallback
}

func normalizedCurrency(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func compactNumber(value float64) string {
	if almostEqual(value, float64(int64(value))) {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func almostEqual(a float64, b float64) bool {
	if a > b {
		return a-b < 0.000001
	}
	return b-a < 0.000001
}
