package iplookup

func parseMaxmind(p *ProviderResult, ip string, o object) error {
	traits, anonymizer := nested(o, "traits"), nested(o, "anonymizer")
	if !validReturnedIP(traits, "ip_address", ip) {
		return &CodeError{"IP_LOOKUP_INVALID_RESPONSE"}
	}
	// Insights explicitly omits false anonymizer flags; legacy traits remain supported.
	value := func(key string) *bool {
		invalid := false
		for _, source := range []object{anonymizer, traits} {
			v := flag(source, key)
			if _, exists := source[key]; exists && v == nil {
				invalid = true
			}
			if v != nil && *v {
				return boolean(true)
			}
		}
		if invalid {
			return nil
		}
		return boolean(false)
	}
	p.Signals = Signals{Datacenter: value("is_hosting_provider"), VPN: value("is_anonymous_vpn"), Proxy: value("is_anonymous"), Tor: value("is_tor_exit_node")}
	for _, key := range []string{"is_public_proxy", "is_residential_proxy"} {
		v := value(key)
		if v == nil && p.Signals.Proxy != nil && !*p.Signals.Proxy {
			p.Signals.Proxy = nil
		}
		if v != nil && *v {
			p.Signals.Proxy = boolean(true)
		}
	}
	if confidence := numeric(nested(anonymizer, "residential"), "confidence"); confidence != nil && *confidence > 0 {
		p.Signals.Proxy = boolean(true)
	}
	city := nested(o, "city")
	p.Facts = Facts{Country: textField(nested(o, "country"), "iso_code"), City: textField(nested(city, "names"), "en"), ISP: textField(traits, "isp"), ASN: textField(traits, "autonomous_system_number"), NetworkType: networkType(textField(traits, "user_type"))}
	addScore(p, traits, "static_ip_score", "static_ip_score")
	addScore(p, traits, "ip_risk_snapshot", "network_risk_snapshot")
	p.Complete = allKnown(p.Signals.Datacenter, p.Signals.VPN, p.Signals.Proxy, p.Signals.Tor)
	return nil
}

func parseIPQS(p *ProviderResult, o object) error {
	if success := flag(o, "success"); success == nil || !*success {
		return &CodeError{"IP_LOOKUP_PROVIDER_UNAVAILABLE"}
	}
	p.Facts = Facts{Country: textField(o, "country_code"), Region: textField(o, "region"), City: textField(o, "city"), ISP: textField(o, "ISP"), ASN: textField(o, "ASN"), NetworkType: networkType(textField(o, "connection_type"))}
	if mobile := flag(o, "mobile"); mobile != nil && *mobile {
		p.Facts.NetworkType = "mobile"
	}
	p.Signals = Signals{Abuse: flag(o, "recent_abuse"), VPN: flag(o, "vpn"), Proxy: flag(o, "proxy"), Tor: flag(o, "tor")}
	if kind := textField(o, "connection_type"); networkType(kind) != "" {
		p.Signals.Datacenter = boolean(networkType(kind) == "datacenter")
	}
	for _, key := range []string{"bot_status", "frequent_abuser", "high_risk_attacks"} {
		if v := flag(o, key); v != nil && *v {
			p.Signals.Abuse = boolean(true)
		}
	}
	if v := flag(o, "active_vpn"); v != nil && *v {
		p.Signals.VPN = boolean(true)
	}
	if v := flag(o, "active_tor"); v != nil && *v {
		p.Signals.Tor = boolean(true)
	}
	addScore(p, o, "fraud_score", "fraud_score")
	p.Complete = allKnown(p.Signals.Abuse, p.Signals.VPN, p.Signals.Proxy, p.Signals.Tor) && numeric(o, "fraud_score") != nil
	return nil
}

func parseIP2(p *ProviderResult, ip string, o object) error {
	if !validReturnedIP(o, "ip", ip) || len(o["error"]) > 0 {
		return &CodeError{"IP_LOOKUP_INVALID_RESPONSE"}
	}
	proxy := nested(o, "proxy")
	p.Facts = Facts{Country: textField(o, "country_code"), Region: textField(o, "region_name"), City: textField(o, "city_name"), ISP: textField(o, "isp"), ASN: textField(o, "asn"), NetworkType: networkType(textField(o, "usage_type"))}
	p.Signals = Signals{Datacenter: flag(proxy, "is_data_center"), VPN: flag(proxy, "is_vpn"), Proxy: flag(o, "is_proxy"), Tor: flag(proxy, "is_tor")}
	if p.Facts.NetworkType == "datacenter" || textField(proxy, "proxy_type") == "DCH" {
		p.Signals.Datacenter = boolean(true)
	}
	for _, key := range []string{"is_public_proxy", "is_web_proxy", "is_residential_proxy", "is_consumer_privacy_network", "is_enterprise_private_network"} {
		if v := flag(proxy, key); v != nil && *v {
			p.Signals.Proxy = boolean(true)
		}
	}
	p.Signals.Abuse = boolean(false)
	for _, key := range []string{"is_spammer", "is_scanner", "is_botnet"} {
		v := flag(proxy, key)
		if v == nil && p.Signals.Abuse != nil && !*p.Signals.Abuse {
			p.Signals.Abuse = nil
		}
		if v != nil && *v {
			p.Signals.Abuse = boolean(true)
		}
	}
	if threat := textField(proxy, "threat"); threat != "" {
		p.Signals.Abuse = boolean(true)
	}
	addScore(p, o, "fraud_score", "fraud_score")
	p.Complete = allKnown(p.Signals.Abuse, p.Signals.Datacenter, p.Signals.VPN, p.Signals.Proxy, p.Signals.Tor) && numeric(o, "fraud_score") != nil
	return nil
}
