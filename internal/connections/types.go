// Package connections owns member connection scans and signed drop handles.
package connections

import (
	"errors"
	"time"
)

const HandleTTL = 15 * time.Minute

const (
	// ScanTTL bounds provider job metadata and all handles derived from it.
	ScanTTL = 30 * time.Minute
	// ScanRequestOutboxKind starts a metadata-only provider connection scan.
	ScanRequestOutboxKind = "connection_scan_request"
	// DropOperationKind owns durable selected-IP unlink operations.
	DropOperationKind = "connection_drop"
)

var (
	// ErrScanNotFound hides scans owned by another member and expired scans.
	ErrScanNotFound = errors.New("connection scan not found")
	// ErrIdentityRequired means the member has no usable provider identity.
	ErrIdentityRequired = errors.New("Remnawave identity is required")
)

// HandleClaims are the server-authoritative fields carried by a signed handle.
type HandleClaims struct {
	UserID   string
	ScanID   string
	NodeUUID string
	IP       string
	Expires  time.Time
}

// IP is one display-safe connection observation with its drop capability.
type IP struct {
	Address  string    `json:"ip"`
	LastSeen time.Time `json:"lastSeen"`
	Handle   string    `json:"handle"`
}

// Node groups current member connections by provider node.
type Node struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	CountryCode string `json:"countryCode"`
	IPs         []IP   `json:"ips"`
}

// Scan is the transient browser projection; only metadata is durable.
type Scan struct {
	ID              string    `json:"id"`
	Completed       bool      `json:"isCompleted"`
	Failed          bool      `json:"isFailed"`
	ProgressPercent float64   `json:"progressPercent"`
	Nodes           []Node    `json:"nodes"`
	CreatedAt       time.Time `json:"createdAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
}

// Observation is one provider result before a short-lived handle is attached.
type Observation struct {
	Address  string
	LastSeen time.Time
}

// ProviderNode is a provider scan node before browser projection.
type ProviderNode struct {
	UUID        string
	Name        string
	CountryCode string
	IPs         []Observation
}

// ProviderScan is the normalized, non-durable Remnawave scan response.
type ProviderScan struct {
	Completed       bool
	Failed          bool
	ProgressPercent float64
	Nodes           []ProviderNode
}
