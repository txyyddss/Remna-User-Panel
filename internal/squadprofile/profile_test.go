package squadprofile

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNormalizeProfiles(t *testing.T) {
	port := 1000
	dynamic := true
	cases := []struct {
		name  string
		input Profile
		valid bool
	}{
		{name: "broadband", input: Profile{Type: Broadband, ISP: " ISP ", PortMbps: &port, Dynamic: &dynamic, Location: " Miaoli "}, valid: true},
		{name: "china unlimited", input: Profile{Type: ChinaOptimized, CT: " CT ", CU: " CU ", CM: " CM ", CountryCode: "tw"}, valid: true},
		{name: "international carriers", input: Profile{Type: InternationalNetwork, CountryCode: "SG", UpstreamCarriers: []string{"A, B", "A"}}, valid: true},
		{name: "invalid country", input: Profile{Type: InternationalNetwork, CountryCode: "ZZ", UpstreamCarriers: []string{"A"}}, valid: false},
		{name: "missing carrier", input: Profile{Type: InternationalNetwork, CountryCode: "SG"}, valid: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result, err := Normalize(&test.input)
			if test.valid && err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if !test.valid && !errors.Is(err, ErrInvalid) {
				t.Fatalf("Normalize() error = %v, want ErrInvalid", err)
			}
			if test.name == "international carriers" && (result == nil || len(result.UpstreamCarriers) != 2) {
				t.Fatalf("normalized carriers = %+v", result)
			}
		})
	}
}

func TestParseJSONNormalizesPersistedProfile(t *testing.T) {
	profile, err := ParseJSON(`{"type":"international_network","portMbps":null,"countryCode":"us","upstreamCarriers":[" A, B "]}`)
	if err != nil {
		t.Fatalf("ParseJSON() error = %v", err)
	}
	if profile.CountryCode != "US" || len(profile.UpstreamCarriers) != 2 {
		t.Fatalf("profile = %+v", profile)
	}
	if _, err := json.Marshal(profile); err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
}
