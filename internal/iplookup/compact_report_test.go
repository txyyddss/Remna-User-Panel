package iplookup

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func parseFixture(t *testing.T, id, raw string) ProviderResult {
	t.Helper()
	var body object
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	p, err := parseProvider(id, "8.8.8.8", body)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestIPAPINetworkFlagsOverrideOrganizationMetadata(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ flags, want string }{
		{`"is_mobile":false,"is_satellite":false,"is_datacenter":false`, "residential"},
		{`"is_mobile":true,"is_satellite":false,"is_datacenter":false`, "mobile"},
		{`"is_mobile":true,"is_satellite":true,"is_datacenter":false`, "satellite"},
		{`"is_mobile":true,"is_satellite":true,"is_datacenter":true`, "datacenter"},
		{`"is_datacenter":false`, ""},
	} {
		p := parseFixture(t, "ipapi", `{"ip":"8.8.8.8","asn":{"type":"hosting"},"company":{"type":"hosting"},`+tc.flags+`}`)
		if p.Facts.NetworkType != tc.want {
			t.Fatalf("flags=%s type=%s", tc.flags, p.Facts.NetworkType)
		}
		if tc.want != "datacenter" && p.Signals.Datacenter != nil && *p.Signals.Datacenter {
			t.Fatal("ASN organization caused refusal")
		}
	}
}

func TestMaxMindCountThresholdAndNullableRequestedFields(t *testing.T) {
	t.Parallel()
	for _, count := range []string{"0", "5", "6", "null", "5.5", "-1"} {
		p := parseFixture(t, "maxmind", `{"traits":{"ip_address":"8.8.8.8","user_type":"residential","user_count":`+count+`,"static_ip_score":0,"ip_risk_snapshot":8.5}}`)
		if (len(RiskReasons(p)) > 0) != (count == "6") {
			t.Fatalf("count=%s reasons=%v", count, RiskReasons(p))
		}
		if (p.MaxMind.UserCount != nil) != (count == "0" || count == "5" || count == "6") {
			t.Fatalf("count=%s parsed=%v", count, p.MaxMind.UserCount)
		}
		if p.MaxMind.StaticIPScore == nil || *p.MaxMind.StaticIPScore != 0 || p.MaxMind.IPRiskSnapshot == nil || *p.MaxMind.IPRiskSnapshot != 8.5 || p.MaxMind.UserType == nil {
			t.Fatal("requested fields lost")
		}
	}
}

func TestStrictMajorityAndOutageCompletion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		kinds []string
		want  string
	}{
		{[]string{"residential", "residential", "mobile"}, "residential"},
		{[]string{"residential", "mobile"}, ""},
		{[]string{"residential", "residential", "mobile", "business"}, ""},
		{[]string{"", "mobile"}, "mobile"},
	} {
		r := NewReport("r", "8.8.8.8", DefaultConfig())
		for i, kind := range tc.kinds {
			RecordProvider(&r, ProviderResult{ID: IDs[i], Status: "success", Complete: true, Facts: Facts{NetworkType: kind}}, DefaultConfig())
		}
		if r.Facts.NetworkType != tc.want {
			t.Fatalf("votes=%v type=%s", tc.kinds, r.Facts.NetworkType)
		}
	}
	r := NewReport("r", "8.8.8.8", DefaultConfig())
	RecordProvider(&r, ProviderResult{ID: "ipapi", Status: "error"}, DefaultConfig())
	r = Aggregate(r, time.Now())
	if r.Status != "partial" || r.Verdict != "inconclusive" || r.RefundRequired || !ReportAttempted(r) {
		t.Fatalf("outage report=%+v", r)
	}
}

func TestCompactCheckpointAndLegacyOutcomePreservation(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"id":"r","ip":"8.8.8.8","status":"processing","verdict":"inconclusive","providers":[{"id":"ipapi","status":"success","complete":true,"facts":{"country":"US","city":"A","networkType":"residential"},"signals":{"abuse":false,"datacenter":false,"vpn":false,"proxy":false,"tor":false},"scores":{"company_abuse_ratio":0.01}},{"id":"maxmind","status":"queued"}]}`)
	r, err := DecodeReport(raw)
	if err != nil {
		t.Fatal(err)
	}
	if r.Checkpoint == nil || r.Checkpoint.Stages[0].Status != "success" || !ReportAttempted(r) {
		t.Fatalf("lost completed stage: %+v", r)
	}
	encoded, err := EncodeReport(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"providers"`) || strings.Contains(string(encoded), "company_abuse_ratio") || strings.Contains(string(encoded), `"signals"`) || !strings.Contains(string(encoded), `"_checkpoint"`) {
		t.Fatalf("noncompact persistence: %s", encoded)
	}
	terminal, err := DecodeReport([]byte(strings.Replace(string(raw), `"status":"processing"`, `"status":"failed","refundRequired":true`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if terminal.Status != "failed" || !terminal.RefundRequired || terminal.Checkpoint != nil {
		t.Fatalf("rewrote legacy outcome: %+v", terminal)
	}
	encoded, err = EncodeReport(terminal)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "_checkpoint") || strings.Contains(string(encoded), `"providers"`) {
		t.Fatalf("terminal retained private data: %s", encoded)
	}
	public, err := json.Marshal(PublicReport(r))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(public), "_checkpoint") {
		t.Fatal("private checkpoint leaked")
	}
}
