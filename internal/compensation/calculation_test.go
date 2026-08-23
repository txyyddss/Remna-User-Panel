package compensation

import "testing"

func TestRecoveryOutcomeUsesStrictThresholdFloorAndCap(t *testing.T) {
	tests := []struct {
		name       string
		seconds    int64
		threshold  int
		multiplier int
		recipients int
		minutes    int
		capped     bool
		reason     string
	}{
		{name: "equal threshold", seconds: 60, threshold: 1, multiplier: 10_000, recipients: 1, reason: "below_threshold"},
		{name: "floor", seconds: 119, threshold: 1, multiplier: 10_000, recipients: 1, minutes: 1},
		{name: "no recipients", seconds: 120, threshold: 1, multiplier: 10_000, reason: "no_recipients"},
		{name: "cap", seconds: 400_000_000, threshold: 1, multiplier: 1_000_000, recipients: 1, minutes: MaxMinutes, capped: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			minutes, capped, reason := RecoveryOutcome(test.seconds, test.threshold, test.multiplier, test.recipients)
			if minutes != test.minutes || capped != test.capped || reason != test.reason {
				t.Fatalf("recoveryOutcome() = (%d,%v,%q)", minutes, capped, reason)
			}
		})
	}
}
