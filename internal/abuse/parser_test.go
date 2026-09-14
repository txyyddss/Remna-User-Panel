package abuse

import (
	"testing"
	"time"
)

func TestParseLineRetainsOnlyConfiguredOutboundDomainAccepts(t *testing.T) {
	fallback := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	valid, ok := parseLine("2026/08/24 12:00:01 accepted tcp://example.com:443 email: 42 >> proxy-main", fallback, []string{"direct", "proxy-main"})
	if !ok || valid.RemoteID != "42" || valid.Domain != "example.com" {
		t.Fatalf("valid configured-tag record = %#v, %v", valid, ok)
	}
	if _, accepted := parseLine("accepted tcp://example.com:443 email: 42 >> direct", fallback, []string{"direct", "proxy-main"}); !accepted {
		t.Fatal("expected second configured tag to be accepted")
	}
	for _, line := range []string{"accepted tcp://203.0.113.8:443 email: 42 >> proxy-main", "accepted tcp://example.com:443 email: 42 >> bypass", "error accepted tcp://example.com:443 email: 42 >> proxy-main"} {
		if _, accepted := parseLine(line, fallback, []string{"direct", "proxy-main"}); accepted {
			t.Fatalf("unexpected accepted record: %s", line)
		}
	}
}
