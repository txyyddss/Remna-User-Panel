package squadprofile

// Type identifies the customer-facing route profile of an internal squad.
type Type string

const (
	Broadband            Type = "broadband"
	ChinaOptimized       Type = "china_optimized"
	InternationalNetwork Type = "international_network"
)

// Profile is a compact discriminated union. Only fields belonging to Type are
// emitted after Normalize, keeping the local payload free of duplicate data.
type Profile struct {
	Type             Type     `json:"type"`
	ISP              string   `json:"isp,omitempty"`
	PortMbps         *int     `json:"portMbps"`
	Dynamic          *bool    `json:"dynamic,omitempty"`
	Location         string   `json:"location,omitempty"`
	CountryCode      string   `json:"countryCode,omitempty"`
	CT               string   `json:"ct,omitempty"`
	CU               string   `json:"cu,omitempty"`
	CM               string   `json:"cm,omitempty"`
	UpstreamCarriers []string `json:"upstreamCarriers,omitempty"`
}

// NewInternational returns the default editor profile for an unconfigured squad.
func NewInternational() *Profile {
	return &Profile{Type: InternationalNetwork, UpstreamCarriers: []string{}}
}
