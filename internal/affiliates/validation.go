package affiliates

import (
	"errors"
	"strings"
)

var (
	ErrInvalidInput    = errors.New("invalid affiliate configuration")
	ErrVersionConflict = errors.New("affiliate configuration version conflict")
)

func NormalizeLocale(language string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(language)), "zh") {
		return LocaleChinese
	}
	return LocaleEnglish
}

func ValidateTiers(tiers []Tier) error {
	if len(tiers) == 0 || len(tiers) > 50 {
		return ErrInvalidInput
	}
	last, enabled := -1, 0
	for i := range tiers {
		tier := tiers[i]
		if strings.TrimSpace(tier.Name) == "" || len([]rune(tier.Name)) > 48 || tier.Threshold < 0 || tier.CommissionBPS < 0 || tier.CommissionBPS > 10000 {
			return ErrInvalidInput
		}
		if !tier.CommissionEnabled && tier.CommissionBPS != 0 || validateReward(tier.Reward) != nil {
			return ErrInvalidInput
		}
		if !tier.Enabled {
			continue
		}
		if tier.Threshold <= last || (enabled == 0 && tier.Threshold != 0) || (tier.Threshold == 0 && tier.Reward.Kind != "none") {
			return ErrInvalidInput
		}
		last, enabled = tier.Threshold, enabled+1
	}
	if enabled == 0 {
		return ErrInvalidInput
	}
	return nil
}

func validateReward(reward Reward) error {
	switch reward.Kind {
	case "none":
		if reward.CouponID != "" || reward.TXBMinor != 0 || reward.ExtensionDays != 0 {
			return ErrInvalidInput
		}
	case "coupon":
		if strings.TrimSpace(reward.CouponID) == "" || reward.TXBMinor != 0 || reward.ExtensionDays != 0 {
			return ErrInvalidInput
		}
	case "txb":
		if reward.TXBMinor <= 0 || reward.TXBMinor > 1_000_000_000_00 || reward.CouponID != "" || reward.ExtensionDays != 0 {
			return ErrInvalidInput
		}
	case "subscription_extension":
		if reward.ExtensionDays < 1 || reward.ExtensionDays > 3650 || reward.CouponID != "" || reward.TXBMinor != 0 {
			return ErrInvalidInput
		}
	default:
		return ErrInvalidInput
	}
	return nil
}
