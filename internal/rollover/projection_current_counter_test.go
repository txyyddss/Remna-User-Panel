package rollover

import (
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestCalculateUsageUsesAuthoritativeCurrentCounterAcrossCadences(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	currentUsed := int64(250)
	tests := []struct {
		name     string
		strategy string
		reset    time.Time
		end      time.Time
	}{
		{name: "no reset", strategy: "NO_RESET", reset: start, end: start.AddDate(0, 0, 2)},
		{name: "daily", strategy: "DAY", reset: start.AddDate(0, 0, 1), end: start.AddDate(0, 0, 2)},
		{name: "weekly", strategy: "WEEK", reset: start.AddDate(0, 0, 7), end: start.AddDate(0, 0, 14)},
		{name: "monthly", strategy: "MONTH", reset: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), end: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{name: "rolling monthly", strategy: "MONTH_ROLLING", reset: start.AddDate(0, 0, 30), end: start.AddDate(0, 0, 60)},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reset := test.reset
			summary := CalculateUsage(model.Purchase{ValidFrom: start, ValidUntil: test.end}, 0, UsageSnapshot{
				LimitBytes: 1_000, Strategy: test.strategy, LastResetAt: &reset,
				CurrentUsedBytes: &currentUsed, Daily: []DailyUsage{
					{Date: start, Bytes: 100}, {Date: test.reset, Bytes: 900},
				},
			})
			if summary.AllocatedBytes != 2_000 || summary.UsedBytes != 350 || summary.EligibleUnusedBytes != 1_650 {
				t.Fatalf("summary = %+v, want allocated=2000 used=350 eligible=1650", summary)
			}
		})
	}
}

func TestCalculateUsageTreatsZeroCurrentCounterAsAuthoritative(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	reset := start.AddDate(0, 0, 1)
	currentUsed := int64(0)
	summary := CalculateUsage(model.Purchase{ValidFrom: start, ValidUntil: start.AddDate(0, 0, 2)}, 0, UsageSnapshot{
		LimitBytes: 1_000, Strategy: "DAY", LastResetAt: &reset, CurrentUsedBytes: &currentUsed,
		Daily: []DailyUsage{{Date: start, Bytes: 100}, {Date: reset, Bytes: 900}},
	})
	if summary.UsedBytes != 100 || summary.EligibleUnusedBytes != 1_900 {
		t.Fatalf("summary = %+v, want used=100 eligible=1900", summary)
	}
}

func TestCalculateUsageFallsBackToDailyBucketsWithoutCurrentCounter(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	reset := start.AddDate(0, 0, 1)
	summary := CalculateUsage(model.Purchase{ValidFrom: start, ValidUntil: start.AddDate(0, 0, 2)}, 0, UsageSnapshot{
		LimitBytes: 1_000, Strategy: "DAY", LastResetAt: &reset,
		Daily: []DailyUsage{{Date: start, Bytes: 100}, {Date: reset, Bytes: 900}},
	})
	if summary.UsedBytes != 1_000 || summary.EligibleUnusedBytes != 1_000 {
		t.Fatalf("summary = %+v, want used=1000 eligible=1000", summary)
	}
}

func TestCadenceBoundaries(t *testing.T) {
	monthEnd := time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC)
	if got, want := cadenceAdvance(monthEnd, "MONTH"), time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("monthly advance = %s, want %s", got, want)
	}
	if got, want := cadenceAdvance(monthEnd, "MONTH_ROLLING"), time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Fatalf("rolling monthly advance = %s, want %s", got, want)
	}
	februaryEnd := time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC)
	if got, want := cadencePrevious(februaryEnd, "MONTH"), monthEnd; !got.Equal(want) {
		t.Fatalf("monthly previous = %s, want %s", got, want)
	}
}
