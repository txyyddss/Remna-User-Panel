package iplookup

import (
	"context"
	"net/http"
	"testing"
)

func TestLiveInternetDBObservationsAndNoInformation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		state  string
		ports  int
	}{
		{"observed", 200, `{"ip":"1.1.1.1","hostnames":["one.one.one.one","one.one.one.one"],"ports":[443,53,443]}`, "success", 2},
		{"not scanned", 404, `{"detail":"No information available"}`, "empty", 0},
		{"provider failure", 500, `{}`, "error", 0},
		{"wrong address", 200, `{"ip":"8.8.8.8","hostnames":[],"ports":[443]}`, "error", 0},
		{"missing ports", 200, `{"ip":"1.1.1.1","hostnames":[]}`, "error", 0},
		{"invalid port", 200, `{"ip":"1.1.1.1","hostnames":[],"ports":[0,65536,443]}`, "partial", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := liveTestService(t, "shodan", func(r *http.Request) (int, string) {
				if r.URL.Host != "internetdb.shodan.io" || r.URL.Path != "/1.1.1.1" || r.URL.RawQuery != "" {
					t.Error("InternetDB request mismatch")
				}
				return tc.status, tc.body
			})
			r := s.shodan(context.Background(), "1.1.1.1")
			if r.Status != tc.state || len(r.Ports) != tc.ports || r.ObservedAt != "" {
				t.Fatalf("%+v", r)
			}
			if tc.name == "observed" && (r.Ports[0] != 53 || len(r.Hostnames) != 1) {
				t.Fatal("sorting/deduplication failed")
			}
		})
	}
}
