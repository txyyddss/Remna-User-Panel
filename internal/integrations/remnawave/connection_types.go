package remnawave

import "time"

// ConnectionScan is the documented asynchronous user-connection result.
type ConnectionScan struct {
	Completed bool
	Failed    bool
	Progress  float64
	Nodes     []ConnectionNode
}

// ConnectionNode groups one user's live IPs by provider node.
type ConnectionNode struct {
	UUID        string         `json:"nodeUuid"`
	Name        string         `json:"nodeName"`
	CountryCode string         `json:"countryCode"`
	IPs         []ConnectionIP `json:"ips"`
}

// ConnectionIP is one provider-observed address and last-seen time.
type ConnectionIP struct {
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"lastSeen"`
}

// StatsDigest is Remnawave's bounded range digest.
type StatsDigest struct {
	CreatedUsers int64
	ExpiredUsers int64
	TrafficBytes int64
}

// NodeMetric is the live online-user portion of the metrics response.
type NodeMetric struct {
	UUID        string `json:"nodeUuid"`
	Name        string `json:"nodeName"`
	UsersOnline int64  `json:"usersOnline"`
}

// Host is the identity, remark, and linked-node subset used by remark upkeep.
type Host struct {
	UUID   string   `json:"uuid"`
	Remark string   `json:"remark"`
	Nodes  []string `json:"nodes"`
}

// NodesUsage is the documented raw per-node range series.
type NodesUsage struct {
	Categories []string          `json:"categories"`
	Series     []NodeUsageSeries `json:"series"`
}
