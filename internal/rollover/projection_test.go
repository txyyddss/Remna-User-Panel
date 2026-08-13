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
