package iplookup

// Descriptor documents an adapter's risk coverage and credential requirements.
type Descriptor struct {
	ID                string
	Capabilities      []string
	AccountIDRequired bool
}

// Registry is the adapter extension point and legacy default check order.
var Registry = []Descriptor{
	{ID: "abuseipdb", Capabilities: []string{"abuse", "datacenter", "tor", "reports"}},
	{ID: "scamalytics", Capabilities: []string{"abuse", "datacenter", "vpn", "proxy", "tor", "fraud_score"}, AccountIDRequired: true},
	{ID: "ipapi", Capabilities: []string{"abuse", "datacenter", "vpn", "proxy", "tor"}},
	{ID: "maxmind", Capabilities: []string{"datacenter", "vpn", "proxy", "tor"}, AccountIDRequired: true},
	{ID: "ipqs", Capabilities: []string{"abuse", "vpn", "proxy", "tor", "fraud_score"}},
	{ID: "ip2location", Capabilities: []string{"abuse", "datacenter", "vpn", "proxy", "tor", "fraud_score"}},
}
