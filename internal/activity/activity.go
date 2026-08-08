// Package activity implements member games, daily check-ins, and lucky draws.
package activity

import (
	"context"
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

// Validate rejects an invalid betting-game configuration.
func (input GameInput) Validate() error {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(name) > 80 {
		return fmt.Errorf("%w: game name must be 1 to 80 bytes", ErrInvalidInput)
	}
	if len(strings.TrimSpace(input.Icon)) > 64 {
		return fmt.Errorf("%w: game icon is too long", ErrInvalidInput)
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
	Timezone    string
	RewardMinor int64
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

// PrizeInput is one weighted entry in a lucky draw.
type PrizeInput struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name"`
	Weight         int64  `json:"weight"`
	StockRemaining *int64 `json:"stockRemaining,omitempty"`
	Reward         Reward `json:"reward"`
}

// Prize is a persisted lucky-draw entry.
type Prize struct {
	PrizeInput
	Position int `json:"position"`
}

// LuckyDrawInput is the complete replace-on-save lucky-draw configuration.
type LuckyDrawInput struct {
	ID          string       `json:"id,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Enabled     bool         `json:"enabled"`
	FeeMinor    int64        `json:"feeMinor"`
	Prizes      []PrizeInput `json:"prizes"`
}

// Validate rejects an invalid lucky-draw configuration.
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

// MaximumPrizeDeduction returns the largest possible negative TXB reward.
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

// LuckyDraw is a persisted draw and its ordered prize pool.
type LuckyDraw struct {
	LuckyDrawInput
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DrawResult is an immutable lucky-draw outcome.
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

// ExtensionCredit is a subscription extension waiting for the next purchase.
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

// History is the bounded member-visible Activity result stream.
type History struct {
	Bets     []BetResult    `json:"bets"`
	CheckIns []DailyCheckIn `json:"checkIns"`
	Draws    []DrawResult   `json:"draws"`
}

// Store is the narrow persistence contract required by Service.
type Store interface {
	SaveActivityGame(context.Context, GameInput, time.Time) (Game, error)
	ListActivityGames(context.Context, bool) ([]Game, error)
	PlaceActivityBet(context.Context, string, string, int64, string, RandomSource, time.Time) (BetResult, error)
	ClaimDailyActivity(context.Context, string, string, string, int64, time.Time) (DailyCheckIn, error)
	SaveLuckyDraw(context.Context, LuckyDrawInput, time.Time) (LuckyDraw, error)
	ListLuckyDraws(context.Context, bool) ([]LuckyDraw, error)
	PlayLuckyDraw(context.Context, string, string, string, RandomSource, time.Time) (DrawResult, error)
	ListActivityHistory(context.Context, string, int) (History, error)
}

// Service applies request validation, trusted time, and cryptographic randomness.
type Service struct {
	store Store
	rng   RandomSource
	now   func() time.Time
}

// NewService constructs the Activity application service.
func NewService(store Store, rng RandomSource, now func() time.Time) *Service {
	if rng == nil {
		rng = CryptoRandom{}
	}
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, rng: rng, now: now}
}

// SaveGame validates and persists one administrator-authored game.
func (service *Service) SaveGame(ctx context.Context, input GameInput) (Game, error) {
	if err := input.Validate(); err != nil {
		return Game{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Icon = strings.TrimSpace(input.Icon)
	input.Description = strings.TrimSpace(input.Description)
	return service.store.SaveActivityGame(ctx, input, service.now().UTC())
}

// Games returns enabled member games or all administrator-visible games.
func (service *Service) Games(ctx context.Context, enabledOnly bool) ([]Game, error) {
	return service.store.ListActivityGames(ctx, enabledOnly)
}

// PlayBet atomically debits a stake and persists the random outcome.
func (service *Service) PlayBet(ctx context.Context, userID, gameID string, stakeMinor int64, idempotencyKey string) (BetResult, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(gameID) == "" || stakeMinor <= 0 || !validIdempotencyKey(idempotencyKey) {
		return BetResult{}, fmt.Errorf("%w: incomplete bet request", ErrInvalidInput)
	}
	return service.store.PlaceActivityBet(ctx, userID, gameID, stakeMinor, idempotencyKey, service.rng, service.now().UTC())
}

// CheckIn claims the configured reward for the current local calendar date.
func (service *Service) CheckIn(ctx context.Context, userID string, config CheckInConfig) (DailyCheckIn, error) {
	if strings.TrimSpace(userID) == "" || config.RewardMinor < 0 {
		return DailyCheckIn{}, fmt.Errorf("%w: invalid check-in request", ErrInvalidInput)
	}
	location, err := time.LoadLocation(strings.TrimSpace(config.Timezone))
	if err != nil {
		return DailyCheckIn{}, fmt.Errorf("%w: unknown timezone", ErrInvalidInput)
	}
	now := service.now()
	localDate := now.In(location).Format(time.DateOnly)
	return service.store.ClaimDailyActivity(ctx, userID, localDate, location.String(), config.RewardMinor, now.UTC())
}

// SaveDraw validates and atomically replaces one lucky-draw configuration.
func (service *Service) SaveDraw(ctx context.Context, input LuckyDrawInput) (LuckyDraw, error) {
	if err := input.Validate(); err != nil {
		return LuckyDraw{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	for index := range input.Prizes {
		input.Prizes[index].Name = strings.TrimSpace(input.Prizes[index].Name)
		input.Prizes[index].Reward.CouponID = strings.TrimSpace(input.Prizes[index].Reward.CouponID)
	}
	return service.store.SaveLuckyDraw(ctx, input, service.now().UTC())
}

// Draws returns enabled member draws or all administrator-visible draws.
func (service *Service) Draws(ctx context.Context, enabledOnly bool) ([]LuckyDraw, error) {
	return service.store.ListLuckyDraws(ctx, enabledOnly)
}

// Draw atomically charges a fee and applies the selected typed reward.
func (service *Service) Draw(ctx context.Context, userID, drawID, idempotencyKey string) (DrawResult, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(drawID) == "" || !validIdempotencyKey(idempotencyKey) {
		return DrawResult{}, fmt.Errorf("%w: incomplete lucky-draw request", ErrInvalidInput)
	}
	return service.store.PlayLuckyDraw(ctx, userID, drawID, idempotencyKey, service.rng, service.now().UTC())
}

// History returns a bounded member Activity stream.
func (service *Service) History(ctx context.Context, userID string, limit int) (History, error) {
	if strings.TrimSpace(userID) == "" {
		return History{}, fmt.Errorf("%w: missing user", ErrInvalidInput)
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return service.store.ListActivityHistory(ctx, userID, limit)
}

func validIdempotencyKey(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && len(value) <= 128
}
