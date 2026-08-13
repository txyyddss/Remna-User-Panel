package rollover

import (
	"math/big"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// ProjectUsage builds the aggregate projection shown on the member dashboard.
// The calculation intentionally shares the cadence evaluator used by expiry
// settlement and uses the purchase's charged, net TXB amount.
func ProjectUsage(purchase model.Purchase, threshold int, snapshot UsageSnapshot, asOf time.Time) model.RolloverProjection {
	termStart := purchase.ValidFrom.UTC()
	termEnd := asOf.UTC()
	if termEnd.After(purchase.ValidUntil.UTC()) {
		termEnd = purchase.ValidUntil.UTC()
	}
	if termEnd.Before(termStart) {
		termEnd = termStart
	}
	anchor := termStart
	if snapshot.LastResetAt != nil && snapshot.LastResetAt.Before(termEnd) {
		anchor = snapshot.LastResetAt.UTC()
	}
	termSummary := calculateUsageRange(threshold, snapshot, anchor, termStart, termEnd)
	lastStart, lastEnd := latestResetBounds(termStart, termEnd, snapshot)
	lastSummary := calculateUsageRange(threshold, snapshot, lastStart, lastStart, lastEnd)

	term := buildWindow(termStart, termEnd, termSummary, purchase.PriceTXBMinor, purchase.RolloverMaxTXBMinor, int64(threshold))
	last := buildWindow(lastStart, lastEnd, lastSummary, purchase.PriceTXBMinor, purchase.RolloverMaxTXBMinor, int64(threshold))
	return model.RolloverProjection{
		PurchaseID:          purchase.ID,
		Paid:                model.TXBMoney(purchase.PriceTXBMinor),
		Maximum:             model.TXBMoney(purchase.RolloverMaxTXBMinor),
		MinimumRemainingBPS: threshold,
		SavedBPS:            savedBPS(term.Rollover.Minor, purchase.PriceTXBMinor),
		Term:                term,
		LastResetPeriod:     last,
	}
}

func latestResetBounds(termStart, termEnd time.Time, snapshot UsageSnapshot) (time.Time, time.Time) {
	if !termEnd.After(termStart) || snapshot.Strategy == "NO_RESET" {
		return termStart, termEnd
	}
	periodStart := termStart
	if snapshot.LastResetAt != nil && snapshot.LastResetAt.Before(termEnd) {
		periodStart = snapshot.LastResetAt.UTC()
	} else {
		periods := cadencePeriods(termStart, termStart, termEnd, snapshot.Strategy)
		if len(periods) > 0 {
			periodStart = periods[len(periods)-1].start
		}
	}
	periodEnd := cadenceAdvance(periodStart, snapshot.Strategy)
	if periodStart.Before(termStart) {
		periodStart = termStart
	}
	if periodEnd.After(termEnd) {
		periodEnd = termEnd
	}
	return periodStart, periodEnd
}

func buildWindow(start, end time.Time, summary model.RolloverUsageSummary, paid, maximum, threshold int64) model.RolloverWindow {
	remaining := summary.AllocatedBytes - summary.UsedBytes
	if remaining < 0 {
		remaining = 0
	}
	credit := proportionalFloor(paid, summary.EligibleUnusedBytes, summary.AllocatedBytes)
	if credit > maximum {
		credit = maximum
	}
	required, reachable := trafficToMaximum(summary.AllocatedBytes, paid, maximum, threshold)
	return model.RolloverWindow{
		Start:                 start.UTC(),
		End:                   end.UTC(),
		AllocatedTrafficBytes: summary.AllocatedBytes,
		UsedTrafficBytes:      summary.UsedBytes,
		RemainingTrafficBytes: remaining,
		EligibleUnusedBytes:   summary.EligibleUnusedBytes,
		Rollover:              model.TXBMoney(credit),
		TrafficToMaximumBytes: required,
		MaximumReachable:      reachable,
	}
}

func proportionalFloor(paid, remaining, allocated int64) int64 {
	if paid <= 0 || remaining <= 0 || allocated <= 0 {
		return 0
	}
	value := new(big.Int).Mul(big.NewInt(paid), big.NewInt(remaining))
	value.Quo(value, big.NewInt(allocated))
	if !value.IsInt64() {
		return 1<<63 - 1
	}
	return value.Int64()
}

func trafficToMaximum(allocated, paid, maximum int64, threshold int64) (*int64, bool) {
	if allocated <= 0 || paid <= 0 || maximum <= 0 {
		return nil, false
	}
	product := new(big.Int).Mul(big.NewInt(maximum), big.NewInt(allocated))
	required := ceilDiv(product, big.NewInt(paid))
	minimum := new(big.Int).Mul(big.NewInt(allocated), big.NewInt(threshold))
	minimum.Quo(minimum, big.NewInt(10000))
	minimum.Add(minimum, big.NewInt(1))
	if required.Cmp(minimum) < 0 {
		required = minimum
	}
	if !required.IsInt64() || required.Sign() <= 0 || required.Cmp(big.NewInt(allocated)) > 0 {
		return nil, false
	}
	value := required.Int64()
	return &value, true
}

func ceilDiv(numerator, denominator *big.Int) *big.Int {
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, denominator, remainder)
	if remainder.Sign() > 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient
}

func savedBPS(rolloverMinor string, paid int64) int {
	if paid <= 0 {
		return 0
	}
	rollover := new(big.Int)
	if _, ok := rollover.SetString(rolloverMinor, 10); !ok || rollover.Sign() <= 0 {
		return 0
	}
	value := new(big.Int).Mul(rollover, big.NewInt(10000))
	value.Quo(value, big.NewInt(paid))
	if value.Cmp(big.NewInt(10000)) > 0 {
		return 10000
	}
	return int(value.Int64())
}
