package activity

import (
	"errors"
	"math"
	"testing"
)

type drawRandomFunc func(int64) (int64, error)

func (random drawRandomFunc) Int63n(bound int64) (int64, error) { return random(bound) }

func TestPowerLawCouponQuantiles(t *testing.T) {
	t.Parallel()
	value := ValueRange{Min: 3_000, Max: 10_000, Distribution: "power_law"}
	for _, test := range []struct {
		name string
		u    int64
		want int64
	}{
		{"minimum", 0, 3_000},
		{"25th percentile", (1 << 53) / 4, 3_437},
		{"median", (1 << 53) / 2, 4_750},
		{"75th percentile", 3 * (1 << 53) / 4, 6_938},
		{"90th percentile", 9 * (1 << 53) / 10, 8_670},
		{"maximum is reachable", (1 << 53) - 1, 10_000},
	} {
		t.Run(test.name, func(t *testing.T) {
			random := drawRandomFunc(func(bound int64) (int64, error) {
				if bound != 1<<53 {
					t.Fatalf("random bound = %d, want exactly representable uniform ticks", bound)
				}
				return test.u, nil
			})
			got, err := value.Roll(random)
			if err != nil || got != test.want {
				t.Fatalf("Roll() = (%d, %v), want %d", got, err, test.want)
			}
		})
	}
}

func TestPowerLawDistributionAcrossPrecisions(t *testing.T) {
	t.Parallel()
	// Stratified uniform inputs make these distribution checks deterministic.
	// A power law with shape 1/2 has mean 1/3 and CDF(1/4)=1/2.
	const samples = 10_000
	for _, span := range []int64{100, 7_001, 60_001, 660_001, 10_000_000_001} {
		var sum float64
		var lowerQuarter, minimum int
		for index := range samples {
			random := drawRandomFunc(func(bound int64) (int64, error) {
				return int64((float64(index) + 0.5) / samples * float64(bound)), nil
			})
			rank, err := drawCostRank(random, span, "power_law")
			if err != nil || rank < 0 || rank >= span {
				t.Fatalf("rank for span %d = (%d, %v)", span, rank, err)
			}
			sum += float64(rank) / float64(span)
			if float64(rank) < float64(span)/4 {
				lowerQuarter++
			}
			if rank == 0 {
				minimum++
			}
		}
		if mean := sum / samples; math.Abs(mean-1.0/3) > 1/float64(span)+0.0002 {
			t.Errorf("span %d normalized mean = %f, want approximately 1/3", span, mean)
		}
		if fraction := float64(lowerQuarter) / samples; math.Abs(fraction-0.5) > 2/float64(span)+0.0002 {
			t.Errorf("span %d lower quarter fraction = %f, want approximately 1/2", span, fraction)
		}
		if span == 7_001 && (minimum < 118 || minimum > 121) {
			t.Errorf("30-100 percent minimum hits = %d/%d, want approximately 1.2%%", minimum, samples)
		}
	}
}

func TestPowerLawSignedAndMultiplierRanges(t *testing.T) {
	t.Parallel()
	chance := 4_000
	for _, test := range []struct {
		name           string
		lo, hi, pivot  int64
		signRoll, want int64
	}{
		{"positive money", -100, 100, 0, 3_999, 26},
		{"negative money", -100, 100, 0, 4_000, -75},
		{"multiplier gain", 5_000, 30_000, 10_000, 3_999, 15_001},
		{"multiplier loss", 5_000, 30_000, 10_000, 4_000, 6_250},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			random := drawRandomFunc(func(bound int64) (int64, error) {
				calls++
				if calls == 1 {
					if bound != 10_000 {
						t.Fatalf("sign bound = %d", bound)
					}
					return test.signRoll, nil
				}
				return bound / 2, nil
			})
			value := ValueRange{Min: test.lo, Max: test.hi, Distribution: "power_law", PositiveChanceBPS: &chance}
			var got int64
			var err error
			if test.pivot == 0 {
				got, err = value.Roll(random)
			} else {
				got, err = value.RollAround(random, test.pivot)
			}
			if err != nil || got != test.want || calls != 2 {
				t.Fatalf("roll = (%d, %v), calls=%d, want %d with 2 draws", got, err, calls, test.want)
			}
		})
	}
}

func TestPowerLawFixedRangeAndRandomFailure(t *testing.T) {
	t.Parallel()
	failure := errors.New("random source unavailable")
	random := drawRandomFunc(func(int64) (int64, error) { return 0, failure })
	if got, err := (ValueRange{Min: 168, Max: 168, Distribution: "power_law"}).Roll(random); err != nil || got != 168 {
		t.Fatalf("fixed reward = (%d, %v)", got, err)
	}
	if _, err := (ValueRange{Min: 3_000, Max: 10_000, Distribution: "power_law"}).Roll(random); !errors.Is(err, failure) {
		t.Fatalf("random failure = %v, want %v", err, failure)
	}
}
