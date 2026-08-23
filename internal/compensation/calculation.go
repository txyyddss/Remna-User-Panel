package compensation

func calculateExtension(seconds int64, multiplierBPS int) (minutes int, capped bool) {
	if seconds <= 0 || multiplierBPS <= 0 {
		return 0, false
	}
	capThreshold := (int64(MaxMinutes+1)*600_000 - 1) / int64(multiplierBPS)
	if seconds > capThreshold {
		return MaxMinutes, true
	}
	value := seconds * int64(multiplierBPS) / 600_000
	if value > MaxMinutes {
		return MaxMinutes, true
	}
	return int(value), false
}

// RecoveryOutcome classifies one exact UTC observation interval.
func RecoveryOutcome(seconds int64, thresholdMinutes, multiplierBPS, recipients int) (int, bool, string) {
	if seconds <= int64(thresholdMinutes)*60 {
		return 0, false, "below_threshold"
	}
	if recipients == 0 {
		return 0, false, "no_recipients"
	}
	minutes, capped := calculateExtension(seconds, multiplierBPS)
	if minutes == 0 {
		return 0, false, "computed_zero"
	}
	return minutes, capped, ""
}
