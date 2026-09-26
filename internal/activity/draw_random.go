package activity

import (
	"fmt"
	"math"
)

// ValueRange stores inclusive integral ticks; the reward kind defines the unit.
type ValueRange struct {
	Min               int64  `json:"min"`
	Max               int64  `json:"max"`
	Distribution      string `json:"distribution"`
	PositiveChanceBPS *int   `json:"positiveChanceBps,omitempty"`
}

func (value ValueRange) Validate(minimum, maximum int64, signed bool) error {
	if value.Min < minimum || value.Max > maximum || value.Min > value.Max || (!signed && value.Min <= 0) || (signed && value.Min == 0 && value.Max == 0) {
		return fmt.Errorf("%w: invalid random range", ErrInvalidInput)
	}
	if value.Distribution != "uniform" && value.Distribution != "gaussian" && value.Distribution != "power_law" {
		return fmt.Errorf("%w: unsupported distribution", ErrInvalidInput)
	}
	if value.Min < 0 && value.Max > 0 {
		if value.PositiveChanceBPS == nil || *value.PositiveChanceBPS < 0 || *value.PositiveChanceBPS > 10000 {
			return fmt.Errorf("%w: signed range needs positive chance", ErrInvalidInput)
		}
	} else if value.PositiveChanceBPS != nil {
		return fmt.Errorf("%w: positive chance only applies to mixed signs", ErrInvalidInput)
	}
	return nil
}

// Roll returns one tick, first selecting sign and then biasing toward lower cost.
func (value ValueRange) Roll(rng RandomSource) (int64, error) {
	lo, hi := value.Min, value.Max
	if lo < 0 && hi > 0 {
		roll, err := rng.Int63n(10000)
		if err != nil {
			return 0, err
		}
		if roll < int64(*value.PositiveChanceBPS) {
			lo = 1
		} else {
			hi = -1
		}
	} else if lo == 0 {
		lo = 1
	} else if hi == 0 {
		hi = -1
	}
	span := hi - lo + 1
	if span <= 0 {
		return 0, ErrInvalidInput
	}
	rank, err := drawCostRank(rng, span, value.Distribution)
	if err != nil {
		return 0, err
	}
	// Cost rises from the lower numeric endpoint for both signed money and traffic.
	return lo + rank, nil
}

// RollAround splits a multiplier around 1x before sampling a side.
func (value ValueRange) RollAround(rng RandomSource, pivot int64) (int64, error) {
	if value.Min >= pivot || value.Max <= pivot {
		return value.Roll(rng)
	}
	if value.PositiveChanceBPS == nil {
		return 0, ErrInvalidInput
	}
	roll, err := rng.Int63n(10000)
	if err != nil {
		return 0, err
	}
	lo, hi := value.Min, value.Max
	if roll < int64(*value.PositiveChanceBPS) {
		lo = pivot + 1
	} else {
		hi = pivot - 1
	}
	span := hi - lo + 1
	rank, err := drawCostRank(rng, span, value.Distribution)
	if err != nil {
		return 0, err
	}
	return lo + rank, nil
}

func drawCostRank(rng RandomSource, span int64, distribution string) (int64, error) {
	if span == 1 {
		return 0, nil
	}
	if distribution == "uniform" {
		return rng.Int63n(span)
	}
	roll, err := rng.Int63n(math.MaxInt64)
	if err != nil {
		return 0, err
	}
	u := (float64(roll) + 0.5) / float64(math.MaxInt64)
	if distribution == "power_law" {
		value := 1/(1-u*(1-1/float64(span))) - 1
		return min(int64(value), span-1), nil
	}
	// Half-normal, truncated to the configured range, with mean at its cheapest tick.
	for attempt := 0; attempt < 32; attempt++ {
		raw, sampleErr := rng.Int63n(math.MaxInt64)
		if sampleErr != nil {
			return 0, sampleErr
		}
		v := (float64(raw) + 0.5) / float64(math.MaxInt64)
		z := math.Abs(math.Sqrt(-2*math.Log(u)) * math.Cos(2*math.Pi*v))
		rank := int64(z * float64(span) / 3)
		if rank < span {
			return rank, nil
		}
		u = v
	}
	return span - 1, nil
}
