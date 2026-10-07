package activity

import (
	"fmt"
	"strings"
	"time"
)

// GroupMessageRewardConfig combines base settings with a server-verified count.
type GroupMessageRewardConfig struct {
	Timezone    string
	Threshold   int
	RewardMinor int64
	BoostCount  int
}

// Validate rejects invalid settings and negative trusted boost counts.
func (config GroupMessageRewardConfig) Validate() error {
	if strings.TrimSpace(config.Timezone) == "" || config.Threshold < 0 || config.RewardMinor < 0 || config.BoostCount < 0 {
		return fmt.Errorf("%w: invalid group-message reward configuration", ErrInvalidInput)
	}
	if _, err := time.LoadLocation(config.Timezone); err != nil {
		return fmt.Errorf("%w: unknown group-message reward timezone", ErrInvalidInput)
	}
	return nil
}

// GroupMessageRewardStatus contains daily progress and the authoritative reward.
type GroupMessageRewardStatus struct {
	Enabled      bool       `json:"enabled"`
	LocalDate    string     `json:"localDate"`
	MessageCount int        `json:"messageCount"`
	Threshold    int        `json:"threshold"`
	RewardMinor  int64      `json:"rewardMinor,string"`
	Rewarded     bool       `json:"rewarded"`
	RewardedAt   *time.Time `json:"rewardedAt,omitempty"`
}

// GroupMessageRewardResult records whether a delivery advanced daily progress.
type GroupMessageRewardResult struct {
	Status   GroupMessageRewardStatus
	Counted  bool
	Replayed bool
}
