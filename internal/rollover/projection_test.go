package rollover

import (
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestProjectUsageUsesNetPaidAndTermProjection(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	purchase := model.Purchase{
		ID: "purchase-1", PriceTXBMinor: 10_000, GrossPriceTXBMinor: 12_000,
		ValidFrom: start, ValidUntil: start.AddDate(0, 0, 10), RolloverMaxTXBMinor: 2_000,
		Price: model.TXBMoney(10_000), RolloverMax: model.TXBMoney(2_000),
	}
	daily := make([]DailyUsage, 0, 10)
	for day := 0; day < 10; day++ {
		daily = append(daily, DailyUsage{Date: start.AddDate(0, 0, day), Bytes: 90})
	}
	projection := ProjectUsage(purchase, 0, UsageSnapshot{LimitBytes: 1_000, Strategy: "NO_RESET", Daily: daily}, purchase.ValidUntil)
	if projection.Term.Rollover.Minor != "1000" || projection.SavedBPS != 1000 {
		t.Fatalf("projection = %+v, want 1000 minor and 1000 bps", projection)
	}
	if projection.Paid.Minor != "10000" || projection.Maximum.Minor != "2000" {
		t.Fatalf("money facts = paid %q maximum %q", projection.Paid.Minor, projection.Maximum.Minor)
	}
}

func TestCalculateUsageClipsDateBucketsToEachResetPeriod(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	reset := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		strategy string
		reset    time.Time
		end      time.Time
		usageDay time.Time
		wantUsed int64
	}{
		{name: "daily boundary", strategy: "DAY", reset: reset, end: start.AddDate(0, 0, 1), usageDay: start, wantUsed: 240},
		{name: "weekly boundary", strategy: "WEEK", reset: start.AddDate(0, 0, 3).Add(12 * time.Hour), end: start.AddDate(0, 0, 7), usageDay: start.AddDate(0, 0, 3), wantUsed: 240},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			purchase := model.Purchase{ValidFrom: start, ValidUntil: test.end}
			resetAt := test.reset
			summary := CalculateUsage(purchase, 0, UsageSnapshot{
				LimitBytes: 1_000, Strategy: test.strategy, LastResetAt: &resetAt,
				Daily: []DailyUsage{{Date: test.usageDay, Bytes: 240}},
			})
			if summary.AllocatedBytes != 1_000 || summary.UsedBytes != test.wantUsed || summary.EligibleUnusedBytes != 760 {
				t.Fatalf("summary = %+v, want allocated=1000 used=%d eligible=760", summary, test.wantUsed)
			}
		})
	}
}

func TestCalculateUsagePartialTermMatchesProjection(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 8, 1, 6, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 2, 18, 0, 0, 0, time.UTC)
	reset := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	purchase := model.Purchase{ID: "partial-term", PriceTXBMinor: 1_000, ValidFrom: start, ValidUntil: end, RolloverMaxTXBMinor: 2_000}
	snapshot := UsageSnapshot{LimitBytes: 1_000, Strategy: "DAY", LastResetAt: &reset}
	summary := CalculateUsage(purchase, 0, snapshot)
	projection := ProjectUsage(purchase, 0, snapshot, end)
	if summary.AllocatedBytes != 1_500 || summary.UsedBytes != 0 || summary.EligibleUnusedBytes != 1_500 {
		t.Fatalf("summary = %+v, want allocated=1500 used=0 eligible=1500", summary)
	}
	if projection.Term.AllocatedTrafficBytes != summary.AllocatedBytes || projection.Term.UsedTrafficBytes != summary.UsedBytes || projection.Term.EligibleUnusedBytes != summary.EligibleUnusedBytes {
		t.Fatalf("projection term = %+v, summary = %+v", projection.Term, summary)
	}
}

func TestCalculateUsageThresholdIsStrictPerPeriod(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name         string
		used         int64
		wantEligible int64
	}{
		{name: "equal threshold", used: 500, wantEligible: 0},
		{name: "above threshold", used: 499, wantEligible: 501},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			purchase := model.Purchase{ValidFrom: start, ValidUntil: start.AddDate(0, 0, 1)}
			summary := CalculateUsage(purchase, 5_000, UsageSnapshot{
				LimitBytes: 1_000, Strategy: "DAY", LastResetAt: &start,
				Daily: []DailyUsage{{Date: start, Bytes: test.used}},
			})
			if summary.EligibleUnusedBytes != test.wantEligible {
				t.Fatalf("eligible = %d, want %d", summary.EligibleUnusedBytes, test.wantEligible)
			}
		})
	}
}

func TestProjectUsageSelectsLatestResetPeriodAndStrictThreshold(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	reset := time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)
	purchase := model.Purchase{ID: "purchase-2", PriceTXBMinor: 10_000, ValidFrom: start, ValidUntil: start.AddDate(0, 0, 10), RolloverMaxTXBMinor: 2_000}
	projection := ProjectUsage(purchase, 9_000, UsageSnapshot{LimitBytes: 1_000, Strategy: "DAY", LastResetAt: &reset,
		Daily: []DailyUsage{{Date: start, Bytes: 100}, {Date: start.AddDate(0, 0, 1), Bytes: 100}, {Date: reset, Bytes: 100}}}, asOf)
	if !projection.LastResetPeriod.Start.Equal(reset) || !projection.LastResetPeriod.End.Equal(asOf) {
		t.Fatalf("latest period = %s to %s", projection.LastResetPeriod.Start, projection.LastResetPeriod.End)
	}
	if projection.LastResetPeriod.Rollover.Minor != "0" {
		t.Fatalf("strict threshold credited %s", projection.LastResetPeriod.Rollover.Minor)
	}
}

func TestProjectUsageMaximumUsesAllCadenceAllowance(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name             string
		strategy         string
		days             int
		usage            map[int]int64
		wantAllocated    int64
		wantEligible     int64
		wantTrafficToMax int64
	}{
		{name: "daily", strategy: "DAY", days: 3, usage: map[int]int64{0: 200, 1: 800, 2: 200}, wantAllocated: 3_000, wantEligible: 1_600, wantTrafficToMax: 1_501},
		{name: "weekly", strategy: "WEEK", days: 14, usage: map[int]int64{0: 200, 7: 800}, wantAllocated: 2_000, wantEligible: 800, wantTrafficToMax: 1_001},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			reset := start
			daily := make([]DailyUsage, 0, test.days)
			for day := 0; day < test.days; day++ {
				daily = append(daily, DailyUsage{Date: start.AddDate(0, 0, day), Bytes: test.usage[day]})
			}
			purchase := model.Purchase{
				ID: "purchase-maximum-" + test.name, PriceTXBMinor: 10_000,
				ValidFrom: start, ValidUntil: start.AddDate(0, 0, test.days), RolloverMaxTXBMinor: 2_000,
			}
			projection := ProjectUsage(purchase, 5_000, UsageSnapshot{
				LimitBytes: 1_000, Strategy: test.strategy, LastResetAt: &reset, Daily: daily,
			}, purchase.ValidUntil)
			if projection.Term.AllocatedTrafficBytes != test.wantAllocated || projection.Term.EligibleUnusedBytes != test.wantEligible {
				t.Fatalf("term allowance = (%d, %d), want (%d, %d)", projection.Term.AllocatedTrafficBytes, projection.Term.EligibleUnusedBytes, test.wantAllocated, test.wantEligible)
			}
			if projection.Term.TrafficToMaximumBytes == nil || *projection.Term.TrafficToMaximumBytes != test.wantTrafficToMax {
				t.Fatalf("traffic to maximum = %v, want %d", projection.Term.TrafficToMaximumBytes, test.wantTrafficToMax)
			}
			if projection.Term.Rollover.Minor != "2000" {
				t.Fatalf("rollover = %s, want 2000", projection.Term.Rollover.Minor)
			}
		})
	}
}

func TestTrafficToMaximumReportsUnreachableStates(t *testing.T) {
	if value, reachable := trafficToMaximum(1_000, 10_000, 1_000, 0); !reachable || value == nil || *value != 100 {
		t.Fatalf("trafficToMaximum() = (%v, %t), want 100/true", value, reachable)
	}
	if value, reachable := trafficToMaximum(1_000, 0, 1_000, 0); reachable || value != nil {
		t.Fatalf("zero-paid maximum = (%v, %t), want nil/false", value, reachable)
	}
	if value, reachable := trafficToMaximum(1_000, 10_000, 20_000, 0); reachable || value != nil {
		t.Fatalf("over-paid maximum = (%v, %t), want nil/false", value, reachable)
	}
}
