package iplookup

import (
	"strconv"
	"strings"
)

func enrichProvider(p *ProviderResult, o object) {
	p.Sources, p.Details = map[string]string{}, map[string]string{}
	g := &GeoCandidate{Country: p.Facts.Country, City: p.Facts.City, Source: p.ID, Provider: p.ID}
	switch p.ID {
	case "ipapi":
		location, asn := nested(o, "location"), nested(o, "asn")
		g.Latitude, g.Longitude = coordinates(location, "latitude", "longitude")
		g.Band = namedBand(textField(location, "accuracy"))
		if v := flag(o, "is_anycast"); v != nil && *v {
			g.Band = 1
		}
		p.Facts.ASNName = textField(asn, "org")
		p.Details["vpn"] = textField(nested(o, "vpn"), "service")
		p.Details["datacenter"] = textField(nested(o, "datacenter"), "datacenter")
	case "maxmind":
		location, city, traits := nested(o, "location"), nested(o, "city"), nested(o, "traits")
		g.Latitude, g.Longitude = coordinates(location, "latitude", "longitude")
		g.RadiusKM = validRadius(numeric(location, "accuracy_radius"))
		confidence := numeric(city, "confidence")
		if g.City == "" {
			confidence = nil
		}
		if g.City == "" && g.Latitude == nil {
			confidence = numeric(nested(o, "country"), "confidence")
		}
		g.Band = qualityBand(confidence, g.RadiusKM)
		if v := flag(traits, "is_anycast"); v != nil && *v {
			g.Band = 1
		}
		p.Facts.ASNName = textField(traits, "autonomous_system_organization")
		p.MaxMind = MaxMindFacts{IPRiskSnapshot: rangedNumber(traits, "ip_risk_snapshot", 0.01, 99), StaticIPScore: rangedNumber(traits, "static_ip_score", 0, 99.99), UserCount: integerField(traits, "user_count")}
		if kind := textField(traits, "user_type"); kind != "" {
			p.MaxMind.UserType = &kind
		}
		anonymizer := nested(o, "anonymizer")
		p.Details["vpn"] = textField(anonymizer, "provider_name")
		p.Details["proxy"] = textField(nested(anonymizer, "residential"), "provider_name")
	case "ip2location":
		g.Latitude, g.Longitude = coordinates(o, "latitude", "longitude")
		if (g.City == "" && g.Latitude != nil) || textField(o, "address_type") == "A" {
			g.Band = 1
		}
		p.Facts.ASNName = textField(o, "as")
		p.Details["vpn"], p.Details["proxy"] = textField(nested(o, "proxy"), "provider"), textField(nested(o, "proxy"), "provider")
		p.Details["abuse"] = textField(nested(o, "proxy"), "threat")
	case "scamalytics":
		g = scamalyticsGeo(p, nested(o, "external_datasources"))
		scamalyticsSources(p, nested(o, "external_datasources"))
		scamalyticsDetails(p, nested(o, "external_datasources"))
	case "abuseipdb":
		g.Source = "abuseipdb:ipinfo"
		p.Sources["networkType"] = "abuseipdb:ipinfo"
	case "ipqs":
		return // IPQS remains a reputation check, outside the location preference list.
	}
	p.Geo = g
	p.Facts.Latitude, p.Facts.Longitude = g.Latitude, g.Longitude
}

func integerField(o object, key string) *int {
	v := numeric(o, key)
	if v == nil || *v < 0 || *v > 1_000_000_000 || *v != float64(int(*v)) {
		return nil
	}
	n := int(*v)
	return &n
}

func rangedNumber(o object, key string, minimum, maximum float64) *float64 {
	value := numeric(o, key)
	if value == nil || *value < minimum || *value > maximum {
		return nil
	}
	return value
}

func coordinateString(o object, key string) (*float64, *float64) {
	parts := strings.Split(textField(o, key), ",")
	if len(parts) != 2 {
		return nil, nil
	}
	lat, latErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lon, lonErr := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if latErr != nil || lonErr != nil {
		return nil, nil
	}
	return coordinates(object{"latitude": []byte(strconv.FormatFloat(lat, 'f', -1, 64)), "longitude": []byte(strconv.FormatFloat(lon, 'f', -1, 64))}, "latitude", "longitude")
}

func scamalyticsGeo(p *ProviderResult, sources object) *GeoCandidate {
	var selected *GeoCandidate
	for _, id := range []string{"maxmind_geolite2", "dbip", "ip2proxy_lite", "ipinfo"} {
		o := nested(sources, id)
		g := &GeoCandidate{Country: textField(o, "ip_country_code"), City: textField(o, "ip_city"), Source: "scamalytics:" + id, Provider: "scamalytics"}
		g.Latitude, g.Longitude = coordinateString(o, "ip_geolocation")
		if raw := textField(o, "ip_location_accuracy_km"); raw != "" {
			if value, err := strconv.ParseFloat(raw, 64); err == nil {
				g.RadiusKM = validRadius(&value)
			}
		}
		g.Band = qualityBand(nil, g.RadiusKM)
		if betterGeo(g, selected, DefaultGeolocationOrder()) {
			selected = g
		}
		if p.Facts.ASN != "" && strings.TrimPrefix(textField(o, "asn"), "AS") == p.Facts.ASN {
			p.Sources["asn"] = "scamalytics:" + id
			if p.Facts.ASNName == "" && textField(o, "as_name") != "" {
				p.Facts.ASNName = textField(o, "as_name")
				p.Sources["asnName"] = "scamalytics:" + id
			}
		}
	}
	if selected == nil {
		selected = &GeoCandidate{Source: "scamalytics", Provider: "scamalytics"}
	}
	return selected
}
