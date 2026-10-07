package activity

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

func TestBoostRewardMinor(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		base    int64
		count   int
		want    int64
		invalid bool
	}{
		{"zero boosts", 125, 0, 0, false},
		{"one boost half cent", 125, 1, 63, false},
		{"two boosts", 125, 2, 125, false},
		{"three boosts half cent", 125, 3, 188, false},
		{"four boosts", 125, 4, 250, false},
		{"zero base", 0, 3, 0, false},
		{"exact maximum", math.MaxInt64, 2, math.MaxInt64, false},
		{"large base halved", math.MaxInt64, 1, math.MaxInt64/2 + 1, false},
		{"overflow", math.MaxInt64, 3, 0, true},
		{"negative base", -1, 2, 0, true},
		{"negative count", 1, -1, 0, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := BoostRewardMinor(test.base, test.count)
			if got != test.want || errors.Is(err, ErrInvalidInput) != test.invalid {
				t.Fatalf("BoostRewardMinor() = (%d, %v), want (%d, invalid=%t)", got, err, test.want, test.invalid)
			}
		})
	}
}

type fixedBoostSource struct {
	count int
	err   error
}

func (source fixedBoostSource) GroupBoost(context.Context, string) (GroupBoostStatus, error) {
	state := "required"
	if source.count > 0 {
		state = "boosted"
	}
	return GroupBoostStatus{State: state, Count: &source.count}, source.err
}

type boostClaimStore struct {
	Store
	calls, count int
}

func (store *boostClaimStore) ClaimBoostedDailyActivityRange(_ context.Context, _, _, _ string, _, _ int64, count int, _ RandomSource, _ time.Time) (DailyCheckIn, error) {
	store.calls++
	store.count = count
	return DailyCheckIn{RewardMinor: 188}, nil
}

func TestCheckInEnforcesBoostBeforeSettlement(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		source GroupBoostSource
		want   error
		calls  int
	}{
		{"unconfigured", nil, ErrGroupBoostUnavailable, 0},
		{"unboosted", fixedBoostSource{}, ErrGroupBoostRequired, 0},
		{"Telegram failure", fixedBoostSource{err: errors.New("permission denied")}, ErrGroupBoostUnavailable, 0},
		{"boosted", fixedBoostSource{count: 3}, nil, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &boostClaimStore{}
			service := NewService(store, nil, nil)
			service.SetGroupBoostSource(test.source)
			_, err := service.CheckIn(context.Background(), "member", CheckInConfig{Timezone: "Asia/Shanghai", RewardMinMinor: 100, RewardMaxMinor: 200})
			if !errors.Is(err, test.want) || store.calls != test.calls || test.calls > 0 && store.count != 3 {
				t.Fatalf("CheckIn() error=%v calls=%d count=%d", err, store.calls, store.count)
			}
		})
	}
}
