// Package statistics owns cached product metrics and provider maintenance.
package statistics

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

// Digest is the remote range growth summary.
type Digest struct {
	CreatedUsers int64
	ExpiredUsers int64
}

// TrafficSeries is one raw per-node traffic series.
type TrafficSeries struct {
	UUID        string
	Name        string
	CountryCode string
	DailyBytes  []int64
}

// Traffic contains the provider's range labels and raw series.
type Traffic struct {
	Categories []string
	Series     []TrafficSeries
}

// Node joins stable node metadata with live system rates.
type Node struct {
	UUID             string
	Name             string
	CountryCode      string
	Online           bool
	UsersOnline      int64
	RXBytesPerSecond int64
	TXBytesPerSecond int64
	XrayVersion      string
	Multiplier       float64
}

// Host is the remark and linked-node subset used by scheduled upkeep.
type Host struct {
	UUID   string
	Remark string
	Nodes  []string
}

// Provider exposes only documented Remnawave reads and the host remark patch.
type Provider interface {
	Digest(context.Context, time.Time, time.Time) (Digest, error)
	Traffic(context.Context, time.Time, time.Time) (Traffic, error)
	UsageSnapshotForRollover(context.Context, string, time.Time, time.Time) (rollover.UsageSnapshot, error)
	Nodes(context.Context) ([]Node, error)
	Hosts(context.Context) ([]Host, error)
	UpdateHostRemark(context.Context, string, string) error
}
