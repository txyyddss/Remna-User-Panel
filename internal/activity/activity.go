// Package activity implements member games, daily check-ins, and lucky draws.
package activity

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// ErrInvalidInput indicates that an activity configuration or request is unsafe.
var ErrInvalidInput = errors.New("invalid activity input")

// RandomSource provides an unbiased integer in [0, upperBound).
type RandomSource interface {
	Int63n(upperBound int64) (int64, error)
}

// RewardKind identifies the effect of a lucky-draw prize.
type RewardKind string

const (
	// RewardNone records a draw without an additional reward.
	RewardNone RewardKind = "none"
	// RewardTXBDelta changes the member's TXB balance by a signed amount.
	RewardTXBDelta RewardKind = "txb_delta"
	// RewardCouponGrant adds a coupon grant to the member's wallet.
	RewardCouponGrant RewardKind = "coupon_grant"
	// RewardSubscriptionExtension extends the current or next subscription.
	RewardSubscriptionExtension RewardKind = "subscription_extension"
)

// Reward is the typed payload applied by a lucky-draw prize.
type Reward struct {
	Kind          RewardKind `json:"kind"`
	TXBDeltaMinor int64      `json:"txbDeltaMinor,omitempty"`
	CouponID      string     `json:"couponId,omitempty"`
	ExtensionDays int        `json:"extensionDays,omitempty"`
}

// Validate rejects ambiguous or unsafe reward payloads.
func (reward Reward) Validate() error {
	switch reward.Kind {
	case RewardNone:
		if reward.TXBDeltaMinor != 0 || reward.CouponID != "" || reward.ExtensionDays != 0 {
			return fmt.Errorf("%w: no-prize reward has a payload", ErrInvalidInput)
		}
	case RewardTXBDelta:
		if reward.TXBDeltaMinor == 0 || reward.TXBDeltaMinor == math.MinInt64 || reward.CouponID != "" || reward.ExtensionDays != 0 {
			return fmt.Errorf("%w: TXB reward must contain only a non-zero delta", ErrInvalidInput)
		}
	case RewardCouponGrant:
		if strings.TrimSpace(reward.CouponID) == "" || reward.TXBDeltaMinor != 0 || reward.ExtensionDays != 0 {
			return fmt.Errorf("%w: coupon reward must contain only a coupon ID", ErrInvalidInput)
		}
	case RewardSubscriptionExtension:
		if reward.ExtensionDays < 1 || reward.ExtensionDays > 3650 || reward.TXBDeltaMinor != 0 || reward.CouponID != "" {
			return fmt.Errorf("%w: extension reward must contain 1 to 3650 days", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: unsupported reward kind %q", ErrInvalidInput, reward.Kind)
	}
	return nil
}

// GameInput is the administrator-authored game configuration.
type GameInput struct {
	ID                  string `json:"id,omitempty"`
	Name                string `json:"name"`
	Icon                string `json:"icon"`
	Description         string `json:"description"`
	Enabled             bool   `json:"enabled"`
	WinChanceBPS        int    `json:"winChanceBps"`
	MinimumStakeMinor   int64  `json:"minimumStakeMinor"`
	MaximumStakeMinor   int64  `json:"maximumStakeMinor"`
	ReturnMultiplierBPS int64  `json:"returnMultiplierBps"`
}

var allowedGameIcons = map[string]struct{}{
	"dice": {}, "coin": {}, "cards": {}, "target": {}, "trophy": {}, "lightning": {}, "sparkle": {},
}

// Validate rejects an invalid betting-game configuration.
func (input GameInput) Validate() error {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 80 {
		return fmt.Errorf("%w: game name must be 1 to 80 bytes", ErrInvalidInput)
	}
	if _, ok := allowedGameIcons[strings.TrimSpace(input.Icon)]; !ok {
		return fmt.Errorf("%w: unsupported game icon", ErrInvalidInput)
	}
	if len(strings.TrimSpace(input.Description)) > 4_000 {
		return fmt.Errorf("%w: game description is too long", ErrInvalidInput)
	}
	if input.WinChanceBPS < 0 || input.WinChanceBPS > 10_000 {
		return fmt.Errorf("%w: win chance must be between 0 and 10000 basis points", ErrInvalidInput)
	}
	if input.MinimumStakeMinor <= 0 || input.MaximumStakeMinor < input.MinimumStakeMinor {
		return fmt.Errorf("%w: invalid stake range", ErrInvalidInput)
	}
	if input.ReturnMultiplierBPS < 10_000 || input.ReturnMultiplierBPS > 1_000_000 {
		return fmt.Errorf("%w: total-return multiplier must be between 1x and 100x", ErrInvalidInput)
	}
	return nil
}

// Game is a persisted betting-game configuration.
type Game struct {
	GameInput
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BetResult is an immutable game outcome and its final balance.
type BetResult struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"userId,omitempty"`
	GameID                string    `json:"gameId"`
	StakeMinor            int64     `json:"stakeMinor"`
	Won                   bool      `json:"won"`
	PayoutMinor           int64     `json:"payoutMinor"`
	BalanceAfterMinor     int64     `json:"balanceAfterMinor"`
	ConfigurationSnapshot string    `json:"configurationSnapshot"`
	IdempotencyKey        string    `json:"idempotencyKey,omitempty"`
	Replayed              bool      `json:"replayed"`
	CreatedAt             time.Time `json:"createdAt"`
}

// CheckInConfig contains the server-owned daily reward and timezone.
type CheckInConfig struct {
	Timezone                string
	RewardMinMinor          int64
	RewardMaxMinor          int64
	GroupMessageThreshold   int
	GroupMessageRewardMinor int64
}

// DailyCheckIn is one member's reward for a local calendar date.
type DailyCheckIn struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId,omitempty"`
	LocalDate         string    `json:"localDate"`
	Timezone          string    `json:"timezone"`
	RewardMinor       int64     `json:"rewardMinor"`
	BalanceAfterMinor int64     `json:"balanceAfterMinor"`
	AlreadyClaimed    bool      `json:"alreadyClaimed"`
	CreatedAt         time.Time `json:"createdAt"`
}
