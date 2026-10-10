package iplookup

import (
	"encoding/json"
	"testing"
)

func TestConfigKeepsCheckOrderAndValidatesGeoPermutation(t *testing.T) {
	t.Parallel()
	c, err := DecodeConfig(`{"providers":[{"id":"maxmind"},{"id":"ipapi"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if c.Providers[0].ID != "maxmind" || c.Providers[1].ID != "ipapi" || len(c.GeolocationOrder) != 5 {
		t.Fatalf("orders=%+v", c)
	}
	for _, order := range [][]string{{"maxmind", "ipapi", "ip2location", "scamalytics", "abuseipdb"}, {"maxmind", "maxmind", "ip2location", "scamalytics", "abuseipdb"}, {"ipqs", "ipapi", "ip2location", "scamalytics", "abuseipdb"}, {}} {
		c.GeolocationOrder = order
		raw, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeConfig(string(raw))
		valid := len(order) == 5 && order[1] == "ipapi" && order[0] == "maxmind"
		if (err == nil) != valid {
			t.Fatalf("order=%v error=%v", order, err)
		}
		if valid && decoded.GeolocationOrder[0] != "maxmind" {
			t.Fatal("geo order reset")
		}
	}
}
