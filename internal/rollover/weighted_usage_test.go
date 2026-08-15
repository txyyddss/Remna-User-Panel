package rollover

import (
	"testing"
	"time"
)

func TestWeightedNodeUsageUsesFixedPointMultipliersWithoutPersistingSeries(t *testing.T) {
	start := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	daily, total := WeightedNodeUsage([]time.Time{start, start.AddDate(0, 0, 1)}, []NodeUsageSeries{
		{UUID: "node-a", MultiplierFP: 1_500_000, Data: []int64{100, 200}},
		{UUID: "node-b", MultiplierFP: 500_000, Data: []int64{100, 200}},
	})
	if total != 300 || len(daily) != 2 || daily[0].Bytes != 200 || daily[1].Bytes != 400 {
		t.Fatalf("weighted usage = (%+v, %d), want 200/400/300", daily, total)
	}
}
