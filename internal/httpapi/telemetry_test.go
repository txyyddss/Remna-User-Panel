package httpapi

import "testing"

func TestFingerprintSimilarity(t *testing.T) {
	base := storedFingerprint{Full: "full-a", Canvas: "canvas", WebGL: "webgl", Fonts: "fonts", Audio: "audio", Network: "network", Browser: "browser", Platform: "platform", Timezone: "timezone", Screen: "screen", Hardware: "hardware", Language: "language"}
	if got := fingerprintSimilarity(base, base); got != 100 {
		t.Fatalf("exact fingerprint score=%d, want 100", got)
	}
	partial := base
	partial.Full = "full-b"
	partial.Audio = "other"
	partial.Fonts = "other"
	partial.Canvas = "other"
	if got := fingerprintSimilarity(base, partial); got != 56 {
		t.Fatalf("partial fingerprint score=%d, want 56", got)
	}
	if got := fingerprintSimilarity(base, storedFingerprint{Full: "unrelated"}); got != 0 {
		t.Fatalf("unrelated fingerprint score=%d, want 0", got)
	}
}

func TestTelemetrySettingRanges(t *testing.T) {
	for _, test := range []struct {
		key   string
		value any
		ok    bool
	}{
		{"TELEMETRY_FINGERPRINT_REJECT_SCORE", 1, true}, {"TELEMETRY_FINGERPRINT_REJECT_SCORE", 70, true}, {"TELEMETRY_FINGERPRINT_REJECT_SCORE", 100, true},
		{"TELEMETRY_FINGERPRINT_REJECT_SCORE", 0, false}, {"TELEMETRY_FINGERPRINT_REJECT_SCORE", 101, false},
		{"TELEMETRY_RETENTION_HOURS", 1, true}, {"TELEMETRY_RETENTION_HOURS", 720, true}, {"TELEMETRY_RETENTION_HOURS", 721, false},
	} {
		_, err := normalizeSettingValue(test.key, test.value)
		if (err == nil) != test.ok {
			t.Errorf("%s=%v err=%v, ok=%v", test.key, test.value, err, test.ok)
		}
	}
}
