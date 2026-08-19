package app

import (
	"math"
	"testing"
)

func TestRoundedNonNegativeInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value float64
		want  int64
	}{
		{name: "fraction rounds down", value: 12.4, want: 12},
		{name: "fraction rounds up", value: 12.5, want: 13},
		{name: "negative becomes zero", value: -1, want: 0},
		{name: "not a number becomes zero", value: math.NaN(), want: 0},
		{name: "positive infinity is bounded", value: math.Inf(1), want: int64(1<<63 - 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := roundedNonNegativeInt64(test.value); got != test.want {
				t.Fatalf("roundedNonNegativeInt64(%v) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}
