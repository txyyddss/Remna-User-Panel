package iplookup

import "time"

// NewReport reserves the immutable provider selection for one shared run.
func NewReport(id, ip string, c Config) Report {
	r := Report{ID: id, IP: ip, Status: "processing", Verdict: "inconclusive", Reasons: []string{}, Sources: map[string]string{}, Providers: []ProviderResult{}, PolicyVersion: Version, ParserVersion: Version}
	for _, p := range c.Providers {
		status := "disabled"
		if p.Enabled {
			status = "queued"
		}
		r.Providers = append(r.Providers, ProviderResult{ID: p.ID, Status: status, Scores: map[string]float64{}})
	}
	return r
}

// RiskReasons identifies only per-IP evidence; network scores are contextual.
func RiskReasons(p ProviderResult) []string {
	result := []string{}
	flags := []struct {
		name  string
		value *bool
	}{{"abuse", p.Signals.Abuse}, {"datacenter", p.Signals.Datacenter}, {"vpn", p.Signals.VPN}, {"proxy", p.Signals.Proxy}, {"tor", p.Signals.Tor}}
	for _, flag := range flags {
		if flag.value != nil && *flag.value {
			result = append(result, p.ID+":"+flag.name)
		}
	}
	if p.Reports != nil && *p.Reports > 0 {
		result = append(result, p.ID+":abuse_reports")
	}
	for _, key := range []string{"abuse_confidence", "fraud_score"} {
		if p.Scores[key] > 0 {
			result = append(result, p.ID+":"+key)
		}
	}
	return result
}

// Aggregate freezes one honest verdict with first-available source-attributed facts.
func Aggregate(r Report, now time.Time) Report {
	r.Reasons = []string{}
	r.Sources = map[string]string{}
	r.Facts = Facts{}
	usable, complete, eligible := 0, true, false
	business := false
	covered := [5]bool{}
	for _, p := range r.Providers {
		if p.Status == "disabled" {
			continue
		}
		if p.Status == "success" || p.Status == "partial" {
			usable++
			mergeFacts(&r, p)
			eligible = eligible || p.Facts.NetworkType == "residential" || p.Facts.NetworkType == "mobile"
			business = business || p.Facts.NetworkType == "business"
			flags := []*bool{p.Signals.Abuse, p.Signals.Datacenter, p.Signals.VPN, p.Signals.Proxy, p.Signals.Tor}
			for i, v := range flags {
				covered[i] = covered[i] || v != nil
			}
			r.Reasons = append(r.Reasons, RiskReasons(p)...)
		}
		complete = complete && p.Status == "success" && p.Complete
	}
	r.CheckedAt = now.UTC().Format(time.RFC3339Nano)
	r.Status, r.Verdict = "succeeded", "inconclusive"
	if usable == 0 {
		r.Status = "failed"
		r.Reasons = []string{"all_providers_failed"}
		return r
	}
	if len(r.Reasons) > 0 {
		r.Verdict = "unsuitable"
		return r
	}
	for _, known := range covered {
		complete = complete && known
	}
	if !complete {
		r.Status = "partial"
		r.Reasons = append(r.Reasons, "incomplete_coverage")
	}
	eligible = eligible && !business
	if !eligible {
		r.Reasons = append(r.Reasons, "residential_unknown")
	}
	if complete && eligible {
		r.Verdict = "suitable"
	}
	return r
}

func mergeFacts(r *Report, p ProviderResult) {
	fields := []struct {
		key    string
		target *string
		value  string
	}{
		{"country", &r.Facts.Country, p.Facts.Country}, {"region", &r.Facts.Region, p.Facts.Region}, {"city", &r.Facts.City, p.Facts.City},
		{"isp", &r.Facts.ISP, p.Facts.ISP}, {"asn", &r.Facts.ASN, p.Facts.ASN}, {"networkType", &r.Facts.NetworkType, p.Facts.NetworkType},
	}
	for _, f := range fields {
		if *f.target == "" && f.value != "" {
			*f.target = f.value
			r.Sources[f.key] = p.ID
		}
	}
}
