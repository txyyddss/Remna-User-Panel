// Package affiliates owns referral eligibility, tier configuration, and member projections.
package affiliates

import "time"

const (
	LocaleEnglish = "en"
	LocaleChinese = "zh-CN"
	PageSize      = 5
)

type Reward struct {
	Kind          string `json:"kind"`
	CouponID      string `json:"couponId,omitempty"`
	CouponName    string `json:"couponName,omitempty"`
	TXBMinor      int64  `json:"txbMinor,omitempty"`
	ExtensionDays int    `json:"extensionDays,omitempty"`
}

type Tier struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Threshold         int    `json:"threshold"`
	Enabled           bool   `json:"enabled"`
	CommissionEnabled bool   `json:"commissionEnabled"`
	CommissionBPS     int    `json:"commissionBps"`
	Reward            Reward `json:"reward"`
}

type Config struct {
	ID      string `json:"-"`
	Version int    `json:"version"`
	Tiers   []Tier `json:"tiers"`
}

type BotIdentity struct {
	Username  string     `json:"username,omitempty"`
	Status    string     `json:"status"`
	CheckedAt *time.Time `json:"checkedAt,omitempty"`
}

type Money struct {
	Minor    string `json:"minor"`
	Currency string `json:"currency"`
	Display  string `json:"display"`
}
