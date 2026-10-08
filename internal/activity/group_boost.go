package activity

import (
	"context"
	"errors"
	"fmt"
	"math/big"
)

var (
	// ErrGroupBoostRequired rejects rewards when verification proves zero boosts.
	ErrGroupBoostRequired = errors.New("group boost required")
	// ErrGroupBoostUnavailable rejects claims whose eligibility cannot be verified.
	ErrGroupBoostUnavailable = errors.New("group boost verification unavailable")
)

// GroupBoostStatus describes live eligibility for the configured Telegram group.
type GroupBoostStatus struct {
	State    string  `json:"state"`
	Count    *int    `json:"count"`
	BoostURL *string `json:"boostUrl"`
}

// GroupBoostSource verifies the member's boosts without persisting upstream data.
type GroupBoostSource interface {
	GroupBoost(context.Context, string) (GroupBoostStatus, error)
}

// SetGroupBoostSource wires the queued provider before the service starts serving.
func (service *Service) SetGroupBoostSource(source GroupBoostSource) {
	service.boosts = source
}

// GroupBoost returns verified eligibility or an explicitly unavailable state.
func (service *Service) GroupBoost(ctx context.Context, userID string) (GroupBoostStatus, error) {
	if service.boosts == nil {
		return GroupBoostStatus{State: "unavailable"}, ErrGroupBoostUnavailable
	}
	status, err := service.boosts.GroupBoost(ctx, userID)
	if err != nil || status.Count == nil || *status.Count < 0 ||
		(status.State != "boosted" && status.State != "required") ||
		(status.State == "boosted") != (*status.Count > 0) {
		status.State, status.Count = "unavailable", nil
		return status, fmt.Errorf("%w: %v", ErrGroupBoostUnavailable, err)
	}
	return status, nil
}

// BoostRewardMinor uses max(1, count/2) for positive counts, rounded half up.
// Exact arithmetic avoids both floating-point cent loss and intermediate overflow.
func BoostRewardMinor(baseMinor int64, count int) (int64, error) {
	if baseMinor < 0 || count < 0 {
		return 0, ErrInvalidInput
	}
	if count == 0 {
		return 0, nil
	}
	count = max(count, 2)
	value := new(big.Int).Mul(big.NewInt(baseMinor), big.NewInt(int64(count)))
	value.Add(value, big.NewInt(1)).Quo(value, big.NewInt(2))
	if !value.IsInt64() {
		return 0, fmt.Errorf("%w: boosted reward exceeds supported amount", ErrInvalidInput)
	}
	return value.Int64(), nil
}
