package iplookup

import "math"

// qualityBand is an application comparison policy, not an upstream equivalence.
func qualityBand(confidence *float64, radius *float64) int {
	if confidence != nil && *confidence >= 0 && *confidence <= 100 {
		for i, threshold := range []float64{95, 80, 60, 30, 0} {
			if *confidence >= threshold {
				return 5 - i
			}
		}
	}
	if radius != nil && *radius >= 0 && !math.IsInf(*radius, 0) && !math.IsNaN(*radius) {
		for i, threshold := range []float64{10, 50, 200, 1000} {
			if *radius <= threshold {
				return 5 - i
			}
		}
		return 1
	}
	return 0
}

func namedBand(value string) int {
	for i, name := range []string{"VERY_LOW", "LOW", "MEDIUM", "HIGH", "VERY_HIGH"} {
		if value == name {
			return i + 1
		}
	}
	return 0
}

func betterGeo(candidate, current *GeoCandidate, order []string) bool {
	if candidate == nil || (candidate.Country == "" && candidate.City == "" && candidate.Latitude == nil) {
		return false
	}
	if current == nil {
		return true
	}
	detailed := func(g *GeoCandidate) bool { return g.City != "" || g.Latitude != nil }
	if detailed(candidate) != detailed(current) {
		return detailed(candidate)
	}
	if candidate.Band != current.Band {
		return candidate.Band > current.Band
	}
	if (candidate.RadiusKM != nil) != (current.RadiusKM != nil) {
		return candidate.RadiusKM != nil
	}
	if candidate.RadiusKM != nil && current.RadiusKM != nil && *candidate.RadiusKM != *current.RadiusKM {
		return *candidate.RadiusKM < *current.RadiusKM
	}
	if len(order) == 0 {
		order = DefaultGeolocationOrder()
	}
	rank := func(id string) int {
		for i, known := range order {
			if known == id {
				return i
			}
		}
		return len(order)
	}
	return rank(candidate.Provider) < rank(current.Provider)
}

func coordinates(o object, latitudeKey, longitudeKey string) (*float64, *float64) {
	lat, lon := numeric(o, latitudeKey), numeric(o, longitudeKey)
	if lat == nil || lon == nil || math.IsNaN(*lat) || math.IsNaN(*lon) || math.IsInf(*lat, 0) || math.IsInf(*lon, 0) || *lat < -90 || *lat > 90 || *lon < -180 || *lon > 180 {
		return nil, nil
	}
	return lat, lon
}

func validRadius(value *float64) *float64 {
	if value == nil || *value < 0 || math.IsNaN(*value) || math.IsInf(*value, 0) {
		return nil
	}
	return value
}
