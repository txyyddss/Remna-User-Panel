package iplookup

import "testing"

func floatPointer(v float64) *float64 { return &v }

func TestApplicationQualityBandBoundaries(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		confidence, radius *float64
		want               int
	}{
		{floatPointer(100), nil, 5}, {floatPointer(95), nil, 5}, {floatPointer(94), nil, 4},
		{floatPointer(80), nil, 4}, {floatPointer(79), nil, 3}, {floatPointer(60), nil, 3},
		{floatPointer(59), nil, 2}, {floatPointer(30), nil, 2}, {floatPointer(29), nil, 1}, {floatPointer(0), nil, 1},
		{nil, floatPointer(10), 5}, {nil, floatPointer(10.1), 4}, {nil, floatPointer(50), 4},
		{nil, floatPointer(50.1), 3}, {nil, floatPointer(200), 3}, {nil, floatPointer(200.1), 2},
		{nil, floatPointer(1000), 2}, {nil, floatPointer(1000.1), 1}, {nil, nil, 0},
		{floatPointer(101), floatPointer(10), 5}, {floatPointer(0), floatPointer(1), 1},
	} {
		if got := qualityBand(tc.confidence, tc.radius); got != tc.want {
			t.Fatalf("band=%d want=%d confidence=%v radius=%v", got, tc.want, tc.confidence, tc.radius)
		}
	}
}

func TestGeolocationWinnerIndependentOfCheckSequence(t *testing.T) {
	t.Parallel()
	candidates := []*GeoCandidate{
		{Provider: "ip2location", Source: "ip2location", City: "A", Band: 4},
		{Provider: "ipapi", Source: "ipapi", City: "B", Band: 4, RadiusKM: floatPointer(50)},
		{Provider: "maxmind", Source: "maxmind", City: "C", Band: 4, RadiusKM: floatPointer(20)},
	}
	for _, order := range [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		var winner *GeoCandidate
		for _, i := range order {
			if betterGeo(candidates[i], winner, DefaultGeolocationOrder()) {
				winner = candidates[i]
			}
		}
		if winner.Source != "maxmind" {
			t.Fatalf("check order %v selected %+v", order, winner)
		}
	}
	if betterGeo(&GeoCandidate{Provider: "maxmind", Country: "US", Band: 5}, candidates[0], DefaultGeolocationOrder()) {
		t.Fatal("country confidence erased city detail")
	}
	if !betterGeo(&GeoCandidate{Provider: "maxmind", City: "C", Band: 4}, candidates[0], []string{"maxmind", "ip2location"}) {
		t.Fatal("admin tie priority ignored")
	}
}

func TestCoherentCoordinatesAndSourceQuality(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		id, raw, city string
		band          int
		coordinates   bool
	}{
		{"ipapi", `{"ip":"8.8.8.8","location":{"country_code":"US","city":"A","latitude":0,"longitude":0,"accuracy":"VERY_HIGH"}}`, "A", 5, true},
		{"ipapi", `{"ip":"8.8.8.8","is_anycast":true,"location":{"city":"A","accuracy":"VERY_HIGH"}}`, "A", 1, false},
		{"maxmind", `{"traits":{"ip_address":"8.8.8.8"},"country":{"confidence":100},"city":{"names":{"en":"A"},"confidence":25},"location":{"latitude":0,"longitude":0,"accuracy_radius":20}}`, "A", 1, true},
		{"maxmind", `{"traits":{"ip_address":"8.8.8.8"},"city":{"names":{"en":"A"}},"location":{"latitude":91,"longitude":0,"accuracy_radius":20}}`, "A", 4, false},
		{"ip2location", `{"ip":"8.8.8.8","country_code":"US","latitude":38,"longitude":-77}`, "", 1, true},
	} {
		p := parseFixture(t, tc.id, tc.raw)
		if p.Geo.City != tc.city || p.Geo.Band != tc.band || (p.Geo.Latitude != nil) != tc.coordinates || (p.Geo.Longitude != nil) != tc.coordinates {
			t.Fatalf("%s geo=%+v", tc.id, p.Geo)
		}
	}
}
