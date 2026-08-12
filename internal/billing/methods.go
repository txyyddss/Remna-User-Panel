package billing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

var ezpayRails = map[string]string{
	"alipay": "Alipay",
	"wxpay":  "WeChat Pay",
	"qqpay":  "QQ Wallet",
	"bank":   "UnionPay",
	"jdpay":  "JD Pay",
}

var bepusdtRails = map[string]string{
	"usdt.trc20":    "USDT TRC20",
	"usdt.erc20":    "USDT ERC20",
	"usdt.polygon":  "USDT Polygon",
	"usdt.bep20":    "USDT BEP20",
	"usdt.aptos":    "USDT Aptos",
	"usdt.solana":   "USDT Solana",
	"usdt.xlayer":   "USDT X-Layer",
	"usdt.arbitrum": "USDT Arbitrum One",
	"usdt.plasma":   "USDT Plasma",
	"usdt.ton":      "USDT TON",
}

var paymentChannelOrder = map[string][]string{
	"ezpay":   {"alipay", "wxpay", "qqpay", "bank", "jdpay"},
	"bepusdt": {"usdt.trc20", "usdt.erc20", "usdt.polygon", "usdt.bep20", "usdt.aptos", "usdt.solana", "usdt.xlayer", "usdt.arbitrum", "usdt.plasma", "usdt.ton"},
}

// PaymentChannels returns the stable channel order accepted by a provider.
func PaymentChannels(provider string) []string {
	return append([]string(nil), paymentChannelOrder[provider]...)
}

// ValidatePaymentChannels validates the independently enabled provider channels.
func ValidatePaymentChannels(provider string, channels []string) error {
	allowed := ezpayRails
	if provider == "bepusdt" {
		allowed = bepusdtRails
	}
	if provider != "ezpay" && provider != "bepusdt" {
		return fmt.Errorf("unsupported provider %q", provider)
	}
	seen := make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		channel = strings.ToLower(strings.TrimSpace(channel))
		if _, ok := allowed[channel]; !ok {
			return fmt.Errorf("unsupported channel %q", channel)
		}
		if _, ok := seen[channel]; ok {
			return fmt.Errorf("duplicate channel %q", channel)
		}
		seen[channel] = struct{}{}
	}
	return nil
}

// ParseMethodID validates a stable payment method identifier and returns its
// provider-owned rail. Stars intentionally has no sub-rail.
func ParseMethodID(methodID string) (provider, rail string, err error) {
	methodID = strings.ToLower(strings.TrimSpace(methodID))
	// Internal compatibility aliases keep older workers/tests readable. The HTTP
	// transport accepts only canonical IDs and never exposes these aliases.
	if methodID == "ezpay" {
		return "ezpay", "alipay", nil
	}
	if methodID == "bepusdt" {
		return "bepusdt", "usdt.trc20", nil
	}
	if methodID == "stars" {
		return "stars", "", nil
	}
	if methodID == "coupon" {
		return "coupon", "", nil
	}
	parts := strings.Split(methodID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%w: malformed method id", ErrInvalidOrder)
	}
	switch parts[0] {
	case "ezpay":
		if _, ok := ezpayRails[parts[1]]; !ok {
			return "", "", fmt.Errorf("%w: unsupported EZPay rail", ErrInvalidOrder)
		}
	case "bepusdt":
		if _, ok := bepusdtRails[parts[1]]; !ok {
			return "", "", fmt.Errorf("%w: unsupported BEPusdt rail", ErrInvalidOrder)
		}
	default:
		return "", "", fmt.Errorf("%w: unsupported provider", ErrInvalidOrder)
	}
	return parts[0], parts[1], nil
}

// CanonicalMethodID reports whether the public identifier is stable and fully
// qualifies its provider rail.
func CanonicalMethodID(methodID string) bool {
	methodID = strings.ToLower(strings.TrimSpace(methodID))
	return methodID == "stars" || methodID == "coupon" || strings.HasPrefix(methodID, "ezpay:") || strings.HasPrefix(methodID, "bepusdt:")
}

func parseEnabledRails(raw string, allowed map[string]string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("at least one rail is required")
	}
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		value = strings.ToLower(strings.TrimSpace(value))
		if _, ok := allowed[value]; !ok {
			return nil, fmt.Errorf("unsupported rail %q", value)
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, fmt.Errorf("duplicate rail %q", value)
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func methodName(provider, rail string) string {
	if provider == "coupon" {
		return "Coupon"
	}
	if provider == "stars" {
		return "Telegram Stars"
	}
	if provider == "ezpay" {
		return ezpayRails[rail]
	}
	return bepusdtRails[rail]
}

func methodModel(provider, rail string, available bool, note string) model.PaymentMethod {
	id := provider
	if rail != "" {
		id += ":" + rail
	}
	return model.PaymentMethod{
		ID: id, Provider: provider, Rail: rail, Name: methodName(provider, rail),
		Currency: strings.ToUpper(currencyCode(provider)), Available: available, Note: note, Mode: "order",
	}
}

func containsRail(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func validateEnabledRails(value string, allowed map[string]string) error {
	_, err := parseEnabledRails(value, allowed)
	return err
}

// ValidateEZPayMethods validates an ordered comma-separated EZPay rail list.
func ValidateEZPayMethods(value string) error { return validateEnabledRails(value, ezpayRails) }

// ValidateBEPusdtMethods validates an ordered comma-separated USDT rail list.
func ValidateBEPusdtMethods(value string) error { return validateEnabledRails(value, bepusdtRails) }

var errRateNotConfigured = errors.New("payment rate is not configured")
