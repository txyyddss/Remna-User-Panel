package activity

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type PrizeInput struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Weight         int64  `json:"weight"`
	StockRemaining *int64 `json:"stockRemaining,omitempty"`
	Reward         Reward `json:"reward"`
}
type Prize struct {
	PrizeInput
	Position int `json:"position"`
}
type LuckyDrawInput struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	FeeMinor    int64        `json:"feeMinor"`
	Prizes      []PrizeInput `json:"prizes"`
}

func (input LuckyDrawInput) Validate() error {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 80 {
		return fmt.Errorf("%w: draw name must be 1 to 80 bytes", ErrInvalidInput)
	}
	if len(strings.TrimSpace(input.Description)) > 4_000 {
		return fmt.Errorf("%w: draw description is too long", ErrInvalidInput)
	}
	if input.FeeMinor < 0 {
		return fmt.Errorf("%w: draw fee cannot be negative", ErrInvalidInput)
	}
	if len(input.Prizes) == 0 || len(input.Prizes) > 200 {
		return fmt.Errorf("%w: draw must contain 1 to 200 prizes", ErrInvalidInput)
	}
	var totalWeight int64
	for index, prize := range input.Prizes {
		if strings.TrimSpace(prize.Name) == "" || len(strings.TrimSpace(prize.Name)) > 80 {
			return fmt.Errorf("%w: prize %d has an invalid name", ErrInvalidInput, index+1)
		}
		if prize.Weight <= 0 || totalWeight > (1<<63-1)-prize.Weight {
			return fmt.Errorf("%w: prize %d has an invalid weight", ErrInvalidInput, index+1)
		}
		totalWeight += prize.Weight
		if prize.StockRemaining != nil && *prize.StockRemaining < 0 {
			return fmt.Errorf("%w: prize %d has negative stock", ErrInvalidInput, index+1)
		}
		if err := prize.Reward.Validate(); err != nil {
			return fmt.Errorf("prize %d: %w", index+1, err)
		}
	}
	return nil
}

func (input LuckyDrawInput) MaximumPrizeDeduction() int64 {
	var maximum int64
	for _, prize := range input.Prizes {
		if prize.Reward.Kind == RewardTXBDelta && prize.Reward.TXBDeltaMinor < 0 {
			if prize.Reward.TXBDeltaMinor == math.MinInt64 {
				return math.MaxInt64
			}
			if -prize.Reward.TXBDeltaMinor > maximum {
				maximum = -prize.Reward.TXBDeltaMinor
			}
		}
	}
	return maximum
}

type LuckyDraw struct {
	LuckyDrawInput
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type DrawResult struct {
	ID                    string    `json:"id"`
	UserID                string    `json:"userId,omitempty"`
	DrawID                string    `json:"drawId"`
	PrizeID               string    `json:"prizeId"`
	PrizeName             string    `json:"prizeName"`
	FeeMinor              int64     `json:"feeMinor"`
	Reward                Reward    `json:"reward"`
	BalanceAfterMinor     int64     `json:"balanceAfterMinor"`
	ConfigurationSnapshot string    `json:"configurationSnapshot"`
	IdempotencyKey        string    `json:"idempotencyKey,omitempty"`
	Replayed              bool      `json:"replayed"`
	CreatedAt             time.Time `json:"createdAt"`
}
type ExtensionCredit struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"userId,omitempty"`
	Days                 int        `json:"days"`
	SourceType           string     `json:"sourceType"`
	SourceID             string     `json:"sourceId"`
	CreatedAt            time.Time  `json:"createdAt"`
	ConsumedAt           *time.Time `json:"consumedAt,omitempty"`
	ConsumedByPurchaseID *string    `json:"consumedByPurchaseId,omitempty"`
}
type History struct {
	Bets     []BetResult    `json:"bets"`
	CheckIns []DailyCheckIn `json:"checkIns"`
	Draws    []DrawResult   `json:"draws"`
}

type GroupMessageRewardConfig struct {
	Timezone    string
	Threshold   int
	RewardMinor int64
}

func (config GroupMessageRewardConfig) Validate() error {
	if strings.TrimSpace(config.Timezone) == "" || config.Threshold < 0 || config.RewardMinor < 0 {
		return fmt.Errorf("%w: invalid group-message reward configuration", ErrInvalidInput)
	}
	if _, err := time.LoadLocation(config.Timezone); err != nil {
		return fmt.Errorf("%w: unknown group-message reward timezone", ErrInvalidInput)
	}
	return nil
}

type GroupMessageRewardStatus struct {
	Enabled      bool       `json:"enabled"`
	LocalDate    string     `json:"localDate"`
	MessageCount int        `json:"messageCount"`
	Threshold    int        `json:"threshold"`
	RewardMinor  int64      `json:"rewardMinor,string"`
	Rewarded     bool       `json:"rewarded"`
	RewardedAt   *time.Time `json:"rewardedAt,omitempty"`
}
type GroupMessageRewardResult struct {
	Status   GroupMessageRewardStatus
	Counted  bool
	Replayed bool
}
