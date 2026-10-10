package iplookup

import "strings"

func parseScamalytics(p *ProviderResult, ip string, o object) error {
	main, sources := nested(o, "scamalytics"), nested(o, "external_datasources")
	if textField(main, "status") != "ok" {
		return &CodeError{"IP_LOOKUP_PROVIDER_UNAVAILABLE"}
	}
	if !validReturnedIP(main, "ip", ip) {
		return &CodeError{"IP_LOOKUP_INVALID_RESPONSE"}
	}
	proxy, firehol, x4b := nested(main, "scamalytics_proxy"), nested(sources, "firehol"), nested(sources, "x4bnet")
	geo, info, lite := nested(sources, "maxmind_geolite2"), nested(sources, "ipinfo"), nested(sources, "ip2proxy_lite")
	p.Signals = Signals{Abuse: flag(main, "is_blacklisted_external"), Datacenter: flag(proxy, "is_datacenter"), VPN: flag(proxy, "is_vpn"), Proxy: flag(firehol, "is_proxy"), Tor: flag(x4b, "is_tor")}
	for _, source := range []object{firehol, nested(sources, "ipsum"), nested(sources, "spamhaus_drop"), lite} {
		for _, key := range []string{"ip_blacklisted", "ip_blacklisted_30", "ip_blacklisted_1day"} {
			if value := flag(source, key); value != nil && *value {
				p.Signals.Abuse = boolean(true)
			}
		}
	}
	if value := flag(x4b, "is_blacklisted_spambot"); value != nil && *value {
		p.Signals.Abuse = boolean(true)
	}
	if value := flag(x4b, "is_vpn"); value != nil && *value {
		p.Signals.VPN = boolean(true)
	}
	if value := flag(x4b, "is_datacenter"); value != nil && *value {
		p.Signals.Datacenter = boolean(true)
	}
	if value := flag(proxy, "is_amazon_aws"); value != nil && *value {
		p.Signals.Datacenter = boolean(true)
	}
	if value := flag(nested(sources, "google"), "is_google_cloud"); value != nil && *value {
		p.Signals.Datacenter = boolean(true)
	}
	for _, key := range []string{"is_resproxy", "is_apple_icloud_private_relay"} {
		if value := flag(proxy, key); value != nil && *value {
			p.Signals.Proxy = boolean(true)
		}
	}
	residential := nested(sources, "residential_proxy_DBI")
	if value := flag(residential, "is_resproxy"); value != nil && *value {
		p.Signals.Proxy = boolean(true)
	}
	for _, source := range []object{lite, nested(sources, "ip2proxy")} {
		switch textField(source, "proxy_type") {
		case "VPN":
			p.Signals.VPN = boolean(true)
		case "TOR":
			p.Signals.Tor = boolean(true)
		case "DCH":
			p.Signals.Datacenter = boolean(true)
		case "PUB", "WEB", "RES":
			p.Signals.Proxy = boolean(true)
		}
	}
	p.Facts = Facts{Country: textField(geo, "ip_country_code"), City: textField(geo, "ip_city"), ASN: textField(geo, "asn"), NetworkType: networkType(textField(lite, "usage_type"))}
	if p.Facts.ASN == "" {
		p.Facts.ASN = strings.TrimPrefix(textField(info, "asn"), "AS")
	}
	if p.Facts.Country == "" {
		p.Facts.Country = textField(info, "ip_country_code")
	}
	if p.Signals.Datacenter != nil && *p.Signals.Datacenter {
		p.Facts.NetworkType = "datacenter"
	}
	if p.Facts.NetworkType == "" {
		switch textField(nested(sources, "dbip"), "connection_type") {
		case "dialup", "isdn", "cable", "dsl", "fttx":
			p.Facts.NetworkType = "residential"
		}
	}
	addScore(p, main, "scamalytics_score", "fraud_score")
	addScore(p, main, "scamalytics_isp_score", "isp_risk_score")
	p.RiskLevel = strings.ReplaceAll(strings.ToLower(textField(main, "scamalytics_risk")), " ", "_")
	switch p.RiskLevel {
	case "low", "medium", "high", "very_high":
	default:
		p.RiskLevel = ""
	}
	p.Complete = allKnown(p.Signals.Abuse, p.Signals.Datacenter, p.Signals.VPN, p.Signals.Proxy, p.Signals.Tor) && numeric(main, "scamalytics_score") != nil && p.RiskLevel != ""
	return nil
}
