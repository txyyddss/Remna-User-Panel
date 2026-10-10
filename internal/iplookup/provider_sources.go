package iplookup

func scamalyticsSources(p *ProviderResult, sources object) {
	for _, item := range []struct{ kind, id, key, proxyType string }{
		{"datacenter", "x4bnet", "is_datacenter", ""}, {"datacenter", "google", "is_google_cloud", ""},
		{"datacenter", "ip2proxy", "", "DCH"}, {"datacenter", "ip2proxy_lite", "", "DCH"},
		{"vpn", "x4bnet", "is_vpn", ""}, {"tor", "x4bnet", "is_tor", ""},
		{"tor", "ip2proxy", "", "TOR"}, {"tor", "ip2proxy_lite", "", "TOR"},
		{"proxy", "firehol", "is_proxy", ""}, {"proxy", "residential_proxy_DBI", "is_resproxy", ""},
	} {
		o := nested(sources, item.id)
		flagged := flag(o, item.key)
		if (flagged != nil && *flagged) || (item.proxyType != "" && textField(o, "proxy_type") == item.proxyType) {
			p.Sources[item.kind] = "scamalytics:" + item.id
		}
	}
	if p.Facts.NetworkType == "datacenter" && p.Sources["datacenter"] != "" {
		p.Sources["networkType"] = p.Sources["datacenter"]
	} else if p.Facts.NetworkType == networkType(textField(nested(sources, "ip2proxy_lite"), "usage_type")) && p.Facts.NetworkType != "" {
		p.Sources["networkType"] = "scamalytics:ip2proxy_lite"
	} else if p.Facts.NetworkType == "residential" {
		p.Sources["networkType"] = "scamalytics:dbip"
	}
}

func scamalyticsDetails(p *ProviderResult, sources object) {
	for _, id := range []string{"residential_proxy_DBI", "ip2proxy", "ip2proxy_lite"} {
		o := nested(sources, id)
		provider := textField(o, "ip_provider")
		if id == "residential_proxy_DBI" {
			if detected := flag(o, "is_resproxy"); detected == nil || !*detected {
				continue
			}
			provider = textField(o, "last_proxy_provider")
		} else if kind := textField(o, "proxy_type"); kind != "PUB" && kind != "WEB" && kind != "RES" {
			continue
		}
		if provider != "" {
			p.Details["proxy"] = provider
			p.Sources["proxy"] = "scamalytics:" + id
			break
		}
	}
	for _, id := range []string{"ip2proxy", "ip2proxy_lite"} {
		o := nested(sources, id)
		if textField(o, "proxy_type") == "VPN" {
			p.Details["vpn"] = textField(o, "ip_provider")
			p.Sources["vpn"] = "scamalytics:" + id
		}
	}
	for _, id := range []string{"firehol", "ipsum", "spamhaus_drop", "ip2proxy_lite", "x4bnet"} {
		o := nested(sources, id)
		for _, key := range []string{"ip_blacklisted", "ip_blacklisted_30", "ip_blacklisted_1day", "is_blacklisted_spambot"} {
			if v := flag(o, key); v != nil && *v {
				p.Sources["abuse"] = "scamalytics:" + id
				p.Details["abuse"] = textField(o, "ip_blacklist_type")
				return
			}
		}
	}
}
