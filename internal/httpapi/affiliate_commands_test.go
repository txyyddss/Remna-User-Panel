package httpapi

import "testing"

func TestParsePositiveTelegramID(t *testing.T) {
	for _, valid := range []string{"1", "123456789"} {
		if _, err := parsePositiveTelegramID(valid); err != nil {
			t.Errorf("parsePositiveTelegramID(%q) error = %v", valid, err)
		}
	}
	for _, invalid := range []string{"", "0", "-1", "+1", "1.0", "abc"} {
		if _, err := parsePositiveTelegramID(invalid); err == nil {
			t.Errorf("parsePositiveTelegramID(%q) returned nil", invalid)
		}
	}
}
